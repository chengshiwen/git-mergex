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

Flags:
  -a, --abort      abort the current conflict resolution process
  -c, --continue   continue to merge after a git merge stops due to conflicts
  -d, --dry-run    simulate to merge two development histories together
  -h, --help       help for git-mergex
  -v, --version    version for git-mergex
```
Specifically, there are three usages:

```bash
  git-mergex [--dry-run] <branch|commit>
  git-mergex --abort
  git-mergex --continue
```
