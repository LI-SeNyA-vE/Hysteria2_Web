#!/usr/bin/env bash
# Hysteria2 VPN Panel — установка одной командой:
#   bash <(curl -fsSL https://raw.githubusercontent.com/LI-SeNyA-vE/Hysteria2_Web/main/install.sh)
#
# С доменом сразу:
#   PANEL_GITHUB=LI-SeNyA-vE/Hysteria2_Web PANEL_SUB_DOMAIN=https://sub.example.com \
#   bash <(curl -fsSL https://raw.githubusercontent.com/LI-SeNyA-vE/Hysteria2_Web/main/install.sh)
#
# Переменные:
#   PANEL_GITHUB         LI-SeNyA-vE/Hysteria2_Web   (owner/repo на GitHub)
#   PANEL_BRANCH         main
#   PANEL_INSTALL_DIR    /opt/hysteria2-panel
#   PANEL_RELEASE_BASE   URL каталога с zip (для тестов: http://127.0.0.1:8765)
#   PANEL_SUB_DOMAIN     https://sub.example.com
#   PANEL_SUB_PATH       subtoken
#   PANEL_HTTP_ADDR      127.0.0.1:8787
#   PANEL_FORCE=1        переустановить без вопросов
#   PANEL_SKIP_SYSTEMD=1 пропустить systemd (локальный Docker-тест)

set -euo pipefail

PANEL_GITHUB="${PANEL_GITHUB:-LI-SeNyA-vE/Hysteria2_Web}"
PANEL_BRANCH="${PANEL_BRANCH:-main}"
PANEL_INSTALL_DIR="${PANEL_INSTALL_DIR:-/opt/hysteria2-panel}"
PANEL_SUB_DOMAIN="${PANEL_SUB_DOMAIN:-}"
PANEL_SUB_PATH="${PANEL_SUB_PATH:-subtoken}"
PANEL_HTTP_ADDR="${PANEL_HTTP_ADDR:-127.0.0.1:8787}"
PANEL_SERVICE="${PANEL_SERVICE:-hysteria2-panel}"
PANEL_REPO="${PANEL_REPO:-https://github.com/${PANEL_GITHUB}.git}"
PANEL_RELEASE_BASE="${PANEL_RELEASE_BASE:-https://github.com/${PANEL_GITHUB}/releases/latest/download}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[1;94m'
NC='\033[0m'
BOLD='\033[1m'

CHECK_MARK="[✓]"
CROSS_MARK="[✗]"
INFO_MARK="[i]"
WARNING_MARK="[!]"

log_info()    { echo -e "${BLUE}${INFO_MARK} ${1}${NC}"; }
log_success() { echo -e "${GREEN}${CHECK_MARK} ${1}${NC}"; }
log_warning() { echo -e "${YELLOW}${WARNING_MARK} ${1}${NC}"; }
log_error()   { echo -e "${RED}${CROSS_MARK} ${1}${NC}" >&2; }

handle_error() {
    log_error "Ошибка на строке $1"
    exit 1
}
trap 'handle_error $LINENO' ERR

check_root() {
    if [[ "$(id -u)" -ne 0 ]]; then
        log_error "Запустите от root: sudo bash install.sh"
        exit 1
    fi
    log_info "Запуск от root"
}

