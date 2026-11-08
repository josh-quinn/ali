package cli

import "github.com/spf13/cobra"

// Help topics explain things that are not commands. They have no Run, so cobra
// lists them under "Additional help topics" and shows them with "ali help <topic>".
func topicCmds() []*cobra.Command {
	return []*cobra.Command{{
		Use:   "files",
		Short: "Where ali keeps your aliases and what it writes",
		Long: `ali keeps everything in one config directory:

  $ALI_CONFIG_DIR                if set, otherwise
  $XDG_CONFIG_HOME/ali           if XDG_CONFIG_HOME is an absolute path, otherwise
  ~/.config/ali

Files in that directory:

  aliases.toml        Your aliases, one [group] table each. Plain TOML: edit it, diff
                      it, keep it in a dotfiles repo (symlinks are followed).
  aliases.zsh         Generated from aliases.toml on every change and loaded by your
                      shell. Do not edit; your changes would be overwritten.
  aliases.toml.lock   Lock file that stops two ali commands writing at once.

ali creates the directory with mode 0700 and its files with mode 0600, because
commands often contain tokens or passwords.

Outside that directory ali only touches your .zshrc ($ZDOTDIR/.zshrc if ZDOTDIR is
an absolute path), and only when you run "ali init zsh". It writes a timestamped copy
next to it first, named .zshrc.ali-backup-<date>. "ali import" also backs up the
file it reads, but never changes it.

If HOME is not set and ALI_CONFIG_DIR is not given, ali refuses to run rather than
write into the current directory.`,
	}}
}
