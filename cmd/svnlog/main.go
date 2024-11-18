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
	flag.Parse()

	r := svn.NewRepository(*repository)

	log.Printf("finding log entries")

	for _, url := range flag.Args() {
		entries, err := r.Log(url, io.Discard)
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
