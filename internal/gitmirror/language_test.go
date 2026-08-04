package gitmirror

import "testing"

func TestDetectLanguagesUsesMirrorTreeBytes(t *testing.T) {
	statistics := DetectLanguages([]TreeEntry{
		{Path: "src/compiler.rs", Size: 600},
		{Path: "std/io.wave", Size: 300},
		{Path: "tools/release.py", Size: 100},
		{Path: "assets/logo.png", Size: 500, Binary: true},
		{Path: "generated/output.rs", Size: 900, Generated: true},
	})

	if len(statistics) != 3 {
		t.Fatalf("statistics = %d, want 3", len(statistics))
	}
	if statistics[0].Name != "Rust" || statistics[0].Percentage != 60 {
		t.Fatalf("primary = %#v", statistics[0])
	}
	if statistics[1].Name != "Wave" || statistics[1].Percentage != 30 {
		t.Fatalf("wave = %#v", statistics[1])
	}
}

func TestDetectLanguagesRecognizesSpecialFiles(t *testing.T) {
	statistics := DetectLanguages([]TreeEntry{
		{Path: "Makefile", Size: 10},
		{Path: "containers/Dockerfile", Size: 10},
		{Path: "boot/start.S", Size: 10},
	})
	if len(statistics) != 3 {
		t.Fatalf("statistics = %#v", statistics)
	}
}
