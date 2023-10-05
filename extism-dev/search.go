package main

import (
	"bufio"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/extism/cli"
	"github.com/iriri/minimal/gitignore"
)

type Search struct {
	paths           []string
	rx              *regexp.Regexp
	filterFilenames *regexp.Regexp
	args            *devFindArgs
}

func NewSearch(args *devFindArgs, query string, paths ...string) *Search {
	var rx *regexp.Regexp
	if query != "" {
		rx = regexp.MustCompile(query)
	}
	return &Search{
		paths:           paths,
		rx:              rx,
		filterFilenames: nil,
		args:            args,
	}
}

func (search *Search) FilterFilenames(filter string) *Search {
	search.filterFilenames = regexp.MustCompile(filter)
	return search
}

func (search *Search) Iter(f func(string) error) error {
	wg := sync.WaitGroup{}
	for _, path := range search.paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			ignore, err := gitignore.FromGit()
			if err != nil {
				ignore, _ = gitignore.New()
			}
			err = ignore.Walk(path, func(path string, entry fs.FileInfo, err error) error {
				if entry.IsDir() || !entry.Mode().IsRegular() {
					return nil
				}

				abs, err := filepath.Abs(path)
				if err != nil {
					return err
				}

				file, err := os.Open(abs)
				if err != nil {
					return err
				}
				defer file.Close()

				if search.filterFilenames != nil {
					if !search.filterFilenames.Match([]byte(abs)) {
						return nil
					}
				}

				reader := bufio.NewReader(file)
				if search.rx == nil || search.rx.MatchReader(reader) {
					return f(abs)
				}

				return nil
			})
			if err != nil {
				cli.Print("Error in", path+":", err)
			}
		}(path)
	}
	wg.Wait()
	return nil
}

func (search *Search) Replace(r string) error {
	f := func(path string) {
		ignore, err := gitignore.FromGit()
		if err != nil {
			ignore, _ = gitignore.New()
		}
		ignore.Walk(path, func(path string, entry fs.FileInfo, err error) error {
			if entry.IsDir() || !entry.Mode().IsRegular() {
				return nil
			}

			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			data, err := ioutil.ReadFile(abs)
			if err != nil {
				return err
			}

			if search.filterFilenames != nil {
				if !search.filterFilenames.Match([]byte(abs)) {
					return nil
				}
			}

			if search.rx.Match(data) {
				if !search.args.prompt("Update ", path) {
					return nil
				}
				cli.Print("Updating", abs)
				data := search.rx.ReplaceAll(data, []byte(r))
				return ioutil.WriteFile(abs, data, entry.Mode().Perm())
			}

			return nil
		})
	}

	if search.args.interactive {
		for _, path := range search.paths {
			f(path)
		}
		return nil
	}

	wg := sync.WaitGroup{}
	for _, path := range search.paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			f(path)
		}(path)
	}
	wg.Wait()

	return nil
}
