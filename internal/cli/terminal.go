package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var initialTermState *term.State

func initTerminal() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return
	}
	state, err := term.GetState(fd)
	if err != nil {
		return
	}
	initialTermState = state
}

func restoreTerminal() {
	if initialTermState == nil {
		return
	}
	_ = term.Restore(int(os.Stdin.Fd()), initialTermState)
}

func isTTYOut() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// termPrintln пишет строку с CR+LF — иначе после raw-ввода текст «съезжает» лесенкой.
func termPrintln(args ...any) {
	if isTTYOut() {
		fmt.Fprint(os.Stdout, fmt.Sprint(args...)+"\r\n")
		return
	}
	fmt.Println(args...)
}

func termPrintf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	if isTTYOut() {
		s = strings.ReplaceAll(s, "\n", "\r\n")
		if !strings.HasSuffix(s, "\r\n") {
			s += "\r\n"
		}
		fmt.Fprint(os.Stdout, s)
		return
	}
	fmt.Print(s)
}

func termPrint(args ...any) {
	if isTTYOut() {
		fmt.Fprint(os.Stdout, fmt.Sprint(args...))
		return
	}
	fmt.Print(args...)
}
