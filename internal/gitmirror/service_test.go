package gitmirror

import "testing"

func TestSortTreeEntriesDirectoriesFirstThenAlphabetically(t *testing.T) {
	entries := []TreeEntry{
		{Name: "zeta.wave", Type: "blob"},
		{Name: "tools", Type: "tree"},
		{Name: "README.md", Type: "blob"},
		{Name: "Examples", Type: "tree"},
		{Name: "alpha.wave", Type: "blob"},
		{Name: "std", Type: "tree"},
	}

	sortTreeEntries(entries)

	want := []string{"Examples", "std", "tools", "alpha.wave", "README.md", "zeta.wave"}
	for index, name := range want {
		if entries[index].Name != name {
			t.Fatalf("entry %d = %q, want %q", index, entries[index].Name, name)
		}
	}
}
