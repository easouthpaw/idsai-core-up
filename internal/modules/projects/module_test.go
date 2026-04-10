package projectsmodule

import "testing"

func TestNew_WiresProjectsModule(t *testing.T) {
	out := New(nil, nil)

	if out.Repo == nil || out.Service == nil {
		t.Fatalf("expected projects module output to be fully wired")
	}
}
