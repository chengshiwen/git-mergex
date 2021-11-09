# git-mergex

git-mergex - git merge extension for aoneflow

## Usage

```bash
$ git-mergex --help
# or
$ git mergex --help

git merge extension for aoneflow

Usage:
  git-mergex <branch|commit> [flags]
  git-mergex [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command

Flags:
  -a, --abort      abort the current conflict resolution process
  -c, --continue   continue to merge after a git merge stops due to conflicts
  -d, --dry-run    simulate to merge two development histories together
  -h, --help       help for git-mergex
  -v, --version    version for git-mergex

Use "git-mergex [command] --help" for more information about a command
```
Specifically, there are three usages for merge:

```bash
  git-mergex [--dry-run] <branch|commit>
  git-mergex --abort
  git-mergex --continue
```

## Shell Completion

git-mergex can generate shell completions for multiple shells. The currently supported shells are:

- Bash
- Zsh
- fish
- PowerShell

For more detailed usage, please execute `git-mergex completion -h`
