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
#   PANEL_CERT_EMAIL     email для Let's Encrypt (опционально)
#   PANEL_FORCE=1        переустановить без вопросов
#   PANEL_SKIP_NGINX=1   не трогать nginx (Docker-тест)
#   PANEL_SKIP_SYSTEMD=1 пропустить systemd (локальный Docker-тест)

set -euo pipefail

PANEL_GITHUB="${PANEL_GITHUB:-LI-SeNyA-vE/Hysteria2_Web}"
PANEL_BRANCH="${PANEL_BRANCH:-main}"
PANEL_INSTALL_DIR="${PANEL_INSTALL_DIR:-/opt/hysteria2-panel}"
PANEL_SUB_DOMAIN="${PANEL_SUB_DOMAIN:-}"
PANEL_SUB_PATH="${PANEL_SUB_PATH:-subtoken}"
PANEL_HTTP_ADDR="${PANEL_HTTP_ADDR:-127.0.0.1:8787}"
PANEL_CERT_EMAIL="${PANEL_CERT_EMAIL:-}"
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

PANEL_NGINX_SNIPPET="/etc/nginx/snippets/hysteria2-panel.conf"
PANEL_NGINX_SITE="/etc/nginx/sites-available/hysteria2-panel.conf"

parse_sub_host() {
    local url="${1:-$PANEL_SUB_DOMAIN}"
    url="${url#https://}"
    url="${url#http://}"
    url="${url%%/*}"
    echo "${url%%:*}"
}

escape_regex() {
    echo "$1" | sed 's/[.[\*^$()+?{|]/\\&/g'
}

panel_upstream() {
    local addr="${PANEL_HTTP_ADDR}"
    addr="${addr#http://}"
    addr="${addr#https://}"
    if [[ "$addr" == 0.0.0.0:* ]]; then
        addr="127.0.0.1:${addr#*:}"
    fi
    echo "$addr"
}

install_nginx_packages() {
    if ! command -v apt-get >/dev/null 2>&1; then
        log_warning "apt не найден — nginx не установлен автоматически"
        return 1
    fi
    log_info "Установка nginx + certbot..."
    apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq nginx certbot python3-certbot-nginx
    systemctl enable nginx
    systemctl start nginx
    log_success "nginx установлен"
}

write_nginx_snippet() {
    local upstream
    upstream="$(panel_upstream)"
    log_info "Nginx snippet → ${PANEL_NGINX_SNIPPET}"

    mkdir -p /etc/nginx/snippets
    cat >"${PANEL_NGINX_SNIPPET}" <<EOF
# Hysteria2 VPN Panel — подписки (install.sh)
location /${PANEL_SUB_PATH}/ {
    proxy_pass http://${upstream};
    proxy_http_version 1.1;
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;
}

location = /healthz {
    proxy_pass http://${upstream}/healthz;
}
EOF
    log_success "Snippet записан (${PANEL_SUB_PATH}/ → http://${upstream})"
}

