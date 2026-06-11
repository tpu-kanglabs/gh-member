# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`gh-member` is a GitHub CLI (`gh`) extension for managing GitHub Organization members. It is written in Go and communicates with the GitHub API via `github.com/cli/go-gh/v2`.

## Commands

The task runner is [mise](https://mise.jdx.dev/).

```bash
mise run build    # build the binary
mise run lint     # run golangci-lint
mise run format   # run golangci-lint fmt
mise run install  # install locally via gh extension install .
mise run remove   # uninstall via gh extension remove member
```

## Architecture

- Entry point is `main.go` only — currently a single-file skeleton implementation.
- GitHub API access uses `api.DefaultRESTClient()`, which automatically picks up the active `gh` login session.
- Releases are triggered by pushing a `v*` tag; `cli/gh-extension-precompile` handles cross-platform builds and generates attestations.
