package main

import (
	"cli/internal/adapter/cli"
	containerClient "cli/internal/adapter/client/container"
	"cli/internal/adapter/client/imagebuilder"
	"cli/internal/service/builder"
	"cli/internal/service/network"
	"cli/internal/service/provision"
	"cli/internal/service/resolver"
	"cli/internal/service/run"
	"os"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {

	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.ErrorLevel)

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	))
	defer logger.Sync()

	cc := containerClient.NewContainer(dockerClient)
	resolverService := resolver.NewService()
	imageClient := imagebuilder.NewClient(dockerClient, logger)

	builderService := builder.NewService().
		WithImageClient(imageClient).
		WithLogger(logger)

	networkService := network.NewService(cc)
	containerService := run.NewService(cc)

	provisioningService := provision.NewService().
		WithBuilderService(builderService).
		WithNetworkService(networkService).
		WithResolver(resolverService).
		WithContainerService(containerService).
		WithLogger(logger)

	cliClient := cli.NewClient(provisioningService)
	commands := cliClient.GetCommands()

	var rootCmd = &cobra.Command{Use: "sedona"}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	if err = rootCmd.Execute(); err != nil {
		rootCmd.PrintErrf("Error executing command: %v\n", err)
		return
	}
}
