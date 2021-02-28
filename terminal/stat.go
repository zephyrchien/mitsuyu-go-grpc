package terminal

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"time"
)

func (t *Terminal) renderStat() {
	c := time.NewTicker(1 * time.Second).C
	stats := t.manager.GetStatistician()
	stat := t.element.stat
	maxlen := int(t.xratio*t.xmax) - 2
	str1fmt := fmt.Sprintf("up %%%ddk", maxlen-4)
	str2fmt := fmt.Sprintf("down %%%ddk", maxlen-6)
	str3fmt := fmt.Sprintf("upload %%%ddm", maxlen-8)
	str4fmt := fmt.Sprintf("download %%%ddm", maxlen-10)
	var lastup, lastdown uint64
	for {
		select {
		case <-c:
			up, down := stats.GetTraffic()
			upspeed := (up - lastup) / 1000
			downspeed := (down - lastdown) / 1000
			str1 := fmt.Sprintf(str1fmt, upspeed)
			str2 := fmt.Sprintf(str2fmt, downspeed)
			str3 := fmt.Sprintf(str3fmt, up/1000000)
			str4 := fmt.Sprintf(str4fmt, down/1000000)
			lastup, lastdown = up, down
			stat.Rows = []string{str1, str2, "", str3, str4}
			ui.Render(t.grid)
		}
	}

}
