package play

import (
	"context"
	"os/exec"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/proc"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Cmd struct {
	Thread *proc.Thread

	ParentLoopRow *LoopRow
	CfgCmd        *CfgCmd

	Command string
	Args    []string
	Vars    map[string]string
	Sudo    bool
	SSH     bool

	IsMD   bool
	Logger *logrus.Entry
	Depth  int
}

func CreateCmd(ccmd *CfgCmd, ctx *RunCtx, parentLoopRow *LoopRow) *Cmd {
	parentPlay := parentLoopRow.ParentPlay
	app := ccmd.CfgPlay.App

	cmd := &Cmd{
		ParentLoopRow: parentLoopRow,
		CfgCmd:        ccmd,
		Command:       ccmd.Command,
		Args:          ccmd.Args,
		IsMD:          ccmd.IsMD,
		Sudo:          parentPlay.Sudo,
		SSH:           parentPlay.SSH,
		Thread:        proc.CreateThread(app),
	}

	depth := ccmd.Depth
	if parentLoopRow.IsLoopRowItem {
		depth = depth + 1
	}
	cmd.Depth = depth

	vars := make(map[string]string)
	for k, v := range ctx.VarsDefault.Items() {
		vars[k] = v.(string)
	}
	for k, v := range ctx.Vars.Items() {
		vars[k] = v.(string)
	}
	cmd.Vars = vars

	logKey := cmd.GetTreeKey()
	logger := logrus.WithFields(logrus.Fields{
		"tree": logKey,
	})
	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), cmd.Depth+1)
	logger = logger.WithContext(loggerCtx)
	cmd.Logger = logger
	cmd.Thread.Logger = logger

	return cmd
}

func (cmd *Cmd) GetTreeKey() string {
	parts := cmd.GetTreeKeyParts()
	for i, v := range parts {
		parts[i] = strings.ReplaceAll(v, "|", "-")
	}
	return strings.Join(parts, "|")

}
func (cmd *Cmd) GetTreeKeyParts() []string {
	var parts []string
	var parent interface{}
	parent = cmd.ParentLoopRow
	for {
		var part string
		switch p := parent.(type) {
		case *LoopRow:
			if p == nil {
				parent = nil
				break
			}
			part = p.GetKey()
			parent = p.ParentPlay
		case *Play:
			if p == nil {
				parent = nil
				break
			}
			part = p.GetKey()
			parent = p.ParentLoopRow
		case nil:
			parent = nil
		}
		if parent == nil {
			break
		}
		parts = append([]string{part}, parts...)
	}
	return parts
}

func (cmd *Cmd) Run() error {
	return cmd.Thread.Run(cmd.Main)
}

func (cmd *Cmd) Main() error {

	var labels []string
	if cmd.SSH {
		labels = append(labels, "ssh")
	}
	if cmd.Sudo {
		labels = append(labels, "sudo")
	}

	labelsStr := ""
	for _, label := range labels {
		labelsStr = labelsStr + "[" + label + "]"
	}

	// cmd.Logger.Info(cmd.Indent + "▶️  playing " + labelsStr)
	cmd.Logger.Info("⮞ playing " + labelsStr)
	cmd.Logger.Debugf("vars: %v", tools.JsonEncode(cmd.Vars))

	commandSlice := append([]string{cmd.Command}, cmd.Args...)
	// if cmd.Sudo {
	// 	commandSlice = append([]string{"sudo"}, commandSlice...)
	// }

	commandHook := func(c *exec.Cmd) error {
		c.Env = tools.EnvToPairs(cmd.Vars)
		return nil
	}

	cmd.Logger.Debugf("command: %v", shellquote.Join(commandSlice...))

	cmd.Thread.RunCmd(commandSlice, cmd.Logger, commandHook)

	return cmd.Thread.Error
}
