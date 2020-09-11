package loader

import (
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	LoaderVars map[string]string

	Command                 []string
	DefaultsPlayProps       map[string]interface{}
	RequiredFiles           map[string]string
	RequiredFilesProcessors map[string][]func(*runner.Config, *string) (func(), error)

	RegisterVars map[string]*registry.VarDef

	CfgPlaySubstitutionMap map[string]interface{}

	ParentBuildFile string
	BuildFile       string
}
