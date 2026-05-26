package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"hysteria2-web/internal/app"
	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/domain/user"
	applog "hysteria2-web/internal/log"
)

func RunInteractive() {
	db := config.EnvOrDefault("DB_PATH", "./panel.db")
	logPath := config.EnvOrDefault("LOG_PATH", "./panel.log")

	logger, closeLog, err := applog.Open(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка открытия лога: %v\n", err)
		os.Exit(1)
	}
	defer closeLog.Close()

	a, err := app.OpenWithLogger(db, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
	defer a.Close()

	intervalStr := config.EnvOrDefault("SYNC_INTERVAL", "30s")
	syncInterval, err := time.ParseDuration(intervalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка SYNC_INTERVAL: %v\n", err)
		os.Exit(1)
	}
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	a.BlitzSvc.StartTrafficSyncWorker(workerCtx, syncInterval)
	logger.Info("panel started", "log_path", logPath, "sync_interval", syncInterval.String())

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	for {
		clearScreen()
		printMenu(syncInterval, logPath)
		choice := strings.TrimSpace(readLine(reader, "Выберите действие: "))

		if choice == "0" || choice == "q" || choice == "exit" {
			clearScreen()
			fmt.Println("Выход.")
			return
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
			actionErr = interactiveListUsers(a, ctx)
		case "6":
			actionErr = interactiveAddUser(reader, a, ctx)
		case "7":
			actionErr = interactiveKickUser(reader, a, ctx)
		case "8":
			actionErr = interactiveUserURI(reader, a, ctx)
		case "9":
			actionErr = interactiveSync(reader, a, ctx)
		default:
			fmt.Println("Неизвестный пункт меню.")
		}

		if actionErr != nil {
			printFail(actionErr)
		}

		fmt.Println()
		readLine(reader, "Нажмите Enter для продолжения...")
	}
}

func readLine(reader *bufio.Reader, prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func readRequired(reader *bufio.Reader, label string) (string, error) {
	for {
		v := readLine(reader, label+": ")
		if v != "" {
			return v, nil
		}
		fmt.Println("Значение не может быть пустым.")
	}
}

func readInt(reader *bufio.Reader, label string, defaultVal int) (int, error) {
	for {
		raw := readLine(reader, label+": ")
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

type blitzUserRow struct {
	ServerID   uint
	ServerName string
	Username   string
	LimitGB    int
	UsedGB     int
	Active     string
	Online     int
	Blocked    string
}

func printBlitzUsers(rows []blitzUserRow) {
	if len(rows) == 0 {
		fmt.Println("Пользователей нет.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
	choice := readLine(reader, "Выберите: ")

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
	confirm := strings.ToLower(readLine(reader, "Удалить сервер? (y/n): "))
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

func interactiveListUsers(a *app.App, ctx context.Context) error {
	servers, err := a.ServerSvc.ListServers(ctx)
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		fmt.Println("Серверов нет — добавьте сервер (пункт 2).")
		return nil
	}

	localUsers, err := a.UserRepo.ListAll()
	if err != nil {
		return err
	}
	localByKey := make(map[string]*user.User, len(localUsers))
	for i := range localUsers {
		key := fmt.Sprintf("%d:%s", localUsers[i].ServerID, localUsers[i].Username)
		localByKey[key] = &localUsers[i]
	}

	var rows []blitzUserRow
	for _, srv := range servers {
		client, err := a.ServerSvc.GetClient(srv.ID)
		if err != nil {
			return fmt.Errorf("сервер %q: %w", srv.Name, err)
		}
		blitzUsers, err := client.ListUsers(ctx)
		if err != nil {
			return fmt.Errorf("сервер %q: %w", srv.Name, err)
		}
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
				ServerID:   srv.ID,
				ServerName: srv.Name,
				Username:   bu.Username,
				LimitGB:    limitGB,
				UsedGB:     usedGB,
				Active:     active,
				Online:     bu.OnlineCount,
				Blocked:    blocked,
			})
		}
	}

	printBlitzUsers(rows)
	return nil
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
	confirm := strings.ToLower(readLine(reader, "Kick пользователя? (y/n): "))
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
	uri, err := client.ShowUserURI(ctx, username)
	if err != nil {
		return err
	}
	if uri.Error != nil && *uri.Error != "" {
		return fmt.Errorf("blitz: %s", *uri.Error)
	}
	if uri.IPv4 != nil {
		fmt.Printf("IPv4: %s\n", *uri.IPv4)
	}
	if uri.IPv6 != nil {
		fmt.Printf("IPv6: %s\n", *uri.IPv6)
	}
	if uri.NormalSub != nil {
		fmt.Printf("Sub:  %s\n", *uri.NormalSub)
	}
	for _, node := range uri.Nodes {
		fmt.Printf("%s: %s\n", node.Name, node.URI)
	}
	return nil
}

func interactiveSync(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	fmt.Println("  1. Все серверы")
	fmt.Println("  2. Один сервер")
	choice := readLine(reader, "Выберите: ")

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
