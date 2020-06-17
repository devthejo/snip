package loader

import (
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
)

type Config struct {
	AppConfig         *snipplugin.AppConfig
	Command           []string
	DefaultsPlayProps map[string]interface{}
	RequiredFiles     map[string]string
}
