package frontend

import (
	"io/fs"
	"strings"
	"testing"
)

func TestFrontendHTMLDefaultsToKazakh(t *testing.T) {
	files, err := fs.Glob(Files, "*.html")
	if err != nil {
		t.Fatalf("glob frontend html: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected embedded html files")
	}

	for _, name := range files {
		data, err := Files.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(data)
		if !strings.Contains(content, `<html lang="kk">`) {
			t.Fatalf("%s does not declare kk as default html lang", name)
		}
		if !strings.Contains(content, `const lang = stored || "kk";`) {
			t.Fatalf("%s does not default bootstrap language to kk", name)
		}
		if !strings.Contains(content, `if (!stored) localStorage.setItem("idsai_site_lang", lang);`) {
			t.Fatalf("%s does not persist kk as initial language", name)
		}
	}
}
