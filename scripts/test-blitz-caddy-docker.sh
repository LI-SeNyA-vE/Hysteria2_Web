#!/usr/bin/env bash
# Тест авто-настройки Caddy как на Blitz (hysteria-caddy + Caddyfile в /etc/hysteria/...).
#
#   ./scripts/test-blitz-caddy-docker.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE="${TEST_IMAGE:-ubuntu:22.04}"
HOST="subpnv.bpo.travel"

echo "==> Blitz Caddy auto-config test (${IMAGE})"
echo

docker run --rm \
    -v "${ROOT}/install.sh:/src/install.sh:ro" \
    "${IMAGE}" \
    bash -euxo pipefail -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y -qq bash grep coreutils

        mkdir -p /etc/hysteria/core/scripts/webpanel /etc/caddy/snippets /usr/local/bin

        cat >/etc/hysteria/core/scripts/webpanel/Caddyfile <<'EOF'
${HOST} {
    reverse_proxy localhost:8080
}
EOF

        # Имитация systemctl: hysteria-caddy «активен», restart только логируется
        cat >/usr/local/bin/systemctl <<'MOCK'
#!/bin/bash
case \"\$*\" in
    'is-active --quiet hysteria-caddy') exit 0 ;;
    'restart hysteria-caddy') echo '[mock] hysteria-caddy restarted' >> /tmp/caddy-restarts; exit 0 ;;
    *) exit 1 ;;
esac
MOCK
        chmod +x /usr/local/bin/systemctl

        export PATH=\"/usr/local/bin:\$PATH\"
        export PANEL_TEST_CADDY_ONLY=1
        export PANEL_SUB_DOMAIN=https://${HOST}
        export PANEL_SUB_PATH=subtoken
        export PANEL_HTTP_ADDR=127.0.0.1:8787

        bash /src/install.sh

        grep -qF 'hysteria2-panel.caddy' /etc/hysteria/core/scripts/webpanel/Caddyfile
        test -f /etc/caddy/snippets/hysteria2-panel.caddy
        grep -qF '/subtoken/' /etc/caddy/snippets/hysteria2-panel.caddy
        test -f /tmp/caddy-restarts

        echo '[OK] Blitz Caddy: snippet + import + hysteria-caddy restart'
    "

echo
echo "Готово: install.sh сам добавляет import и перезапускает hysteria-caddy."
