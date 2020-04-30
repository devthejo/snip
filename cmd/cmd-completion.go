package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func CmdCompletion(app App, rootCmd *cobra.Command) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

	. <(snip completion)

	To configure your bash shell to load completions for each session add to your bashrc

	# ~/.bashrc or ~/.profile
	. <(snip completion)
	`,
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}
	return cmd
}
