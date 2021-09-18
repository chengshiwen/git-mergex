/*
Copyright 2021 Shiwen Cheng

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var Version = "unknown"

type command struct {
	cobraCmd *cobra.Command
	dryRun   bool
	abort    bool
}

func Execute() {
	cmd := NewCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func NewCommand() *cobra.Command {
	cmd := &command{}
	cmd.cobraCmd = &cobra.Command{
		Use:           "git-mergex <branch|commit>",
		Short:         "git merge extension for aoneflow",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.runE(args)
		},
	}
	pflags := cmd.cobraCmd.PersistentFlags()
	pflags.BoolVarP(&cmd.dryRun, "dry-run", "d", false, "simulate to merge two development histories together")
	pflags.BoolVarP(&cmd.abort, "abort", "a", false, "abort the current conflict resolution process")
	return cmd.cobraCmd
}

func (cmd *command) runE(args []string) (err error) {
	if (cmd.abort && len(args) > 0) || (!cmd.abort && len(args) == 0) {
		return fmt.Errorf("only one of --abort and <branch|commit> can be specified")
	}
	_, err = exec.LookPath("git")
	if err != nil {
		return
	}
	if cmd.abort {
		abort()
		branch, err := getCurrentBranch()
		if err != nil {
			return err
		}
		reset := exec.Command("git", "reset", "--hard", "origin/"+branch)
		err = reset.Run()
		if err != nil {
			return fmt.Errorf("%s: %s", reset.String(), err)
		}
		return nil
	}
	fetch := exec.Command("git", "fetch", "-f", "origin", args[0])
	out, err := fetch.CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
		return fmt.Errorf("%s: %s", fetch.String(), err)
	}
	if cmd.dryRun {
		merge := exec.Command("git", "merge", "--no-commit", "--no-ff", "origin/"+args[0])
		out, _ = merge.CombinedOutput()
		fmt.Print(strings.ReplaceAll(string(out), "; stopped before committing as requested", ""))
		abort()
	} else {
		branch, err := getCurrentBranch()
		if err != nil {
			return err
		}
		push := exec.Command("git", "push", "-f", "origin", branch)
		err = push.Run()
		if err != nil {
			return fmt.Errorf("%s: %s", push.String(), err)
		}
		reset := exec.Command("git", "reset", "--hard", "origin/"+args[0])
		err = reset.Run()
		if err != nil {
			return fmt.Errorf("%s: %s", reset.String(), err)
		}
		merge := exec.Command("git", "merge", "origin/"+branch)
		out, _ = merge.CombinedOutput()
		fmt.Print(string(out))
	}
	return
}

func getCurrentBranch() (string, error) {
	revParse := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := revParse.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %s", revParse.String(), err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "master" || strings.HasPrefix(branch, "release") {
		return branch, fmt.Errorf("branch %s is forbidden", branch)
	}
	return branch, nil
}

func abort() {
	mergeAbort := exec.Command("git", "merge", "--abort")
	_ = mergeAbort.Run()
}
