package variable

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type FromType int

const (
	FromValue FromType = iota
	FromVar
	FromFile
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
	switch(fromType) {
	case FromValue:
		runVar.LazyLoad = false
	case FromVar:
		runVar.LazyLoad = true
	case FromFile:
		runVar.LazyLoad = true
	default:
		runVar.LazyLoad = true
	}
}

func (runVar *RunVar) GetValue(ctxs ...RunCtx) string {

	var r string
	switch runVar.FromType {
	case FromValue:
		r = runVar.Param
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
	}
	return r
}

func (runVar *RunVar) getValueOfCtx(ctx RunCtx, ctxs []RunCtx) string {
	var r string
	k := runVar.Param
	vars := ctx.GetVars()
	varsDefault := ctx.GetVarsDefault()
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
