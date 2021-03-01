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
	//conns    *common.Connector
	//stats    *common.Statistician
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Add(worker Worker) {
	m.worker = worker
}

func (m *Manager) SetRecorder(r *LogRecorder) {
	m.recorder = r
}

func (m *Manager) GetRecorder() *LogRecorder {
	return m.recorder
}

func (m *Manager) GetWorker() Worker {
	return m.worker
}

func (m *Manager) GetClient() *client.Client {
	if c, ok := m.worker.(*client.Client); ok {
		return c
	}
	return nil
}

func (m *Manager) GetConnector() *common.Connector {
	if c, ok := m.worker.(*client.Client); ok {
		return c.GetConnector()
	}
	return nil
}

func (m *Manager) GetStatistician() *common.Statistician {
	if c, ok := m.worker.(*client.Client); ok {
		return c.GetStatistician()
	}
	return nil
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
	if conns := m.GetConnector(); conns != nil {
		go conns.StartRecord()
	}
}

func (m *Manager) StopConnector() {
	if conns := m.GetConnector(); conns != nil {
		conns.StopRecord()
	}
}

func (m *Manager) StartStatistician() {
	if stats := m.GetStatistician(); stats != nil {
		go stats.StartRecord()
	}
}

func (m *Manager) StopStatistician() {
	if stats := m.GetStatistician(); stats != nil {
		stats.StopRecord()
	}
}
