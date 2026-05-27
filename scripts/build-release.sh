#!/usr/bin/env bash
# Сборка release zip для install.sh (как Blitz-amd64.zip)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/dist"

mkdir -p "$OUT"
cd "$ROOT"

build_one() {
    local goarch="$1"
    local cc="${2:-gcc}"
    local name="panel-${goarch}"

    echo "Building ${name}.zip ..."
    export CGO_ENABLED=1 GOOS=linux GOARCH="$goarch"
    export CC="$cc"

    local tmp
    tmp=$(mktemp -d)
    go build -trimpath -ldflags="-s -w" -o "$tmp/panel" ./cmd/panel/
    (cd "$tmp" && zip -q "${OUT}/${name}.zip" panel)
    rm -rf "$tmp"
    echo "  -> ${OUT}/${name}.zip"
}

build_one amd64 gcc

if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
    build_one arm64 aarch64-linux-gnu-gcc
elif [[ "$(uname -m)" == "aarch64" ]]; then
    build_one arm64 gcc
else
    echo "  skip arm64 (нет aarch64-linux-gnu-gcc; соберётся в GitHub Actions)"
fi

echo
echo "Загрузите в GitHub Release:"
echo "  dist/panel-amd64.zip"
echo "  dist/panel-arm64.zip"
