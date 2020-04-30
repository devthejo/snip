package decode

import (
	"github.com/mitchellh/mapstructure"
)

// ToStruct decodes a raw interface{} into the target struct
func ToStruct(raw interface{}, result interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           result,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(raw)
}
