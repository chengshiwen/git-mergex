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

	"github.com/chengshiwen/git-mergex/cmd/completion"
	"github.com/spf13/cobra"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

const (
	remote = "origin"
	mergex = "_mergex"
)

type command struct {
	cobraCmd *cobra.Command
	dryRun   bool
	abort    bool
	cont     bool
	remove   bool
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
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return cmd.comp(args, toComplete)
		},
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.runE(args)
		},
	}
	cmd.cobraCmd.SetVersionTemplate(`{{.Version}}`)
	pflags := cmd.cobraCmd.Flags()
	pflags.BoolVarP(&cmd.dryRun, "dry-run", "d", false, "simulate to merge two development histories together")
	pflags.BoolVarP(&cmd.abort, "abort", "a", false, "abort the current conflict resolution process")
	pflags.BoolVarP(&cmd.cont, "continue", "c", false, "continue to merge after a git merge stops due to conflicts")
	pflags.BoolVarP(&cmd.remove, "remove", "r", false, "remove all temporary mergex branches")
	cmd.cobraCmd.AddCommand(completion.NewCommand())
	return cmd.cobraCmd
}

func (cmd *command) comp(args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cmd.abort || cmd.cont || cmd.remove || len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	branchCmd := exec.Command("git", "branch", "-r")
	out, err := branchCmd.Output()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	branchSet := make(map[string]bool)
	for _, item := range strings.Split(string(out), "\n") {
		branch := strings.TrimSpace(item)
		if len(branch) > 0 {
			if strings.HasPrefix(branch, remoteBranch("HEAD")) {
				continue
			}
			branch = strings.TrimPrefix(branch, remoteBranch(""))
			branchSet[branch] = true
		}
	}
	var choices []string
	for branch := range branchSet {
		choices = append(choices, branch)
	}
	return choices, cobra.ShellCompDirectiveNoFileComp
}

func (cmd *command) runE(args []string) (err error) {
	if len(args)+boolSum(cmd.abort, cmd.cont, cmd.remove) != 1 {
		return fmt.Errorf("only one of <branch|commit>, --abort, --continue and --remove can be specified")
	}
	_, err = exec.LookPath("git")
	if err != nil {
		return
	}

	branch, err := headBranch()
	if err != nil {
		return err
	}

	// --abort
	if cmd.abort {
		abortMerge()
		resetCmd := exec.Command("git", "reset", "--hard", mergexBranch(branch))
		err = resetCmd.Run()
		if err != nil {
			return commandError(resetCmd, err)
		}
		deleteBranch(mergexBranch(branch))
		return nil
	}

	// --continue
	if cmd.cont {
		mergeCmd := exec.Command("git", "merge", "--continue")
		mergeCmd.Stdin = os.Stdin
		mergeCmd.Stdout = os.Stdout
		_ = mergeCmd.Run()
		deleteBranch(mergexBranch(branch))
		return nil
	}

	// --remove
	if cmd.remove {
		branchCmd := exec.Command("git", "branch")
		out, err := branchCmd.Output()
		if err != nil {
			return commandError(branchCmd, err)
		}
		branches := make([]string, 0)
		for _, item := range strings.Split(string(out), "\n") {
			_branch := strings.TrimSpace(item)
			if len(_branch) > 0 {
				if strings.HasPrefix(_branch, mergex) {
					branches = append(branches, _branch)
				}
			}
		}
		if len(branches) > 0 {
			rmCmd := &exec.Cmd{
				Path: "git",
				Args: append([]string{"git", "branch", "-D"}, branches...),
			}
			if lp, err := exec.LookPath("git"); err == nil {
				rmCmd.Path = lp
			}
			rmCmd.Stdin = os.Stdin
			rmCmd.Stdout = os.Stdout
			_ = rmCmd.Run()
		}
		return nil
	}

	// fetch
	fetchCmd := exec.Command("git", "fetch", "-f", remote, args[0])
	out, err := fetchCmd.CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
		if strings.Contains(strings.ToLower(string(out)), "couldn't find remote ref") && strings.HasPrefix(args[0], remote) {
			fmt.Printf("it seems that the branch '%s' should not start with '%s'\n", args[0], remote)
		}
		return commandError(fetchCmd, err)
	}

	// --dry-run
	if cmd.dryRun {
		mergeCmd := exec.Command("git", "merge", "--no-ff", "--no-commit", remoteBranch(args[0]))
		out, _ = mergeCmd.CombinedOutput()
		fmt.Print(strings.ReplaceAll(string(out), "; stopped before committing as requested", ""))
		abortMerge()
		return nil
	}

	// status
	statusCmd := exec.Command("git", "status", "--porcelain", "-uno")
	out, _ = statusCmd.Output()
	outs := strings.TrimSpace(string(out))
	if len(outs) > 0 {
		return fmt.Errorf("Changes not committed before merge:\n%s", outs)
	}

	// merge
	branchCmd := exec.Command("git", "branch", "-f", mergexBranch(branch))
	err = branchCmd.Run()
	if err != nil {
		return commandError(branchCmd, err)
	}
	resetCmd := exec.Command("git", "reset", "--hard", remoteBranch(args[0]))
	err = resetCmd.Run()
	if err != nil {
		return commandError(resetCmd, err)
	}
	mergeCmd := exec.Command("git", "merge", "--no-ff", "-m", fmt.Sprintf("Merge branch '%s' into %s", branch, args[0]), mergexBranch(branch))
	out, err = mergeCmd.CombinedOutput()
	if err == nil {
		deleteBranch(mergexBranch(branch))
	}
	outs = string(out)
	if strings.Contains(outs, "up to date") {
		fmt.Printf("Fast-forward to %s\n", args[0])
	} else {
		fmt.Print(outs)
	}
	return nil
}

func headBranch() (string, error) {
	revParseCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := revParseCmd.Output()
	if err != nil {
		return "", commandError(revParseCmd, err)
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

func mergexBranch(branch string) string {
	return fmt.Sprintf("%s/%s", mergex, branch)
}

func abortMerge() {
	mergeCmd := exec.Command("git", "merge", "--abort")
	_ = mergeCmd.Run()
}

func deleteBranch(branch string) {
	branchCmd := exec.Command("git", "branch", "-D", branch)
	_ = branchCmd.Run()
}

func commandError(c *exec.Cmd, e error) error {
	s := c.String()
	i := strings.Index(s, "git")
	if i > -1 {
		s = s[i:]
	}
	return fmt.Errorf("%s: %s", s, e)
}

func boolSum(items ...bool) int {
	sum := 0
	for _, item := range items {
		if item {
			sum++
		}
	}
	return sum
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
