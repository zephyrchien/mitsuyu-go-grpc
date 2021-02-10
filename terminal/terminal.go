package terminal

import (
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/manager"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

type element struct {
	grid *ui.Grid

	conn     *widgets.List      // left top
	stat     *widgets.List      // left bottom
	info     *widgets.List      // right top
	shell    *widgets.List      // right top
	cmd      *widgets.Paragraph // right bottom
	infoData []string
}

type Terminal struct {
	color   ui.Color
	xratio  float64
	yratio  float64
	manager *manager.Manager
	element *element
}

func NewTerminal(m *manager.Manager, c string, xratio, yratio float64) (*Terminal, error) {
	color, ok := ui.StyleParserColorMap[c]
	if !ok {
		return nil, fmt.Errorf("Terminal: No such color")
	}
	if err := ui.Init(); err != nil {
		return nil, fmt.Errorf("Terminal: %v", err)
	}
	conn := newList(color, true, true, true, true)
	stat := newList(color, true, true, true, true)
	info := newList(color, true, true, true, true)
	shell := newList(color, true, true, true, false)
	conn.Title = "connection"
	stat.Title = "statistic"
	info.Title = "information"
	shell.Title = "command"
	cmd := newParagraph(color, true, true, false, true)
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(
		ui.NewCol(xratio,
			ui.NewRow(yratio, conn),
			ui.NewRow(1.0-yratio, stat),
		),
		ui.NewCol(1-xratio,
			ui.NewRow(yratio, info),
			ui.NewRow(0.9-yratio, shell),
			ui.NewRow(0.1, cmd),
		),
	)
	infoData := make([]string, 0, 500)
	elem := &element{
		grid:     grid,
		conn:     conn,
		stat:     stat,
		info:     info,
		shell:    shell,
		cmd:      cmd,
		infoData: infoData,
	}
	return &Terminal{
		color:   color,
		xratio:  xratio,
		yratio:  yratio,
		manager: m,
		element: elem,
	}, nil
}

func newGrid(color ui.Color) *ui.Grid {
	grid := &ui.Grid{
		Block: ui.Block{
			Border:       true,
			BorderStyle:  ui.NewStyle(color),
			BorderLeft:   true,
			BorderRight:  true,
			BorderTop:    true,
			BorderBottom: true,
			TitleStyle:   ui.NewStyle(color),
		},
	}
	grid.Border = false
	return grid
}

func newList(color ui.Color, l, r, t, b bool) *widgets.List {
	block := &ui.Block{
		Border:       true,
		BorderStyle:  ui.NewStyle(color),
		BorderLeft:   l,
		BorderRight:  r,
		BorderTop:    t,
		BorderBottom: b,
		TitleStyle:   ui.NewStyle(color),
	}
	return &widgets.List{
		Block:            *block,
		TextStyle:        ui.NewStyle(color),
		SelectedRowStyle: ui.Theme.List.Text,
	}
}

func newParagraph(color ui.Color, l, r, t, b bool) *widgets.Paragraph {
	block := &ui.Block{
		Border:       true,
		BorderStyle:  ui.NewStyle(color),
		BorderLeft:   l,
		BorderRight:  r,
		BorderTop:    t,
		BorderBottom: b,
		TitleStyle:   ui.NewStyle(color),
	}
	return &widgets.Paragraph{
		Block:     *block,
		TextStyle: ui.NewStyle(color),
		WrapText:  true,
	}
}

func (t *Terminal) Run() {
	idch := make(chan string, 5)
	event := ui.PollEvents()
	defer ui.Close()
	ui.Render(t.element.grid)
	go t.renderConn()
	go t.renderStat()
	go t.renderInfo(idch)
	for e := range event {
		switch id := e.ID; id {
		case "q", "<C-c>":
			return
		default:
			idch <- id
		}
	}
}
