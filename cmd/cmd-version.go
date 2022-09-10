package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"io"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/devthejo/snip/tools"
)

type BuildInfo struct {
	Version   string `json:"version,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
}

type versionOptions struct {
	short    bool
	template string
}

func CmdVersion(app App) *cobra.Command {
	o := &versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version informations",
		Long:  "show the version informations of snip build in json format",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(os.Stdout, app)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&o.short, "short", false, "print the version number")
	f.StringVar(&o.template, "template", "", "template for version string format")

	return cmd
}

func getBuildInfo(app App) BuildInfo {
	var vcsRevision = func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					return setting.Value
				}
			}
		}
		return ""
	}()
	return BuildInfo{
		Version:   app.GetVersion(),
		GitCommit: vcsRevision,
		GoVersion: runtime.Version(),
	}
}

func (o *versionOptions) run(out io.Writer, app App) error {
	if o.template != "" {
		tt, err := template.New("_").Parse(o.template)
		if err != nil {
			return err
		}
		return tt.Execute(out, getBuildInfo(app))
	}
	if o.short {
		getBuildInfo(app)
		fmt.Fprintln(out, app.GetVersion())
	} else {
		fmt.Fprintln(out, fmt.Sprintf("%s", tools.JsonEncode(getBuildInfo(app))))
	}
	return nil
}