check_github_config() {
    if [[ "$PANEL_GITHUB" == USER/* ]]; then
        log_error "Укажите репозиторий: PANEL_GITHUB=your-user/Hysteria2_Web"
        exit 1
    fi
}

check_os_version() {
    local os_name="" os_version=""

    log_info "Проверка ОС..."
    if [[ -f /etc/os-release ]]; then
        # shellcheck disable=SC1091
        source /etc/os-release
        os_name="${ID:-}"
        os_version="${VERSION_ID:-}"
    else
        log_warning "Не удалось определить ОС — продолжаем на свой риск"
        return 0
    fi

    if ! command -v bc >/dev/null 2>&1; then
        if command -v apt-get >/dev/null 2>&1; then
            apt-get update -qq
            DEBIAN_FRONTEND=noninteractive apt-get install -y -qq bc
        fi
    fi

    if command -v bc >/dev/null 2>&1; then
        if [[ "$os_name" == "ubuntu" && $(echo "${os_version} >= 22" | bc) -eq 1 ]] ||
           [[ "$os_name" == "debian" && $(echo "${os_version} >= 12" | bc) -eq 1 ]]; then
            log_success "ОС: ${os_name} ${os_version}"
            return 0
        fi
        log_warning "Рекомендуется Ubuntu 22+ или Debian 12+ (сейчас: ${os_name} ${os_version})"
    else
        log_info "ОС: ${os_name} ${os_version}"
    fi
}

detect_arch() {
    case "$(uname -m)" in
        x86_64)  echo "amd64" ;;
        aarch64) echo "arm64" ;;
        *)
            log_error "Неподдерживаемая архитектура: $(uname -m)"
            exit 1
            ;;
    esac
}

install_packages() {
    local required=(curl ca-certificates unzip)
    local missing=()

    if ! command -v apt-get >/dev/null 2>&1; then
        log_warning "Поддержка apt-only; на других дистрибутивах установите curl, unzip вручную"
        return 0
    fi

    log_info "Проверка пакетов..."
    for pkg in "${required[@]}"; do
        if ! dpkg -l 2>/dev/null | grep -q "^ii  ${pkg} "; then
            missing+=("$pkg")
        else
            log_success "Пакет ${pkg} установлен"
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_info "Установка: ${missing[*]}"
        apt-get update -qq
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "${missing[@]}"
    fi
}

install_build_deps() {
    local missing=()
    for pkg in git gcc; do
        if command -v "${pkg}" >/dev/null 2>&1; then
            log_success "${pkg} установлен"
            continue
        fi
        missing+=("$pkg")
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_info "Для сборки из исходников: ${missing[*]}"
        apt-get update -qq
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "${missing[@]}"
    fi

    command -v gcc >/dev/null 2>&1 || { log_error "gcc не установлен (нужен для SQLite)"; exit 1; }

    local go_ver=""
    if command -v go >/dev/null 2>&1; then
        go_ver="$(go version | awk '{print $3}' | tr -d 'go')"
    fi
    if [[ -z "$go_ver" ]] || [[ "$(printf '%s\n' "1.23" "$go_ver" | sort -V | head -1)" != "1.23" ]]; then
        local arch goarch go_tar
        arch="$(detect_arch)"
        goarch="$arch"
        go_tar="go1.23.2.linux-${goarch}.tar.gz"
        log_info "Установка Go 1.23 (${go_tar})..."
        curl -fsSL "https://go.dev/dl/${go_tar}" | tar -C /usr/local -xz
        export PATH="/usr/local/go/bin:${PATH}"
    fi
    log_success "Go: $(go version)"
}

confirm_reinstall() {
    if [[ "${PANEL_FORCE:-0}" == "1" ]]; then
        return 0
    fi
    if [[ ! -d "${PANEL_INSTALL_DIR}" ]]; then
        return 0
    fi
    if [[ ! -f "${PANEL_INSTALL_DIR}/panel" && ! -f "${PANEL_INSTALL_DIR}/panel.db" ]]; then
        return 0
    fi

    log_warning "Уже установлено в ${PANEL_INSTALL_DIR}"
    if [[ -t 0 ]]; then
        read -r -p "Переустановить/обновить? (y/n): " reply
        if [[ ! "$reply" =~ ^[Yy]$ ]]; then
            log_info "Обновление отменено"
            exit 0
        fi
    else
        log_info "Неинтерактивный режим — обновляем на месте"
    fi
}

install_from_release() {
    local arch="$1"
    local asset="panel-${arch}.zip"
    local url="${PANEL_RELEASE_BASE}/${asset}"
    local tmp="/tmp/${asset}"

    log_info "Скачиваем release: ${url}"
    if ! curl -fsSL -o "$tmp" "$url"; then
        log_warning "Release не найден — соберём из исходников (нужны git, gcc, go)"
        return 1
    fi

    mkdir -p "${PANEL_INSTALL_DIR}"
    unzip -qo "$tmp" -d "${PANEL_INSTALL_DIR}"
    rm -f "$tmp"
    chmod +x "${PANEL_INSTALL_DIR}/panel"
    log_success "Бинарник установлен из GitHub Release (${asset})"
    return 0
}

clone_or_update() {
    log_info "Каталог: ${PANEL_INSTALL_DIR}"
    mkdir -p "${PANEL_INSTALL_DIR}"

    if [[ -d "${PANEL_INSTALL_DIR}/.git" ]]; then
        log_info "Обновление git..."
        git -C "${PANEL_INSTALL_DIR}" fetch origin
        git -C "${PANEL_INSTALL_DIR}" checkout "${PANEL_BRANCH}"
        git -C "${PANEL_INSTALL_DIR}" pull --ff-only origin "${PANEL_BRANCH}"
    else
        local tmp="${PANEL_INSTALL_DIR}.src"
        rm -rf "$tmp"
        git clone --depth 1 --branch "${PANEL_BRANCH}" "${PANEL_REPO}" "$tmp"
        cp -a "$tmp/." "${PANEL_INSTALL_DIR}/"
        rm -rf "$tmp"
        log_success "Исходники получены"
    fi
}

build_panel() {
    log_info "Сборка panel (CGO)..."
    cd "${PANEL_INSTALL_DIR}"
    export CGO_ENABLED=1
    go build -trimpath -ldflags="-s -w" -o panel ./cmd/panel/
    chmod +x panel
    log_success "Сборка завершена"
}

install_panel_binary() {
    local arch
    arch="$(detect_arch)"
    log_info "Архитектура: ${arch}"

    if install_from_release "$arch"; then
        return 0
    fi

    install_build_deps
    clone_or_update
    build_panel
}

write_config() {
    local cfg="${PANEL_INSTALL_DIR}/panel.json"
    if [[ -f "$cfg" ]]; then
        log_warning "Конфиг уже есть — не перезаписываю: ${cfg}"
        return 0
    fi

    log_info "Создание ${cfg}..."
    cat >"$cfg" <<EOF
{
  "db_path": "${PANEL_INSTALL_DIR}/panel.db",
  "log_path": "${PANEL_INSTALL_DIR}/panel.log",
  "http_addr": "${PANEL_HTTP_ADDR}",
  "sync_interval": "30s",
  "sub_domain": "${PANEL_SUB_DOMAIN}",
  "sub_path": "${PANEL_SUB_PATH}"
}
EOF
    log_success "Конфиг создан"
}

install_systemd() {
    log_info "Systemd-служба ${PANEL_SERVICE}..."
    cat >/etc/systemd/system/${PANEL_SERVICE}.service <<EOF
[Unit]
Description=Hysteria2 VPN Panel (HTTP subscriptions + traffic sync)
After=network.target

[Service]
Type=simple
WorkingDirectory=${PANEL_INSTALL_DIR}
ExecStart=${PANEL_INSTALL_DIR}/panel serve -config ${PANEL_INSTALL_DIR}/panel.json
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "${PANEL_SERVICE}"
    systemctl restart "${PANEL_SERVICE}"

    sleep 1
    if systemctl is-active --quiet "${PANEL_SERVICE}"; then
        log_success "Служба ${PANEL_SERVICE} запущена"
    else
        log_error "Служба не стартовала: journalctl -u ${PANEL_SERVICE} -n 30 --no-pager"
        exit 1
    fi
}

add_alias() {
    local line="alias hvpn='cd ${PANEL_INSTALL_DIR} && ./panel -config ${PANEL_INSTALL_DIR}/panel.json'"
    local rc="/root/.bashrc"

    log_info "Alias hvpn → меню панели"
    if [[ -f "$rc" ]] && grep -q "alias hvpn=" "$rc"; then
        log_info "Alias hvpn уже есть"
        return 0
    fi
    echo "$line" >>"$rc"
    log_success "Добавлен alias hvpn в ${rc}"
}

health_url() {
    local addr="${PANEL_HTTP_ADDR}"
    addr="${addr/0.0.0.0/127.0.0.1}"
    [[ "$addr" == *"://"* ]] || addr="http://${addr}"
    echo "${addr}/healthz"
}

print_summary() {
    echo
    echo -e "${BOLD}${GREEN}======== Установка завершена ========${NC}"
    echo
    echo "  Служба:   systemctl status ${PANEL_SERVICE}"
    echo "  Лог:      tail -f ${PANEL_INSTALL_DIR}/panel.log"
    echo "  Health:   curl -fsS $(health_url)"
    echo "  Меню:     hvpn   (или cd ${PANEL_INSTALL_DIR} && ./panel)"
    echo
    if [[ -z "${PANEL_SUB_DOMAIN}" ]]; then
        log_warning "sub_domain пуст — задайте в меню (п. 11), затем: systemctl restart ${PANEL_SERVICE}"
    else
        log_info "Подписка: ${PANEL_SUB_DOMAIN}/${PANEL_SUB_PATH}/{SubToken}"
    fi
    echo
    log_info "Nginx: location /${PANEL_SUB_PATH}/ → proxy_pass http://127.0.0.1:8787;"
    echo
}

run_menu() {
    if [[ ! -t 0 ]] || [[ ! -t 1 ]]; then
        log_info "Не TTY — меню не запускаем автоматически"
        return 0
    fi

    echo -e "${YELLOW}Запуск меню через 3 сек... (Ctrl+C — пропустить)${NC}"
    sleep 3 || return 0

    cd "${PANEL_INSTALL_DIR}"
    echo -e "\n${BOLD}${GREEN}======== Hysteria2 VPN Panel ========${NC}\n"
    ./panel -config "${PANEL_INSTALL_DIR}/panel.json"
}

main() {
    echo -e "\n${BOLD}${BLUE}======== Hysteria2 VPN Panel Setup ========${NC}\n"

    check_root
    check_github_config
    check_os_version
    install_packages
    confirm_reinstall
    install_panel_binary
    write_config

    if [[ "${PANEL_SKIP_SYSTEMD:-0}" == "1" ]]; then
        log_warning "PANEL_SKIP_SYSTEMD=1 — systemd пропущен (тестовый режим)"
    else
        install_systemd
    fi

    add_alias
    print_summary
    run_menu
}

main "$@"
