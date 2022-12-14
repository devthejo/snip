package variable

import (
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
)

type FromType int

const (
	FromValue FromType = iota
	FromVar
	FromFile
	FromRegister
)

type RunVar struct {
	FromType FromType
	LazyLoad bool
	Param    string
}

func CreateRunVar() *RunVar {
	runVar := &RunVar{
		FromType: FromValue,
	}
	return runVar
}

func (runVar *RunVar) Set(fromType FromType, param string) {
	runVar.FromType = fromType
	runVar.Param = param
	switch fromType {
	case FromValue:
		runVar.LazyLoad = false
	case FromVar:
		runVar.LazyLoad = true
	case FromRegister:
		runVar.LazyLoad = true
	case FromFile:
		runVar.LazyLoad = true
	default:
		runVar.LazyLoad = true
	}
}

func (runVar *RunVar) GetValue(ctxs ...RunVars) string {

	var r string
	switch runVar.FromType {
	case FromValue:
		r = runVar.Param
	case FromRegister:
		for _, c := range ctxs {
			r = runVar.getDefaultOfCtx(c, ctxs)
			if r != "" {
				break
			}
		}
	case FromVar:
		for _, c := range ctxs {
			r = runVar.getValueOfCtx(c, ctxs)
			if r != "" {
				break
			}
		}
	case FromFile:
		content, err := ioutil.ReadFile(runVar.Param)
		if err != nil {
			logrus.Debugf("unable to read from_file file %v, %v", runVar.Param, err)
		}
		r = string(content)
		r = strings.TrimRight(r, "\n")
	}
	return r
}

func (runVar *RunVar) getDefaultOfCtx(ctx RunVars, ctxs []RunVars) string {
	var r string
	k := runVar.Param
	varsDefault := ctx.GetDefaults()
	if v, ok := varsDefault.Get(k); ok {
		rv := v.(*RunVar)
		if rv != runVar {
			r = rv.GetValue(ctxs...)
		}
	}
	return r
}

func (runVar *RunVar) getValueOfCtx(ctx RunVars, ctxs []RunVars) string {
	var r string
	k := runVar.Param
	vars := ctx.GetValues()
	varsDefault := ctx.GetDefaults()
	if v, ok := vars.Get(k); ok {
		rv := v.(*RunVar)
		r = rv.GetValue(ctxs...)
	}
	if r == "" {
		if v, ok := varsDefault.Get(k); ok {
			rv := v.(*RunVar)
			r = rv.GetValue(ctxs...)
		}
	}
	return r
}
