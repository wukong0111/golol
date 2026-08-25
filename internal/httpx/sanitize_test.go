package httpx

import (
	"strings"
	"testing"
)

func TestRewriteRiotTags(t *testing.T) {
	in := `<mainText><stats><attention>15</attention> de armadura</stats><br><br></mainText>`
	got := rewriteRiotTags(in)
	want := `<div class="maintext"><div class="stats"><span class="attention">15</span> de armadura</div><br><br></div>`
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSanitizeKeepsStats(t *testing.T) {
	p := descriptionPolicy()
	in := `<mainText><stats><attention>75</attention> de armadura</stats><passive>Espinas</passive></mainText><script>alert(1)</script>`
	got := sanitizeDescription(p, in)
	for _, part := range []string{`class="stats"`, `class="attention"`, `class="passive"`, "75", "Espinas"} {
		if !strings.Contains(got, part) {
			t.Fatalf("missing %q in %s", part, got)
		}
	}
	if strings.Contains(got, "<script") || strings.Contains(got, "alert") {
		t.Fatalf("script survived: %s", got)
	}
}
