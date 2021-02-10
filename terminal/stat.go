package terminal

import (
	"fmt"
	"time"
	ui "github.com/gizak/termui"
)

func (t *Terminal) renderStat() {
	c := time.NewTicker(1*time.Second).C
	stats := t.manager.GetStatistician()
	stat := t.element.stat
	var lastup, lastdown uint64
	for {
		select {
		case <-c:
			up, down := stats.GetTraffic()
			upspeed := (up - lastup) / 1000
			downspeed := (down - lastdown) / 1000
			str1 := fmt.Sprintf("upload: %8dk", upspeed)
			str2 := fmt.Sprintf("download: %6dk", downspeed)
			str3 := fmt.Sprintf("traffic: %3d/%3dm", up/1000000, down/1000000)
			stat.Rows = []string{str1, str2, str3}
			ui.Render(stat)
		}
	}

}
