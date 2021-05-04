package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	// "github.com/docker/docker/api/types/network"
	"github.com/nektos/act/pkg/common"
)

func NewDockerNetworkCreateExecutor(name, subnet, ipRange, gateway string) common.Executor {
	return func(ctx context.Context) error {
		if common.Dryrun(ctx) {
			return nil
		}

		cli, err := GetDockerClient(ctx)
		if err != nil {
			return err
		}

		if exists := DockerNetworkExists(ctx, name); exists {
			return nil
			// return fmt.Errorf("network '%s' already exists", name)
		}

		if _, err = cli.NetworkCreate(ctx, name, types.NetworkCreate{}); err != nil {
			return err
		}

		return nil
	}
}

func NewDockerNetworkRemoveExecutor(name string) common.Executor {
	return func(ctx context.Context) error {
		if common.Dryrun(ctx) {
			return nil
		}

		cli, err := GetDockerClient(ctx)
		if err != nil {
			return err
		}

		err = cli.NetworkRemove(ctx, name)
		if err != nil {
			return err
		}

		return nil
	}
}

func DockerNetworkExists(ctx context.Context, name string) bool {
	if exists, err := dockerNetworkExists(ctx, name); err != nil && !exists || err == nil && !exists {
		return false
	} else if exists && err == nil || exists && err != nil {
		return true
	}
	return false
}

func dockerNetworkExists(ctx context.Context, name string) (bool, error) {
	log := common.Logger(ctx)

	cli, err := GetDockerClient(ctx)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	_, err = cli.NetworkInspect(ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		if err.Error() == fmt.Sprintf("Error: No such %s: %s", "network", name) {
			log.Debug(err)
			return false, err
		}
		return false, err
	}

	return true, nil
}
