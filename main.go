package main

import (
	snipApp "github.com/devthejo/snip/app"
	"github.com/devthejo/snip/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version string

func main() {
	app := snipApp.NewApp(Version)
	cobra.OnInitialize(app.OnInitialize)

	RootCmd := cmd.NewCmd(app)
	app.RootCmd = RootCmd
	app.ConfigLoader.RootCmd = RootCmd

	if err := RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
