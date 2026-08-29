package main

import "testing"

// hasTarget decides whether the runner appends the default spec directory. A
// false positive drops the target and atago then searches the whole repository;
// a false negative appends e2e/atago next to a target the caller already named,
// running those specs twice. Both are silent, so the flag/value split is pinned
// here.
func TestHasTarget(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args []string
		want bool
	}{
		"no arguments":                     {args: nil, want: false},
		"flag without a value":             {args: []string{"--verbose"}, want: false},
		"flag whose value follows":         {args: []string{"--filter", "generate"}, want: false},
		"flag with an inline value":        {args: []string{"--filter=generate"}, want: false},
		"value flag then another flag":     {args: []string{"--parallel", "2", "--verbose"}, want: false},
		"spec path":                        {args: []string{"e2e/atago/cli.atago.yaml"}, want: true},
		"spec directory":                   {args: []string{"e2e/atago"}, want: true},
		"flag with a value then spec path": {args: []string{"--filter", "generate", "e2e/atago"}, want: true},
		"value that looks like a path":     {args: []string{"--report", "junit"}, want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := hasTarget(tt.args); got != tt.want {
				t.Errorf("hasTarget(%q) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestFirstLine(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   string
		want string
	}{
		"single line without a newline": {in: "career v1.0.0", want: "career v1.0.0"},
		"multi-line banner":             {in: "career v1.0.0\nbuilt with go1.27\n", want: "career v1.0.0"},
		"leading newline":               {in: "\ncareer v1.0.0", want: ""},
		"empty":                         {in: "", want: ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := firstLine(tt.in); got != tt.want {
				t.Errorf("firstLine(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// findRepoRoot walks up looking for go.mod. Running it from the package
// directory must land on the checkout root, not on the package itself.
func TestFindRepoRoot(t *testing.T) {
	t.Parallel()

	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("findRepoRoot() returned an error: %v", err)
	}
	if root == "" {
		t.Fatal("findRepoRoot() returned an empty path")
	}
}
