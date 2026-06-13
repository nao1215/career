package pdf

import "testing"

func TestAccent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setting     string
		wantEnabled bool
		wantColor   rgb
		wantErr     bool
	}{
		{name: "default", setting: "", wantEnabled: true, wantColor: defaultAccent},
		{name: "none", setting: "none", wantEnabled: false, wantColor: black},
		{name: "none mixed case", setting: "None", wantEnabled: false, wantColor: black},
		{name: "hex with hash", setting: "#1f4e79", wantEnabled: true, wantColor: rgb{0x1f, 0x4e, 0x79}},
		{name: "hex without hash", setting: "ff0000", wantEnabled: true, wantColor: rgb{0xff, 0, 0}},
		{name: "padded", setting: "  #00ff00  ", wantEnabled: true, wantColor: rgb{0, 0xff, 0}},
		{name: "too short", setting: "#fff", wantErr: true},
		{name: "non hex", setting: "#zzzzzz", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			color, enabled, err := accent(tt.setting)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("accent(%q) error = nil, want error", tt.setting)
				}
				return
			}
			if err != nil {
				t.Fatalf("accent(%q) error = %v", tt.setting, err)
			}
			if enabled != tt.wantEnabled {
				t.Errorf("enabled = %v, want %v", enabled, tt.wantEnabled)
			}
			if enabled && color != tt.wantColor {
				t.Errorf("color = %v, want %v", color, tt.wantColor)
			}
		})
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		query string
		want  string
		found bool
	}{
		{"cv", "cv", true},
		{"japanese-resume", "japanese-resume", true},
		{"履歴書", "japanese-resume", true},
		{"career-history", "career-history", true},
		{"職務経歴書", "career-history", true},
		{"bogus", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			t.Parallel()
			got, ok := Lookup(tt.query)
			if ok != tt.found {
				t.Fatalf("Lookup(%q) found = %v, want %v", tt.query, ok, tt.found)
			}
			if ok && got.Name != tt.want {
				t.Errorf("Lookup(%q).Name = %q, want %q", tt.query, got.Name, tt.want)
			}
		})
	}
}
