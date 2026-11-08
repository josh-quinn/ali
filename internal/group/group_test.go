package group

import "testing"

func TestInfer(t *testing.T) {
	cases := []struct{ in, want string }{
		{"git pull", "git"},
		{"helm get pods", "helm"},
		{"sudo helm get pods", "helm"},
		{"sudo -E helm ls", "helm"},
		{"env FOO=1 BAR=2 kubectl get po", "kubectl"},
		{"FOO=1 docker ps", "docker"},
		{"time npm test", "npm"},
		{"command ls -la", "ls"},
		{"./run.sh --fast", "run-sh"},
		{"/usr/local/bin/Terraform plan", "terraform"},
		{"~/bin/deploy", "deploy"},
		{"  cd ..  ", "cd"},
		{"", "misc"},
		{"sudo", "misc"},
		{"FOO=1", "misc"},
		{`"git" status`, "git"},
		{"!!", "misc"},
	}
	for _, c := range cases {
		if got := Infer(c.in); got != c.want {
			t.Errorf("Infer(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
