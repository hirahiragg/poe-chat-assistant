package hotkey

import (
	"log"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	mu         sync.Mutex
	hotkeyStr  string
	gameTitle  string
	onPress    func()
	registered bool
	stopCh     chan struct{}
}

func NewManager(gameTitle string) *Manager {
	return &Manager{gameTitle: gameTitle}
}

func (m *Manager) Update(hotkeyStr string, onPress func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hotkeyStr = hotkeyStr
	m.onPress = onPress

	if m.registered {
		Unregister()
		m.registered = false
	}
}

func (m *Manager) Start() {
	m.mu.Lock()
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	go m.poll()
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopCh != nil {
		close(m.stopCh)
		m.stopCh = nil
	}
	if m.registered {
		Unregister()
		m.registered = false
	}
}

func (m *Manager) poll() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.check()
		}
	}
}

func (m *Manager) check() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.hotkeyStr == "" {
		if m.registered {
			Unregister()
			m.registered = false
		}
		return
	}

	name := foregroundAppName()
	active := strings.Contains(strings.ToLower(name), strings.ToLower(m.gameTitle)) || isSelfFocused()

	if active && !m.registered {
		if err := Register(m.hotkeyStr, m.onPress); err != nil {
			log.Printf("hotkey register: %v", err)
		} else {
			m.registered = true
		}
	} else if !active && m.registered {
		Unregister()
		m.registered = false
	}
}
