package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/taylormonacelli/dotchin"
	"github.com/taylormonacelli/goldbug"
)

var (
	verbose   bool
	noCache   bool
	logFormat string
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")

	flag.BoolVar(&noCache, "no-cache", false, "Dont fetch from cache.  Update cache.")

	flag.StringVar(&logFormat, "log-format", "", "Log format (text or json)")

	flag.Parse()

	if verbose || logFormat != "" {
		if logFormat == "json" {
			goldbug.SetDefaultLoggerJson(slog.LevelDebug)
		} else {
			goldbug.SetDefaultLoggerText(slog.LevelDebug)
		}
	}

	code := dotchin.Main(noCache)
	os.Exit(code)
}
