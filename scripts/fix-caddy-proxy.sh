#!/usr/bin/env bash
# Починка HTTPS подписок, когда 443 занят Caddy (Blitz).
# Запуск на сервере: bash scripts/fix-caddy-proxy.sh
set -euo pipefail

PANEL_SUB_PATH="${PANEL_SUB_PATH:-subtoken}"
PANEL_HTTP_ADDR="${PANEL_HTTP_ADDR:-127.0.0.1:8787}"
PANEL_SUB_DOMAIN="${PANEL_SUB_DOMAIN:-}"

if [[ -f /opt/hysteria2-panel/panel.json ]]; then
    PANEL_SUB_PATH="$(python3 -c "import json; print(json.load(open('/opt/hysteria2-panel/panel.json')).get('sub_path','subtoken'))" 2>/dev/null || echo subtoken)"
    PANEL_HTTP_ADDR="$(python3 -c "import json; print(json.load(open('/opt/hysteria2-panel/panel.json')).get('http_addr','127.0.0.1:8787'))" 2>/dev/null || echo 127.0.0.1:8787)"
    if [[ -z "$PANEL_SUB_DOMAIN" ]]; then
        PANEL_SUB_DOMAIN="$(python3 -c "import json; print(json.load(open('/opt/hysteria2-panel/panel.json')).get('sub_domain',''))" 2>/dev/null || true)"
    fi
fi

if [[ -z "$PANEL_SUB_DOMAIN" ]]; then
    read -r -p "sub_domain (https://subpnv.bpo.travel): " PANEL_SUB_DOMAIN
fi

host="${PANEL_SUB_DOMAIN#https://}"
host="${host#http://}"
host="${host%%/*}"
host="${host%%:*}"

upstream="${PANEL_HTTP_ADDR#http://}"
upstream="${upstream#https://}"
upstream="${upstream/0.0.0.0/127.0.0.1}"

SNIPPET="/etc/caddy/snippets/hysteria2-panel.caddy"

echo "==> host=${host} path=/${PANEL_SUB_PATH}/ upstream=http://${upstream}"

mkdir -p /etc/caddy/snippets
cat >"$SNIPPET" <<EOF
# hysteria2-panel subscriptions
handle /${PANEL_SUB_PATH}/* {
    reverse_proxy http://${upstream}
}
handle /healthz {
    reverse_proxy http://${upstream}
}
EOF
echo "==> snippet: $SNIPPET"

find_caddyfile() {
    local h="$1"
    local f
    while IFS= read -r f; do
        [[ -f "$f" ]] || continue
        grep -qF "$h" "$f" && echo "$f" && return 0
    done < <(find /etc/caddy /etc/hysteria /opt -name 'Caddyfile' -o -name '*.caddy' 2>/dev/null | sort -u)
    grep -rlF "$h" /etc/caddy /etc/hysteria /opt 2>/dev/null | head -1 || true
}

cf="$(find_caddyfile "$host")"
if [[ -z "$cf" || ! -f "$cf" ]]; then
    echo "ERROR: Caddyfile с ${host} не найден."
    echo "Найдите вручную: grep -r ${host} /etc/caddy /opt"
    exit 1
fi

echo "==> Caddyfile: $cf"

if grep -qF "$SNIPPET" "$cf"; then
    echo "==> import уже есть"
else
    cp -a "$cf" "${cf}.bak.hysteria2-panel-$(date +%s)"
    tmp="$(mktemp)"
    inserted=0
    while IFS= read -r line || [[ -n "$line" ]]; do
        printf '%s\n' "$line"
        if [[ "$inserted" == "0" && "$line" == *"$host"* && "$line" == *"{"* ]]; then
            echo "    import ${SNIPPET}"
            inserted=1
        fi
    done <"$cf" >"$tmp"
    if [[ "$inserted" != "1" ]]; then
        echo "ERROR: не нашли блок site { для ${host} в ${cf}"
        echo "Добавьте вручную внутрь блока: import ${SNIPPET};"
        exit 1
    fi
    mv "$tmp" "$cf"
    echo "==> import добавлен"
fi

if command -v caddy >/dev/null 2>&1; then
    caddy validate --config /etc/caddy/Caddyfile 2>/dev/null || caddy validate --config "$cf" 2>/dev/null || true
fi

if systemctl is-active --quiet hysteria-caddy 2>/dev/null; then
    caddy validate --config "$cf" 2>/dev/null || true
    systemctl restart hysteria-caddy
    echo "==> hysteria-caddy restarted (VPN не трогаем)"
elif systemctl is-active --quiet caddy 2>/dev/null; then
    systemctl reload caddy
    echo "==> systemd caddy reload"
else
    echo "ERROR: не найден hysteria-caddy — перезапустите Caddy вручную"
    exit 1
fi

echo "==> проверка"
sleep 1
curl -fsS "http://127.0.0.1:8787/healthz" >/dev/null && echo "[ok] panel local"
curl -fsS --http1.1 "https://${host}/healthz" && echo "[ok] https healthz" || echo "[!!] https healthz FAIL — см. journalctl -u caddy -n 30"
