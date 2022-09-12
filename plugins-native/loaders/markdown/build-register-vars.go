package mainNative

import (
	"strings"

	"github.com/devthejo/snip/registry"
)

func BuildRegisterVars(registerVars map[string]*registry.VarDef, langCode string) string {

	var buildRegisterVarsFunc func(map[string]*registry.VarDef) string
	switch langCode {
	case "sh":
		buildRegisterVarsFunc = BuildRegisterVarsSh
	case "bash":
		buildRegisterVarsFunc = BuildRegisterVarsSh
	case "js":
		buildRegisterVarsFunc = BuildRegisterVarsJs
	case "node":
		buildRegisterVarsFunc = BuildRegisterVarsJs
	}
	content := ""
	if buildRegisterVarsFunc != nil && len(registerVars) > 0 {
		content += buildRegisterVarsFunc(registerVars)
	}
	return content
}

func BuildRegisterVarsSh(registerVars map[string]*registry.VarDef) string {
	content := "\n\n# snip vars export \n"
	for _, vr := range registerVars {
		if !vr.Enable {
			continue
		}
		content += `echo "${` + vr.GetSource() + `}">${SNIP_VARS_TREEPATH}/` + vr.GetFrom() + "\n"
	}
	return content
}

func BuildRegisterVarsJs(registerVars map[string]*registry.VarDef) string {
	content := "\n\n// snip vars export \n"
	content += ";(async () => {\n"
	content += "const fs = require('fs/promises')\n"
	content += "const promises = []\n"
	content += "const exported = Object.fromEntries(Object.entries(module.exports).map(([k, v]) => [k.toLowerCase(), v]))\n"
	for _, vr := range registerVars {
		if !vr.Enable {
			continue
		}
		content += "promises.push(fs.writeFile(`${process.env.SNIP_VARS_TREEPATH}/" + vr.GetFrom() + "`, exported[" + `"` + strings.ToLower(vr.GetSource()) + `"` + "] || '' ))\n"
	}
	content += "await Promise.all(promises)\n"
	content += "})()"
	return content
}
