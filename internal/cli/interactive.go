package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"hysteria2-web/internal/app"
	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/domain/user"
	applog "hysteria2-web/internal/log"
)

func RunInteractive(initial config.Config, configPath string) {
	config.SetConfigPath(configPath)
	cfg := initial

	for {
		if !runPanelSession(cfg) {
			return
		}

		loaded, err := config.Load(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка перезагрузки конфигурации: %v\n", err)
			os.Exit(1)
		}
		cfg = loaded
		clearScreen()
		fmt.Println("Панель перезагружена с новыми настройками.")
		fmt.Println("Если меняли http_addr, sync_interval или sub_path — перезапустите службу:")
		fmt.Println("  systemctl restart hysteria2-panel")
		fmt.Println()
	}
}

func runPanelSession(cfg config.Config) (reload bool) {
	config.Set(cfg)

	logger, closeLog, err := applog.Open(cfg.LogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка открытия лога: %v\n", err)
		os.Exit(1)
	}
	defer closeLog.Close()

	a, err := app.OpenWithLogger(cfg.DBPath, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
	defer a.Close()

	serviceOK := serviceRunning(cfg)
	if serviceOK {
		fmt.Println("Служба: работает (HTTP подписок + sync)")
	} else {
		fmt.Fprintf(os.Stderr, "\n⚠ Служба не запущена — подписки и авто-sync не работают.\n")
		fmt.Fprintf(os.Stderr, "  Запуск: panel serve\n")
		fmt.Fprintf(os.Stderr, "  или:    systemctl start hysteria2-panel\n\n")
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()
	initInputSignals()

	for {
		clearScreen()
		printMenu(cfg, serviceOK)
		setMainMenu(true)
		choice, err := readLine(reader, "Выберите действие: ")
		setMainMenu(false)
		if isInputCancelled(err) {
			continue
		}

		if choice == "0" || choice == "q" || choice == "exit" {
			clearScreen()
			fmt.Println("Выход.")
			return false
		}

		clearScreen()
		title := menuTitle(choice)
		if title == "" {
			title = "Неизвестный пункт меню"
		}
		printScreenHeader(title)
		fmt.Println()

		var actionErr error
		switch choice {
		case "1":
			actionErr = interactiveListServers(a, ctx)
		case "2":
			actionErr = interactiveAddServer(reader, a, ctx)
		case "3":
			actionErr = interactiveDeleteServer(reader, a, ctx)
		case "4":
			actionErr = interactiveServerStatus(reader, a, ctx)
		case "5":
			actionErr = interactiveListUsers(reader, a, ctx)
		case "6":
			actionErr = interactiveAddUser(reader, a, ctx)
		case "7":
			actionErr = interactiveKickUser(reader, a, ctx)
		case "8":
			actionErr = interactiveUserURI(reader, a, ctx)
		case "9":
			actionErr = interactiveSync(reader, a, ctx)
		case "10":
			actionErr = interactiveSubscriptionQR(reader, a, ctx)
		case "11":
			actionErr = interactiveSettings(reader, cfg)
		default:
			fmt.Println("Неизвестный пункт меню.")
		}

		if errors.Is(actionErr, ErrReloadPanel) {
			serviceOK = serviceRunning(config.Get())
			return true
		}

		if isInputCancelled(actionErr) {
			fmt.Println("Действие отменено.")
		} else if actionErr != nil {
			printFail(actionErr)
		}

		if !isInputCancelled(actionErr) {
			fmt.Println()
			if _, err := readLine(reader, "Нажмите Enter для продолжения..."); isInputCancelled(err) {
				continue
			}
		}
	}
}

func readRequired(reader *bufio.Reader, label string) (string, error) {
	for {
		v, err := readLine(reader, label+": ")
		if isInputCancelled(err) {
			return "", err
		}
		if v != "" {
			return v, nil
		}
		fmt.Println("Значение не может быть пустым.")
	}
}

func readInt(reader *bufio.Reader, label string, defaultVal int) (int, error) {
	for {
		raw, err := readLine(reader, label+": ")
		if isInputCancelled(err) {
			return 0, err
		}
		if raw == "" && defaultVal >= 0 {
			return defaultVal, nil
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			fmt.Println("Введите число.")
			continue
		}
		return n, nil
	}
}

const bytesPerGB = 1024 * 1024 * 1024

func printServers(servers []server.Server) {
	if len(servers) == 0 {
		fmt.Println("Серверов нет.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tBASE_URL\tACTIVE")
	for _, s := range servers {
		active := "yes"
		if !s.IsActive {
			active = "no"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", s.ID, s.Name, s.BaseURL, active)
	}
	_ = w.Flush()
}

type panelUserRow struct {
	Username string
	Servers  string
	LimitGB  string
	UsedGB   int
	Active   string
	Expires  string
}

type blitzUserRow struct {
	ServerName string
	Username   string
	LimitGB    int
	UsedGB     int
	Active     string
	Online     int
	Blocked    string
}

func printPanelUsers(rows []panelUserRow) {
	if len(rows) == 0 {
		fmt.Println("Пользователей панели нет.")
		return
	}
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 2, ' ', 0)
	fmt.Fprintln(w, "USERNAME\tSERVERS\tLIMIT_GB\tUSED_GB\tACTIVE\tEXPIRES")
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			r.Username, r.Servers, r.LimitGB, r.UsedGB, r.Active, r.Expires)
	}
	_ = w.Flush()
}

func formatUserExpires(u user.User) string {
	if u.ExpiresAt.IsZero() {
		return "—"
	}
	return u.ExpiresAt.Format("2006-01-02")
}

func printBlitzUsers(rows []blitzUserRow) {
	if len(rows) == 0 {
		fmt.Println("Пользователей нет.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 2, ' ', 0)
	fmt.Fprintln(w, "SERVER\tUSERNAME\tLIMIT_GB\tUSED_GB\tACTIVE\tONLINE\tBLOCKED")
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%d\t%s\n",
			r.ServerName, r.Username, r.LimitGB, r.UsedGB, r.Active, r.Online, r.Blocked)
	}
	_ = w.Flush()
}

func blitzUserBytes(u blitz.UserInfo) int64 {
	var total int64
	if u.UploadBytes != nil {
		total += *u.UploadBytes
	}
	if u.DownloadBytes != nil {
		total += *u.DownloadBytes
	}
	return total
}

func pickServerScope(reader *bufio.Reader, a *app.App, ctx context.Context) (all bool, serverID uint, err error) {
	servers, err := a.ServerSvc.ListServers(ctx)
	if err != nil {
		return false, 0, err
	}
	if len(servers) == 0 {
		return false, 0, fmt.Errorf("нет серверов — сначала добавьте сервер (пункт 2)")
	}

	fmt.Println("  1. Все серверы")
	fmt.Println("  2. Один сервер")
	choice, err := readLine(reader, "Выберите: ")
	if isInputCancelled(err) {
		return false, 0, err
	}

	switch choice {
	case "1":
		return true, 0, nil
	case "2":
		id, err := pickServer(reader, a, ctx)
		return false, id, err
	default:
		return false, 0, fmt.Errorf("неверный выбор")
	}
}

func pickServer(reader *bufio.Reader, a *app.App, ctx context.Context) (uint, error) {
	servers, err := a.ServerSvc.ListServers(ctx)
	if err != nil {
		return 0, err
	}
	if len(servers) == 0 {
		return 0, fmt.Errorf("нет серверов — сначала добавьте сервер (пункт 2)")
	}

	fmt.Println("Серверы:")
	for i, s := range servers {
		fmt.Printf("  %d. [%d] %s\n", i+1, s.ID, s.Name)
	}

	for {
		idx, err := readInt(reader, "Выберите номер сервера", -1)
		if err != nil {
			return 0, err
		}
		if idx >= 1 && idx <= len(servers) {
			return servers[idx-1].ID, nil
		}
		fmt.Printf("Введите число от 1 до %d.\n", len(servers))
	}
}

func interactiveListServers(a *app.App, ctx context.Context) error {
	servers, err := a.ServerSvc.ListServers(ctx)
	if err != nil {
		return err
	}
	printServers(servers)
	return nil
}

func interactiveAddServer(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	name, err := readRequired(reader, "Имя сервера")
	if err != nil {
		return err
	}
	url, err := readRequired(reader, "Blitz URL (с path prefix)")
	if err != nil {
		return err
	}
	key, err := readRequired(reader, "API Key")
	if err != nil {
		return err
	}

	srv, err := a.ServerSvc.CreateServer(ctx, name, url, key)
	if err != nil {
		return err
	}
	fmt.Printf("Сервер создан: id=%d name=%s\n", srv.ID, srv.Name)
	return nil
}

func interactiveDeleteServer(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	id, err := pickServer(reader, a, ctx)
	if err != nil {
		return err
	}
	confirm, err := readLine(reader, "Удалить сервер? (y/n): ")
	if isInputCancelled(err) {
		return err
	}
	confirm = strings.ToLower(confirm)
	if confirm != "y" && confirm != "yes" && confirm != "д" && confirm != "да" {
		fmt.Println("Отменено.")
		return nil
	}
	if err := a.ServerSvc.DeleteServer(ctx, id); err != nil {
		return err
	}
	fmt.Printf("Сервер id=%d удалён.\n", id)
	return nil
}

func interactiveServerStatus(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	id, err := pickServer(reader, a, ctx)
	if err != nil {
		return err
	}
	client, err := a.ServerSvc.GetClient(id)
	if err != nil {
		return err
	}
	status, err := client.GetServerStatus(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Uptime:        %s\n", status.Uptime)
	fmt.Printf("Online users:  %d\n", status.OnlineUsers)
	fmt.Printf("CPU:           %s\n", status.CPUUsage)
	fmt.Printf("RAM:           %s / %s\n", status.RAMUsage, status.TotalRAM)
	fmt.Printf("Traffic total: %s\n", status.UserTotalTraffic)
	return nil
}

func interactiveListUsers(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	servers, err := a.ServerSvc.ListServers(ctx)
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		fmt.Println("Серверов нет — добавьте сервер (пункт 2).")
		return nil
	}

	fmt.Println("  1. Пользователи панели (созданные через панель)")
	fmt.Println("  2. Все пользователи на сервере (Blitz)")
	choice, err := readLine(reader, "Выберите: ")
	if isInputCancelled(err) {
		return err
	}

	switch choice {
	case "1":
		return listPanelUsers(a, servers)
	case "2":
		return listBlitzUsersOnServer(reader, a, ctx)
	default:
		return fmt.Errorf("неверный выбор")
	}
}

func listPanelUsers(a *app.App, servers []server.Server) error {
	localUsers, err := a.UserRepo.ListAll()
	if err != nil {
		return err
	}
	if len(localUsers) == 0 {
		fmt.Println("Пользователей панели нет.")
		return nil
	}

	serverNames := make(map[uint]string, len(servers))
	for _, srv := range servers {
		serverNames[srv.ID] = srv.Name
	}

	type agg struct {
		servers    []string
		limitGB    int
		limitSet   bool
		limitMixed bool
		usedGB     int
		allActive  bool
		anyActive  bool
		expires    string
	}

	byUser := make(map[string]*agg)
	order := make([]string, 0)

	for _, u := range localUsers {
		srvName := serverNames[u.ServerID]
		if srvName == "" {
			srvName = fmt.Sprintf("id=%d", u.ServerID)
		}

		a, ok := byUser[u.Username]
		if !ok {
			a = &agg{allActive: true, expires: formatUserExpires(u)}
			byUser[u.Username] = a
			order = append(order, u.Username)
		}
		a.servers = append(a.servers, srvName)
		a.usedGB += u.TrafficUsed
		if u.IsActive {
			a.anyActive = true
		} else {
			a.allActive = false
		}
		if !a.limitSet {
			a.limitGB = u.TrafficLimit
			a.limitSet = true
		} else if a.limitGB != u.TrafficLimit {
			a.limitMixed = true
		}
	}

	sort.Strings(order)
	rows := make([]panelUserRow, 0, len(order))
	for _, username := range order {
		a := byUser[username]
		sort.Strings(a.servers)

		active := "no"
		if a.allActive {
			active = "yes"
		} else if a.anyActive {
			active = "частично"
		}

		limit := strconv.Itoa(a.limitGB)
		if a.limitMixed {
			limit = "разн."
		}

		rows = append(rows, panelUserRow{
			Username: username,
			Servers:  strings.Join(a.servers, ", "),
			LimitGB:  limit,
			UsedGB:   a.usedGB,
			Active:   active,
			Expires:  a.expires,
		})
	}

	printPanelUsers(rows)
	return nil
}

func listBlitzUsersOnServer(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	serverID, err := pickServer(reader, a, ctx)
	if err != nil {
		return err
	}

	srv, err := a.ServerSvc.GetServer(ctx, serverID)
	if err != nil {
		return err
	}

	client, err := a.ServerSvc.GetClient(serverID)
	if err != nil {
		return fmt.Errorf("сервер %q: %w", srv.Name, err)
	}
	blitzUsers, err := client.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("сервер %q: %w", srv.Name, err)
	}

	localByKey, err := localUsersByKey(a)
	if err != nil {
		return err
	}

	rows := blitzUsersToRows(*srv, blitzUsers, localByKey)
	fmt.Printf("Пользователи на сервере %q (id=%d):\n\n", srv.Name, srv.ID)
	printBlitzUsers(rows)
	return nil
}

