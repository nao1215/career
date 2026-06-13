// Package cmd implements the career command-line interface: a small,
// dependency-light front end that loads a resume YAML file and renders it into a
// 履歴書 or 職務経歴書 PDF.
package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/nao1215/career/internal/pdf"
	"github.com/nao1215/career/internal/resume"
)

const devVersion = "dev"

// Version is overridden at build time via ldflags.
var Version = devVersion

func resolveVersion() string {
	if Version != devVersion {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return devVersion
}

// App holds the IO streams and working directory for one CLI invocation.
type App struct {
	stdout  io.Writer
	stderr  io.Writer
	stdin   io.Reader
	workDir string
}

// NewApp constructs an App bound to the given streams and working directory.
func NewApp(stdout, stderr io.Writer, stdin io.Reader, workDir string) *App {
	return &App{
		stdout:  stdout,
		stderr:  stderr,
		stdin:   stdin,
		workDir: workDir,
	}
}

// Run dispatches a single command invocation and returns the process exit code.
func (a *App) Run(args []string) int {
	if len(args) == 0 {
		a.printRootHelp(a.stdout)
		return 0
	}

	switch args[0] {
	case "help":
		if len(args) > 1 {
			return a.runHelp(args[1:])
		}
		a.printRootHelp(a.stdout)
		return 0
	case "-h", "--help":
		if len(args) > 1 {
			writef(a.stderr, "%q takes no arguments; use \"career help <command>\"\n\n", args[0])
			a.printRootHelp(a.stderr)
			return 1
		}
		a.printRootHelp(a.stdout)
		return 0
	case "generate", "gen":
		return a.runGenerate(args[1:])
	case "templates":
		return a.runTemplates(args[1:])
	case "version", "-v", "--version":
		if len(args) > 1 {
			writef(a.stderr, "%q takes no arguments\n\n", args[0])
			a.printRootHelp(a.stderr)
			return 1
		}
		writef(a.stdout, "career %s\n", resolveVersion())
		return 0
	default:
		writef(a.stderr, "unknown command: %s\n\n", args[0])
		a.printRootHelp(a.stderr)
		return 1
	}
}

func (a *App) runHelp(args []string) int {
	if len(args) == 0 {
		a.printRootHelp(a.stdout)
		return 0
	}
	name := args[0]
	rest := args[1:]

	switch name {
	case "generate", "gen", "templates":
		forwarded := make([]string, 0, len(args)+1)
		forwarded = append(forwarded, args...)
		forwarded = append(forwarded, "--help")
		return a.Run(forwarded)
	case "help", "version":
		if len(rest) > 0 {
			writef(a.stderr, "%q takes no arguments\n\n", name)
			a.printRootHelp(a.stderr)
			return 1
		}
		a.printRootHelp(a.stdout)
		return 0
	default:
		writef(a.stderr, "unknown command: %s\n\n", name)
		a.printRootHelp(a.stderr)
		return 1
	}
}

func (a *App) runGenerate(args []string) int {
	flagSet := newFlagSet("generate", a.stderr)
	templateName := flagSet.String("template", "rirekisho", "document template: rirekisho or shokumukeirekisho")
	flagSet.StringVar(templateName, "t", "rirekisho", "shorthand for --template")
	input := flagSet.String("input", "", "path to the resume YAML file (defaults to the first argument)")
	flagSet.StringVar(input, "i", "", "shorthand for --input")
	output := flagSet.String("output", "", "output PDF path (defaults to the template's name)")
	flagSet.StringVar(output, "o", "", "shorthand for --output")
	flagSet.Usage = func() {
		writeLine(flagSet.Output(), "Render a resume YAML file into a 履歴書 or 職務経歴書 PDF.")
		writeLine(flagSet.Output(), "Usage: career generate [INPUT] --template NAME [--input PATH] [--output PATH]")
		writeLine(flagSet.Output(), "The input file may be given as the first argument or via --input.")
		writeLine(flagSet.Output(), "Run \"career templates\" to list the available templates.")
		printFlagDefaults(flagSet.Output(), flagSet)
	}
	if code, ok := a.parseFlags(flagSet, args); !ok {
		return code
	}

	tmpl, ok := pdf.Lookup(*templateName)
	if !ok {
		writef(a.stderr, "unknown template: %s\n", *templateName)
		writeLine(a.stderr, "run \"career templates\" to list the available templates")
		return 1
	}

	inputPath := *input
	if inputPath == "" {
		inputPath = flagSet.Arg(0)
	}
	if inputPath == "" {
		writeLine(a.stderr, "no input file: pass a path as the first argument or via --input")
		return 1
	}
	if flagSet.NArg() > 1 || (*input != "" && flagSet.NArg() > 0) {
		writeLine(a.stderr, "too many arguments")
		flagSet.Usage()
		return 1
	}

	outputPath := *output
	if outputPath == "" {
		outputPath = tmpl.DefaultOutput
	}

	res, err := resume.Load(a.resolvePath(inputPath))
	if err != nil {
		writeLine(a.stderr, err)
		return 1
	}

	data, err := tmpl.Render(res)
	if err != nil {
		writeLine(a.stderr, err)
		return 1
	}

	dest := a.resolvePath(outputPath)
	if err := os.WriteFile(dest, data, 0o600); err != nil {
		writef(a.stderr, "write %s: %v\n", outputPath, err)
		return 1
	}
	writef(a.stdout, "wrote %s (%s, %d bytes)\n", outputPath, tmpl.Name, len(data))
	return 0
}

func (a *App) runTemplates(args []string) int {
	flagSet := newFlagSet("templates", a.stderr)
	flagSet.Usage = func() {
		writeLine(flagSet.Output(), "List the available document templates.")
		writeLine(flagSet.Output(), "Usage: career templates")
		printFlagDefaults(flagSet.Output(), flagSet)
	}
	if code, ok := a.parseFlags(flagSet, args); !ok {
		return code
	}
	if flagSet.NArg() > 0 {
		writeLine(a.stderr, "templates takes no arguments")
		return 1
	}

	for _, t := range pdf.Templates() {
		writef(a.stdout, "%s\n", t.Name)
		if len(t.Aliases) > 0 {
			writef(a.stdout, "  aliases: %s\n", strings.Join(t.Aliases, ", "))
		}
		writef(a.stdout, "  %s\n", t.Description)
		writef(a.stdout, "  default output: %s\n", t.DefaultOutput)
	}
	return 0
}

// resolvePath turns a relative path into one rooted at the working directory so
// the tool behaves predictably regardless of the process CWD.
func (a *App) resolvePath(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(a.workDir, path)
}

func (a *App) printRootHelp(w io.Writer) {
	writeLine(w, "Generate Japanese 履歴書 and 職務経歴書 PDFs from a YAML file.")
	writeLine(w, "")
	writeLine(w, "Usage:")
	writeLine(w, "  career <command> [arguments] [flags]")
	writeLine(w, "")
	writeLine(w, "Commands:")
	writeLine(w, "  generate    Render a resume YAML file into a PDF")
	writeLine(w, "  templates   List the available document templates")
	writeLine(w, "  version     Print the version")
	writeLine(w, "  help        Show command help")
	writeLine(w, "")
	writeLine(w, "Example:")
	writeLine(w, "  career generate resume.yaml --template rirekisho --output rireki.pdf")
}

// --- flag plumbing (shared by every subcommand) ---

type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	flagSet := flag.NewFlagSet(name, flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	return flagSet
}

func (a *App) parseFlags(flagSet *flag.FlagSet, args []string) (int, bool) {
	if helpRequested(args) {
		flagSet.SetOutput(a.stdout)
		flagSet.Usage()
		return 0, false
	}
	return parseArgs(flagSet, reorderArgs(flagSet, args))
}

func parseArgs(flagSet *flag.FlagSet, args []string) (int, bool) {
	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, false
		}
		return 1, false
	}
	return 0, true
}

