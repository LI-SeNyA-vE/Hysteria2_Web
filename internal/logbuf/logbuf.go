// Package logbuf реализует потокобезопасный кольцевой буфер строк для логов.
package logbuf

import (
	"bytes"
	"sync"
)

const maxLines = 1000

// Buffer — thread-safe кольцевой буфер последних maxLines строк.
// Реализует io.Writer: принимает произвольные байты, нарезает по '\n'.
type Buffer struct {
	mu    sync.Mutex
	lines []string
}

func New() *Buffer {
	return &Buffer{lines: make([]string, 0, 256)}
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, chunk := range bytes.Split(p, []byte("\n")) {
		line := string(bytes.TrimRight(chunk, "\r"))
		if line == "" {
			continue
		}
		if len(b.lines) >= maxLines {
			b.lines = b.lines[1:]
		}
		b.lines = append(b.lines, line)
	}
	return len(p), nil
}

// Lines возвращает копию всех накопленных строк (старые → новые).
func (b *Buffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}
