package manager

import (
	"io"
	"mitsuyu/client"
	"mitsuyu/common"
)

type Worker interface {
	Run()
	Stop()
	GetLogger() *common.Logger
}

type Manager struct {
	worker   Worker
	recorder *LogRecorder
	conns    *common.Connector
	stats    *common.Statistician
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Add(worker Worker, terminal bool) {
	m.worker = worker
	if c, ok := worker.(*client.Client); ok && terminal {
		m.conns = c.GetConnector()
		m.stats = c.GetStatistician()
		m.conns.Config(true)
		m.stats.Config(true)
	}
}

func (m *Manager) SetRecorder(r *LogRecorder) {
	m.recorder = r
}

func (m *Manager) GetWorker() Worker {
	return m.worker
}

func (m *Manager) GetClient() (*client.Client, bool) {
	c, ok := m.worker.(*client.Client)
	return c, ok
}

func (m *Manager) GetRecorder() *LogRecorder {
	return m.recorder
}

func (m *Manager) GetConnector() *common.Connector {
	return m.conns
}

func (m *Manager) GetStatistician() *common.Statistician {
	return m.stats
}

func (m *Manager) Start() {
	if m.worker != nil {
		go m.worker.Run()
	}
}

func (m *Manager) Stop() {
	if m.worker != nil {
		m.worker.Stop()
	}
}

func (m *Manager) StartLog(dst io.Writer) {
	if m.worker != nil {
		go m.worker.GetLogger().StartLog(dst)
	}
}

func (m *Manager) StopLog() {
	if m.worker != nil {
		m.worker.GetLogger().StopLog()
	}
}

func (m *Manager) StartConnector() {
	if m.conns != nil {
		go m.conns.StartRecord()
	}
}

func (m *Manager) StopConnector() {
	if m.conns != nil {
		m.conns.StopRecord()
	}
}

func (m *Manager) StartStatistician() {
	if m.stats != nil {
		go m.stats.StartRecord()
	}
}
func (m *Manager) StopStatistician() {
	if m.stats != nil {
		m.stats.StopRecord()
	}
}
