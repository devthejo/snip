package loader

import (
	snipplugin "github.com/devthejo/snip/plugin"
	"github.com/devthejo/snip/plugin/processor"
	"github.com/devthejo/snip/registry"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	LoaderVars map[string]string

	Command                    []string
	DefaultsPlayProps          map[string]interface{}
	RequiredFiles              map[string]string
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error

	RegisterVars map[string]*registry.VarDef

	CfgPlaySubstitutionMap map[string]interface{}

	ParentBuildFile string
	BuildFile       string
}
