package notificationsmodule

import "testing"

func TestNew_WiresNotificationsModule(t *testing.T) {
	out := New(nil)

	if out.Repo == nil || out.Service == nil || out.Handler == nil {
		t.Fatalf("expected notifications module output to be fully wired")
	}
}
