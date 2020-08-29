package mainNative

import (
	"path/filepath"
	"strings"
	// "path/filepath"

	"gopkg.in/yaml.v2"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

var (
	Loader = loader.Plugin{
		Check: func(cfg *loader.Config) bool {
			file := cfg.Command[0]
			return strings.HasSuffix(file, ".yml") ||
				strings.HasSuffix(file, ".yaml") ||
				strings.HasSuffix(file, ".json")
		},
		Load: func(cfg *loader.Config) error {
			s := loader.GetFileContent(cfg, cfg.Command[0])
			m := make(map[string]interface{}, 0)
			err := yaml.Unmarshal(s, &m)
			errors.Check(err)

			if m["key"] == nil {
				m["key"] = cfg.Command[0]
			}
			if m["title"] == nil {
				title := cfg.Command[0]
				title = strings.TrimSuffix(title, filepath.Ext(title))
				title = strings.ReplaceAll(title, "-", " ")
				title = strings.ReplaceAll(title, "/", " ")
				title = "snippet: " + title
				m["title"] = title

			}
			cfg.AddPlays = append(cfg.AddPlays, m)
			cfg.SkipItself = true

			return nil
		},
	}
)
