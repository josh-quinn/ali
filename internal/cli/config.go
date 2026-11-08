package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/settings"
	"github.com/josh-quinn/ali/internal/store"
)

func (a *app) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or change ali's settings for this machine",
		Long: `Settings live in settings.toml next to your aliases and stay on this machine
(they are not part of aliases.toml, so they do not travel with your dotfiles).

  expand   inline (default): sub-aliases and placeholders expand in your command
                    line before you press Enter, and history keeps the full command.
           runtime: the line stays as you typed it and expands when it runs. Only
                    lines that start with an ali alias expand in this mode.
  prompt   right (default): show the active environment in the right prompt.
           off: do not touch the prompt (use $(ali_prompt_info) in your theme).`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "get KEY",
		Short:   "Print a setting",
		Long:    `Print the value of setting KEY (expand or prompt).`,
		Example: `  ali config get expand`,
		Args:    args(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := settings.Load(a.st.Dir)
			if err != nil {
				return err
			}
			v, err := s.GetKey(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(a.out, v)
			return nil
		},
	}, &cobra.Command{
		Use:   "set KEY VALUE",
		Short: "Change a setting",
		Long:  `Change setting KEY to VALUE and regenerate your aliases file so the change applies at your next prompt.`,
		Example: `  ali config set expand runtime
  ali config set prompt off`,
		Args: args(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := settings.Load(a.st.Dir)
			if err != nil {
				return err
			}
			if err := s.SetKey(args[0], args[1]); err != nil {
				return err
			}
			if err := settings.Save(a.st.Dir, s); err != nil {
				return err
			}
			// regenerate aliases.zsh with the new setting
			if err := a.st.Update(func(*store.Data) error { return nil }); err != nil {
				return err
			}
			fmt.Fprintf(a.out, "%s = %s\n", args[0], args[1])
			return nil
		},
	}, &cobra.Command{
		Use:     "ls",
		Short:   "List all settings",
		Long:    `List every setting and its current value.`,
		Example: `  ali config ls`,
		Args:    args(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := settings.Load(a.st.Dir)
			if err != nil {
				return err
			}
			for _, k := range settings.Keys {
				v, _ := s.GetKey(k)
				fmt.Fprintf(a.out, "%s = %s\n", k, v)
			}
			return nil
		},
	})
	return cmd
}
