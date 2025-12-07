package provision

import (
	"cli/internal/domain/dto"
	"cli/internal/domain/entity"
	"context"
	"time"

	"go.uber.org/zap"
)

const (
	SedonaNetworkName   = "sedona"
	SedonaLabelName     = "sedona"
	HealthCheckTimeout  = 120 * time.Second
	HealthCheckTickTime = time.Second * 2
)

type (
	ContainerService interface {
		Clear(ctx context.Context) error
		RunContainers(ctx context.Context, request *dto.StartContainersRequest) error
		ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error)
	}

	BuilderService interface {
		Build(ctx context.Context, request *dto.StartContainersRequest) error
	}

	Resolver interface {
		Resolve(ctx context.Context, request *entity.ResolveRequest) (entity.ContainerRunRequests, error)
	}

	NetworkService interface {
		CreateOrGetNetwork(ctx context.Context, networkName string) (*entity.Network, error)
	}
)

type Service struct {
	containerService ContainerService
	builderService   BuilderService
	resolver         Resolver
	networkService   NetworkService
	logger           *zap.Logger
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) WithContainerService(containerService ContainerService) *Service {
	s.containerService = containerService
	return s
}

func (s *Service) WithBuilderService(builderService BuilderService) *Service {
	s.builderService = builderService
	return s
}

func (s *Service) WithResolver(resolver Resolver) *Service {
	s.resolver = resolver
	return s
}

func (s *Service) WithNetworkService(networkService NetworkService) *Service {
	s.networkService = networkService
	return s
}

func (s *Service) WithLogger(logger *zap.Logger) *Service {
	s.logger = logger
	return s
}

func (s *Service) Provision(ctx context.Context, request *dto.ProvisionRequest) (*dto.StartContainersRequest, error) {
	err := s.containerService.Clear(ctx)
	if err != nil {
		return nil, err
	}

	images, err := s.resolver.Resolve(ctx, &entity.ResolveRequest{
		Chapter:    request.Chapter,
		SubChapter: request.SubChapter,
	})
	if err != nil {
		return nil, err
	}

	network, err := s.networkService.CreateOrGetNetwork(ctx, SedonaNetworkName)
	if err != nil {
		return nil, err
	}

	errorsChan := make(chan error, 2)

	runContainersRequest := &dto.StartContainersRequest{
		BuildingImages: make(map[string]bool),
		StartingImages: make(map[string]bool),
		NetworkID:      network.ID,
		Images:         images,
		Errors:         errorsChan,
	}

	go func() {
		// build images
		err = s.builderService.Build(ctx, runContainersRequest)
		if err != nil {
			errorsChan <- err
			return
		}

		// start containers
		err = s.containerService.RunContainers(ctx, runContainersRequest)
		if err != nil {
			s.logger.Error("can't start container", zap.Error(err))
			errorsChan <- err
			return
		}
	}()

	return runContainersRequest, nil
}

func (s *Service) Clear(ctx context.Context) error {
	return s.containerService.Clear(ctx)
}

func (s *Service) ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error) {
	return s.containerService.ListContainers(ctx)
}
