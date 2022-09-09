package loader

import (
	snipplugin "github.com/devthejo/snip/plugin"
	"github.com/devthejo/snip/plugin/processor"
	"github.com/devthejo/snip/registry"
	cmap "github.com/orcaman/concurrent-map"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	LoaderVars map[string]string

	Command                    []string
	DefaultsPlayProps          map[string]interface{}
	RequiredFiles              cmap.ConcurrentMap
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error

	RegisterVars map[string]*registry.VarDef

	CfgPlaySubstitutionMap map[string]interface{}

	ParentBuildFile string
	BuildFile       string
}
