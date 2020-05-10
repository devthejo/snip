package play

import (
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/sirupsen/logrus"
)

func PromptVars(app App, playbook []*Play) {

	// cfg := app.GetConfig()

	for _, p := range playbook {
		for _, v := range p.Vars {
			if (v.Required && v.Default == "" && v.DefaultFromVar == "") || v.PromptForce {
				switch v.Prompt {
				default:
					askInput(v)
				case PromptInput:
					askInput(v)
				case PromptMultiline:
					askMultiline(v)
				case PromptPassword:
					askPassword(v)
				case PromptConfirm:
					askConfirm(v)
				case PromptSelect:
					askSelect(v)
				case PromptMultiSelect:
					askMultiSelect(v)
				case PromptEditor:
					askEditor(v)
				}
			}
		}
	}

}

func promptMessage(v *Var) string {
	msg := v.PromptMessage
	if v.DefaultFromVar != "" {
		msg += " $(" + v.DefaultFromVar + ")"

	}
	if v.Default != "" {
		msg += " (" + v.Default + ")"
	}
	return msg
}

func handleAnswer(v *Var, err error) bool {
	if err == terminal.InterruptErr {
		os.Exit(0)
	} else if err != nil {
		logrus.Error(err)
	}
	if v.PromptAnswer == "" {
		v.PromptAnswer = v.Default
	}
	if v.Required && v.PromptAnswer == "" {
		return false
	}
	return true
}

func askInput(v *Var) {
	prompt := &survey.Input{
		Message: promptMessage(v),
	}
	err := survey.AskOne(prompt, &v.PromptAnswer)
	if !handleAnswer(v, err) {
		askInput(v)
	}
}

func askMultiline(v *Var) {
	prompt := &survey.Multiline{
		Message: promptMessage(v),
	}
	err := survey.AskOne(prompt, &v.PromptAnswer)
	if !handleAnswer(v, err) {
		askMultiline(v)
	}
}

func askPassword(v *Var) {
	prompt := &survey.Password{
		Message: promptMessage(v),
	}
	err := survey.AskOne(prompt, &v.PromptAnswer)
	if !handleAnswer(v, err) {
		askPassword(v)
	}
}

func askEditor(v *Var) {
	prompt := &survey.Editor{
		Message: promptMessage(v),
	}
	err := survey.AskOne(prompt, &v.PromptAnswer)
	if !handleAnswer(v, err) {
		askEditor(v)
	}
}

func askConfirm(v *Var) {
	prompt := &survey.Confirm{
		Message: promptMessage(v),
	}
	err := survey.AskOne(prompt, &v.PromptAnswer)
	if !handleAnswer(v, err) {
		askConfirm(v)
	}
}
func askSelect(v *Var) {
	prompt := &survey.Select{
		Message: promptMessage(v),
		Options: v.PromptSelectOptions,
	}
	err := survey.AskOne(prompt, &v.PromptAnswer)
	if !handleAnswer(v, err) {
		askSelect(v)
	}
}
func askMultiSelect(v *Var) {
	answer := []string{}
	prompt := &survey.MultiSelect{
		Message:  v.PromptMessage,
		Options:  v.PromptSelectOptions,
		PageSize: 10,
	}
	err := survey.AskOne(prompt, answer)
	v.PromptAnswer = strings.Join(answer, v.PromptMultiSelectGlue)
	if !handleAnswer(v, err) {
		askMultiSelect(v)
	}
}
