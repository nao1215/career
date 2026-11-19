package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"
)

// starterYAML is the scaffold written by "career init". The __DATE__ token is
// replaced with today's date so a freshly created file is not already stale.
//
//go:embed templates/starter.yaml
var starterYAML []byte

// starterContent returns the scaffold with the date placeholder filled in.
func starterContent(now time.Time) []byte {
	date := fmt.Sprintf("%d年%d月%d日現在", now.Year(), int(now.Month()), now.Day())
	return []byte(strings.ReplaceAll(string(starterYAML), "__DATE__", date))
}

func (a *App) runInit(args []string) int {
	flagSet := newFlagSet("init", a.stderr)
	force := flagSet.Bool("force", false, "overwrite the file if it already exists")
	flagSet.BoolVar(force, "f", false, "shorthand for --force")
	flagSet.Usage = func() {
		writeLine(flagSet.Output(), "Write a starter resume YAML file you can edit.")
		writeLine(flagSet.Output(), "Usage: career init [PATH] [--force]")
		writeLine(flagSet.Output(), "PATH defaults to resume.yaml in the current directory.")
		printFlagDefaults(flagSet.Output(), flagSet)
	}
	if code, ok := a.parseFlags(flagSet, args); !ok {
		return code
	}
	if flagSet.NArg() > 1 {
		writeLine(a.stderr, "too many arguments")
		flagSet.Usage()
		return 1
	}

	target := flagSet.Arg(0)
	if target == "" {
		target = "resume.yaml"
	}
	dest := a.resolvePath(target)

	if !*force {
		if _, err := os.Stat(dest); err == nil {
			writef(a.stderr, "%s already exists; pass --force to overwrite\n", target)
			return 1
		}
	}

	if err := os.WriteFile(dest, starterContent(time.Now()), 0o600); err != nil {
		writef(a.stderr, "write %s: %v\n", target, err)
		return 1
	}
	writef(a.stdout, "wrote %s\n", target)
	writef(a.stdout, "edit it, then run: career generate %s --template cv\n", target)
	return 0
}
