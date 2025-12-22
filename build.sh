#!/usr/bin/env bash
set -euo pipefail
go mod download
go build -o nvimwiz ./cmd/nvimwiz
