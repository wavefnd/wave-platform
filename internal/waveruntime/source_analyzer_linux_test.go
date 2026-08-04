//go:build cgo && linux

package waveruntime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeSourceAnalyzer(t *testing.T) {
	library := filepath.Join("..", "..", "build", "wave", "libwave-source-analyzer.so")
	if _, err := os.Stat(library); err != nil {
		t.Skip("Wave source analyzer artifact has not been built")
	}

	analyzer, err := OpenSourceAnalyzer(library)
	if err != nil {
		t.Fatalf("open analyzer: %v", err)
	}
	defer analyzer.Close()

	analysis, err := analyzer.Analyze(context.Background(), []byte(
		"import(\"std::io\");\n// fun ignored() {}\nexport(c) fun main() {}\nstruct Result {}\n",
	))
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if analysis.Engine != "wave" || analysis.ABI != 2 {
		t.Fatalf("unexpected analysis: %+v", analysis)
	}

	source := []byte("import(\"std::io\");\n// fun ignored() {}\nexport(c) fun main() {}\nstruct Result {}\n")
	want := map[string]bool{
		"keyword:import":              false,
		"string:\"std::io\"":          false,
		"comment:// fun ignored() {}": false,
		"keyword:export":              false,
		"keyword:fun":                 false,
		"keyword:struct":              false,
	}
	for _, token := range analysis.Tokens {
		key := token.Kind + ":" + string(source[token.Start:token.End])
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for token, found := range want {
		if !found {
			t.Fatalf("missing token %q in %+v", token, analysis.Tokens)
		}
	}
}
