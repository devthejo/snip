package play

import (
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

func ParseMap(play *Play, playMap map[string]interface{}) {

	parseName(play, playMap)
	parseTitle(play, playMap)
	parseVars(play, playMap)
	parseCheckCommand(play, playMap)
	parseDependencies(play, playMap)
	parsePostInstall(play, playMap)
	parseSudo(play, playMap)
	parseSSH(play, playMap)

}

func hasField(Iface interface{}, FieldName string) bool {
	ValueIface := reflect.ValueOf(Iface)

	// Check if the passed interface is a pointer
	if ValueIface.Type().Kind() != reflect.Ptr {
		// Create a new type of Iface's Type, so we have a pointer to work with
		ValueIface = reflect.New(reflect.TypeOf(Iface))
	}

	// 'dereference' with Elem() and get the field by name
	Field := ValueIface.Elem().FieldByName(FieldName)
	if !Field.IsValid() {
		return false
	}
	return true
}

func parseName(play *Play, playMap map[string]interface{}) {
	switch playMap["name"].(type) {
	case string:
		play.Name = playMap["name"].(string)
	case nil:
	default:
		logrus.Fatalf("unexpected play name type %T value %v", playMap["name"], playMap["name"])
	}
}

func parseTitle(play *Play, playMap map[string]interface{}) {
	switch playMap["title"].(type) {
	case string:
		play.Title = playMap["title"].(string)
	case nil:
		title := play.Name
		title = strings.ReplaceAll(title, "/", " ")
		title = strings.ReplaceAll(title, "-", " ")
		title = strings.ReplaceAll(title, "_", " ")
		play.Title = title
	default:
		logrus.Fatalf("unexpected play title type %T value %v", playMap["name"], playMap["name"])
	}
}

func parseCheckCommand(play *Play, playMap map[string]interface{}) {
	switch playMap["checkCommand"].(type) {
	case string:
		play.CheckCommand = playMap["checkCommand"].(string)
	case nil:
	default:
		logrus.Fatalf("unexpected play checkCommand type %T value %v", playMap["checkCommand"], playMap["checkCommand"])
	}
}

func parseDependencies(play *Play, playMap map[string]interface{}) {
	switch playMap["dependencies"].(type) {
	case []interface{}:
		dependencies, err := decode.ToStrings(playMap["dependencies"])
		errors.Check(err)
		play.Dependencies = dependencies
	case nil:
	default:
		logrus.Fatalf("unexpected play dependencies type %T value %v", playMap["dependencies"], playMap["dependencies"])
	}
}

func parsePostInstall(play *Play, playMap map[string]interface{}) {
	switch playMap["postInstall"].(type) {
	case []interface{}:
		postInstall, err := decode.ToStrings(playMap["postInstall"])
		errors.Check(err)
		play.PostInstall = postInstall
	case nil:
	default:
		logrus.Fatalf("unexpected play postInstall type %T value %v", playMap["postInstall"], playMap["postInstall"])
	}
}

func parseSudo(play *Play, playMap map[string]interface{}) {
	switch playMap["sudo"].(type) {
	case bool:
		play.Sudo = playMap["sudo"].(bool)
	case string:
		s := playMap["sudo"].(string)
		if s == "true" || s == "1" {
			play.Sudo = true
		} else if s == "false" || s == "0" || s == "" {
			play.Sudo = false
		} else {
			logrus.Fatalf("unexpected play var sudo type %T value %v", playMap["sudo"], playMap["sudo"])
		}
	case nil:
	default:
		logrus.Fatalf("unexpected play var sudo type %T value %v", playMap["sudo"], playMap["sudo"])
	}
}

func parseSSH(play *Play, playMap map[string]interface{}) {
	switch playMap["ssh"].(type) {
	case bool:
		play.SSH = playMap["ssh"].(bool)
	case string:
		s := playMap["ssh"].(string)
		if s == "true" || s == "1" {
			play.SSH = true
		} else if s == "false" || s == "0" || s == "" {
			play.SSH = false
		} else {
			logrus.Fatalf("unexpected play var ssh type %T value %v", playMap["ssh"], playMap["ssh"])
		}
	case nil:
	default:
		logrus.Fatalf("unexpected play var ssh type %T value %v", playMap["ssh"], playMap["ssh"])
	}
}

func parseVars(play *Play, playMap map[string]interface{}) {
	switch playMap["vars"].(type) {
	case map[interface{}]interface{}:
		varsI := playMap["vars"].(map[interface{}]interface{})
		vars := make(map[string]*Var)
		for k, v := range varsI {
			key := k.(string)
			var val map[string]interface{}
			switch v.(type) {
			case map[interface{}]interface{}:
				var err error
				val, err = decode.ToMap(v.(map[interface{}]interface{}))
				errors.Check(err)
			case string:
				val = make(map[string]interface{})
				val["default"] = v.(string)
			case nil:
				val = make(map[string]interface{})
				val["default"] = ""
			default:
				logrus.Fatalf("unexpected play var type %T value %v", v, v)
			}
			vars[key] = parseVar(val)
		}
		play.Vars = vars
	case nil:
	default:
		logrus.Fatalf("unexpected play vars type %T value %v", playMap["name"], playMap["name"])
	}
}

func parseVar(v map[string]interface{}) *Var {
	vr := &Var{}
	parseVarRequired(vr, v)
	parseVarDefault(vr, v)
	parseVarDefaultFromVar(vr, v)
	parseVarPrompt(vr, v)
	parseVarPromptMessage(vr, v)
	parseVarPromptSelectOptions(vr, v)
	return vr
}

func parseVarRequired(vr *Var, v map[string]interface{}) {
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
			logrus.Fatalf("unexpected play var required type %T value %v", v["required"], v["required"])
		}
	case nil:
	default:
		logrus.Fatalf("unexpected play var required type %T value %v", v["required"], v["required"])
	}
}

func parseVarDefault(vr *Var, v map[string]interface{}) {
	switch v["default"].(type) {
	case string:
		vr.Default = v["default"].(string)
	case nil:
	default:
		logrus.Fatalf("unexpected play var default type %T value %v", v["default"], v["default"])
	}
}

func parseVarDefaultFromVar(vr *Var, v map[string]interface{}) {
	switch v["defaultFromVar"].(type) {
	case string:
		vr.DefaultFromVar = v["defaultFromVar"].(string)
	case nil:
	default:
		logrus.Fatalf("unexpected play var defaultFromVar type %T value %v", v["defaultFromVar"], v["defaultFromVar"])
	}
}

func parseVarPrompt(vr *Var, v map[string]interface{}) {
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
		logrus.Fatalf("unexpected play var prompt type %T value %v", v["prompt"], v["prompt"])
	}
}

func parseVarPromptMessage(vr *Var, v map[string]interface{}) {
	switch v["promptMessage"].(type) {
	case string:
		vr.PromptMessage = v["promptMessage"].(string)
	case nil:
	default:
		logrus.Fatalf("unexpected play var promptMessage type %T value %v", v["promptMessage"], v["promptMessage"])
	}
}

func parseVarPromptSelectOptions(vr *Var, v map[string]interface{}) {
	switch v["promptSelectOptions"].(type) {
	case []string:
		vr.PromptSelectOptions = v["promptSelectOptions"].([]string)
	case nil:
		if vr.Prompt == PromptSelect || vr.Prompt == PromptMultiSelect {
			logrus.Fatalf("unexpected empty play var promptSelectOptions for %v", v)
		}
	default:
		logrus.Fatalf("unexpected play var promptSelectOptions type %T value %v", v["promptSelectOptions"], v["promptSelectOptions"])
	}
}
