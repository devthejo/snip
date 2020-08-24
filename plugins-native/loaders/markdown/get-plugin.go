package mainNative

import (
	"plugin"

	cmap "github.com/orcaman/concurrent-map"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

func getPlugin(plugins cmap.ConcurrentMap, k string) interface{} {
	plugInterface, ok := plugins.Get(k)
	var plug interface{}
	if ok {
		plug = plugInterface
	} else {
		mod := "./plugins/loaders/markdown/plugins/" + k + ".so"
		var err error
		plug, err = plugin.Open(mod)
		errors.Check(err)
		plugins.Set(k, plug)
	}
	return plug
}
