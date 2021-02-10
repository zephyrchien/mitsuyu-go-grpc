package terminal

import (
	ui "github.com/gizak/termui"
)

func (t *Terminal) renderInfo(idch chan string) {
	ch := t.manager.GetRecorder().GetChan()
	info := t.element.info
	infoData := t.element.infoData
	for {
		select {
		case msg := <-ch:
			infoData = append(infoData, msg)
			info.Rows = infoData
			ui.Render(info)
		case id := <-idch:
			switch id {
			case "<Up>", "<MouseWheelUp>":
				info.ScrollUp()
			case "<Down>", "<MouseWheelDown>":
				info.ScrollDown()
			}
		}
	}
}
