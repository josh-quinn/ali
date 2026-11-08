package store

import (
	"fmt"
	"sort"

	"github.com/josh-quinn/ali/internal/model"
	"github.com/josh-quinn/ali/internal/render"
	"github.com/josh-quinn/ali/internal/settings"
)

// Feature 003 tables (data-model.md). They are written only when used, and
// their presence makes the file version 2 (ADR 0037).
type expansion struct {
	subs   map[string]map[string]string // tool -> name -> expansion
	shared map[string]string
	vars   map[string]model.Variable
	envs   map[string]model.Environment
}

func newExpansion() expansion {
	return expansion{subs: map[string]map[string]string{}, shared: map[string]string{},
		vars: map[string]model.Variable{}, envs: map[string]model.Environment{}}
}

func (e expansion) used() bool {
	return len(e.subs) > 0 || len(e.shared) > 0 || len(e.vars) > 0 || len(e.envs) > 0
}

// Subs returns a copy of tool's sub-aliases.
func (d *Data) Subs(tool string) map[string]string { return copyMap(d.x.subs[tool]) }

// Tools lists tools that have sub-aliases, sorted.
func (d *Data) Tools() []string { return sortedKeys(d.x.subs) }

// SetSub adds or replaces a tool's sub-alias.
func (d *Data) SetSub(tool, name, expansion string) {
	if d.x.subs[tool] == nil {
		d.x.subs[tool] = map[string]string{}
	}
	d.x.subs[tool][name] = expansion
}

// DeleteSub removes a tool's sub-alias and reports whether it existed.
func (d *Data) DeleteSub(tool, name string) bool {
	_, ok := d.x.subs[tool][name]
	delete(d.x.subs[tool], name)
	if len(d.x.subs[tool]) == 0 {
		delete(d.x.subs, tool)
	}
	return ok
}

// Shared returns a copy of the shared sub-aliases.
func (d *Data) Shared() map[string]string { return copyMap(d.x.shared) }

// SetShared adds or replaces a shared sub-alias.
func (d *Data) SetShared(name, expansion string) { d.x.shared[name] = expansion }

// DeleteShared removes a shared sub-alias and reports whether it existed.
func (d *Data) DeleteShared(name string) bool {
	_, ok := d.x.shared[name]
	delete(d.x.shared, name)
	return ok
}

// Var returns a variable.
func (d *Data) Var(name string) (model.Variable, bool) { v, ok := d.x.vars[name]; return v, ok }

// VarNames lists variables, sorted.
func (d *Data) VarNames() []string { return sortedKeys(d.x.vars) }

// SetVar adds or replaces a variable.
func (d *Data) SetVar(name string, v model.Variable) { d.x.vars[name] = v }

// DeleteVar removes a variable and reports whether it existed.
func (d *Data) DeleteVar(name string) bool {
	_, ok := d.x.vars[name]
	delete(d.x.vars, name)
	return ok
}

// Env returns an environment.
func (d *Data) Env(name string) (model.Environment, bool) { e, ok := d.x.envs[name]; return e, ok }

// EnvNames lists environments, sorted.
func (d *Data) EnvNames() []string { return sortedKeys(d.x.envs) }

// SetEnv adds or replaces an environment.
func (d *Data) SetEnv(name string, e model.Environment) { d.x.envs[name] = e }

// DeleteEnv removes an environment and reports whether it existed.
func (d *Data) DeleteEnv(name string) bool {
	_, ok := d.x.envs[name]
	delete(d.x.envs, name)
	return ok
}

// RenderInput collects what aliases.zsh is generated from, including the
// machine-local settings that affect it.
func (s *Store) RenderInput(d *Data) render.Input {
	subs := map[string]map[string]string{}
	for tool := range d.x.subs {
		subs[tool] = d.Subs(tool)
	}
	in := render.Input{Aliases: d.All(), Subs: subs, Shared: d.Shared(), Vars: d.VarNames(), Envs: d.EnvNames()}
	for _, n := range in.Envs {
		if e := d.x.envs[n]; e.Protected {
			in.Protected = append(in.Protected, n)
		}
	}
	// A broken settings file must not stop aliases being written; defaults apply
	// and "ali config" reports the problem.
	st, _ := settings.Load(s.Dir)
	in.PromptOff = st.Prompt == settings.PromptOff
	return in
}

// UsesOf lists what references variable name: "alias X", "sub TOOL X",
// "shared X" and "env X", sorted.
func (d *Data) UsesOf(name string) []string {
	var uses []string
	has := func(text string) bool {
		for _, p := range model.PlaceholderNames(text) {
			if p == name {
				return true
			}
		}
		return false
	}
	for _, a := range d.aliases {
		if has(a.Command) {
			uses = append(uses, "alias "+a.Name)
		}
	}
	for tool, subs := range d.x.subs {
		for n, exp := range subs {
			if has(exp) {
				uses = append(uses, "sub "+tool+" "+n)
			}
		}
	}
	for n, exp := range d.x.shared {
		if has(exp) {
			uses = append(uses, "shared "+n)
		}
	}
	for n, e := range d.x.envs {
		if _, ok := e.Set[name]; ok {
			uses = append(uses, "env "+n)
		}
	}
	sort.Strings(uses)
	return uses
}

