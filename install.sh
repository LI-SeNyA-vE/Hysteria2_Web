#!/usr/bin/env bash
# install.sh — установка Hysteria2 Web Panel на Linux VPS.
# Работает без вопросов: все параметры берутся из переменных окружения PANEL_*
# или из встроенных значений по умолчанию.
#
# Переменные окружения (все опциональны для главной панели):
#   PANEL_ROLE           — роль: main_node1 / node1 / node2 / main  (по умолч.: main_node1)
#   PANEL_HTTP_ADDR      — адрес HTTP панели                         (по умолч.: :8080)
#   PANEL_PUBLIC_IP      — публичный IP сервера                      (авто-определяется)
#   PANEL_HY2_PORT       — UDP-порт Hysteria2                        (по умолч.: 443)
#   PANEL_DATA_DIR       — директория данных                         (по умолч.: /var/lib/hysteria2-panel)
#   PANEL_DB_DSN         — PostgreSQL DSN (если пусто — используется встроенный SQLite)
#   PANEL_MAIN_URL       — URL главной панели    (обязателен для node1/node2)
#   PANEL_NODE_TOKEN     — Node Token            (обязателен для node1/node2)
#   PANEL_CASCADE_TARGET — hostname node2 для каскада (только node1, необязательно)
set -euo pipefail

REPO="LI-SeNyA-vE/Hysteria2_Web"
INSTALL_DIR="/opt/hysteria2-panel"
SERVICE_FILE="/etc/systemd/system/hysteria2-panel.service"
CONFIG_FILE="$INSTALL_DIR/panel.yaml"
BINARY="$INSTALL_DIR/panel"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# ── Предварительные проверки ─────────────────────────────────────────────────
[[ $EUID -eq 0 ]] || error "Запускайте от root: sudo bash install.sh"
[[ "$(uname -s)" == "Linux" ]] || error "Поддерживается только Linux"

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  GOARCH="amd64" ;;
  aarch64) GOARCH="arm64" ;;
  *)       error "Неподдерживаемая архитектура: $ARCH" ;;
esac

# ── Параметры из env или дефолтов ────────────────────────────────────────────
ROLE="${PANEL_ROLE:-main_node1}"
HTTP_ADDR="${PANEL_HTTP_ADDR:-:8080}"
HY2_PORT="${PANEL_HY2_PORT:-443}"
DATA_DIR="${PANEL_DATA_DIR:-/var/lib/hysteria2-panel}"
DB_DSN="${PANEL_DB_DSN:-}"
MAIN_URL="${PANEL_MAIN_URL:-}"
NODE_TOKEN="${PANEL_NODE_TOKEN:-}"
CASCADE_TARGET="${PANEL_CASCADE_TARGET:-}"

# Публичный IP — из env или авто-определение
if [[ -n "${PANEL_PUBLIC_IP:-}" ]]; then
    PUBLIC_IP="$PANEL_PUBLIC_IP"
else
    info "Определяем публичный IP..."
    PUBLIC_IP=$(curl -s --max-time 10 ifconfig.me 2>/dev/null || \
                curl -s --max-time 10 api.ipify.org 2>/dev/null || \
                echo "0.0.0.0")
fi

# ── Валидация ────────────────────────────────────────────────────────────────
case "$ROLE" in
  main|main_node1)
    ;;
  node1|node2)
    [[ -n "$MAIN_URL" ]]    || error "PANEL_MAIN_URL обязателен для роли $ROLE"
    [[ -n "$NODE_TOKEN" ]]  || error "PANEL_NODE_TOKEN обязателен для роли $ROLE"
    ;;
  *)
    error "Неизвестная роль: $ROLE. Допустимые: main, main_node1, node1, node2"
    ;;
esac

# ── Информация об установке ───────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════╗"
echo "║     Hysteria2 Web Panel — Установка      ║"
echo "╚══════════════════════════════════════════╝"
echo ""
info "Роль:       $ROLE"
info "IP:         $PUBLIC_IP"
info "HTTP порт:  $HTTP_ADDR"
info "HY2 порт:   $HY2_PORT"
info "Данные:     $DATA_DIR"
[[ -n "$DB_DSN" ]] && info "БД:         PostgreSQL" || info "БД:         SQLite (встроенная)"
echo ""

# ── Остановка сервиса перед обновлением ──────────────────────────────────────
if systemctl is-active --quiet hysteria2-panel 2>/dev/null; then
    info "Останавливаем службу для обновления..."
    systemctl stop hysteria2-panel
fi

