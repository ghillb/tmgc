package cli

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ghillb/tmgc/internal/config"
	"github.com/ghillb/tmgc/internal/output"
)

var version = "dev"

func SetVersion(v string) {
	if v != "" {
		version = v
	}
}

func Execute() error {
	cmd := newRootCmd()
	return cmd.Execute()
}

func newRootCmd() *cobra.Command {
	var (
		configPath string
		profile    string
		timeout    time.Duration
		jsonOut    bool
		plainOut   bool
		noColor    bool
	)

	cmd := &cobra.Command{
		Use:           "tmgc",
		Short:         "Telegram MTProto Go CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if jsonOut && plainOut {
				return errors.New("--json and --plain are mutually exclusive")
			}

			paths, err := config.ResolvePaths(configPath, profile)
			if err != nil {
				return err
			}
			if err := config.EnsureDirs(paths); err != nil {
				return err
			}

			cfg, err := config.Load(paths.ConfigPath)
			if err != nil {
				return err
			}
			if err := cfg.ApplyEnv(); err != nil {
				return err
			}

			mode := output.ModeHuman
			switch {
			case jsonOut:
				mode = output.ModeJSON
			case plainOut:
				mode = output.ModePlain
			}

			printer := output.NewPrinter(os.Stdout, os.Stderr, mode, noColor)
			rt := &Runtime{
				Paths:   paths,
				Config:  &cfg,
				Printer: printer,
				Timeout: timeout,
			}
			cmd.SetContext(withRuntime(cmd.Context(), rt))
			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&profile, "profile", "default", "profile name")
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", 15*time.Second, "request timeout")
	cmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "JSON output")
	cmd.PersistentFlags().BoolVar(&plainOut, "plain", false, "plain output")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newChatCmd())
	cmd.AddCommand(newContactCmd())
	cmd.AddCommand(newMessageCmd())
	cmd.AddCommand(newSearchCmd())

	cmd.SetHelpTemplate(helpTemplate())

	return cmd
}

func helpTemplate() string {
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:
  {{.UseLine}}

{{if .Aliases}}Aliases:
  {{.NameAndAliases}}

{{end}}{{if .HasAvailableSubCommands}}Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}

{{end}}Flags:
{{.Flags.FlagUsages | trimTrailingWhitespaces}}

{{if .HasAvailableInheritedFlags}}Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}`
}
