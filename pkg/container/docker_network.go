package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/nektos/act/pkg/common"
)

func NewDockerNetworkCreateExecutor(name string, config types.NetworkCreate) common.Executor {
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
	log := common.Logger(ctx)
	if _, exists, err := GetDockerNetwork(ctx, name); !exists {
		log.Error(err)
		return false
	}
	return true
}

func GetDockerNetwork(ctx context.Context, name string) (types.NetworkResource, bool, error) {
	log := common.Logger(ctx)

	cli, err := GetDockerClient(ctx)
	if err != nil {
		log.Debug(err)
		return types.NetworkResource{}, false, err
	}

	res, err := cli.NetworkInspect(ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		if err.Error() == fmt.Sprintf("Error: No such network: %s", name) {
			log.Error(err)
			return types.NetworkResource{}, false, err
		}
		return types.NetworkResource{}, false, err
	}

	return res, true, nil
}
