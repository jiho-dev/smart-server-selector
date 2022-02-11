package selector

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type SshConfig struct {
	HostFile  string
	KeyFile   map[string]string // key: env, data: ssh key file
	SearchKey string
}

var exited bool
var app *tview.Application
var view *ServerUI
var cfg SshConfig

// Start the selector's render loop
func Start(sshCfg SshConfig, show_about bool, a *tview.Application) {
	app = a

	cfg = sshCfg
	server := loadServers(cfg.HostFile)
	view = newServersUI(server)

	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	if show_about {
		topFlex.AddItem(buildAboutUI(), SidebarWidth, 0, false)
	}
	topFlex.AddItem(buildSearchUI(cfg.SearchKey), 0, 1, true)

	btmFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	if show_about {
		btmFlex.AddItem(buildTipsUI(), SidebarWidth, 0, false)
	}
	btmFlex.AddItem(view.flex, 0, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 3, 0, true).
		AddItem(btmFlex, 0, 1, false)

	flex.SetBackgroundColor(tcell.ColorBlack)
	app.SetInputCapture(onKeyEvent)
	app.SetRoot(flex, true)
}

func onKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	if view.onEvent(event) {
		return nil
	}
	switch event.Key() {
	case tcell.KeyEscape:
		searchInput.SetText("")
	case tcell.KeyCtrlC, tcell.KeyCtrlD:
		exited = true // exit
	case tcell.KeyCtrlP:
		startVim() // open editor
	case tcell.KeyEnter:
		startSSH()
	}
	if exited {
		app.Stop()
	}

	return event
}

// start vim subprocess
func startVim() {
	app.Suspend(func() {
		execute("vim", SssFile)
		view.setServers(loadServers(cfg.HostFile)) // reload
	})
}

// start ssh subprocess
func startSSH() {
	if view.offset >= len(view.visible) {
		return
	}
	app.Suspend(func() {
		s := view.visible[view.offset]
		k, ok := cfg.KeyFile[s.env]
		if !ok || k == "" {
			fmt.Printf("no SSH Key file for %s \n", s.env)
			return
		}

		var cmds []string
		if len(s.port) > 0 {
			cmds = append(cmds, "-p"+s.port)
		}
		if len(s.user) > 0 {
			cmds = append(cmds, s.user+"@"+s.ip)
		} else {
			cmds = append(cmds, s.ip)
		}

		cmds = append(cmds, "-i"+k)

		execute("ssh", cmds...)

		// XXX: stop selector menu
		app.Stop()
	})

}

// execute the specified command
func execute(name string, args ...string) {
	// print command
	s := name
	if len(args) > 0 {
		for _, a := range args {
			s += " " + a
		}
	}
	fmt.Println(">", s)
	// start command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	// print error
	if err != nil {
		fmt.Println("> error:", err.Error())
		fmt.Println("> press any key to continue")
		getchar()
	}
}

func StartSSHExt(sshCfg SshConfig, key string) error {
	serverList := loadServers(sshCfg.HostFile)
	kw := strings.ToLower(key)
	var server *server

	for _, s := range serverList {
		if strings.ToLower(s.host_name) == kw {
			server = &s
			break
		} else if strings.ToLower(s.ip) == kw {
			server = &s
			break
		}
	}

	if server == nil {
		return fmt.Errorf("Not server to connect: %s", key)
	}

	k, ok := sshCfg.KeyFile[server.env]
	if !ok || k == "" {
		return fmt.Errorf("no SSH Key file: %s %s(%s) \n", server.env, server.host_name, server.ip)
	}

	var cmds []string
	if len(server.port) > 0 {
		cmds = append(cmds, "-p"+server.port)
	}
	if len(server.user) > 0 {
		cmds = append(cmds, server.user+"@"+server.ip)
	} else {
		cmds = append(cmds, server.ip)
	}

	cmds = append(cmds, "-i"+k)
	execute("ssh", cmds...)

	return nil
}
