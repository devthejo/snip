package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"gitlab.com/youtopia.earth/ops/snip/errors"
)

func CmdMiddlewareSudo(app App) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "sudo",
		Short: "sudo middleware",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// cfg := app.GetConfig()

			info, err := os.Stdin.Stat()
			errors.Check(err)

			if info.Mode()&os.ModeNamedPipe == 0 {
				fmt.Println("The command is intended to work with pipes.")
				fmt.Println("Usage: echo my-command-to-wrap | snip middleware sudo")
				return
			}

			data, err := ioutil.ReadAll(os.Stdin)
			errors.Check(err)

			fmt.Printf("sudo %v", string(data))
		},
	}

	return cmd
}