func localUsersByKey(a *app.App) (map[string]*user.User, error) {
	localUsers, err := a.UserRepo.ListAll()
	if err != nil {
		return nil, err
	}
	localByKey := make(map[string]*user.User, len(localUsers))
	for i := range localUsers {
		key := fmt.Sprintf("%d:%s", localUsers[i].ServerID, localUsers[i].Username)
		localByKey[key] = &localUsers[i]
	}
	return localByKey, nil
}

func blitzUsersToRows(srv server.Server, blitzUsers []blitz.UserInfo, localByKey map[string]*user.User) []blitzUserRow {
	rows := make([]blitzUserRow, 0, len(blitzUsers))
	for _, bu := range blitzUsers {
		key := fmt.Sprintf("%d:%s", srv.ID, bu.Username)
		local := localByKey[key]

		limitGB := int(bu.MaxDownloadBytes / bytesPerGB)
		usedGB := int(blitzUserBytes(bu) / bytesPerGB)
		active := "yes"
		blocked := "no"
		if bu.Blocked {
			blocked = "yes"
			active = "no"
		}
		if local != nil {
			limitGB = local.TrafficLimit
			usedGB = local.TrafficUsed
			if local.IsActive {
				active = "yes"
			} else {
				active = "no"
			}
		}

		rows = append(rows, blitzUserRow{
			ServerName: srv.Name,
			Username:   bu.Username,
			LimitGB:    limitGB,
			UsedGB:     usedGB,
			Active:     active,
			Online:     bu.OnlineCount,
			Blocked:    blocked,
		})
	}
	return rows
}

