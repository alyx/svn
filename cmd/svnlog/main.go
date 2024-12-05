package main

import (
	"flag"
	"fmt"
	"github.com/jhinrichsen/svn"
	"io"
	"log/slog"
)

func main() {
	logger := slog.Default()
	slog.SetDefault(logger)
	slog.Info("starting svnlog")
	repository := flag.String("repository", "", "svn repository")
	firstCommit := flag.String("first", "", "first commit")
	lastCommit := flag.String("last", "", "last commit")
	flag.Parse()

	r := svn.NewRepository(*repository)

	slog.Info("finding log entries")

	for _, url := range flag.Args() {
		var entries *svn.LogElement
		var err error
		if *firstCommit != "" && *lastCommit != "" {
			entries, err = r.LogByRange(url, io.Discard, *firstCommit, *lastCommit)
		} else {
			entries, err = r.Log(url, io.Discard)
		}
		if err != nil {
			slog.Error("error checking url", "url", url, "err", err)
		}
		es := entries.Logentry
		for _, newEntry := range es {
			slog.Debug("found new entry", "url", fmt.Sprintf("%s/%s", r.Location, url), "name", newEntry.Text)
			for _, path := range newEntry.Paths.Path {
				slog.Debug("found path", "kind", path.Kind, "action", path.Action, "path", path.Path)
			}
		}
	}
}
