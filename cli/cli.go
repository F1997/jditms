package cli

import (
	"jditms/cli/upgrade"
)

func Upgrade(configFile string) error {
	return upgrade.Upgrade(configFile)
}
