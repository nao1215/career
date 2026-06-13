// Command career renders Japanese 履歴書 and 職務経歴書 PDFs from a YAML file.
package main

import (
	"os"

	"github.com/nao1215/career/cmd"
)

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	app := cmd.NewApp(os.Stdout, os.Stderr, os.Stdin, workDir)
	os.Exit(app.Run(os.Args[1:]))
}
