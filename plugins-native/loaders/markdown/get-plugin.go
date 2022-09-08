package mainNative

import (
	"plugin"

	"github.com/devthejo/snip/errors"
)

func getPlugin(k string) interface{} {
	plugInterface, ok := Plugins.Get(k)
	var plug interface{}
	if ok {
		plug = plugInterface
	} else {
		mod := "./plugins/loaders/markdown/plugins/" + k + ".so"
		var err error
		plug, err = plugin.Open(mod)
		errors.Check(err)
		Plugins.Set(k, plug)
	}
	return plug
}