# ── Скачивание бинаря ────────────────────────────────────────────────────────
ASSET_NAME="panel-linux-$GOARCH"
VERSION="${PANEL_VERSION:-latest}"

if [[ "$VERSION" == "latest" ]]; then
    DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"
else
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET_NAME"
fi

mkdir -p "$INSTALL_DIR" "$DATA_DIR"
info "Скачиваем $DOWNLOAD_URL ..."
TMP_BINARY="$BINARY.tmp"
curl -fL --location "$DOWNLOAD_URL" -o "$TMP_BINARY"
chmod +x "$TMP_BINARY"
mv -f "$TMP_BINARY" "$BINARY"
info "Бинарь установлен в $BINARY"

# ── Скачивание hysteria2 (только для ролей, запускающих VPN-сервер) ──────────
if [[ "$ROLE" == "main_node1" || "$ROLE" == "node1" || "$ROLE" == "node2" ]]; then
    HY2_VERSION="v2.9.3"
    HY2_TAG="app%2F${HY2_VERSION}"
    HY2_ASSET="hysteria-linux-${GOARCH}"
    HY2_URL="https://github.com/apernet/hysteria/releases/download/${HY2_TAG}/${HY2_ASSET}"
    HY2_BIN_DIR="${DATA_DIR}/bin"
    HY2_BIN="${HY2_BIN_DIR}/hysteria"

    mkdir -p "$HY2_BIN_DIR"
    if [[ ! -f "$HY2_BIN" ]]; then
        info "Скачиваем hysteria2 ${HY2_VERSION} ..."
        curl -fL --location "$HY2_URL" -o "${HY2_BIN}.tmp"
        chmod +x "${HY2_BIN}.tmp"
        mv -f "${HY2_BIN}.tmp" "$HY2_BIN"
        info "hysteria2 установлен в $HY2_BIN"
    else
        info "hysteria2 уже установлен: $HY2_BIN"
    fi
fi

# ── Конфиг ───────────────────────────────────────────────────────────────────
cat > "$CONFIG_FILE" <<YAML
role: $ROLE
httpAddr: "$HTTP_ADDR"
dataDir: "$DATA_DIR"
publicIp: "$PUBLIC_IP"

hy2:
  port: $HY2_PORT
YAML

if [[ -n "$DB_DSN" ]]; then
    cat >> "$CONFIG_FILE" <<YAML

db:
  dsn: "$DB_DSN"
YAML
fi

if [[ "$ROLE" == "node1" || "$ROLE" == "node2" ]]; then
    cat >> "$CONFIG_FILE" <<YAML

main:
  url: "$MAIN_URL"
  token: "$NODE_TOKEN"
YAML
fi

if [[ "$ROLE" == "node1" && -n "$CASCADE_TARGET" ]]; then
    cat >> "$CONFIG_FILE" <<YAML

cascadeTarget: "$CASCADE_TARGET"
YAML
fi

info "Конфиг записан в $CONFIG_FILE"

# ── systemd юнит ─────────────────────────────────────────────────────────────
cat > "$SERVICE_FILE" <<UNIT
[Unit]
Description=Hysteria2 Web Panel
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
ExecStart=$BINARY -config $CONFIG_FILE
WorkingDirectory=$INSTALL_DIR
Restart=always
RestartSec=5
LimitNOFILE=65536

# QUIC / UDP буферы
ExecStartPre=/sbin/sysctl -w net.core.rmem_max=7500000
ExecStartPre=/sbin/sysctl -w net.core.wmem_max=7500000

[Install]
WantedBy=multi-user.target
UNIT

systemctl daemon-reload
systemctl enable hysteria2-panel
systemctl restart hysteria2-panel

sleep 2
if systemctl is-active --quiet hysteria2-panel; then
    info "Служба hysteria2-panel запущена успешно!"
else
    warn "Служба не запустилась, проверьте: journalctl -u hysteria2-panel -n 50"
fi

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║          Установка завершена             ║"
echo "╠══════════════════════════════════════════╣"
printf "║  Панель: http://%-25s║\n" "${PUBLIC_IP}${HTTP_ADDR}"
echo "║  Логи:   journalctl -u hysteria2-panel  ║"
echo "╚══════════════════════════════════════════╝"
echo ""
if [[ "$ROLE" == "main" || "$ROLE" == "main_node1" ]]; then
    echo "  Пароль admin показывается ОДИН РАЗ при первом старте:"
    echo "  journalctl -u hysteria2-panel | grep Пароль"
    echo ""
fi
