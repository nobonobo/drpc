package naming

import "time"

// DefaultDeadline ...
var DefaultDeadline = 10 * time.Second

// Monitor ...
type Monitor struct {
	done        chan struct{}
	unsubscribe chan string
	dead        chan string
}

// NewMonitor ...
func NewMonitor(update <-chan string) *Monitor {
	m := &Monitor{
		done:        make(chan struct{}),
		unsubscribe: make(chan string),
		dead:        make(chan string),
	}
	go func() {
		defer close(m.done)
		peers := map[string]time.Time{}
		tick := time.NewTicker(time.Second)
		for {
			select {
			case <-m.done:
				return
			case addr := <-update:
				peers[addr] = time.Now()
			case addr := <-m.unsubscribe:
				delete(peers, addr)
			case <-tick.C:
				now := time.Now()
				for addr, tm := range peers {
					if now.Sub(tm) > DefaultDeadline {
						delete(peers, addr)
						m.dead <- addr
					}
				}
			}
		}
	}()
	return m
}

// Remove 監視停止
func (m *Monitor) Remove(target string) {
	m.unsubscribe <- target
}

// Dead DefaultDeadline到達
func (m *Monitor) Dead() <-chan string {
	return m.dead
}

// Close ...
func (m *Monitor) Close() error {
	m.done <- struct{}{}
	<-m.done
	close(m.dead)
	return nil
}
