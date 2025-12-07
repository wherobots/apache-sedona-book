package container

import (
	"cli/internal/domain/entity"
	domainerrors "cli/internal/domain/errors"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	SedonaNetworkName = "sedona"
	SedonaLabelName   = "sedona"
)

var (
	SedonaLabels = map[string]string{
		"app": SedonaLabelName,
	}
)

type Container struct {
	cli *client.Client
}

func NewClient(cli *client.Client) *Container {
	return &Container{
		cli: cli,
	}
}

func (c *Container) getFilterArgs() filters.Args {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("app=%s", SedonaLabelName))
	return filterArgs
}

func (c *Container) GetNetworkID(ctx context.Context, networkName string) (*entity.Network, error) {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(networks) == 0 {
		return nil, domainerrors.ErrNetworkNotFound
	}

	for _, n := range networks {
		if n.Name == networkName {
			return &entity.Network{
				ID: n.ID,
			}, nil
		}
	}

	return nil, domainerrors.ErrNetworkNotFound
}

func (c *Container) RemoveNetwork(ctx context.Context, networkID string) error {
	err := c.cli.NetworkRemove(ctx, networkID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) CreateNetwork(ctx context.Context, request *entity.CreateNetworkRequest) (*entity.CreateNetworkResponse, error) {
	n, err := c.cli.NetworkCreate(ctx, request.Name, network.CreateOptions{
		Driver:     "bridge",
		Internal:   false,
		Attachable: true,
		Labels:     SedonaLabels,
	})
	if err != nil {
		return nil, err
	}

	return &entity.CreateNetworkResponse{
		ID: n.ID,
	}, nil
}

func (c *Container) Run(ctx context.Context, request *entity.ContainerRunRequest) (*entity.CreateContainerResponse, error) {
	portMap := make(nat.PortMap, len(request.ExposedPorts))
	exposedPorts := make(nat.PortSet, len(request.ExposedPorts))
	imageName := strings.Split(request.Image, ":")[0]
	parts := strings.Split(imageName, "/")
	imageName = parts[len(parts)-1]

	for _, containerPort := range request.ExposedPorts {
		exposedPorts[nat.Port(containerPort)] = struct{}{}
	}

	for hostPort, containerPort := range request.ExposedPorts {
		portMap[nat.Port(containerPort)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Env:          request.EnvVariables.ToEnvList(),
		Image:        request.Image,
		ExposedPorts: exposedPorts,
		Labels:       SedonaLabels,
		Cmd:          request.Command,
		Healthcheck: &container.HealthConfig{
			Test:        request.HealthCheck.Test,
			Interval:    5 * time.Second,
			Timeout:     20 * time.Second,
			Retries:     3,
			StartPeriod: 2 * time.Second,
		},
	},
		&container.HostConfig{
			PortBindings: portMap,
			Mounts:       createMounts(request.MountPathHost, request.MountFiles),
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				SedonaNetworkName: {
					Aliases:   []string{imageName},
					NetworkID: request.NetworkID,
				},
			},
		},
		nil,
		request.ContainerName)
	if err != nil {
		return nil, err
	}

	if err = c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	return &entity.CreateContainerResponse{
		ID: resp.ID,
	}, nil
}

func createMounts(hostRoot string, mountsInput []string) []mount.Mount {
	mounts := make([]mount.Mount, 0)

	for _, m := range mountsInput {
		source := strings.Split(m, ":")[0]
		target := strings.Split(m, ":")[1]

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: fmt.Sprintf("%s/%s", hostRoot, source),
			Target: target,
		})
	}

	return mounts
}

func (c *Container) ReadLogs(ctx context.Context, containerID string) (string, error) {
	result, err := c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", err
	}

	defer result.Close()
	logs, err := io.ReadAll(result)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

func (c *Container) Wait(ctx context.Context, containerID string) (<-chan container.WaitResponse, <-chan error) {
	return c.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
}

func NewContainer(cli *client.Client) *Container {
	return &Container{
		cli: cli,
	}
}

func (c *Container) ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error) {
	result, err := c.cli.ContainerList(ctx, container.ListOptions{
		Filters: c.getFilterArgs(),
	})
	if err != nil {
		return nil, err
	}

	containers := make([]*entity.ContainerMetadata, 0, len(result))
	for _, cn := range result {
		containers = append(containers, &entity.ContainerMetadata{
			ID:    cn.ID,
			Name:  cn.Image,
			State: cn.State,
		})
	}

	return containers, nil
}

func (c *Container) Clear(ctx context.Context, metadata *entity.ContainerMetadata) error {
	err := c.cli.ContainerKill(ctx, metadata.ID, "")
	if err != nil {
		return err
	}

	err = c.cli.ContainerRemove(ctx, metadata.ID, container.RemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) IsHealthy(ctx context.Context, containerID string) (bool, error) {
	result, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, err
	}

	state := result.State
	if state == nil {
		return false, nil
	}

	if state.Status == "running" && state.Health == nil {
		return true, nil
	}

	if state.Health == nil {
		return false, nil
	}

	return result.State.Health.Status == "healthy" || result.State.Health.Status == "running", nil
}

func (c *Container) PullImage(ctx context.Context, img string) error {
	out, err := c.cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	line := make([]byte, 20)

	for {
		rs, err := out.Read(line)
		if rs == 0 {
			break
		}

		if err != nil {
			break
		}

	}

	return nil
}

func (c *Container) VerifyImageExists(ctx context.Context, img string) (bool, error) {
	images, err := c.cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", img)),
	})
	if err != nil {
		return false, err
	}

	return len(images) > 0, nil
}

type ImagePullStatus struct {
	Status string `json:"status"`
}
