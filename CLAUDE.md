# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`gh-member` is a GitHub CLI (`gh`) extension for managing GitHub Organization members. It is written in Go and communicates with the GitHub API via `github.com/cli/go-gh/v2`.

## Development commands

The task runner is [mise](https://mise.jdx.dev/).

```bash
mise run build    # build the binary
mise run lint     # run golangci-lint
mise run format   # run golangci-lint fmt
mise run install  # install locally via gh extension install .
mise run remove   # uninstall via gh extension remove member
```

## User-facing commands

For the full description of `gh member` subcommands, flags, output modes, and usage examples, see **[docs/commands.md](docs/commands.md)**.

## Architecture

- Entry point is `main.go` — delegates to `internal/cmd`.
- GitHub API access uses `api.DefaultGraphQLClient()`, which automatically picks up the active `gh` login session.
- Releases are triggered by pushing a `v*` tag; `cli/gh-extension-precompile` handles cross-platform builds and generates attestations.
