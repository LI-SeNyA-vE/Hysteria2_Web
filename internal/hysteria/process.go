package hysteria

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type supervisor struct {
	binPath   string
	args      []string // аргументы процесса, например ["server","-c","server.yaml"]
	logWriter io.Writer
	mu        sync.Mutex
	cmd       *exec.Cmd
	isRunning bool
	cancel    context.CancelFunc
}

func newSupervisor(binPath, configPath string, lw io.Writer) *supervisor {
	return &supervisor{binPath: binPath, args: []string{"server", "-c", configPath}, logWriter: lw}
}

func newClientSupervisor(binPath, configPath string, lw io.Writer) *supervisor {
	return &supervisor{binPath: binPath, args: []string{"client", "-c", configPath}, logWriter: newPrefixWriter("[client] ", lw)}
}

// Строки hysteria2 которые являются шумом и не несут полезной информации.
var noiseLines = []string{"client mode", "server mode"}

// filterWriter пропускает строки, содержащие любую из подстрок filters.
type filterWriter struct {
	dst     io.Writer
	filters []string
	buf     []byte
}

func newFilterWriter(dst io.Writer, filters []string) *filterWriter {
	return &filterWriter{dst: dst, filters: filters}
}

func (f *filterWriter) Write(b []byte) (int, error) {
	f.buf = append(f.buf, b...)
	sc := bufio.NewScanner(bytes.NewReader(f.buf))
	consumed := 0
	for sc.Scan() {
		raw := sc.Bytes()
		if !f.shouldFilter(raw) {
			line := append(raw, '\n')
			if _, err := f.dst.Write(line); err != nil {
				return consumed, err
			}
		}
		consumed += len(raw) + 1
	}
	if consumed > 0 && consumed <= len(f.buf) {
		f.buf = f.buf[consumed:]
	}
	return len(b), nil
}

func (f *filterWriter) shouldFilter(line []byte) bool {
	for _, s := range f.filters {
		if bytes.Contains(line, []byte(s)) {
			return true
		}
	}
	return false
}

// prefixWriter добавляет префикс к каждой строке вывода.
type prefixWriter struct {
	prefix []byte
	dst    io.Writer
	buf    []byte
}

func newPrefixWriter(prefix string, dst io.Writer) *prefixWriter {
	return &prefixWriter{prefix: []byte(prefix), dst: dst}
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	sc := bufio.NewScanner(bytes.NewReader(p.buf))
	consumed := 0
	for sc.Scan() {
		raw := sc.Bytes()
		line := p.injectPrefix(raw)
		line = append(line, '\n')
		if _, err := p.dst.Write(line); err != nil {
			return consumed, err
		}
		consumed += len(raw) + 1
	}
	if consumed > 0 && consumed <= len(p.buf) {
		p.buf = p.buf[consumed:]
	}
	return len(b), nil
}

// injectPrefix вставляет [client] после второго таба (после LEVEL) в формате hysteria2:
// "TIMESTAMP\tLEVEL\tMESSAGE" → "TIMESTAMP\tLEVEL\t[client] MESSAGE"
func (p *prefixWriter) injectPrefix(line []byte) []byte {
	tabs := 0
	for i, c := range line {
		if c == '\t' {
			tabs++
			if tabs == 2 {
				result := make([]byte, 0, len(line)+len(p.prefix))
				result = append(result, line[:i+1]...)
				result = append(result, p.prefix...)
				result = append(result, line[i+1:]...)
				return result
			}
		}
	}
	// Формат без табов — ставим в начало
	result := make([]byte, 0, len(line)+len(p.prefix))
	result = append(result, p.prefix...)
	result = append(result, line...)
	return result
}

func (s *supervisor) start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.isRunning = true
	go s.supervise(ctx)
}

func (s *supervisor) supervise(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cmd := exec.Command(s.binPath, s.args...) //nolint:gosec
		out := io.MultiWriter(os.Stdout, newFilterWriter(s.logWriter, noiseLines))
		cmd.Stdout = out
		cmd.Stderr = out

		s.mu.Lock()
		s.cmd = cmd
		s.mu.Unlock()

		log.Printf("hysteria2: запуск (%v)", s.args)
		if err := cmd.Start(); err != nil {
			log.Printf("hysteria2 %v: не удалось запустить: %v, повтор через %v", s.args[0], err, backoff)
			s.mu.Lock()
			s.cmd = nil
			s.mu.Unlock()
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
			backoff = min(backoff*2, 30*time.Second)
			continue
		}

		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			<-done
			s.mu.Lock()
			s.cmd = nil
			s.mu.Unlock()
			log.Printf("hysteria2 %v: остановлен", s.args[0])
			return
		case err := <-done:
			s.mu.Lock()
			s.cmd = nil
			s.mu.Unlock()
			if err != nil {
				log.Printf("hysteria2 %v: упал (%v), перезапуск через %v", s.args[0], err, backoff)
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return
				}
				backoff = min(backoff*2, 30*time.Second)
			} else {
				log.Printf("hysteria2 %v: завершился штатно", s.args[0])
				backoff = time.Second
			}
		}
	}
}

func (s *supervisor) stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	cancel := s.cancel
	cmd := s.cmd
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (s *supervisor) running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}
