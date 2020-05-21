package play

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Var struct {
	Name string

	DefaultFromVar string
	Default        string
	Required       bool

	Prompt                PromptType
	ForcePrompt           bool
	PromptMessage         string
	PromptSelectOptions   []string
	PromptMultiSelectGlue string
	PromptAnswer          string
}

type PromptType int

const (
	PromptInput PromptType = iota
	PromptMultiline
	PromptPassword
	PromptConfirm
	PromptSelect
	PromptMultiSelect
	PromptEditor
	Prompt
)

func ParsesVarsMap(varsI map[string]interface{}) map[string]*Var {
	vars := make(map[string]*Var)
	for key, val := range varsI {
		var value map[string]interface{}
		switch v := val.(type) {
		case map[interface{}]interface{}:
			var err error
			value, err = decode.ToMap(v)
			errors.Check(err)
		case string:
			value = make(map[string]interface{})
			value["default"] = v
		case bool:
			if v {
				value["default"] = "true"
			} else {
				value["default"] = "false"
			}
		case nil:
			value = make(map[string]interface{})
			value["default"] = ""
		default:
			unexpectedTypeVarValue(key, v)
		}
		vr := &Var{}
		vr.Parse(key, value)
		vars[key] = vr
	}
	return vars
}

func (vr *Var) Parse(k string, m map[string]interface{}) {
	vr.Name = k
	vr.ParseDefault(m)
	vr.ParseDefaultFromVar(m)
	vr.ParseRequired(m)
	vr.ParsePrompt(m)
	vr.ParseForcePrompt(m)
	vr.ParsePromptMessage(m)
	vr.ParsePromptSelectOptions(m)
	vr.ParsePromptMultiSelectGlue(m)
}

func (vr *Var) ParseRequired(v map[string]interface{}) {
	switch v["required"].(type) {
	case bool:
		vr.Required = v["required"].(bool)
	case string:
		s := v["required"].(string)
		if s == "true" || s == "1" {
			vr.Required = true
		} else if s == "false" || s == "0" || s == "" {
			vr.Required = false
		} else {
			unexpectedTypeVar(v, "required")
		}
	case nil:
	default:
		unexpectedTypeVar(v, "required")
	}
}

func (vr *Var) ParseDefault(v map[string]interface{}) {
	switch v["default"].(type) {
	case string:
		vr.Default = v["default"].(string)
	case nil:
	default:
		unexpectedTypeVar(v, "default")
	}
}

func (vr *Var) ParseDefaultFromVar(v map[string]interface{}) {
	switch v["defaultFromVar"].(type) {
	case string:
		vr.DefaultFromVar = v["defaultFromVar"].(string)
	case nil:
	default:
		unexpectedTypeVar(v, "defaultFromVar")
	}
}

func (vr *Var) ParsePrompt(v map[string]interface{}) {
	switch v["prompt"].(type) {
	case string:
		promptString := v["prompt"].(string)
		var prompt PromptType
		switch promptString {
		case "input":
			prompt = PromptInput
		case "multiline":
			prompt = PromptMultiline
		case "password":
			prompt = PromptPassword
		case "confirm":
			prompt = PromptConfirm
		case "select":
			prompt = PromptSelect
		case "multiselect":
			prompt = PromptMultiSelect
		case "editor":
			prompt = PromptEditor
		}
		vr.Prompt = prompt
	case nil:
	default:
		unexpectedTypeVar(v, "prompt")
	}
}

func (vr *Var) ParseForcePrompt(v map[string]interface{}) {
	switch v["forcePrompt"].(type) {
	case bool:
		vr.ForcePrompt = v["forcePrompt"].(bool)
	case string:
		s := v["forcePrompt"].(string)
		if s == "true" || s == "1" {
			vr.ForcePrompt = true
		} else if s == "false" || s == "0" || s == "" {
			vr.ForcePrompt = false
		} else {
			unexpectedTypeVar(v, "forcePrompt")
		}
	case nil:
	default:
		unexpectedTypeVar(v, "forcePrompt")
	}
	// logrus.Infof("vr.ForcePrompt %v", vr.ForcePrompt)
}

func (vr *Var) ParsePromptMessage(v map[string]interface{}) {
	switch v["promptMessage"].(type) {
	case string:
		vr.PromptMessage = v["promptMessage"].(string)
	case nil:
	default:
		unexpectedTypeVar(v, "promptMessage")
	}
	if vr.PromptMessage == "" {
		vr.PromptMessage = vr.Name
	}
}

func (vr *Var) ParsePromptSelectOptions(v map[string]interface{}) {
	switch v["promptSelectOptions"].(type) {
	case []string:
		vr.PromptSelectOptions = v["promptSelectOptions"].([]string)
	case nil:
		if vr.Prompt == PromptSelect || vr.Prompt == PromptMultiSelect {
			logrus.Fatalf("unexpected empty play var promptSelectOptions for %v", v)
		}
	default:
		unexpectedTypeVar(v, "promptSelectOptions")
	}
}
func (vr *Var) ParsePromptMultiSelectGlue(v map[string]interface{}) {
	switch v["promptMultiSelectGlue"].(type) {
	case string:
		vr.PromptMultiSelectGlue = v["promptMultiSelectGlue"].(string)
	case nil:
		if vr.Prompt == PromptMultiSelect {
			vr.PromptMultiSelectGlue = ","
		}
	default:
		unexpectedTypeVar(v, "promptMultiSelectGlue")
	}
}

func unexpectedTypeVarValue(k string, v interface{}) {
	logrus.Fatalf("unexpected var type %T value %v for key %v", v, v, k)
}
func unexpectedTypeVar(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "var")
}
