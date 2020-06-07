package ssh

import (
	shellquote "github.com/kballard/go-shellquote"
	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
)

func Exec(cfg *sshclient.Config, mutableCmd *middleware.MutableCmd) error {
	client, err := sshclient.CreateClient(cfg)
	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	commandSlice := append([]string{}, mutableCmd.Args...)
	runCmd := shellquote.Join(commandSlice...)

	if err := client.Run(runCmd); err != nil {
		return err
	}

	return nil
}
