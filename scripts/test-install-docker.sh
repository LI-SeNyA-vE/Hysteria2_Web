#!/usr/bin/env bash
# Локальный тест install.sh + zip в Ubuntu 22.04 (Docker), без systemd и без прод-сервера.
#
#   ./scripts/test-install-docker.sh
#
# Что проверяется:
#   1) сборка panel-{arch}.zip (как в CI)
#   2) раздача zip через локальный HTTP
#   3) install.sh скачивает zip и распаковывает
#   4) panel serve отвечает на /healthz

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE="${TEST_IMAGE:-ubuntu:22.04}"

echo "==> Docker-тест установки (образ: ${IMAGE})"
echo "    Проект: ${ROOT}"
echo

docker run --rm \
    -v "${ROOT}:/src:ro" \
    "${IMAGE}" \
    bash -euxo pipefail -c '
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y -qq curl ca-certificates unzip bc zip gcc python3

        case "$(uname -m)" in
            x86_64)  GOARCH=amd64; PANEL_ARCH=amd64 ;;
            aarch64) GOARCH=arm64; PANEL_ARCH=arm64 ;;
            *) echo "unsupported arch: $(uname -m)"; exit 1 ;;
        esac
        ZIP="panel-${PANEL_ARCH}.zip"

        echo "==> Установка Go 1.23 (в Ubuntu 22.04 apt даёт Go 1.18 — слишком старый)"
        GO_TAR="go1.23.2.linux-${GOARCH}.tar.gz"
        curl -fsSL "https://go.dev/dl/${GO_TAR}" | tar -C /usr/local -xz
        export PATH="/usr/local/go/bin:${PATH}"
        go version

        echo "==> 1/4 Сборка ${ZIP} (CGO)"
        export CGO_ENABLED=1
        cd /src
        go build -trimpath -ldflags="-s -w" -o /tmp/panel ./cmd/panel/
        (cd /tmp && zip -q "${ZIP}" panel)

        echo "==> 2/4 Локальный HTTP-сервер с zip (имитация GitHub Release)"
        mkdir -p /tmp/release
        cp "/tmp/${ZIP}" /tmp/release/
        python3 -m http.server 8765 --directory /tmp/release &
        HTTP_PID=$!
        sleep 1
        curl -fsSI "http://127.0.0.1:8765/${ZIP}" | head -1

        echo "==> 3/4 install.sh (zip, без systemd)"
        export PANEL_GITHUB=LI-SeNyA-vE/Hysteria2_Web
        export PANEL_RELEASE_BASE=http://127.0.0.1:8765
        export PANEL_INSTALL_DIR=/opt/hysteria2-panel-test
        export PANEL_HTTP_ADDR=127.0.0.1:8787
        export PANEL_SUB_DOMAIN=https://test.example.com
        export PANEL_SUB_PATH=subtoken
        export PANEL_FORCE=1
        export PANEL_SKIP_SYSTEMD=1
        export PANEL_SKIP_NGINX=1
        rm -rf "${PANEL_INSTALL_DIR}"
        bash /src/install.sh

        test -x "${PANEL_INSTALL_DIR}/panel"
        test -f "${PANEL_INSTALL_DIR}/panel.json"
        "${PANEL_INSTALL_DIR}/panel" help >/dev/null

        echo "==> 4/4 panel serve + healthz"
        cd "${PANEL_INSTALL_DIR}"
        ./panel serve -config panel.json &
        SERVE_PID=$!
        sleep 2
        curl -fsS "http://127.0.0.1:8787/healthz"
        kill "${SERVE_PID}" 2>/dev/null || true
        kill "${HTTP_PID}" 2>/dev/null || true

        echo
        echo "[OK] Тест пройден: ${ZIP} → install.sh → panel serve → healthz"
    '

echo
echo "Готово. Можно деплоить на сервер после публикации zip в GitHub Release."
