package run

import (
	"cli/internal/domain/dto"
	"cli/internal/domain/entity"
	domainerrors "cli/internal/domain/errors"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	SedonaNetworkName   = "sedona"
	HealthCheckTimeout  = 120 * time.Second
	HealthCheckTickTime = time.Second * 2
)

type Client interface {
	ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error)
	Clear(ctx context.Context, metadata *entity.ContainerMetadata) error
	Run(ctx context.Context, arg *entity.ContainerRunRequest) (*entity.CreateContainerResponse, error)
	IsHealthy(ctx context.Context, containerID string) (bool, error)
	CreateNetwork(ctx context.Context, request *entity.CreateNetworkRequest) (*entity.CreateNetworkResponse, error)
	GetNetworkID(ctx context.Context, networkName string) (*entity.Network, error)
	RemoveNetwork(ctx context.Context, networkID string) error
}

type Service struct {
	client Client
}

func NewService(client Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) RunContainers(ctx context.Context, request *dto.StartContainersRequest) error {
	imageToConnect := ""
	resolvedURL := request.Images.ResolveURL()
	request.OpenUrl = resolvedURL
	errGroup := &errgroup.Group{}

	for _, image := range request.Images {
		errGroup.Go(func() error {
			request.UpdateContainerStatus(image.Image, false)
			image.NetworkID = request.NetworkID

			containerID, err := s.startContainer(ctx, image)
			if err != nil {
				return err
			}

			request.UpdateContainerStatus(image.Image, true)

			if image.OpenTerminal {
				imageToConnect = containerID
			}

			return nil
		})
	}

	err := errGroup.Wait()
	if err != nil {
		return err
	}

	if imageToConnect != "" {
		cmd := exec.Command("docker", "exec", "-it", imageToConnect, "/bin/bash")

		// Attach current terminal’s stdin/stdout/stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		request.OpenTerminal = func() error {
			return cmd.Run()
		}

		return nil
	}

	return nil
}

func (s *Service) startContainer(ctx context.Context, image *entity.ContainerRunRequest) (string, error) {
	containerInfo, err := s.client.Run(ctx, image)
	if err != nil {
		return "", err
	}

	healthy, err := s.WaitUntilHealthy(ctx, containerInfo.ID)
	if err != nil {
		return "", err
	}

	if !healthy {
		return "", err
	}

	return containerInfo.ID, nil
}

func (s *Service) WaitUntilHealthy(ctx context.Context, containerID string) (bool, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, HealthCheckTimeout)
	defer cancel()

	sleepTicker := time.NewTicker(HealthCheckTickTime)

	for {
		select {
		case <-ctxWithTimeout.Done():
			return false, nil
		case <-sleepTicker.C:
			healthy, err := s.client.IsHealthy(ctxWithTimeout, containerID)
			if err != nil {
				return false, err
			}

			if healthy {
				return true, nil
			}
		}
	}
}

func (s *Service) Clear(ctx context.Context) error {
	containers, err := s.client.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, c := range containers {
		err = s.client.Clear(ctx, c)
		if err != nil {
			return err
		}
	}

	network, err := s.client.GetNetworkID(ctx, SedonaNetworkName)
	if err != nil && errors.Is(err, domainerrors.ErrNetworkNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	err = s.client.RemoveNetwork(ctx, network.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error) {
	return s.client.ListContainers(ctx)
}
