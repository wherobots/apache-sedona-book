package cli

import (
	"cli/internal/adapter/cli/animation"
	"cli/internal/adapter/cli/browser"
	"cli/internal/adapter/cli/transform"
	"cli/internal/domain/dto"
	"cli/internal/domain/entity"
	"context"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type ProvisionService interface {
	Provision(ctx context.Context, request *dto.ProvisionRequest) (*dto.StartContainersRequest, error)
	Clear(ctx context.Context) error
	ListContainers(ctx context.Context) ([]*entity.ContainerMetadata, error)
}

type Client struct {
	provisionService ProvisionService
}

func NewClient(provisionService ProvisionService) *Client {
	return &Client{
		provisionService: provisionService,
	}
}

func (c *Client) GetCommands() []*cobra.Command {
	return []*cobra.Command{
		c.listContainers(context.Background()),
		c.clear(context.Background()),
		c.provision(context.Background()),
	}
}

func (c *Client) clear(ctx context.Context) *cobra.Command {
	var command = &cobra.Command{
		Use:  "clear",
		Args: cobra.NoArgs,
	}

	command.Run = func(cmd *cobra.Command, args []string) {
		go animation.BroomSweepMultipleAnimation(ctx)
		time.Sleep(time.Millisecond * 500)

		err := c.provisionService.Clear(ctx)
		if err != nil {
			cmd.PrintErrf("Error clearing containers: %v\n", err)
			return
		}

		cmd.Println("\n ✅ Containers cleared successfully.")
	}

	return command
}

func (c *Client) listContainers(ctx context.Context) *cobra.Command {
	var command = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
	}

	command.Run = func(cmd *cobra.Command, args []string) {
		containers, err := c.provisionService.ListContainers(ctx)
		if err != nil {
			cmd.PrintErrf("Error listing containers: %v\n", err)
		}

		for _, container := range containers {
			cmd.Printf("Name: %s State: %s 🐳\n", container.Name, container.State)
		}
	}

	return command
}

func (c *Client) provision(ctx context.Context) *cobra.Command {
	var chapter string
	var copyData bool
	var subChapter int

	var command = &cobra.Command{
		Use:  "provision",
		Args: cobra.NoArgs,
	}

	command.Flags().StringVarP(&chapter, "chapter", "c", "", "Chapter to provision")
	command.Flags().BoolVarP(&copyData, "copy", "l", false, "Copy the data to minio bucket if applicable")
	command.Flags().IntVarP(&subChapter, "sub-chapter", "s", 0, "Selected Sub chapter, if empty full chapter will be provisioned")

	ctxWithCancel, cancel := context.WithCancel(ctx)

	command.Run = func(cmd *cobra.Command, args []string) {
		var chapterDomain entity.Chapter

		if chapter != "" {
			chapterDomainRaw, err := transform.ChapterToDomain(chapter)
			if err != nil {
				command.PrintErrf("Error transforming chapter: %v\n", err)
				return
			}

			chapterDomain = chapterDomainRaw
		}

		startContainersResponse, err := c.provisionService.Provision(ctx, &dto.ProvisionRequest{
			Chapter:    chapterDomain,
			SubChapter: entity.SubChapter(subChapter),
		})
		if err != nil {
			return
		}

		go func() {
			for {
				select {
				case <-ctxWithCancel.Done():
					return
				case givenErr := <-startContainersResponse.Errors:
					if givenErr != nil {
						cancel()
						os.Exit(0)
						return
					}
				case <-time.Tick(time.Millisecond * 1000):
				}
			}
		}()

		ctxWithCancel, cancel = context.WithCancel(ctxWithCancel)

		go animation.BuildingImages(ctxWithCancel, startContainersResponse)

		c.waitImagesToBeReady(ctxWithCancel, startContainersResponse)
		cancel()

		ctxWithCancel, cancel = context.WithCancel(ctx)

		go animation.Provisioning(ctxWithCancel, startContainersResponse)

		c.waitContainersToBeReady(ctxWithCancel, startContainersResponse)
		cancel()

		cmd.Println("\n 🚀 Containers started successfully.")
		if startContainersResponse.OpenUrl != nil {
			print("\n 🌐 Opening browser...")
			browser.Open(*startContainersResponse.OpenUrl)
		}

		if startContainersResponse.OpenTerminal != nil {
			cmd.Println("\n 🖥️ Starting terminal...")
			cancel()

			err = startContainersResponse.OpenTerminal()
			if err != nil {
				cmd.PrintErrf("Error starting terminal: %v\n", err)
				return
			}
		}
	}

	return command
}

func (c *Client) waitImagesToBeReady(ctx context.Context, request *dto.StartContainersRequest) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	getReadyContainers := func(in *dto.StartContainersRequest) int {
		readyCount := 0
		for _, ready := range in.BuildingImages {
			if ready {
				readyCount++
			}
		}

		return readyCount
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Millisecond * 1000):
				if getReadyContainers(request) == len(request.BuildingImages) {
					wg.Done()
				}
			}

		}
	}()

	wg.Wait()
}

func (c *Client) waitContainersToBeReady(ctx context.Context, request *dto.StartContainersRequest) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	getReadyContainers := func(in *dto.StartContainersRequest) int {
		readyCount := 0
		for _, ready := range in.StartingImages {
			if ready {
				readyCount++
			}
		}

		return readyCount
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Millisecond * 1000):
				if getReadyContainers(request) == len(request.StartingImages) {
					wg.Done()
				}
			}

		}
	}()

	wg.Wait()
}
