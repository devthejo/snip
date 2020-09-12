package loader

import (
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/plugin/processor"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	LoaderVars map[string]string

	Command                 []string
	DefaultsPlayProps       map[string]interface{}
	RequiredFiles           map[string]string
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error

	RegisterVars map[string]*registry.VarDef

	CfgPlaySubstitutionMap map[string]interface{}

	ParentBuildFile string
	BuildFile       string
}
