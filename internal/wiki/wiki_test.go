package wiki

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewHandler(t *testing.T) {
	handler, mount, err := NewHandler("", "test", "https://github.com/nerdneilsfield/RSSHub-Gateway", zap.NewNop())
	if err != nil {
		t.Fatalf("expected handler, got error: %v", err)
	}
	if handler == nil {
		t.Fatalf("expected handler to be non-nil")
	}
	if mount != "/wiki" {
		t.Fatalf("unexpected mount: %s", mount)
	}
}
