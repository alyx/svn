package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jhinrichsen/svn"
)

func main() {
	log.Println("starting svnlog")
	repository := flag.String("repository", "", "svn repository")
	firstCommit := flag.String("first", "", "first commit")
	lastCommit := flag.String("last", "", "last commit")
	flag.Parse()

	r := svn.NewRepository(*repository)

	log.Printf("finding log entries")

	for _, url := range flag.Args() {
		var entries *svn.LogElement
		var err error
		if *firstCommit != "" && *lastCommit != "" {
			entries, err = r.LogByRange(url, io.Discard, *firstCommit, *lastCommit)
		} else {
			entries, err = r.Log(url, io.Discard)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error checking %q: %s\n", url, err)
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
