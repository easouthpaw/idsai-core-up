package adminmodule

import "testing"

func TestNew_WiresAdminModule(t *testing.T) {
	out := New(nil)

	if out.Repo == nil || out.Service == nil || out.Handler == nil {
		t.Fatalf("expected admin module output to be fully wired")
	}
}
