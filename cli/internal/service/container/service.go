package container

import (
	"cli/internal/domain/entity"
	domainerrors "cli/internal/domain/errors"
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
)

const (
	SedonaNetworkName   = "sedona"
	SedonaLabelName     = "sedona"
	HealthCheckTimeout  = 120 * time.Second
	HealthCheckTickTime = time.Second * 2
)

type Client interface {
	ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error)
	Clear(ctx context.Context, metadata *entity.ContainerMetadata) error
	Run(ctx context.Context, arg *entity.ContainerRunRequest) (*entity.CreateContainerResponse, error)
	ReadLogs(ctx context.Context, containerID string) (string, error)
	Wait(ctx context.Context, containerID string) (<-chan container.WaitResponse, <-chan error)
	IsHealthy(ctx context.Context, containerID string) (bool, error)
	RunScript(ctx context.Context, containerID string, command []string) error
	CreateNetwork(ctx context.Context, request *entity.CreateNetworkRequest) (*entity.CreateNetworkResponse, error)
	GetNetworkID(ctx context.Context) (*entity.Network, error)
	RemoveNetwork(ctx context.Context, networkID string) error
}

type Resolver interface {
	Resolve(ctx context.Context, request *entity.ResolveRequest) ([]*entity.ContainerRunRequest, error)
}

type Service struct {
	client   Client
	resolver Resolver
}

func NewService(client Client, resolver Resolver) *Service {
	return &Service{
		client:   client,
		resolver: resolver,
	}
}

func (s *Service) ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error) {
	containers, err := s.client.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (s *Service) Clear(ctx context.Context) error {
	containers, err := s.ListContainers(ctx)
	if err != nil {
		return err
	}

	for _, c := range containers {
		err = s.client.Clear(ctx, c)
		if err != nil {
			return err
		}
	}

	network, err := s.client.GetNetworkID(ctx)
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

func (s *Service) StartContainers(ctx context.Context, request *entity.RunPreRequisiteRequest) (*entity.StartContainerResponse, error) {
	images, err := s.resolver.Resolve(ctx, &entity.ResolveRequest{
		Chapter:    request.Chapter,
		LoadData:   request.CopyData,
		SubChapter: request.SubChapter,
	})
	if err != nil {
		return nil, err
	}

	networkResponse, err := s.client.CreateNetwork(ctx, &entity.CreateNetworkRequest{
		Name: SedonaNetworkName,
	})
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(images))

	var url *string

	imageToConnect := ""

	for _, image := range images {
		go func() {
			defer wg.Done()
			image.NetworkID = networkResponse.ID
			if image.OpenWebUI != nil {
				url = &image.OpenWebUI.Path
			}

			containerInfo, err := s.client.Run(ctx, image)
			if err != nil {
				return
			}

			healthy, err := s.WaitUntilHealthy(ctx, containerInfo.ID)
			if err != nil {
				return
			}

			if !healthy {
				return
			}

			if image.OpenTerminal {
				imageToConnect = containerInfo.ID
			}

			return
		}()
	}

	wg.Wait()

	if imageToConnect != "" {
		cmd := exec.Command("docker", "exec", "-it", imageToConnect, "/bin/bash")

		// Attach current terminal’s stdin/stdout/stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return &entity.StartContainerResponse{
			OpenUrl: nil,
			OpenTerminal: func() error {
				return cmd.Run()
			},
		}, nil
	}

	return &entity.StartContainerResponse{
		OpenUrl: url,
	}, nil
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