func parseExpansion(path, key string, val any, d *Data) error {
	table, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: [%s] must be a table", path, key)
	}
	strs := func(where string, t any) (map[string]string, error) {
		m, ok := t.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: %s must be a table", path, where)
		}
		out := map[string]string{}
		for k, v := range m {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%s: %s.%s must be a string", path, where, k)
			}
			out[k] = s
		}
		return out, nil
	}
	switch key {
	case "_subs":
		for tool, t := range table {
			if err := model.ValidateTool(tool); err != nil {
				return fmt.Errorf("%s: [_subs.%s]: %w", path, tool, err)
			}
			subs, err := strs("[_subs."+tool+"]", t)
			if err != nil {
				return err
			}
			for n, exp := range subs {
				if err := checkSub(n, exp); err != nil {
					return fmt.Errorf("%s: [_subs.%s]: %w", path, tool, err)
				}
				d.SetSub(tool, n, exp)
			}
		}
	case "_shared":
		subs, err := strs("[_shared]", table)
		if err != nil {
			return err
		}
		for n, exp := range subs {
			if err := checkSub(n, exp); err != nil {
				return fmt.Errorf("%s: [_shared]: %w", path, err)
			}
			d.SetShared(n, exp)
		}
	case "_vars":
		for name, t := range table {
			if err := model.ValidateVarName(name); err != nil {
				return fmt.Errorf("%s: [_vars.%s]: %w", path, name, err)
			}
			body, _ := t.(map[string]any)
			v := model.Variable{Values: map[string]string{}}
			if def, ok := body["default"]; ok {
				if v.Default, ok = def.(string); !ok {
					return fmt.Errorf("%s: [_vars.%s] default must be a string", path, name)
				}
			}
			if vals, ok := body["values"]; ok {
				m, err := strs("[_vars."+name+".values]", vals)
				if err != nil {
					return err
				}
				v.Values = m
			}
			if err := v.Validate(); err != nil {
				return fmt.Errorf("%s: [_vars.%s]: %w", path, name, err)
			}
			d.SetVar(name, v)
		}
	case "_envs":
		for name, t := range table {
			if err := model.ValidateEnvName(name); err != nil {
				return fmt.Errorf("%s: [_envs.%s]: %w", path, name, err)
			}
			body, _ := t.(map[string]any)
			e := model.Environment{Set: map[string]string{}}
			if p, ok := body["protected"]; ok {
				if e.Protected, ok = p.(bool); !ok {
					return fmt.Errorf("%s: [_envs.%s] protected must be true or false", path, name)
				}
			}
			if set, ok := body["set"]; ok {
				m, err := strs("[_envs."+name+".set]", set)
				if err != nil {
					return err
				}
				e.Set = m
			}
			d.SetEnv(name, e)
		}
	}
	return nil
}

// checkRefs runs after the whole file is read, because environments may refer
// to variables defined later in the file.
func checkRefs(path string, d *Data) error {
	for name, e := range d.x.envs {
		for v, val := range e.Set {
			if _, ok := d.x.vars[v]; !ok {
				return fmt.Errorf("%s: [_envs.%s] sets unknown variable %q", path, name, v)
			}
			if err := model.ValidateValue(val); err != nil {
				return fmt.Errorf("%s: [_envs.%s] %s: %w", path, name, v, err)
			}
		}
	}
	return nil
}

func checkSub(name, exp string) error {
	if err := model.ValidateSubName(name); err != nil {
		return err
	}
	if exp == "" {
		return fmt.Errorf("sub-alias %q has an empty expansion", name)
	}
	return nil
}

func (e expansion) marshalInto(top map[string]any) {
	if len(e.subs) > 0 {
		subs := map[string]any{}
		for tool, m := range e.subs {
			subs[tool] = m
		}
		top["_subs"] = subs
	}
	if len(e.shared) > 0 {
		top["_shared"] = e.shared
	}
	if len(e.vars) > 0 {
		vars := map[string]any{}
		for n, v := range e.vars {
			body := map[string]any{"values": v.Values}
			if v.Default != "" {
				body["default"] = v.Default
			}
			vars[n] = body
		}
		top["_vars"] = vars
	}
	if len(e.envs) > 0 {
		envs := map[string]any{}
		for n, env := range e.envs {
			body := map[string]any{"set": env.Set}
			if env.Protected {
				body["protected"] = true
			}
			envs[n] = body
		}
		top["_envs"] = envs
	}
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
