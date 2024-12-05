// Error codes:
// 1: general error
// 2: bad commandline invocation (follow usage to resolve)
//

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"time"

	"github.com/jhinrichsen/fio"
	"github.com/jhinrichsen/svn"
)

func main() {
	logger := slog.Default()
	slog.SetDefault(logger)
	slog.Info("starting svnentries")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [--since|--sincefile] uri [uri]*\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "    Default timestamp format is RFC 3339, e.g. 'date --rfc-3339=seconds'")
		flag.PrintDefaults()
	}
	repository := flag.String("repository", "https://svn.apache.org/repos/asf/subversion",
		"Subversion repository to check")
	since := flag.String("since", DefaultSince().Format(time.RFC3339), "Use since timestamp")
	sincefile := flag.String("sincefile", "", "Use timestamp from file to check for new entries, takes precedence over since")
	sinceformat := flag.String("sinceformat", time.RFC3339, "Default timestamp format (RFC 3339)")
	flag.Parse()

	r := svn.NewRepository(*repository)

	// Prefer sincefile over since
	var t time.Time
	var err error
	if *sincefile != "" {
		t, err = fio.ModTime(*sincefile)
		if err != nil {
			slog.Error("cannot determine timestamp of file", "file", *sincefile, "err", err.Error())
		}
	} else {
		t, err = time.Parse(*sinceformat, *since)
		if err != nil {
			slog.Error("error parsing timestamp", "timestamp", *since, "format", *sinceformat, "err", err.Error())
			os.Exit(2)
		}
	}
	logger.Info("looking for new entries", "since", t)
	for _, url := range flag.Args() {
		es, err := r.List(url, ioutil.Discard)
		if err != nil {
			slog.Error("error checking", "url", url, "err", err.Error())
		}
		for _, newEntry := range svn.Since(es, t) {
			slog.Debug("found new entry", "url", fmt.Sprintf("%s/%s", r.Location, url), "name", newEntry.Name)
		}
	}
}

// DefaultSince returns the timestamp 24 hours ago
func DefaultSince() time.Time {
	t := time.Now()
	// minus one day
	return t.AddDate(0, 0, -1)
}
