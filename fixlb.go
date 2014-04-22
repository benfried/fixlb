package main

import (
	"bufio"
	"fmt"
	"github.com/daviddengcn/go-algs/ed"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type file struct {
	name       string
	matched    bool
	xlatedName string
	season     int
	episode    int
}

type entry struct {
	name    string
	season  int
	episode int
}

var (
	tivo_to_plex   = make(map[string]entry)
	guide_file     = "/Users/bf/Google Drive/lb.html"
	tivo_dir       = "/Volumes/media/Videos/Little Bear"
	athome         = false
	scriptfilename = "/Users/bf/fix-lb.sh"
	season         = 0
	episode        = 0
	renames        = make(map[string]string)
	unmatched      = make([]string, 0, 200)
	files          = make([]file, 0, 200)
)

type appearanceOrder []file

func (f appearanceOrder) Len() int      { return len(f) }
func (f appearanceOrder) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f appearanceOrder) Less(i, j int) bool {
	return ((f[i].season*26)+(f[i].episode) < ((f[j].season * 26) + f[j].episode))
}

func loadEpisodeMappings() {

	f, err := os.Open(guide_file)
	if err != nil {
		fmt.Printf("Open error: %v\n", err)
		return
	}

	scanner := bufio.NewScanner(f)

	seasonRe := regexp.MustCompile("Season ([0-9]+)")
	episodeRe := regexp.MustCompile(`"(.+) / (.+) / (.+)"`)

	for scanner.Scan() {
		line := scanner.Text()
		saysseason := seasonRe.FindStringSubmatch(line)
		if saysseason != nil {
			season, _ = strconv.Atoi(saysseason[1])
			episode = 0
			continue
		}

		cln := regexp.MustCompile("[?!]")
		newep := cln.ReplaceAllString(line, "")

		episodeTitles := episodeRe.FindStringSubmatch(newep)
		if episodeTitles == nil {
			continue
		}
		episode++
		tivo_filename := fmt.Sprintf("Little Bear - ''%s; %s; %s''", episodeTitles[1], episodeTitles[2], episodeTitles[3])
		plex_filename := fmt.Sprintf("Little Bear s%de%d - %s; %s; %s.mp4",
			season, episode, episodeTitles[1], episodeTitles[2], episodeTitles[3])

		tivo_to_plex[tivo_filename] = entry{plex_filename, season, episode}
	}
	f.Close()
}

func loadFiles() {
	f, err := os.Open("/Users/bf/lb-ls.txt")
	if err != nil {
		fmt.Printf("error opening listing lb-ls.txt\n")
		os.Exit(0)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		files = append(files, file{scanner.Text(), false, "", 0, 0})
	}
	f.Close()
}

func main() {

	loadEpisodeMappings()
	loadFiles()

	// find mappings - move this loop to a separate function?
	for i := range files {
		f := &files[i]

		for t, v := range tivo_to_plex {
			if strings.Contains(strings.ToLower(f.name), strings.ToLower(t)) {
				f.xlatedName = v.name
				f.matched = true
				f.season = v.season
				f.episode = v.episode
			}
		}

		// can i combine this loop with the one above? do i need a second pass for inexact matches?

		if !f.matched {
			position := strings.Index(f.name, " (R")
			if position > -1 {
				trimmed_filename := f.name[0:position]
				for t, v := range tivo_to_plex {
					if ed.String(trimmed_filename, t) < 15 {
						f.xlatedName = v.name
						f.matched = true
						f.season = v.season
						f.episode = v.episode
					}

				}
			}
		}

	}

	scriptf, err := os.Create(scriptfilename)
	if err != nil {
		fmt.Printf("Could not create scriptfile %v: %v\n", scriptfilename, err)
		return
	}

	sort.Sort(appearanceOrder(files))

	w := bufio.NewWriter(scriptf)
	fmt.Fprintln(w, "#!/bin/sh\n")
	for i := range files {
		if files[i].matched {
			fmt.Fprintf(w, "mv \"%v\" \"%v\"\n", files[i].name, files[i].xlatedName)
		}
		fmt.Printf("files[%v]: %v\n", i, files[i].name)
	}
	w.Flush()
	scriptf.Close()

}
