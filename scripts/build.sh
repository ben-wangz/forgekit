#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT_DIR="${REPO_ROOT}/dist/bin"

export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

mkdir -p "${OUT_DIR}"
go build -C "${REPO_ROOT}" -o "${OUT_DIR}/forgekit" ./cmd/forgekit

echo "Built ${OUT_DIR}/forgekit"
