#!/usr/bin/env bash
# Тест merge Caddy snippet (имитация Blitz) без полного Blitz.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

HOST="subpnv.bpo.travel"
CF="${TMP}/Caddyfile"
SNIP="${TMP}/hysteria2-panel.caddy"

cat >"$CF" <<EOF
${HOST} {
    reverse_proxy localhost:8080
}
EOF

cat >"$SNIP" <<'EOF'
handle /subtoken/* {
    reverse_proxy http://127.0.0.1:8787
}
EOF

inserted=0
while IFS= read -r line || [[ -n "$line" ]]; do
    printf '%s\n' "$line"
    if [[ "$inserted" == "0" && "$line" == *"$HOST"* && "$line" == *"{"* ]]; then
        echo "    import ${SNIP}"
        inserted=1
    fi
done <"$CF" >"${CF}.new"
mv "${CF}.new" "$CF"

grep -qF "$SNIP" "$CF"
grep -F "import ${SNIP}" "$CF" >/dev/null

echo "[OK] caddy merge test passed"
