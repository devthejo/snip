package play

import (
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Var struct {
	Name string

	Default string
	Value   string
	Depth   int

	Required bool

	PromptDefault bool
	PromptValue   bool

	Prompt                PromptType
	PromptMessage         string
	PromptIcons           survey.AskOpt
	PromptSelectOptions   []string
	PromptMultiSelectGlue string
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

func ParsesVarsMap(varsI map[string]interface{}, depth int) map[string]*Var {
	vars := make(map[string]*Var)
	for key, val := range varsI {
		var value map[string]interface{}
		switch v := val.(type) {
		case map[string]interface{}:
			value = v
		case map[interface{}]interface{}:
			var err error
			value, err = decode.ToMap(v)
			errors.Check(err)
		case string:
			value = make(map[string]interface{})
			value["value"] = v
		case bool:
			if v {
				value["value"] = "true"
			} else {
				value["value"] = "false"
			}
		case nil:
			value = make(map[string]interface{})
			value["value"] = ""
		default:
			unexpectedTypeVarValue(key, v)
		}
		vr := &Var{
			Depth: depth,
		}
		key = strings.ToUpper(key)
		vr.Parse(key, value)
		vars[key] = vr
	}
	return vars
}

func (vr *Var) Parse(k string, m map[string]interface{}) {
	vr.Name = k
	vr.ParseDefault(m)
	vr.ParseValue(m)
	vr.ParseRequired(m)
	vr.ParsePrompt(m)
	vr.ParsePromptDefault(m)
	vr.ParsePromptValue(m)
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
func (vr *Var) ParseValue(v map[string]interface{}) {
	switch v["value"].(type) {
	case string:
		vr.Value = v["value"].(string)
	case nil:
	default:
		unexpectedTypeVar(v, "value")
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

func (vr *Var) ParsePromptDefault(v map[string]interface{}) {
	switch v["prompt_default"].(type) {
	case bool:
		vr.PromptDefault = v["prompt_default"].(bool)
	case string:
		s := v["prompt_default"].(string)
		if s == "true" || s == "1" {
			vr.PromptDefault = true
		} else if s == "false" || s == "0" || s == "" {
			vr.PromptDefault = false
		} else {
			unexpectedTypeVar(v, "prompt_default")
		}
	case nil:
	default:
		unexpectedTypeVar(v, "prompt_default")
	}
}

func (vr *Var) ParsePromptValue(v map[string]interface{}) {
	switch v["prompt_value"].(type) {
	case bool:
		vr.PromptValue = v["prompt_value"].(bool)
	case string:
		s := v["prompt_value"].(string)
		if s == "true" || s == "1" {
			vr.PromptValue = true
		} else if s == "false" || s == "0" || s == "" {
			vr.PromptValue = false
		} else {
			unexpectedTypeVar(v, "prompt_value")
		}
	case nil:
	default:
		unexpectedTypeVar(v, "prompt_value")
	}
}

func (vr *Var) ParsePromptMessage(v map[string]interface{}) {
	switch v["prompt_message"].(type) {
	case string:
		vr.PromptMessage = v["prompt_message"].(string)
	case nil:
	default:
		unexpectedTypeVar(v, "prompt_message")
	}
	if vr.PromptMessage == "" {
		vr.PromptMessage = vr.Name
	}
}

func (vr *Var) ParsePromptSelectOptions(v map[string]interface{}) {
	switch v["prompt_select_options"].(type) {
	case []string:
		vr.PromptSelectOptions = v["prompt_select_options"].([]string)
	case nil:
		if vr.Prompt == PromptSelect || vr.Prompt == PromptMultiSelect {
			logrus.Fatalf("unexpected empty play var prompt_select_options for %v", v)
		}
	default:
		unexpectedTypeVar(v, "prompt_select_options")
	}
}
func (vr *Var) ParsePromptMultiSelectGlue(v map[string]interface{}) {
	switch v["prompt_multi_select_glue"].(type) {
	case string:
		vr.PromptMultiSelectGlue = v["prompt_multi_select_glue"].(string)
	case nil:
		if vr.Prompt == PromptMultiSelect {
			vr.PromptMultiSelectGlue = ","
		}
	default:
		unexpectedTypeVar(v, "prompt_multi_select_glue")
	}
}

func (v *Var) PromptVarDefault() {
	msg := v.GetPromptMessageDefault()
	v.PromptVar(&v.Default, msg)
}

func (v *Var) PromptVarValue() {
	msg := v.GetPromptMessageValue()
	v.PromptVar(&v.Value, msg)
	if v.Value == "" && v.Default != "" {
		v.Value = v.Default
	}
}

func (v *Var) PromptVar(ref *string, msg string) {
	if v.PromptIcons == nil {
		v.PromptIcons = survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = strings.Repeat("  ", v.Depth+4) + " •"
			icons.Question.Format = "default+hb"
		})
	}

	switch v.Prompt {
	default:
		v.AskInput(ref, msg)
	case PromptInput:
		v.AskInput(ref, msg)
	case PromptMultiline:
		v.AskMultiline(ref, msg)
	case PromptPassword:
		v.AskPassword(ref, msg)
	case PromptConfirm:
		v.AskConfirm(ref, msg)
	case PromptSelect:
		v.AskSelect(ref, msg)
	case PromptMultiSelect:
		v.AskMultiSelect(ref, msg)
	case PromptEditor:
		v.AskEditor(ref, msg)
	}
}

func (v *Var) GetPromptMessageDefault() string {
	msg := v.PromptMessage
	if v.Default != "" {
		msg += " (" + v.Default + ")"
	}
	msg += " :"
	return msg
}
func (v *Var) GetPromptMessageValue() string {
	msg := v.PromptMessage
	if v.Default != "" {
		msg += " (" + v.Default + ")"
	}
	msg += " :"
	return msg
}

func (v *Var) HandleAnswer(err error) {
	if err == terminal.InterruptErr {
		os.Exit(0)
	} else if err != nil {
		logrus.Error(err)
	}
}

func (v *Var) AskInput(ref *string, msg string) {
	prompt := &survey.Input{
		Message: msg,
	}
	err := survey.AskOne(prompt, ref, v.PromptIcons)
	v.HandleAnswer(err)
}

func (v *Var) AskMultiline(ref *string, msg string) {
	prompt := &survey.Multiline{
		Message: msg,
	}
	err := survey.AskOne(prompt, ref, v.PromptIcons)
	v.HandleAnswer(err)
}

func (v *Var) AskPassword(ref *string, msg string) {
	prompt := &survey.Password{
		Message: msg,
	}
	err := survey.AskOne(prompt, ref, v.PromptIcons)
	v.HandleAnswer(err)
}

func (v *Var) AskEditor(ref *string, msg string) {
	prompt := &survey.Editor{
		Message: msg,
	}
	err := survey.AskOne(prompt, ref, v.PromptIcons)
	v.HandleAnswer(err)
}

func (v *Var) AskConfirm(ref *string, msg string) {
	prompt := &survey.Confirm{
		Message: msg,
	}
	err := survey.AskOne(prompt, ref, v.PromptIcons)
	v.HandleAnswer(err)
}
func (v *Var) AskSelect(ref *string, msg string) {
	prompt := &survey.Select{
		Message: msg,
		Options: v.PromptSelectOptions,
	}
	err := survey.AskOne(prompt, ref, v.PromptIcons)
	v.HandleAnswer(err)
}
func (v *Var) AskMultiSelect(ref *string, msg string) {
	answer := []string{}
	prompt := &survey.MultiSelect{
		Message:  msg,
		Options:  v.PromptSelectOptions,
		PageSize: 10,
	}
	err := survey.AskOne(prompt, answer, v.PromptIcons)
	*ref = strings.Join(answer, v.PromptMultiSelectGlue)
	v.HandleAnswer(err)
}

func (v *Var) RegisterValueTo(vars cmap.ConcurrentMap) {
	if v.PromptValue && v.Value == "" {
		v.PromptVarValue()
	}
	if v.Value != "" {
		vars.Set(v.Name, v.Value)
	}
}

func (v *Var) RegisterDefaultTo(varsDefault cmap.ConcurrentMap) {
	if v.PromptDefault && v.Default == "" {
		v.PromptVarDefault()
	}
	val, ok := varsDefault.Get(v.Name)
	if !ok || val == nil || val.(string) == "" {
		varsDefault.Set(v.Name, v.Default)
	}
}

func (v *Var) HandleRequired(varsDefault cmap.ConcurrentMap, vars cmap.ConcurrentMap) {
	if !v.Required || v.Default != "" || v.Value != "" {
		return
	}

	val, ok := varsDefault.Get(v.Name)
	if ok && val != nil && val.(string) != "" {
		return
	}

	val, ok = vars.Get(v.Name)
	if ok && val != nil && val.(string) != "" {
		return
	}

	for {
		v.PromptVarDefault()
		if v.Default != "" {
			varsDefault.Set(v.Name, v.Default)
			break
		}
		logrus.Warnf(strings.Repeat("  ", v.Depth+2)+` variable "%v" is required and cannot be empty`, v.Name)
	}

}

func unexpectedTypeVarValue(k string, v interface{}) {
	logrus.Fatalf("unexpected var type %T value %v for key %v", v, v, k)
}
func unexpectedTypeVar(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "var")
}
