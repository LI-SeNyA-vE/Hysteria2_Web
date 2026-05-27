#!/usr/bin/env bash
# Проверка установки hysteria2-panel на сервере.
#   bash scripts/verify-install.sh
set -euo pipefail

PANEL_DIR="${PANEL_INSTALL_DIR:-/opt/hysteria2-panel}"
FAIL=0

ok()  { echo "[OK]  $*"; }
bad() { echo "[!!]  $*"; FAIL=1; }

echo "==> hysteria2-panel verify"
echo

if [[ -x "${PANEL_DIR}/panel" ]]; then ok "binary: ${PANEL_DIR}/panel"; else bad "нет бинарника panel"; fi
if [[ -f "${PANEL_DIR}/panel.json" ]]; then ok "config: panel.json"; else bad "нет panel.json"; fi
if command -v hvpn >/dev/null 2>&1; then ok "команда: hvpn"; else bad "нет hvpn (/usr/local/bin/hvpn)"; fi

if systemctl is-active --quiet hysteria2-panel 2>/dev/null; then
    ok "служба: hysteria2-panel active"
else
    bad "служба hysteria2-panel не запущена"
fi

if curl -fsS --max-time 3 "http://127.0.0.1:8787/healthz" >/dev/null 2>&1; then
    ok "local: http://127.0.0.1:8787/healthz"
else
    bad "panel serve не отвечает на :8787"
fi

if [[ -f "${PANEL_DIR}/panel.json" ]]; then
    sub_domain="$(python3 -c "import json; print(json.load(open('${PANEL_DIR}/panel.json')).get('sub_domain',''))" 2>/dev/null || true)"
    sub_path="$(python3 -c "import json; print(json.load(open('${PANEL_DIR}/panel.json')).get('sub_path','subtoken'))" 2>/dev/null || true)"
    if [[ -n "$sub_domain" ]]; then
        host="${sub_domain#https://}"; host="${host#http://}"; host="${host%%/*}"; host="${host%%:*}"
        if curl -fsS --max-time 10 "https://${host}/healthz" >/dev/null 2>&1; then
            ok "public: https://${host}/healthz"
        else
            bad "https://${host}/healthz недоступен (Caddy/nginx)"
        fi
        cf=""
        if [[ -f /etc/hysteria/core/scripts/webpanel/Caddyfile ]]; then
            cf="/etc/hysteria/core/scripts/webpanel/Caddyfile"
        fi
        if [[ -n "$cf" ]] && grep -qF "hysteria2-panel.caddy" "$cf" 2>/dev/null; then
            ok "caddy import в ${cf}"
        elif [[ -f /etc/nginx/snippets/hysteria2-panel.conf ]]; then
            ok "nginx snippet есть"
        else
            bad "нет import hysteria2-panel в Caddy/nginx"
        fi
    else
        echo "[i]   sub_domain пуст — публичная подписка не проверялась"
    fi
fi

echo
if [[ "$FAIL" == "0" ]]; then
    echo "==> Всё OK"
else
    echo "==> Есть проблемы — переустановите:"
    echo "    PANEL_FORCE=1 PANEL_SUB_DOMAIN=... bash install.sh"
    exit 1
fi
