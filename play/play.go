package play

import (
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Play struct {
	App App

	RunCtx *RunCtx

	ParentLoopRow *LoopRow

	Index int
	Key   string
	Title string

	LoopRow []*LoopRow

	Vars map[string]*Var

	LoopSequential bool

	RegisterVars []string

	CheckCommand []string

	Dependencies []string
	PostInstall  []string

	Sudo bool
	SSH  bool

	Depth       int
	HasChildren bool

	State StateType
}

func (p *Play) GetTitle() string {
	title := p.Title
	if title == "" {
		title = p.GetKey()
	}
	return title
}

func (p *Play) GetKey() string {
	key := p.Key
	if key == "" {
		key = strconv.Itoa(p.Index)
	}
	return key
}

func (p *Play) Run() {

	var icon string
	if p.ParentLoopRow == nil {
		icon = `🠞`
	} else if !p.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}

	logrus.Info(strings.Repeat("  ", p.Depth+1) + icon + " " + p.GetTitle())

	runLoopSeq := func(loop *LoopRow) {
		if loop.IsLoopRowItem {
			logrus.Info(strings.Repeat("  ", p.Depth+2) + "⦿ " + loop.Name)
		}

		switch pl := loop.Play.(type) {
		case []*Play:
			for _, child := range pl {
				child.Run()
			}
		case *Cmd:
			pl.Run()
		}
	}

	var wg sync.WaitGroup
	var runLoopRow func(loop *LoopRow)

	if p.LoopSequential {
		runLoopRow = runLoopSeq
	} else {
		runLoopRow = func(loop *LoopRow) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runLoopSeq(loop)
			}()
		}
	}

	for _, loop := range p.LoopRow {
		runLoopRow(loop)
	}
	wg.Wait()

}

func (p *Play) Start() {
	logrus.Infof("🚀 running playbook")
	p.Run()
}

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
