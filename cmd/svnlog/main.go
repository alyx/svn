package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/jhinrichsen/svn"
)

func main() {
	logger := slog.Default()
	logger.Info("starting svnlog")
	repository := flag.String("repository", "", "svn repository")
	firstCommit := flag.String("first", "", "first commit")
	lastCommit := flag.String("last", "", "last commit")
	flag.Parse()

	r := svn.NewRepository(*repository, logger)

	logger.Info("finding log entries")

	for _, url := range flag.Args() {
		var entries *svn.LogElement
		var err error
		if *firstCommit != "" && *lastCommit != "" {
			entries, err = r.LogByRange(url, io.Discard, *firstCommit, *lastCommit)
		} else {
			entries, err = r.Log(url, io.Discard)
		}
		if err != nil {
			logger.Error("error checking url", "url", url, "err", err)
		}
		es := entries.Logentry
		for _, newEntry := range es {
			fmt.Fprintf(os.Stdout, fmt.Sprintf("%s/%s@%s\n", r.Location, url, newEntry.Revision))
			for _, path := range newEntry.Paths.Path {
				fmt.Fprintf(os.Stdout, fmt.Sprintf("--> %s/%s: %s\n", path.Kind, path.Action, path.Path))
			}
		}
	}
}
