package builder

import (
	"cli/internal/domain/dto"
	domainerrors "cli/internal/domain/errors"
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ImageClient interface {
	BuildImage(ctx context.Context, image, tag string) error
	Exists(ctx context.Context, image, tag string) error
}

type Service struct {
	imageClient ImageClient
	logger      *zap.Logger
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) WithImageClient(imageClient ImageClient) *Service {
	s.imageClient = imageClient
	return s
}

func (s *Service) WithLogger(logger *zap.Logger) *Service {
	s.logger = logger
	return s
}

func (s *Service) Build(ctx context.Context, request *dto.StartContainersRequest) error {
	for _, img := range request.Images {
		request.UpdateBuildImageStatus(img.Image, false)
	}

	err := s.buildImages(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) buildImages(ctx context.Context, request *dto.StartContainersRequest) (err error) {
	errGroup := &errgroup.Group{}

	for _, img := range request.Images.GetImagesNames() {
		errGroup.Go(func() error {
			err = s.imageClient.Exists(ctx, img.Name, img.Tag)
			if err != nil && !errors.Is(err, domainerrors.ErrImageNotFound) {
				s.logger.Error(err.Error(), zap.String("image", img.FullName()))
				return err
			}

			err = s.imageClient.BuildImage(ctx, img.Name, img.Tag)
			if err != nil {
				s.logger.Debug(fmt.Sprint("Failed to build image: ", img.FullName(), " error: ", err.Error()))
				return err
			}

			request.UpdateBuildImageStatus(img.FullName(), true)

			return nil
		})
	}

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	return nil
}
