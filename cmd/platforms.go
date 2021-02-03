package cmd

import (
	"strings"

	"github.com/nektos/act/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (i *Input) newPlatforms() map[string]*model.Platform {
	platforms := map[string]*model.Platform{
		"ubuntu-latest":  i.createPlatform("ubuntu-latest", false, "node:12-buster-slim"),
		"ubuntu-20.04":   i.createPlatform("ubuntu-18.04", false, "node:12-buster-slim"),
		"ubuntu-18.04":   i.createPlatform("ubuntu-18.04", false, "node:12-buster-slim"),
		"ubuntu-16.04":   i.createPlatform("ubuntu-16.04", false, "node:12-stretch-slim"),
		"windows-latest": i.createPlatform("windows-latest", false, ""),
		"windows-2019":   i.createPlatform("windows-2019", false, ""),
		"macos-latest":   i.createPlatform("macos-latest", false, ""),
		"macos-10.15":    i.createPlatform("macos-10.15", false, ""),
	}

	for _, p := range i.platforms {
		pParts := strings.Split(p, "=")
		if len(pParts) == 2 {
			platform := platforms[pParts[0]]

			if platform.Supported == false {
				log.Warnf("\U0001F6A7  Unable to set custom image for non supported platform '%+v'", platform.Platform)
				continue
			}

			if platform.UseHost {
				log.Warnf("\U0001F6A7  Unable to set custom image for non-docker based platforms '%+v'", platform.Platform)
				continue
			}

			platform.Image = pParts[1]
		}
	}
	return platforms
}

func (i *Input) createPlatform(platform string, useHost bool, image string) *model.Platform {
	return &model.Platform{
		Platform: platform,
		UseHost:  useHost,
		Image:    image,
	}
}
