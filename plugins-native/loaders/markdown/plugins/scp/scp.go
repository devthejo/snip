package mainNative

import (
	"github.com/sirupsen/logrus"

	loaderMardownMod "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/mod"
)

var (
	Mod = loaderMardownMod.Plugin{
		Mod: func(cfg *loaderMardownMod.Config) error {
			logrus.Warn("HELLO SCP")
			return nil
		},
	}
)
