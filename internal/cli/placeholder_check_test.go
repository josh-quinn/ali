package cli

import (
	"strings"
	"testing"
)

func TestSaveRejectsPlaceholdersInsideQuotes(t *testing.T) {
	e := newEnv(t)
	e.must("var", "add", "ns", "one=x")
	for _, args := range [][]string{
		{"add", "a", "echo '{{ns}}'"},
		{"add", "b", `kubectl -n "{{ns}}"`},
		{"sub", "add", "kubectl", "n", "-n '{{ns}}'"},
	} {
		if _, _, err := e.run(args...); err == nil || !strings.Contains(err.Error(), "inside quotes") {
			t.Errorf("%q: want inside-quotes error, got %v", args, err)
		}
	}
	e.must("add", "ok", "kubectl -n {{ns}}")
	if _, _, err := e.run("update", "ok", "echo '{{ns}}'"); err == nil {
		t.Error("update should apply the same check")
	}
}

func TestSaveNotesUndefinedPlaceholders(t *testing.T) {
	e := newEnv(t)
	_, errs, err := e.run("add", "tpl", "kubectl get po -o go-template={{end}}")
	if err != nil || !strings.Contains(errs, "{{end}} is not a variable") {
		t.Errorf("want a note, got err=%v stderr=%q", err, errs)
	}
	// quoted and undefined: fine, literal, no note needed
	_, errs, err = e.run("add", "tpl2", "kubectl get po -o 'go-template={{range .items}}{{end}}'")
	if err != nil || strings.Contains(errs, "not a variable") {
		t.Errorf("err=%v stderr=%q", err, errs)
	}
}
