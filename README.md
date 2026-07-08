# Hysteria2 Panel

Веб-панель для управления VPN Hysteria2 с поддержкой каскадных серверов и подписок.

---

## Требования

| Компонент | Версия |
|---|---|
| Go | 1.23+ |
| Node.js | 20+ |
| PostgreSQL | 15+ |

---

## Разработка

### 1. Запуск фронтенда (без бэкенда)

```bash
cd frontend
npm install
npm run dev
```

Открыть: **http://localhost:5173**

> В режиме разработки авторизация не требует бэкенда — нажми «Sign In» с любым паролем.

---

### 2. Запуск с бэкендом

Сначала запускаем бэкенд (Go-панель):

```bash
# Создать файл конфига панели
cp panel.example.yaml panel.yaml

# Применить миграции и запустить
go run ./cmd/panel --config panel.yaml
```

Затем запускаем фронтенд с проксированием на бэкенд:

```bash
cd frontend
npm run dev
```

Фронтенд автоматически проксирует `/api/*` и `/sub/*` на `http://localhost:8080`.

---

### 3. Сборка фронтенда для продакшена

```bash
cd frontend
npm run build
```

Файлы появятся в `frontend/dist/`. Go-бинарник встраивает их через `embed.FS` и отдаёт сам — отдельный nginx не нужен.

---

## Деплой на сервер

```bash
bash <(curl https://raw.githubusercontent.com/LI-SeNyA-vE/Hysteria2_Web/main/install.sh)
```

Скрипт задаст вопросы:
1. Выбор роли сервера (Главный / Node 1 / Node 2 / Главный + Node 1)
2. Настройки PostgreSQL
3. Порт панели

После установки панель будет доступна по адресу, который покажет скрипт.
**Hysteria2 при установке НЕ запускается** — это делается через веб-панель в разделе Settings.

---

## Структура проекта

```
hysteria2-web/
├── cmd/panel/          — точка входа Go-приложения
├── internal/
│   ├── domain/         — бизнес-сущности (Server, User, Subscription)
│   ├── service/        — бизнес-логика
│   ├── repository/     — работа с БД (PostgreSQL + GORM)
│   └── httpapi/        — HTTP API (chi router)
├── frontend/           — React + TypeScript + Vite + Tailwind
│   └── src/
│       ├── api/        — запросы к бэкенду
│       ├── components/ — UI компоненты и layout
│       ├── pages/      — страницы панели
│       └── types/      — TypeScript типы
└── install.sh          — скрипт установки на сервер
```

---

## Роли серверов

| Роль | Описание |
|---|---|
| **main** | Только панель управления и база данных |
| **node1** | Принимает трафик от клиентов, форвардит через hy2-client на node2 |
| **node2** | Конечный узел, выходит в интернет. 1 системный пользователь |
| **main + node1** | Совмещённый режим — панель + приём трафика |

**Схема каскада:**
```
Клиент → Node1:443 (hy2 server)
              ↓ SOCKS5 127.0.0.1:1080
         Node1 (hy2 client) → Node2:443 (hy2 server) → Internet
```

---

## API

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/login` | Авторизация |
| GET | `/api/stats` | Статистика для дашборда |
| GET/POST | `/api/users` | Пользователи |
| PUT/DELETE | `/api/users/{id}` | Редактирование / удаление |
| GET/POST | `/api/servers` | Серверы |
| GET/POST | `/api/subscriptions` | Подписки |
| GET | `/sub/{token}` | Публичная ссылка подписки (base64) |
| POST | `/api/hysteria/install` | Установить Hysteria2 |
| POST | `/api/hysteria/start` | Запустить |
| POST | `/api/hysteria/stop` | Остановить |
| GET | `/api/hysteria/status` | Статус службы |
| GET/PUT | `/api/hysteria/config` | Конфигурация |
| POST | `/api/hysteria/cert/regenerate` | Перегенерировать сертификат |