func helpRequested(args []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}

// reorderArgs moves positional arguments after flags so that flags may appear in
// any position (e.g. "career generate resume.yaml -t rireki").
func reorderArgs(flagSet *flag.FlagSet, args []string) []string {
	valueFlags := map[string]bool{}
	flagSet.VisitAll(func(f *flag.Flag) {
		if bf, ok := f.Value.(boolFlag); !ok || !bf.IsBoolFlag() {
			valueFlags[f.Name] = true
		}
	})
	return reorder(args, valueFlags)
}

func reorder(args []string, valueFlags map[string]bool) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			flags = append(flags, arg)
			name := strings.TrimLeft(arg, "-")
			if !strings.Contains(name, "=") && valueFlags[name] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positionals = append(positionals, arg)
	}
	if len(positionals) == 0 {
		return flags
	}
	result := make([]string, 0, len(flags)+1+len(positionals))
	result = append(result, flags...)
	result = append(result, "--")
	result = append(result, positionals...)
	return result
}

func printFlagDefaults(w io.Writer, flagSet *flag.FlagSet) {
	flagSet.VisitAll(func(f *flag.Flag) {
		writef(w, "  -%s=%s\n      %s\n", f.Name, f.DefValue, f.Usage)
	})
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func writeLine(w io.Writer, v any) {
	_, _ = fmt.Fprintln(w, v)
}
