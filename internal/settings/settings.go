// Package settings holds machine-local preferences that must not travel with a
// synced aliases.toml (ADR 0037): the expansion mode, the prompt indicator and
// the global environment.
package settings

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/josh-quinn/ali/internal/fsx"
	"github.com/josh-quinn/ali/internal/model"
)

// Files in the config directory.
const (
	FileName      = "settings.toml"
	GlobalEnvFile = "global-env"
)

// Expansion modes (ADR 0030) and prompt indicator placements.
const (
	Inline      = "inline"
	Runtime     = "runtime"
	PromptRight = "right"
	PromptOff   = "off"
)

// Settings are the machine-local preferences.
type Settings struct {
	Expand string `toml:"expand"`
	Prompt string `toml:"prompt"`
}

// Defaults returns the settings used when nothing is configured.
func Defaults() Settings { return Settings{Expand: Inline, Prompt: PromptRight} }

// Validate checks every field.
func (s Settings) Validate() error {
	if s.Expand != Inline && s.Expand != Runtime {
		return fmt.Errorf("expand must be %q or %q, not %q", Inline, Runtime, s.Expand)
	}
	if s.Prompt != PromptRight && s.Prompt != PromptOff {
		return fmt.Errorf("prompt must be %q or %q, not %q", PromptRight, PromptOff, s.Prompt)
	}
	return nil
}

// Keys lists the settable keys.
var Keys = []string{"expand", "prompt"}

// GetKey returns a setting by name.
func (s Settings) GetKey(key string) (string, error) {
	switch key {
	case "expand":
		return s.Expand, nil
	case "prompt":
		return s.Prompt, nil
	}
	return "", fmt.Errorf("unknown setting %q (known: %s)", key, strings.Join(Keys, ", "))
}

// SetKey changes a setting by name, validating the value.
func (s *Settings) SetKey(key, value string) error {
	next := *s
	switch key {
	case "expand":
		next.Expand = value
	case "prompt":
		next.Prompt = value
	default:
		return fmt.Errorf("unknown setting %q (known: %s)", key, strings.Join(Keys, ", "))
	}
	if err := next.Validate(); err != nil {
		return err
	}
	*s = next
	return nil
}

// Load reads settings.toml, filling defaults for anything unset.
func Load(dir string) (Settings, error) {
	s := Defaults()
	raw, err := os.ReadFile(filepath.Join(dir, FileName)) // #nosec G304 -- reading the user's own config file is the purpose
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	dec := toml.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return Defaults(), fmt.Errorf("%s: %w", filepath.Join(dir, FileName), err)
	}
	if err := s.Validate(); err != nil {
		return Defaults(), fmt.Errorf("%s: %w", filepath.Join(dir, FileName), err)
	}
	return s, nil
}

// Save writes settings.toml with private permissions.
func Save(dir string, s Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	body, err := toml.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return fsx.WriteAtomic(filepath.Join(dir, FileName), body, 0o600)
}

// GlobalEnv returns the global environment name, or "" if none. The prompt
// hook reads the same file directly in zsh, so it holds just the name.
func GlobalEnv(dir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dir, GlobalEnvFile)) // #nosec G304 -- reading the user's own config file is the purpose
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(raw))
	if name == "" {
		return "", nil
	}
	if err := model.ValidateEnvName(name); err != nil {
		return "", fmt.Errorf("%s: %w", filepath.Join(dir, GlobalEnvFile), err)
	}
	return name, nil
}

// SetGlobalEnv writes the global environment name ("" clears it).
func SetGlobalEnv(dir, name string) error {
	if name != "" {
		if err := model.ValidateEnvName(name); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	body := ""
	if name != "" {
		body = name + "\n"
	}
	return fsx.WriteAtomic(filepath.Join(dir, GlobalEnvFile), []byte(body), 0o600)
}