func interactiveAddUser(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	allServers, serverID, err := pickServerScope(reader, a, ctx)
	if err != nil {
		return err
	}

	username, err := readRequired(reader, "Username (a-z, A-Z, 0-9, _)")
	if err != nil {
		return err
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	limit, err := readInt(reader, "Лимит трафика (GB, 0 — без лимита)", 0)
	if err != nil {
		return err
	}
	if limit < 0 {
		return fmt.Errorf("лимит не может быть отрицательным")
	}
	days, err := readInt(reader, "Срок (дней, 0 — без срока) [30]", 30)
	if err != nil {
		return err
	}
	if days < 0 {
		return fmt.Errorf("срок не может быть отрицательным")
	}

	password, err := generatePassword(33)
	if err != nil {
		return fmt.Errorf("генерация пароля: %w", err)
	}
	fmt.Printf("Сгенерированный пароль: %s\n", password)

	if allServers {
		servers, err := a.ServerSvc.ListServers(ctx)
		if err != nil {
			return err
		}
		return addUserOnServers(ctx, a, servers, username, password, limit, days)
	}

	srv, err := a.ServerSvc.GetServer(ctx, serverID)
	if err != nil {
		return err
	}

	printStep("Создание пользователя %q на сервере %q (лимит %d GB, %d дней)...", username, srv.Name, limit, days)
	printStep("Запрос POST /api/v1/users/ → %s", srv.BaseURL)

	if err := a.BlitzSvc.AddUser(ctx, serverID, username, password, limit, days); err != nil {
		return err
	}
	printOK("Пользователь %q создан на сервере %q (id=%d)", username, srv.Name, serverID)
	printSubscriptionLink(a, username)
	return nil
}

func addUserOnServers(ctx context.Context, a *app.App, servers []server.Server, username, password string, limit, days int) error {
	var failures []string
	successCount := 0

	for _, srv := range servers {
		printStep("Создание пользователя %q на сервере %q (лимит %d GB, %d дней)...", username, srv.Name, limit, days)
		printStep("Запрос POST /api/v1/users/ → %s", srv.BaseURL)

		if err := a.BlitzSvc.AddUser(ctx, srv.ID, username, password, limit, days); err != nil {
			printFail(fmt.Errorf("%s: %w", srv.Name, err))
			failures = append(failures, fmt.Sprintf("%s: %s", srv.Name, humanError(err)))
			continue
		}
		printOK("Пользователь %q создан на сервере %q (id=%d)", username, srv.Name, srv.ID)
		successCount++
	}

	if successCount == 0 {
		fmt.Println("Пользователь не создан ни на одном сервере.")
		return nil
	}
	if len(failures) > 0 {
		fmt.Printf("Создано на %d из %d серверов. Ошибки:\n", successCount, len(servers))
		for _, msg := range failures {
			fmt.Printf("  • %s\n", msg)
		}
	}
	printSubscriptionLink(a, username)
	return nil
}

func interactiveKickUser(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	allServers, serverID, err := pickServerScope(reader, a, ctx)
	if err != nil {
		return err
	}

	username, err := readRequired(reader, "Username (a-z, A-Z, 0-9, _)")
	if err != nil {
		return err
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	confirm, err := readLine(reader, "Kick пользователя? (y/n): ")
	if isInputCancelled(err) {
		return err
	}
	confirm = strings.ToLower(confirm)
	if confirm != "y" && confirm != "yes" && confirm != "д" && confirm != "да" {
		fmt.Println("Отменено.")
		return nil
	}

	if allServers {
		servers, err := a.ServerSvc.ListServers(ctx)
		if err != nil {
			return err
		}
		return kickUserOnServers(ctx, a, servers, username)
	}

	if err := a.BlitzSvc.KickUser(ctx, serverID, username); err != nil {
		return err
	}
	fmt.Printf("Пользователь %q отключён.\n", username)
	return nil
}

func kickUserOnServers(ctx context.Context, a *app.App, servers []server.Server, username string) error {
	var failures []string
	successCount := 0

	for _, srv := range servers {
		printStep("Kick пользователя %q на сервере %q...", username, srv.Name)
		if err := a.BlitzSvc.KickUser(ctx, srv.ID, username); err != nil {
			printFail(fmt.Errorf("%s: %w", srv.Name, err))
			failures = append(failures, fmt.Sprintf("%s: %s", srv.Name, humanError(err)))
			continue
		}
		printOK("Пользователь %q отключён на сервере %q", username, srv.Name)
		successCount++
	}

	if successCount == 0 {
		fmt.Println("Пользователь не удалён ни на одном сервере.")
		return nil
	}
	if len(failures) > 0 {
		fmt.Printf("Удалено на %d из %d серверов. Ошибки:\n", successCount, len(servers))
		for _, msg := range failures {
			fmt.Printf("  • %s\n", msg)
		}
	}
	return nil
}

func interactiveUserURI(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	serverID, err := pickServer(reader, a, ctx)
	if err != nil {
		return err
	}
	username, err := readRequired(reader, "Username")
	if err != nil {
		return err
	}
	client, err := a.ServerSvc.GetClient(serverID)
	if err != nil {
		return err
	}
	srv, err := a.ServerSvc.GetServer(ctx, serverID)
	if err != nil {
		return err
	}
	uri, err := client.ShowUserURI(ctx, username)
	if err != nil {
		return err
	}
	if uri.Error != nil && *uri.Error != "" {
		return fmt.Errorf("blitz: %s", *uri.Error)
	}
	for _, line := range blitz.CollectRelabeledHy2URIs(uri, srv.Name) {
		fmt.Println(line)
	}
	return nil
}

func interactiveSync(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	fmt.Println("  1. Все серверы")
	fmt.Println("  2. Один сервер")
	choice, err := readLine(reader, "Выберите: ")
	if isInputCancelled(err) {
		return err
	}

	switch choice {
	case "1":
		if err := a.BlitzSvc.SyncTraffic(ctx); err != nil {
			return err
		}
		fmt.Println("Синхронизация завершена для всех серверов.")
	case "2":
		id, err := pickServer(reader, a, ctx)
		if err != nil {
			return err
		}
		if err := a.BlitzSvc.SyncTrafficForServer(ctx, id); err != nil {
			return err
		}
		fmt.Printf("Синхронизация завершена для сервера id=%d.\n", id)
	default:
		return fmt.Errorf("неверный выбор")
	}
	return nil
}
