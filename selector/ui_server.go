package selector

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/sisyphsu/smart-server-selector/log"
)

const (
	MATCH_USER int = 1 << iota
	MATCH_PORT
	MATCH_DESC
	MATCH_IP
	MATCH_HOST_NAME
	MATCH_HOST_TYPE
	MATCH_ENV
)

type ServerUI struct {
	flex *tview.Flex
	rows []*row

	offset  int
	keyword string
	kws     []string
	visible []server
	all     []server

	Logger *logrus.Logger
}

func init() {
	log.InitLogger(logrus.InfoLevel)
}

func newServersUI(all []server) *ServerUI {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true).
		SetBorderColor(tcell.ColorDarkCyan).
		SetBackgroundColor(tcell.ColorBlack)
	v := &ServerUI{
		flex:   flex,
		all:    all,
		Logger: log.GetLogger(),
	}

	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		v.render()
	})

	return v
}

func (s *ServerUI) onEvent(event *tcell.EventKey) bool {
	l := len(s.visible)

	//s.Logger.Infof("onEvent %s, len=%d, offset=%d", event.Name(), l, s.offset)

	switch event.Key() {
	case tcell.KeyDown, tcell.KeyTab, tcell.KeyPgDn: // select down
		s.selectOffset((s.offset + 1 + l) % l)
		return true
	case tcell.KeyUp, tcell.KeyBacktab, tcell.KeyPgUp: // select up
		s.selectOffset((s.offset - 1 + l) % l)
		return true
	}
	return false
}

func (s *ServerUI) setKeyword(kw string) {
	s.keyword = kw
	s.selectOffset(0)
}

func (s *ServerUI) setServers(all []server) {
	s.all = all
	s.selectOffset(0)
}

func (s *ServerUI) selectOffset(off int) {
	s.offset = off
	s.render()
}

func (s *ServerUI) flushVisible() {
	var kws = make([]string, 0)
	for _, kw := range strings.Split(s.keyword, " ") {
		kw = strings.TrimSpace(kw)
		if len(kw) == 0 {
			continue
		}
		kws = append(kws, kw)
	}

	var matched int
	var result []server
	for _, server := range s.all {
		server.score = 0

		for _, kw := range kws {
			kw = strings.ToLower(kw)
			if strings.Contains(strings.ToLower(server.env), kw) {
				//server.score += 1000
				server.score |= MATCH_ENV
				matched |= MATCH_ENV
			}

			if strings.Contains(strings.ToLower(server.host_type), kw) {
				//server.score += 500
				server.score |= MATCH_HOST_TYPE
				matched |= MATCH_HOST_TYPE
			}
			if strings.Contains(strings.ToLower(server.host_name), kw) {
				//server.score += 300
				server.score |= MATCH_HOST_NAME
			}
			if strings.Contains(strings.ToLower(server.ip), kw) {
				//server.score += 200
				server.score |= MATCH_IP
			}
			if strings.Contains(strings.ToLower(server.desc), kw) {
				//server.score += 100
				server.score |= MATCH_DESC
			}
			if strings.Contains(strings.ToLower(server.port), kw) {
				//server.score += 10
				server.score |= MATCH_PORT
			}
			if strings.Contains(strings.ToLower(server.user), kw) {
				//server.score += 1
				server.score |= MATCH_USER
			}
		}

		if server.score > 0 || len(kws) == 0 {
			result = append(result, server)
		}
	}

	if matched > 0 {
		var ret1 []server
		for _, s := range result {
			if (s.score & matched) == matched {
				ret1 = append(ret1, s)
			}
		}

		result = ret1
	}

	sort.Sort(serverArray(result))

	s.visible = result
	s.kws = kws
}

func (s *ServerUI) flushRows() {
	_, _, _, height := s.flex.GetInnerRect()
	for l := len(s.rows); l > height; {
		l--
		s.flex.RemoveItem(s.rows[l].flex)
		s.rows = s.rows[:l]
	}
	for l := len(s.rows); l < height; {
		l++
		newRow := newRow()
		s.flex.AddItem(newRow.flex, 1, 0, false)
		s.rows = append(s.rows, newRow)
	}
}

func (s *ServerUI) render() {
	s.flushRows()
	s.flushVisible()
	servers := s.visible
	offset := s.offset
	if rowNum := len(s.rows); offset >= rowNum {
		servers = servers[offset-rowNum : offset]
		offset = rowNum - 1
	}

	for i, row := range s.rows {
		var selected = i == offset
		var curr *server
		if i < len(servers) {
			tmp := servers[i]
			curr = &tmp
		}

		row.render(curr, selected, s.kws)
	}
}
