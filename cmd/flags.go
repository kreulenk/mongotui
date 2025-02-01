package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

// namedFlagSet allows us to group sets of flags together so that they can be displayed in groups
// under a common name for each group.
// This workaround is necessary as pflag's FlagSet does not make the name attribute public
type namedFlagSet struct {
	name    string
	flagset *pflag.FlagSet
}

func addFlagsAndSetHelpMenu(cmd *cobra.Command, sets []namedFlagSet) {
	var usages strings.Builder
	for _, set := range sets {
		cmd.Flags().AddFlagSet(set.flagset)
		usages.WriteString(fmt.Sprintf("%s:\n%s\n", set.name, set.flagset.FlagUsages()))
	}
	cmd.SetUsageTemplate(strings.TrimSpace(fmt.Sprintf(usageTemplate, usages.String())))
}

// usageTemplate is a custom template
// Its difference from the default cobra template is that it allows for the grouping of flags by flagSets
// The single %s will have the custom flagUsages from addFlagsAndSetHelpMenu templated in
const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
