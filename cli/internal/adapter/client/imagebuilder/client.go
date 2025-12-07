package imagebuilder

import (
	"archive/tar"
	domainerrors "cli/internal/domain/errors"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"go.uber.org/zap"
)

type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type DockerBuildResponse struct {
	ErrorDetail *ErrorDetail `json:"errorDetail,omitempty"`
}

//go:embed docker
var sedona embed.FS

type DockerClient interface {
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImageInspect(ctx context.Context, imageID string, inspectOpts ...client.ImageInspectOption) (image.InspectResponse, error)
}

type Client struct {
	client DockerClient
	logger *zap.Logger
}

func NewClient(client DockerClient, logger *zap.Logger) *Client {
	return &Client{
		client: client,
		logger: logger,
	}
}

func (c *Client) BuildImage(ctx context.Context, imageName string, tag string) error {
	archive, err := createAnArchive(fmt.Sprintf("docker/%s", imageName))
	if err != nil {
		return err
	}

	imageBuildResponse, err := c.client.ImageBuild(
		context.Background(),
		archive,
		types.ImageBuildOptions{
			Tags:       []string{fmt.Sprintf("%s:%s", imageName, tag)},
			Dockerfile: "Dockerfile",
		})
	if err != nil {
		c.logger.Error("Failed to build image", zap.String("image", imageName), zap.String("tag", tag), zap.Error(err))
		return err
	}

	bodyBytes, err := io.ReadAll(imageBuildResponse.Body)
	if err != nil {
		return err
	}
	defer imageBuildResponse.Body.Close()

	lines := strings.Split(string(bodyBytes), "\r\n")

	for index, l := range lines {
		var msg DockerBuildResponse
		err := json.Unmarshal([]byte(l), &msg)
		if err != nil {
			continue
		}

		if msg.ErrorDetail != nil {
			c.logger.Error("Docker build error", zap.Int("code", msg.ErrorDetail.Code), zap.String("message", msg.ErrorDetail.Message), zap.String("image", imageName), zap.String("tag", tag), zap.Int("line", index))
			return fmt.Errorf("docker build error: %s", msg.ErrorDetail.Message)
		}
	}

	return nil
}

func (c *Client) Exists(ctx context.Context, image, tag string) error {
	_, err := c.client.ImageInspect(context.Background(), fmt.Sprintf("%s:%s", image, tag))
	if err != nil {
		_, ok := (err).(errdefs.ErrNotFound)
		if ok {
			return domainerrors.ErrImageNotFound
		}

		return err
	}

	return nil
}

func createAnArchive(path string) (reader io.Reader, err error) {
	pr, pw := io.Pipe()

	tarWriteHandler := func() (subErr error) {
		defer func() {
			err = pw.Close()
		}()

		tw := tar.NewWriter(pw)
		defer func() {
			subErr = tw.Close()
		}()

		sub, err := fs.Sub(sedona, path)
		if err != nil {
			return err
		}

		if subErr = tw.AddFS(sub); err != nil {
			return err
		}

		return nil
	}

	go func() {
		if err = tarWriteHandler(); err != nil {
			_ = pw.CloseWithError(err)
		}
	}()

	return pr, nil
}
