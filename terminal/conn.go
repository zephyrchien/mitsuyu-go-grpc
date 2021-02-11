package terminal

import (
	//ui "github.com/gizak/termui"
	"time"
)

func (t *Terminal) renderConn() {
	c := time.NewTicker(1 * time.Second).C
	conns := t.manager.GetConnector()
	conn := t.element.conn
	for {
		select {
		case <-c:
			report := conns.GetReport()
			conn.Rows = report
			//ui.Render(t.grid)
		}
	}
}
