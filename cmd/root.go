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
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

const remote = "origin"

type command struct {
	cobraCmd *cobra.Command
	dryRun   bool
	abort    bool
	cont     bool
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
		Version:       version(),
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.runE(args)
		},
	}
	cmd.cobraCmd.SetVersionTemplate(`{{.Version}}`)
	pflags := cmd.cobraCmd.PersistentFlags()
	pflags.BoolVarP(&cmd.dryRun, "dry-run", "d", false, "simulate to merge two development histories together")
	pflags.BoolVarP(&cmd.abort, "abort", "a", false, "abort the current conflict resolution process")
	pflags.BoolVarP(&cmd.cont, "continue", "c", false, "continue to merge after a git merge stops due to conflicts")
	return cmd.cobraCmd
}

func (cmd *command) runE(args []string) (err error) {
	if (len(args) > 0 && (cmd.abort || cmd.cont)) || (len(args) == 0 && (cmd.abort == cmd.cont)) {
		return fmt.Errorf("only one of <branch|commit>, --abort and --continue can be specified")
	}
	_, err = exec.LookPath("git")
	if err != nil {
		return
	}

	// --abort
	if cmd.abort {
		abort()
		branch, err := headBranch()
		if err != nil {
			return err
		}
		reset := exec.Command("git", "reset", "--hard", remoteBranch(branch))
		err = reset.Run()
		if err != nil {
			return commandError(reset, err)
		}
		return nil
	}

	// --continue
	if cmd.cont {
		cont := exec.Command("git", "merge", "--continue")
		cont.Stdin = os.Stdin
		cont.Stdout = os.Stdout
		_ = cont.Run()
		return nil
	}

	// fetch
	fetch := exec.Command("git", "fetch", "-f", remote, args[0])
	out, err := fetch.CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
		if strings.Contains(strings.ToLower(string(out)), "couldn't find remote ref") && strings.HasPrefix(args[0], remote) {
			fmt.Printf("it seems that the branch '%s' should not start with '%s'\n", args[0], remote)
		}
		return commandError(fetch, err)
	}

	// --dry-run
	if cmd.dryRun {
		merge := exec.Command("git", "merge", "--no-ff", "--no-commit", remoteBranch(args[0]))
		out, _ = merge.CombinedOutput()
		fmt.Print(strings.ReplaceAll(string(out), "; stopped before committing as requested", ""))
		abort()
		return nil
	}

	// status
	status := exec.Command("git", "status", "--porcelain", "-uno")
	out, _ = status.Output()
	outs := strings.TrimSpace(string(out))
	if len(outs) > 0 {
		return fmt.Errorf("Changes not committed before merge:\n%s", outs)
	}

	// merge
	branch, err := headBranch()
	if err != nil {
		return err
	}
	push := exec.Command("git", "push", "-f", remote, branch)
	err = push.Run()
	if err != nil {
		return commandError(push, err)
	}
	reset := exec.Command("git", "reset", "--hard", remoteBranch(args[0]))
	err = reset.Run()
	if err != nil {
		return commandError(reset, err)
	}
	merge := exec.Command("git", "merge", "--no-ff", "-m", fmt.Sprintf("Merge branch '%s' into %s", branch, args[0]), remoteBranch(branch))
	out, _ = merge.CombinedOutput()
	outs = string(out)
	if strings.Contains(outs, "up to date") {
		fmt.Printf("Fast-forward to %s\n", args[0])
	} else {
		fmt.Print(outs)
	}
	return nil
}

func headBranch() (string, error) {
	revParse := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := revParse.Output()
	if err != nil {
		return "", commandError(revParse, err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "master" || strings.HasPrefix(branch, "release") {
		return branch, fmt.Errorf("branch %s is forbidden", branch)
	}
	return branch, nil
}

func remoteBranch(branch string) string {
	return fmt.Sprintf("%s/%s", remote, branch)
}

func abort() {
	mergeAbort := exec.Command("git", "merge", "--abort")
	_ = mergeAbort.Run()
}

func commandError(c *exec.Cmd, e error) error {
	s := c.String()
	i := strings.Index(s, "git")
	if i > -1 {
		s = s[i:]
	}
	return fmt.Errorf("%s: %s", s, e)
}

func version() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Version:    %s\n", Version))
	sb.WriteString(fmt.Sprintf("Git commit: %s\n", GitCommit))
	sb.WriteString(fmt.Sprintf("Build time: %s\n", BuildTime))
	sb.WriteString(fmt.Sprintf("Go version: %s\n", runtime.Version()))
	sb.WriteString(fmt.Sprintf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH))
	return sb.String()
}
