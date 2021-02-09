package manager

import (
	"github.com/ZephyrChien/Mitsuyu/common"
	"io"
)

type Worker interface {
	Run()
	Stop()
	GetLogger() *common.Logger
}

type Manager struct {
	workers map[string]Worker
}

func NewManager() *Manager {
	workers := make(map[string]Worker)
	return &Manager{workers: workers}
}

func (m *Manager) Add(tag string, worker Worker) {
	m.workers[tag] = worker
}

func (m *Manager) Delete(tag string) {
	delete(m.workers, tag)
}

func (m *Manager) Start(tag string) {
	if worker, ok := m.workers[tag]; ok {
		go worker.Run()
	}
}

func (m *Manager) Stop(tag string) {
	if worker, ok := m.workers[tag]; ok {
		worker.Stop()
	}
}

func (m *Manager) StartAll() {
	for _, worker := range m.workers {
		go worker.Run()
	}
}

func (m *Manager) StopAll() {
	for _, worker := range m.workers {
		worker.Stop()
	}
}

func (m *Manager) StartLog(dst io.Writer, tag string) {
	if worker, ok := m.workers[tag]; ok {
		go worker.GetLogger().StartLog(dst)
	}
}

func (m *Manager) StopLog(tag string) {
	if worker, ok := m.workers[tag]; ok {
		worker.GetLogger().StopLog()
	}
}

func (m *Manager) StartLogAll(dst io.Writer) {
	for _, worker := range m.workers {
		go worker.GetLogger().StartLog(dst)
	}
}

func (m *Manager) StopLogAll() {
	for _, worker := range m.workers {
		worker.GetLogger().StopLog()
	}
}
