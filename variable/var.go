package variable

import (
	"os"
	"strconv"
	"strings"
	"sync"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
)

type Var struct {
	Name string

	Depth int

	Required bool

	PromptDefault bool
	PromptValue   bool

	Prompt                PromptType
	PromptMessage         string
	PromptIcons           survey.AskOpt
	PromptSelectOptions   []string
	PromptMultiSelectGlue string

	OnPrompt func(*Var)

	DefaultFromType FromType
	DefaultParam    string
	ValueFromType   FromType
	ValueParam      string
}

func (vr *Var) Parse(k string, m map[string]interface{}) {
	vr.Name = k
	vr.ParseDefaultFromVar(m)
	vr.ParseDefaultFromFile(m)
	vr.ParseDefault(m)
	vr.ParseValueFromVar(m)
	vr.ParseValueFromFile(m)
	vr.ParseValue(m)
	vr.ParseRequired(m)
	vr.ParsePrompt(m)
	vr.ParsePromptDefault(m)
	vr.ParsePromptValue(m)
	vr.ParsePromptMessage(m)
	vr.ParsePromptSelectOptions(m)
	vr.ParsePromptMultiSelectGlue(m)
}

func (vr *Var) SetValue(v string) {
	if v == "" {
		return
	}
	vr.ValueFromType = FromValue
	vr.ValueParam = v
}
func (vr *Var) SetDefault(v string) {
	if v == "" {
		return
	}
	vr.DefaultFromType = FromValue
	vr.DefaultParam = v
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
			UnexpectedTypeVar(v, "required")
		}
	case nil:
	default:
		UnexpectedTypeVar(v, "required")
	}
}

func (vr *Var) ParseDefault(v map[string]interface{}) {
	switch val := v["default"].(type) {
	case string:
		vr.SetDefault(val)
	case int:
		vr.SetDefault(strconv.Itoa(val))
	case bool:
		if val {
			vr.SetDefault("true")
		} else {
			vr.SetDefault("false")
		}
	case nil:
	default:
		UnexpectedTypeVar(v, "default")
	}
}
func (vr *Var) ParseValue(v map[string]interface{}) {
	switch val := v["value"].(type) {
	case string:
		vr.SetValue(val)
	case int:
		vr.SetValue(strconv.Itoa(val))
	case bool:
		if val {
			vr.SetValue("true")
		} else {
			vr.SetValue("false")
		}
	case nil:
	default:
		UnexpectedTypeVar(v, "value")
	}
}

func (vr *Var) ParseDefaultFromVar(v map[string]interface{}) {
	switch val := v["default_from_var"].(type) {
	case string:
		vr.DefaultFromType = FromVar
		vr.DefaultParam = strings.ToUpper(val)
	case nil:
	default:
		UnexpectedTypeVar(v, "default_from_var")
	}
}
func (vr *Var) ParseValueFromVar(v map[string]interface{}) {
	switch val := v["value_from_var"].(type) {
	case string:
		vr.ValueFromType = FromVar
		vr.ValueParam = strings.ToUpper(val)
	case nil:
		switch val := v["from_var"].(type) {
		case string:
			vr.ValueFromType = FromVar
			vr.ValueParam = strings.ToUpper(val)
		case nil:
		default:
			UnexpectedTypeVar(v, "from_var")
		}
	default:
		UnexpectedTypeVar(v, "value_from_var")
	}
}

func (vr *Var) ParseDefaultFromFile(v map[string]interface{}) {
	switch val := v["default_from_file"].(type) {
	case string:
		vr.DefaultFromType = FromFile
		vr.DefaultParam = val
	case nil:
	default:
		UnexpectedTypeVar(v, "default_from_file")
	}
}
func (vr *Var) ParseValueFromFile(v map[string]interface{}) {
	switch val := v["value_from_file"].(type) {
	case string:
		vr.ValueFromType = FromFile
		vr.ValueParam = val
	case nil:
		switch val := v["from_file"].(type) {
		case string:
			vr.ValueFromType = FromFile
			vr.ValueParam = val
		case nil:
		default:
			UnexpectedTypeVar(v, "from_file")
		}
	default:
		UnexpectedTypeVar(v, "value_from_file")
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
		UnexpectedTypeVar(v, "prompt")
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
			UnexpectedTypeVar(v, "prompt_default")
		}
	case nil:
	default:
		UnexpectedTypeVar(v, "prompt_default")
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
			UnexpectedTypeVar(v, "prompt_value")
		}
	case nil:
	default:
		UnexpectedTypeVar(v, "prompt_value")
	}
}

