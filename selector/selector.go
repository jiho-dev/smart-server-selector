package selector

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/sisyphsu/smart-server-selector/iterm2"
	"github.com/sisyphsu/smart-server-selector/ssh"
)

var exited bool
var app *tview.Application
var view *ServerUI

// Start the selector's render loop
func Start(sssCfg *SssConfig, searchKey string, server []server, a *tview.Application) {
	app = a

	SssCfg = sssCfg
	//server := LoadServers(sssCfg)
	view = newServersUI(server)

	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	if sssCfg.ShowAbout {
		topFlex.AddItem(buildAboutUI(), SidebarWidth, 0, false)
	}
	topFlex.AddItem(buildSearchUI(searchKey), 0, 1, true)

	btmFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	if sssCfg.ShowAbout {
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
		execute("vim", SssCfg.HostFile)
		view.setServers(LoadServers(SssCfg)) // reload
	})
}

// start ssh subprocess
func startSSH() {
	if view.offset >= len(view.visible) {
		return
	}
	app.Suspend(func() {
		s := view.visible[view.offset]
		k, ok := SssCfg.KeyFile[s.env]
		if !ok || k == "" {
			fmt.Printf("no SSH Key file for %s \n", s.env)
			return
		}

		ExecSSH(SssCfg, &s)
		// XXX: stop selector menu
		//app.Stop()
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

/*
func StartSSHExt(sssCfg *SssConfig, key string) error {
	serverList := LoadServers(sssCfg)
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

	return ExecSSH(sssCfg, server)
}
*/

/*
func ExecSSH(sssCfg *SssConfig, server *server) error {
	if server == nil {
		return fmt.Errorf("server is empty")
	}

	k, ok := sssCfg.KeyFile[server.env]
	if !ok || k == "" {
		return fmt.Errorf("no SSH Key file: %s %s(%s) \n", server.env, server.host_name, server.ip)
	}

	defPort, _ := sssCfg.SshPort[server.env]
	if len(server.port) > 0 {
		defPort = server.port
	}

	defUser, _ := sssCfg.UserName[server.env]
	if len(server.user) > 0 {
		defUser = server.user
	}

	var cmds []string

	if defPort != "" {
		cmds = append(cmds, "-p"+defPort)
	}

	if sssCfg.SshArgs != "" {
		cmds = append(cmds, sssCfg.SshArgs)
	}

	if len(defUser) > 0 {
		cmds = append(cmds, defUser+"@"+server.ip)
	} else {
		cmds = append(cmds, server.ip)
	}

	if sssCfg.ShowBadge {
		badge := server.host_name + ":" + server.ip
		iterm2.PrintBadge(badge)
	}

	iterm2.PrintHostName()

	iterm2.PrintTabTitle(server.host_name)

	var c *iterm2.RgbColor
	switch server.host_type {
	case "blackpearl":
		c = &iterm2.RgbColor{
			Red: 255,
		}
	case "vrouter":
		c = &iterm2.RgbColor{
			Blue: 255,
		}

	default:
		c = &iterm2.RgbColor{
			Green: 205,
		}

	}

	iterm2.PrintTabBGColor(*c)
	iterm2.PrintRemoteHostName(server.host_name)

	cmds = append(cmds, "-i"+k)

	execute("ssh", cmds...)

	if sssCfg.ShowBadge {
		iterm2.PrintBadge("")
	}

	iterm2.PrintRemoteHostName("")
	iterm2.PrintTabTitle("")
	iterm2.PrintResetTabBGColor()

	return nil
}
*/

func SetIterm2Env(sssCfg *SssConfig, server *server) {
	if sssCfg.ShowBadge {
		badge := server.host_name + ":" + server.ip
		iterm2.PrintBadge(badge)
	}

	iterm2.PrintHostName()
	iterm2.PrintTabTitle(server.host_name)
	iterm2.PrintRemoteHostName(server.host_name)

	var c *iterm2.RgbColor
	switch server.host_type {
	case "blackpearl":
		c = &iterm2.RgbColor{
			Red: 255,
		}
	case "vrouter":
		c = &iterm2.RgbColor{
			Blue: 255,
		}

	default:
		c = &iterm2.RgbColor{
			Green: 205,
		}

	}

	iterm2.PrintTabBGColor(*c)
}

func ResetIterm2Env(sssCfg *SssConfig) {
	if sssCfg.ShowBadge {
		iterm2.PrintBadge("")
	}

	iterm2.PrintRemoteHostName("")
	iterm2.PrintTabTitle("")

	var c *iterm2.RgbColor
	tabBgColor := os.Getenv("IT2_TAB_BG_COLOR")
	if tabBgColor != "" {
		cl := strings.Split(tabBgColor, " ")

		if len(cl) >= 3 {
			c = &iterm2.RgbColor{
				Green: 205,
			}

			if i, err := strconv.Atoi(cl[0]); err == nil {
				c.Red = i
			}
			if i, err := strconv.Atoi(cl[1]); err == nil {
				c.Green = i
			}
			if i, err := strconv.Atoi(cl[2]); err == nil {
				c.Blue = i
			}
		}
	}

	if c != nil {
		iterm2.PrintTabBGColor(*c)
	} else {
		iterm2.PrintResetTabBGColor()
	}
}

func ExecSSH(sssCfg *SssConfig, server *server) error {
	if server == nil {
		return fmt.Errorf("server is empty")
	}

	k, ok := sssCfg.KeyFile[server.env]
	if !ok || k == "" {
		return fmt.Errorf("no SSH Key file: %s %s(%s) \n", server.env, server.host_name, server.ip)
	}

	defPort, _ := sssCfg.SshPort[server.env]
	if len(server.port) > 0 {
		defPort = server.port
	}

	defUser, _ := sssCfg.UserName[server.env]
	if len(server.user) > 0 {
		defUser = server.user
	}

	var cmds []string

	if defPort != "" {
		cmds = append(cmds, "-p"+defPort)
	}

	if sssCfg.SshArgs != "" {
		cmds = append(cmds, sssCfg.SshArgs)
	}

	if len(defUser) > 0 {
		cmds = append(cmds, defUser+"@"+server.ip)
	} else {
		cmds = append(cmds, server.ip)
	}

	/*
		signalChan := make(chan os.Signal, 1)
		//signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGUSR1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		for {
			signal := <-signalChan
			switch signal {
			case syscall.SIGHUP:
				fmt.Printf("SIGHUP(%d)\n", signal)
			case syscall.SIGINT:
				fmt.Printf("SIGINT(%d)\n", signal)
			case syscall.SIGTERM:
				fmt.Printf("SIGTERM(%d)\n", signal)
			default:
				fmt.Printf("Unknown signal(%d)\n", signal)
			}
		}
	*/

	sshClient := ssh.Client{
		ServerAddress: server.ip + ":" + defPort,
		KeyFile:       k,
	}

	sshClient.User = defUser
	sshClient.Timeout = time.Second * 5

	fmt.Printf("> connect %s: ssh -i %s %s@%s \n", server.host_name, k, sshClient.User, sshClient.ServerAddress)

	err := sshClient.Connect()
	if err != nil {
		fmt.Printf("ssh err: %s \n", err)
		return err
	}

	SetIterm2Env(sssCfg, server)
	defer ResetIterm2Env(sssCfg)

	err = sshClient.Shell()
	if err != nil {
		fmt.Printf("ssh err: %s \n", err)
	}

	return nil
}