find_nginx_server_conf() {
    local host="$1"
    local host_re f
    host_re="$(escape_regex "$host")"
    for f in /etc/nginx/sites-enabled/* /etc/nginx/conf.d/*; do
        [[ -f "$f" ]] || continue
        [[ "$f" == *hysteria2-panel* ]] && continue
        if grep -qE "server_name[[:space:]].*${host_re}" "$f" 2>/dev/null; then
            echo "$f"
            return 0
        fi
    done
    return 1
}

merge_nginx_snippet() {
    local host="$1"
    local host_re conf
    host_re="$(escape_regex "$host")"
    conf="$(find_nginx_server_conf "$host")" || return 1

    if grep -q "${PANEL_NGINX_SNIPPET}" "$conf"; then
        log_info "Snippet уже подключён в ${conf}"
        return 0
    fi

    cp -a "$conf" "${conf}.bak.hysteria2-panel"
    sed -i "/server_name[[:space:]].*${host_re}/a\\    include ${PANEL_NGINX_SNIPPET};" "$conf"
    log_success "Snippet добавлен в ${conf} (бэкап: ${conf}.bak.hysteria2-panel)"
    return 0
}

write_nginx_standalone_site() {
    local host="$1"
    log_info "Создание ${PANEL_NGINX_SITE} для ${host}..."

    cat >"${PANEL_NGINX_SITE}" <<EOF
# Managed by hysteria2-panel install.sh
server {
    listen 80;
    listen [::]:80;
    server_name ${host};

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    include ${PANEL_NGINX_SNIPPET};
}
EOF

    ln -sf "${PANEL_NGINX_SITE}" /etc/nginx/sites-enabled/hysteria2-panel.conf
    log_success "Сайт ${PANEL_NGINX_SITE} включён"
}

nginx_has_ssl_for_host() {
    local host="$1"
    local conf
    conf="$(find_nginx_server_conf "$host")" || conf=""
    if [[ -n "$conf" ]] && grep -q "ssl_certificate" "$conf" 2>/dev/null; then
        return 0
    fi
    if [[ -f "${PANEL_NGINX_SITE}" ]] && grep -q "ssl_certificate" "${PANEL_NGINX_SITE}" 2>/dev/null; then
        return 0
    fi
    return 1
}

run_certbot() {
    local host="$1"
    local certbot_args=(--nginx -d "$host" --non-interactive --agree-tos --redirect)

    if [[ -n "${PANEL_CERT_EMAIL}" ]]; then
        certbot_args+=(--email "${PANEL_CERT_EMAIL}")
    else
        certbot_args+=(--register-unsafely-without-email)
    fi

    log_info "Let's Encrypt для ${host}..."
    if certbot "${certbot_args[@]}"; then
        log_success "SSL-сертификат получен"
    else
        log_warning "certbot не смог получить сертификат — проверьте DNS и порт 80"
        return 1
    fi
}

reload_nginx() {
    nginx -t
    systemctl reload nginx
    log_success "nginx перезагружен"
}

verify_public_health() {
    local host="$1"
    local code

    if curl -fsS --max-time 10 "https://${host}/healthz" >/dev/null 2>&1; then
        log_success "Публично: https://${host}/healthz"
        return 0
    fi
    code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 10 "http://${host}/healthz" 2>/dev/null || echo 000)"
    if [[ "$code" == "200" ]]; then
        log_warning "Работает по HTTP (без SSL): http://${host}/healthz"
        return 0
    fi
    log_warning "Снаружи /healthz пока недоступен — проверьте DNS и firewall (80/443)"
    return 1
}

install_nginx() {
    if [[ "${PANEL_SKIP_NGINX:-0}" == "1" ]]; then
        log_warning "PANEL_SKIP_NGINX=1 — nginx пропущен"
        return 0
    fi

    if [[ -z "${PANEL_SUB_DOMAIN}" ]]; then
        log_warning "PANEL_SUB_DOMAIN пуст — nginx не настраиваем (задайте домен и переустановите)"
        return 0
    fi

    local host
    host="$(parse_sub_host)"
    if [[ -z "$host" ]]; then
        log_warning "Не удалось разобрать домен из PANEL_SUB_DOMAIN=${PANEL_SUB_DOMAIN}"
        return 0
    fi

    if [[ "${PANEL_SUB_DOMAIN}" == *":"* ]]; then
        log_warning "В sub_domain указан порт — nginx настраивает только 443. Используйте домен без порта или правьте nginx вручную."
    fi

    install_nginx_packages || return 0
    write_nginx_snippet

    if merge_nginx_snippet "$host"; then
        log_info "Подключено к существующему nginx (Blitz/другой сайт)"
    else
        log_info "Сайт для ${host} не найден — создаём отдельный vhost"
        write_nginx_standalone_site "$host"
    fi

    reload_nginx

    if ! nginx_has_ssl_for_host "$host"; then
        run_certbot "$host" || true
        reload_nginx 2>/dev/null || true
    else
        log_info "SSL для ${host} уже есть — certbot пропущен"
    fi

    verify_public_health "$host" || true
}

add_alias() {
    local panel_bin="${PANEL_INSTALL_DIR}/panel"
    local panel_cfg="${PANEL_INSTALL_DIR}/panel.json"
    local hvpn_bin="/usr/local/bin/hvpn"

    log_info "Команда hvpn → меню панели"

    cat >"$hvpn_bin" <<EOF
#!/usr/bin/env bash
cd "${PANEL_INSTALL_DIR}" || exit 1
exec "${panel_bin}" -config "${panel_cfg}" "\$@"
EOF
    chmod +x "$hvpn_bin"
    log_success "Установлено: ${hvpn_bin}"

    local line="alias hvpn='cd ${PANEL_INSTALL_DIR} && ./panel -config ${panel_cfg}'"
    local rc="/root/.bashrc"
    if [[ -f "$rc" ]] && grep -q "alias hvpn=" "$rc"; then
        log_info "Alias hvpn уже есть в ${rc}"
    else
        echo "$line" >>"$rc"
        log_success "Alias hvpn добавлен в ${rc}"
    fi
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
    echo "  Меню:     hvpn"
    echo "            (или: cd ${PANEL_INSTALL_DIR} && ./panel -config panel.json)"
    echo
    if [[ -z "${PANEL_SUB_DOMAIN}" ]]; then
        log_warning "sub_domain пуст — задайте в меню (п. 11), затем: systemctl restart ${PANEL_SERVICE}"
    else
        log_info "Подписка: ${PANEL_SUB_DOMAIN}/${PANEL_SUB_PATH}/{SubToken}"
        if [[ "${PANEL_SKIP_NGINX:-0}" != "1" ]]; then
            log_info "Nginx: https://$(parse_sub_host)/healthz  |  /${PANEL_SUB_PATH}/ → panel"
        fi
    fi
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
    install_nginx
    print_summary
    run_menu
}

main "$@"
