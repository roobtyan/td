#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
export GOCACHE="$ROOT_DIR/.cache/go-build"
export GOMODCACHE="$ROOT_DIR/.cache/gomod"
export TD_HOME="$ROOT_DIR/.tmp/smoke-home"
rm -rf "$TD_HOME"
mkdir -p "$TD_HOME"

go run ./cmd/td add "a" 2>&1
go run ./cmd/td ls 2>&1 | grep "a"
