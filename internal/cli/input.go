package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"

	"golang.org/x/term"
)

var ErrInputCancelled = errors.New("input cancelled")

var onMainMenu atomic.Bool

func initInputSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for range ch {
			if onMainMenu.Load() {
				clearScreen()
				fmt.Println("\nВыход.")
				os.Exit(0)
			}
		}
	}()
}

func setMainMenu(active bool) {
	onMainMenu.Store(active)
}

func readLine(reader *bufio.Reader, prompt string) (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if prompt != "" {
			fmt.Print(prompt)
		}
		line, err := readLineTTY()
		return strings.TrimSpace(line), err
	}

	if prompt != "" {
		fmt.Print(prompt)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readLineTTY() (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	var buf []byte
	for {
		var b [1]byte
		n, err := os.Stdin.Read(b[:])
		if n == 0 {
			if err != nil {
				return "", err
			}
			continue
		}

		switch b[0] {
		case 3: // Ctrl+C
			fmt.Println("^C")
			if onMainMenu.Load() {
				clearScreen()
				fmt.Println("Выход.")
				os.Exit(0)
			}
			return "", ErrInputCancelled
		case '\r', '\n':
			fmt.Println()
			return string(buf), nil
		case 127, 8: // backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}
		default:
			if b[0] >= 32 && b[0] < 127 {
				buf = append(buf, b[0])
				fmt.Print(string([]byte{b[0]}))
			}
		}
	}
}

func isInputCancelled(err error) bool {
	return errors.Is(err, ErrInputCancelled)
}
