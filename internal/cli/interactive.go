package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"hysteria2-web/internal/app"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/domain/user"
)

var errCancelled = errors.New("cancelled")

func RunInteractive() {
	db := config.EnvOrDefault("DB_PATH", "./panel.db")

	a, err := app.Open(db)
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
	fmt.Printf("Фоновая синхронизация трафика: каждые %s\n\n", syncInterval)

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	for {
		printMenu()
		choice := strings.TrimSpace(readLine(reader, "Выберите действие: "))
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
			actionErr = interactiveListUsers(a)
		case "6":
			actionErr = interactiveAddUser(reader, a, ctx)
		case "7":
			actionErr = interactiveKickUser(reader, a, ctx)
		case "8":
			actionErr = interactiveUserURI(reader, a, ctx)
		case "9":
			actionErr = interactiveSync(reader, a, ctx)
		case "0", "q", "exit":
			fmt.Println("Выход.")
			return
		default:
			fmt.Println("Неизвестный пункт меню.")
		}

		if actionErr != nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", actionErr)
		}

		fmt.Println()
		readLine(reader, "Нажмите Enter для продолжения...")
		fmt.Println()
	}
}

func printMenu() {
	fmt.Println("========================================")
	fmt.Println("         Hysteria2 VPN Panel")
	fmt.Println("========================================")
	fmt.Println("  1. Список серверов")
	fmt.Println("  2. Добавить сервер")
	fmt.Println("  3. Удалить сервер")
	fmt.Println("  4. Статус сервера")
	fmt.Println("  5. Список пользователей")
	fmt.Println("  6. Добавить пользователя")
	fmt.Println("  7. Kick пользователя")
	fmt.Println("  8. URI подключения")
	fmt.Println("  9. Синхронизация трафика")
	fmt.Println("  0. Выход")
	fmt.Println("========================================")
}

func readLine(reader *bufio.Reader, prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func isCancel(s string) bool {
	return s == "0"
}

func printCancelled() {
	fmt.Println("Отменено.")
}

func finishOrCancel(err error) error {
	if errors.Is(err, errCancelled) {
		printCancelled()
		return nil
	}
	return err
}

func readRequired(reader *bufio.Reader, prompt string) (string, error) {
	for {
		v := readLine(reader, prompt+cancelHint)
		if isCancel(v) {
			return "", errCancelled
		}
		if v != "" {
			return v, nil
		}
		fmt.Println("Значение не может быть пустым.")
	}
}

func readInt(reader *bufio.Reader, prompt string, defaultVal int, allowCancel bool) (int, error) {
	hint := ""
	if allowCancel {
		hint = cancelHint
	}
	for {
		raw := readLine(reader, prompt+hint)
		if allowCancel && isCancel(raw) {
			return 0, errCancelled
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

const cancelHint = " (0 — отмена)"

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

func printUsers(users []user.User) {
	if len(users) == 0 {
		fmt.Println("Пользователей нет.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVER_ID\tUSERNAME\tLIMIT_GB\tUSED_GB\tACTIVE\tEXPIRES_DAYS")
	for _, u := range users {
		active := "yes"
		if !u.IsActive {
			active = "no"
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%s\t%d\n",
			u.ServerID, u.Username, u.TrafficLimit, u.TrafficUsed, active, u.ExpirationDays)
	}
	_ = w.Flush()
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
	fmt.Println("  0. Отмена")
	for i, s := range servers {
		fmt.Printf("  %d. [%d] %s\n", i+1, s.ID, s.Name)
	}

	for {
		idx, err := readInt(reader, "Выберите номер сервера: ", -1, true)
		if err != nil {
			return 0, err
		}
		if idx == 0 {
			return 0, errCancelled
		}
		if idx >= 1 && idx <= len(servers) {
			return servers[idx-1].ID, nil
		}
		fmt.Printf("Введите 0 для отмены или число от 1 до %d.\n", len(servers))
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
	name, err := readRequired(reader, "Имя сервера: ")
	if err != nil {
		return finishOrCancel(err)
	}
	url, err := readRequired(reader, "Blitz URL (с path prefix): ")
	if err != nil {
		return finishOrCancel(err)
	}
	key, err := readRequired(reader, "API Key: ")
	if err != nil {
		return finishOrCancel(err)
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
		return finishOrCancel(err)
	}
	confirm := strings.ToLower(readLine(reader, "Удалить сервер? (y/n, 0 — отмена): "))
	if isCancel(confirm) {
		printCancelled()
		return nil
	}
	if confirm != "y" && confirm != "yes" && confirm != "д" && confirm != "да" {
		printCancelled()
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
		return finishOrCancel(err)
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

func interactiveListUsers(a *app.App) error {
	users, err := a.UserRepo.ListAll()
	if err != nil {
		return err
	}
	printUsers(users)
	return nil
}

func interactiveAddUser(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	serverID, err := pickServer(reader, a, ctx)
	if err != nil {
		return finishOrCancel(err)
	}
	username, err := readRequired(reader, "Username: ")
	if err != nil {
		return finishOrCancel(err)
	}
	password, err := readRequired(reader, "Password: ")
	if err != nil {
		return finishOrCancel(err)
	}
	limit, err := readInt(reader, "Лимит трафика (GB): ", 0, true)
	if err != nil {
		return finishOrCancel(err)
	}
	if limit <= 0 {
		return fmt.Errorf("лимит должен быть > 0")
	}
	days, err := readInt(reader, "Срок (дней) [30]: ", 30, true)
	if err != nil {
		return finishOrCancel(err)
	}

	if err := a.BlitzSvc.AddUser(ctx, serverID, username, password, limit, days); err != nil {
		return err
	}
	fmt.Printf("Пользователь %q создан на сервере id=%d.\n", username, serverID)
	return nil
}

func interactiveKickUser(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	serverID, err := pickServer(reader, a, ctx)
	if err != nil {
		return finishOrCancel(err)
	}
	username, err := readRequired(reader, "Username: ")
	if err != nil {
		return finishOrCancel(err)
	}
	confirm := strings.ToLower(readLine(reader, "Kick пользователя? (y/n, 0 — отмена): "))
	if isCancel(confirm) {
		printCancelled()
		return nil
	}
	if confirm != "y" && confirm != "yes" && confirm != "д" && confirm != "да" {
		printCancelled()
		return nil
	}
	if err := a.BlitzSvc.KickUser(ctx, serverID, username); err != nil {
		return err
	}
	fmt.Printf("Пользователь %q отключён.\n", username)
	return nil
}

func interactiveUserURI(reader *bufio.Reader, a *app.App, ctx context.Context) error {
	serverID, err := pickServer(reader, a, ctx)
	if err != nil {
		return finishOrCancel(err)
	}
	username, err := readRequired(reader, "Username: ")
	if err != nil {
		return finishOrCancel(err)
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
	fmt.Println("  0. Отмена")
	choice := readLine(reader, "Выберите: ")

	switch choice {
	case "0":
		printCancelled()
		return nil
	case "1":
		if err := a.BlitzSvc.SyncTraffic(ctx); err != nil {
			return err
		}
		fmt.Println("Синхронизация завершена для всех серверов.")
	case "2":
		id, err := pickServer(reader, a, ctx)
		if err != nil {
			return finishOrCancel(err)
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
