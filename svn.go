package svn

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

type LogElement struct {
	XMLName  xml.Name   `xml:"log"`
	Text     string     `xml:",chardata"`
	Logentry []LogEntry `xml:"logentry"`
}

type LogEntry struct {
	Text     string `xml:",chardata"`
	Revision string `xml:"revision,attr"`
	Author   string `xml:"author"`
	Date     string `xml:"date"`
	Paths    Paths  `xml:"paths"`
}

type Paths struct {
	Text string `xml:",chardata"`
	Path []Path `xml:"path"`
}

type Path struct {
	Path     string `xml:",chardata"`
	TextMods string `xml:"text-mods,attr"`
	Kind     string `xml:"kind,attr"`
	Action   string `xml:"action,attr"`
	PropMods string `xml:"prop-mods,attr"`
}

// List represents XML output of an `svn list` subcommand
type ListElement struct {
	XMLName xml.Name `xml:"lists"`
	Entries []Entry  `xml:"list>entry"`
}

// ListEntry represents XML output of an `svn list` subcommand
type Entry struct {
	Kind   string `xml:"kind,attr"`
	Name   string `xml:"name"`
	Commit Commit `xml:"commit"`
}

// Commit represents XML output of an `svn list` subcommand
type Commit struct {
	Revision string    `xml:"revision,attr"`
	Author   string    `xml:"author"`
	Date     time.Time `xml:"date"`
}

// Repository holds information about a (possibly remote) repository
type Repository struct {
	Location string
}

// NewRepository will initialize the internal structure of a possible remote
// repository, usually pointing to the parent of the default trunk/ tags/ branches
// structure.
func NewRepository(l string) *Repository {
	return &Repository{
		Location: l,
	}
}

// FullPath returns the full path into a repository
func (a *Repository) FullPath(relpath string) string {
	return fmt.Sprintf("%s/%s", a.Location, relpath)
}

// List will execute an `svn list` subcommand.
// Any non-nil xmlWriter will receive the XML content
func (a *Repository) List(relpath string, w io.Writer) ([]Entry, error) {
	//log.Printf("listing %s\n", relpath)
	slog.Info("listing", "path", relpath)
	fp := a.FullPath(relpath)
	cmd := exec.Command("svn", "list", "-R", "--xml", fp)
	//log.Printf("executing %+v\n", cmd)
	slog.Info("executing", "cmd", fmt.Sprintf("%+v", cmd))
	buf, err := cmd.CombinedOutput()
	if w != nil {
		io.Copy(w, bytes.NewReader(buf))
	}
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s", buf)
		return nil, fmt.Errorf("cannot list %s: %s", fp, err)
	}
	var l ListElement
	if err := xml.Unmarshal(buf, &l); err != nil {
		return nil, fmt.Errorf("cannot parse XML: %s: %s", buf, err)
	}
	return l.Entries, nil
}

func (a *Repository) Log(relpath string, w io.Writer) (*LogElement, error) {
	slog.Info("reading log", "path", relpath)
	fp := a.FullPath(relpath)
	cmd := exec.Command("svn", "log", "-v", "-q", "--xml", fp)
	slog.Info("executing", "cmd", fmt.Sprintf("%+v", cmd))

	buf, err := cmd.CombinedOutput()
	if w != nil {
		io.Copy(w, bytes.NewReader(buf))
	}
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s", buf)
		return nil, fmt.Errorf("cannot get log for %s: %s", fp, err)
	}

	//var l LogElement
	l := new(LogElement)
	if err := xml.Unmarshal(buf, &l); err != nil {
		return nil, fmt.Errorf("cannot parse XML: %s: %s", buf, err)
	}

	return l, nil
}

func (a *Repository) LogByRange(relpath string, w io.Writer, firstCommit string, lastCommit string) (*LogElement, error) {
	slog.Info("reading ranged log", "path", relpath)
	fp := a.FullPath(relpath)
	cmd := exec.Command("svn", "log", "-v", "-q", "-r", firstCommit+":"+lastCommit, "--xml", fp)
	slog.Info("executing", "cmd", fmt.Sprintf("%+v", cmd))
	buf, err := cmd.CombinedOutput()
	if w != nil {
		io.Copy(w, bytes.NewReader(buf))
	}
	if err != nil {
		//fmt.Fprintf(os.Stdout, "%s", buf)
		return nil, fmt.Errorf("cannot get log for %s: %s", fp, err)
	}

	//var l LogElement
	l := new(LogElement)
	if err := xml.Unmarshal(buf, &l); err != nil {
		return nil, fmt.Errorf("cannot parse XML: %s: %s", buf, err)
	}

	return l, nil
}

// Export will execute an `svn export` subcommand.
// combined output of stdout and stderr will be written to w
// absolute filenames will be written to notifier channel for each exported file
func (a *Repository) Export(relpath string, into string, w io.Writer, notifier chan string) error {
	slog.Info("exporting", "path", relpath)
	fp := a.FullPath(relpath)
	cmd := exec.Command("svn", "export", fp, into)

	// stdout is written to both w and export notifier
	pr, pw := io.Pipe()
	mw := io.MultiWriter(w, pw)
	cmd.Stdout = mw

	// stderr is written to w
	cmd.Stderr = w

	slog.Info("executing", "cmd", fmt.Sprintf("%+v", cmd))
	if err := cmd.Start(); err != nil {
		return err
	}

	go exportNotifier(pr, notifier)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error exporting %q: %s", fp, err)
	}
	pw.Close()
	return nil
}

// Notify  will report incoming exported filenames to notifier channel.
// channel will be closed once EOF is read
func exportNotifier(r io.Reader, c chan string) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 {
			col1 := strings.TrimSpace(parts[0])
			if col1 != "A" {
				log.Printf("ignoring line because of unknown prefix %q\n", col1)
				continue
			}
			filename := strings.TrimSpace(parts[1])
			c <- filename
		} else {
			log.Printf("ignoring line %q\n", line)
		}
	}
	close(c)
}

// Since returns all entries created after t
func Since(entries []Entry, t time.Time) []Entry {
	var es []Entry
	for _, e := range entries {
		if e.Commit.Date.After(t) {
			es = append(es, e)
		}
	}
	return es
}
