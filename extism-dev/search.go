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
	repos           []repo
	rx              *regexp.Regexp
	filterFilenames *regexp.Regexp
	args            *devFindArgs
}

func NewSearch(args *devFindArgs, query string, repos ...repo) *Search {
	if args == nil {
		args = &devFindArgs{}
	}
	var rx *regexp.Regexp
	if query != "" {
		rx = regexp.MustCompile(query)
	}
	return &Search{
		repos:           repos,
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
	for _, r := range search.repos {
		path := r.path()
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
		for _, r := range search.repos {
			f(r.path())
		}
		return nil
	}

	wg := sync.WaitGroup{}
	for _, r := range search.repos {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			f(path)
		}(r.path())
	}
	wg.Wait()

	return nil
}
