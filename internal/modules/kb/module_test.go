package kbmodule

import "testing"

func TestNew_WiresKBModule(t *testing.T) {
	out := New(nil)

	if out.Repo == nil || out.Service == nil || out.Handler == nil {
		t.Fatalf("expected kb module output to be fully wired")
	}
}
