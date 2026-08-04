package account

import "testing"

func TestLocalPartFromDisplayName(t *testing.T) {
	for _, test := range []struct{ name, want string }{
		{name: "John Mark", want: "john-mark"},
		{name: "  Luna__Stev  ", want: "luna-stev"},
		{name: "홍 길동", want: "홍-길동"},
	} {
		got, err := LocalPart(test.name)
		if err != nil {
			t.Fatalf("LocalPart(%q): %v", test.name, err)
		}
		if got != test.want {
			t.Errorf("LocalPart(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}
