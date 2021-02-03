package cmd

import (
	"fmt"
	"strings"

	"github.com/nektos/act/pkg/model"
	log "github.com/sirupsen/logrus"
	reg "golang.org/x/sys/windows/registry"
)

func (i *Input) newPlatforms() map[string]*model.Platform {
	platforms := map[string]*model.Platform{
		"ubuntu-latest":  i.createPlatform("ubuntu-latest", false, "node:12.6-buster-slim"),
		"ubuntu-20.04":   i.createPlatform("ubuntu-18.04", false, "node:12.6-buster-slim"),
		"ubuntu-18.04":   i.createPlatform("ubuntu-18.04", false, "node:12.6-buster-slim"),
		"ubuntu-16.04":   i.createPlatform("ubuntu-16.04", false, "node:12.6-stretch-slim"),
		"windows-latest": i.createPlatform("windows-latest", false, "mcr.microsoft.com/windows/servercore:ltsc2019"),
		"windows-2019":   i.createPlatform("windows-2019", false, "mcr.microsoft.com/windows/servercore:ltsc2019"),
		"macos-latest":   i.createPlatform("macos-latest", false, ""),
		"macos-10.15":    i.createPlatform("macos-10.15", false, ""),
	}

	if i.windowsHostCompatibility {
		key, err := reg.OpenKey(reg.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, reg.QUERY_VALUE)
		if err != nil {
			log.Errorf("Failed to open registry. In order to use Windows platforms, please specify them manually via --platform, -P flag. (%s)", err)
		}
		defer key.Close()

		osversion, _, err := key.GetStringValue("ReleaseId")
		if err != nil {
			log.Errorf("Failed to open registry. In order to use Windows platforms, please specify them manually via --platform, -P flag. (%s)", err)
		}

		if osversion != "" {
			platforms["windows-latest"] = i.createPlatform("windows-2019", false, fmt.Sprintf("mcr.microsoft.com/windows:%s", osversion))
			platforms["windows-2019"] = i.createPlatform("windows-2019", false, fmt.Sprintf("mcr.microsoft.com/windows:%s", osversion))
		}
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
