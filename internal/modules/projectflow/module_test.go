package projectflowmodule

import "testing"

func TestNew_WiresProjectFlowModule(t *testing.T) {
	out := New(nil, nil, nil)

	if out.Repo == nil || out.Service == nil || out.Handler == nil {
		t.Fatalf("expected projectflow module output to be fully wired")
	}
}
