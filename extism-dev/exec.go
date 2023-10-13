package main

import (
	"os"
	"os/exec"
	"regexp"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devExecArgs struct {
	devArgs
	category string
	repo     string
	parallel int
}

func runDevExec(cmd *cobra.Command, args *devExecArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	rx := &regexp.Regexp{}
	if args.repo != "" {
		rx = regexp.MustCompile(args.repo)
	}

	pool := NewPool(args.parallel)
	for i, r := range data.Repos {
		RunTask(pool, func(repo repo) {
			if args.category != "" && repo.Category != args.category {
				return
			}

			if args.repo != "" {
				if !rx.MatchString(repo.Url) {
					return
				}
			}
			p := repo.path()
			cli.Log("Executing", args.args[0], "in", p)
			if args.parallel <= 1 {
				if i > 0 {
					cli.Print()
				}
				cli.Print(p)
			}
			cmd := exec.Command(args.args[0], args.args[1:]...)
			cmd.Dir = p
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "EXTISM_DEV_ROOT="+Root)
			cmd.Env = append(cmd.Env, "EXTISM_DEV_RUNTIME="+args.Path("extism", "extism"))
			cmd.Env = append(cmd.Env, "EXTISM_DEV_REPO_URL="+repo.Url)
			cmd.Env = append(cmd.Env, "EXTISM_DEV_REPO_CATEGORY"+repo.Category)
			cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH")+":"+args.Path(".bin"))
			if err := cmd.Run(); err != nil {
				cli.Print("Error: command failed in", p)
			}
		}, r)
	}
	pool.Wait()
	return nil
}
