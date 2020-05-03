package play

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

func ParseInterface(playI interface{}, app App) *Play {
	play := &Play{}
	switch playI.(type) {
	case string:
		play.Name = playI.(string)
	case map[interface{}]interface{}:
		playMap, err := decode.ToMap(playI)
		errors.Check(err)
		ParseMap(play, playMap)
	default:
		logrus.Fatalf("unexpected play type %T value %v", playI, playI)
	}
	ParseMarkdownFile(play, app)
	return play
}
