package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/youtopia.earth/ops/snip/config"
)

func CmdMiddleware(app App) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "middleware",
		Short: "middleware wrapper for command",
	}
	cmd.AddCommand(CmdMiddlewareSudo(app))
	cmd.AddCommand(CmdMiddlewareSSH(app))

	pFlags := cmd.PersistentFlags()

	pFlags.StringP("build-dir", "", config.FlagBuildDirDefault, config.FlagBuildDirDesc)

	return cmd
}