func (vr *Var) ParsePromptMessage(v map[string]interface{}) {
	switch v["prompt_message"].(type) {
	case string:
		vr.PromptMessage = v["prompt_message"].(string)
	case nil:
	default:
		UnexpectedTypeVar(v, "prompt_message")
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
			logrus.Fatalf("Unexpected empty play var prompt_select_options for %v", v)
		}
	default:
		UnexpectedTypeVar(v, "prompt_select_options")
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
		UnexpectedTypeVar(v, "prompt_multi_select_glue")
	}
}

func (v *Var) PromptVarDefault() {
	msg := v.GetPromptMessageDefault()
	v.DefaultFromType = FromValue
	v.PromptVar(&v.DefaultParam, msg)
}

func (v *Var) PromptVarValue() {
	msg := v.GetPromptMessageValue()
	v.ValueFromType = FromValue
	v.PromptVar(&v.ValueParam, msg)
	if v.ValueParam == "" && v.GetDefault() != "" {
		v.ValueParam = v.GetDefault()
	}
}

func (v *Var) GetDefault() string {
	var r string
	switch v.DefaultFromType {
	case FromValue:
		return v.DefaultParam
	}
	return r
}
func (v *Var) GetValue() string {
	var r string
	switch v.ValueFromType {
	case FromValue:
		return v.ValueParam
	}
	return r
}

func (v *Var) PromptVar(ref *string, msg string) {

	if v.OnPrompt != nil {
		v.OnPrompt(v)
	}

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
	if v.GetDefault() != "" {
		msg += " (" + v.GetDefault() + ")"
	}
	msg += " :"
	return msg
}
func (v *Var) GetPromptMessageValue() string {
	msg := v.PromptMessage
	if v.GetDefault() != "" {
		msg += " (" + v.GetDefault() + ")"
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

func (v *Var) PromptOnEmptyDefault() {
	if v.PromptDefault && v.GetDefault() == "" {
		v.PromptVarDefault()
	}
}

func (v *Var) PromptOnEmptyValue() {
	if v.PromptValue && v.GetValue() == "" {
		v.PromptVarValue()
	}
}

func (v *Var) RegisterValueTo(vars cmap.ConcurrentMap) {
	v.PromptOnEmptyValue()
	val, ok := vars.Get(v.Name)
	var runVar *RunVar
	if ok && val != nil {
		runVar = val.(*RunVar)
	} else {
		runVar = CreateRunVar()
		vars.Set(v.Name, runVar)
	}

	if v.ValueParam != "" {
		runVar.Set(v.ValueFromType, v.ValueParam)
	}
}

func (v *Var) RegisterDefaultTo(varsDefault cmap.ConcurrentMap) {
	v.PromptOnEmptyDefault()
	val, ok := varsDefault.Get(v.Name)
	var runVar *RunVar
	if ok && val != nil {
		runVar = val.(*RunVar)
	} else {
		runVar = CreateRunVar()
		varsDefault.Set(v.Name, runVar)
	}
	if v.DefaultParam != "" {
		runVar.Set(v.DefaultFromType, v.DefaultParam)
	}
}

func (v *Var) HandleRequired(varsDefault cmap.ConcurrentMap, vars cmap.ConcurrentMap) {
	if !v.Required || v.GetDefault() != "" || v.GetValue() != "" {
		return
	}

	var runVarDefault *RunVar
	if varsDefault != nil {
		val, ok := varsDefault.Get(v.Name)
		if ok && val != nil {
			runVarDefault = val.(*RunVar)
		} else {
			runVarDefault = CreateRunVar()
			varsDefault.Set(v.Name, runVarDefault)
		}
		if runVarDefault.Param != "" {
			return
		}
	}

	var runVar *RunVar
	if vars != nil {
		val, ok := vars.Get(v.Name)
		if ok && val != nil {
			runVar = val.(*RunVar)
		} else {
			runVar = CreateRunVar()
			varsDefault.Set(v.Name, runVar)
		}
		if runVar.Param != "" {
			return
		}
	}

	for {
		v.PromptVarDefault()
		if v.GetDefault() != "" {
			if varsDefault != nil {
				runVarDefault.Set(FromValue, v.DefaultParam)
			}
			break
		}
		logrus.Warnf(strings.Repeat("  ", v.Depth+2)+` variable "%v" is required and cannot be empty`, v.Name)
	}

}

func (v *Var) OnPromptMessageOnce(msg string, once *sync.Once) {
	v.OnPrompt = MakeOnPromptOnce(msg, once)
}
func (v *Var) OnPromptMessage(msg string) {
	v.OnPrompt = MakeOnPrompt(msg)
}
