package hysteria

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type supervisor struct {
	binPath   string
	args      []string // аргументы процесса, например ["server","-c","server.yaml"]
	mu        sync.Mutex
	cmd       *exec.Cmd
	isRunning bool
	cancel    context.CancelFunc
}

func newSupervisor(binPath, configPath string) *supervisor {
	return &supervisor{binPath: binPath, args: []string{"server", "-c", configPath}}
}

func newClientSupervisor(binPath, configPath string) *supervisor {
	return &supervisor{binPath: binPath, args: []string{"client", "-c", configPath}}
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
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

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
