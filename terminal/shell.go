package terminal

import (
	"fmt"
	ui "github.com/gizak/termui"
	"os"
	"strconv"
	"strings"
)

const cmdPrefix = "mitsuyu> "
const cmdSuffix = "|"

func (t *Terminal) renderShell(event chan string) {
	history := make([]string, 0, 50)
	shell := t.element.shell
	cmd := ""
	for e := range event {
		switch e {
		case "<Backspace>":
			if cmd != "" {
				b := []byte(cmd)
				cmd = string(b[:len(b)-1])
			}
			shell.Rows = append(history, cmdPrefix+cmd+cmdSuffix)
		case "<Space>":
			cmd += " "
			shell.Rows = append(history, cmdPrefix+cmd+cmdSuffix)
		case "<Enter>":
			result := t.handle(cmd, &history)
			cmd = ""
			history = append(history, result)
			shell.Rows = append(history, cmdPrefix+cmd+cmdSuffix)
			shell.ScrollBottom()
		case "<Up>", "<MouseWheelUp>":
			shell.ScrollUp()
		case "<Down>", "<MouseWheelDown>":
			shell.ScrollDown()
		default:
			b := []byte(e)
			s := string(b[0])
			ss := string(b[len(b)-1])
			if s != "<" || ss != ">" {
				cmd += e
			}
			shell.Rows = append(history, cmdPrefix+cmd+cmdSuffix)
		}
		ui.Render(t.grid)
	}
}

func (t *Terminal) handle(cmd string, history *[]string) string {
	m := t.manager
	cmds := strings.Split(cmd, " ")
	var ret string
	switch cmds[0] {
	case "reboot":
		m.Stop()
		m.Start()
		ret = "service: reboot successfully"
	case "shutdown":
		m.Stop()
		ret = "service: shutdown successfully"
	case "exit":
		ui.Close()
		os.Exit(0)
	case "clear":
		*history = make([]string, 0, 50)
		ret = "history cleared"
	case "ls":
		if cc, ok := m.GetClient(); !ok {
			ret = cmd + ": command not found"
		} else {
			ss := cc.GetSummary()
			for _, s := range ss {
				*history = append(*history, s)
			}
			ret = "===================="
		}

	case "set":
		cc, ok := m.GetClient()
		if len(cmds) != 3 || !ok {
			ret = cmd + ": command not found"
		} else {
			if i, err := strconv.Atoi(cmds[2]); err == nil {
				switch cmds[1] {
				case "log":
					cc.GetLogger().SetLevel(i)
					ret = fmt.Sprintf("service: set log level to %d", i)
				case "conn":
					m.GetConnector().Config(i > 0)
					ret = fmt.Sprintf("service: set conn level to %d", i)
				case "stat":
					m.GetStatistician().Config(i > 0)
					ret = fmt.Sprintf("service: set stat level to %d", i)
				case "compress":
					cc.SetCompress(i > 0)
					ret = fmt.Sprintf("service: set compress level to %d", i)
				default:
					ret = cmd + ": command not found"
				}
			} else {
				switch cmds[1] {
				case "remote":
					cc.SetRemote(cmds[2])
					ret = fmt.Sprintf("service: set remote as %s", cmds[2])
				case "local":
					cc.SetLocal(cmds[2])
					ret = fmt.Sprintf("service: set local as %s", cmds[2])
				case "sni":
					cc.SetTLSSNI(cmds[2])
					ret = fmt.Sprintf("service: set sni as %s", cmds[2])
				default:
					ret = cmd + ": command not found"
				}
			}
		}
	default:
		ret = cmd + ": command not found"
	}
	return ret
}
