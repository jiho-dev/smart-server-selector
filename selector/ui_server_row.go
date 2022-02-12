package selector

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// an word split from str
type word struct {
	txt       string
	highlight bool
}

// row represent an server in server list, it wrapped one server's render logic.
type row struct {
	flex      *tview.Flex
	env       *tview.TextView
	host_type *tview.TextView
	host_name *tview.TextView
	ip        *tview.TextView
	desc      *tview.TextView
	//score     *tview.TextView
}

// create an new row
func newRow() *row {
	r := &row{
		flex:      tview.NewFlex(),
		env:       tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorLawnGreen),
		host_type: tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorLawnGreen),
		host_name: tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorLawnGreen),
		ip:        tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorLawnGreen),
		desc:      tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorLawnGreen),
		//score:     tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorLawnGreen),
	}
	r.flex.SetDirection(tview.FlexColumn)
	r.flex.AddItem(r.env, 7, 1, false)
	r.flex.AddItem(r.host_type, 12, 1, false)
	r.flex.AddItem(r.host_name, 20, 1, false)
	r.flex.AddItem(r.ip, 30, 1, false)
	r.flex.AddItem(r.desc, 0, 10, false)
	//r.flex.AddItem(r.score, 0, 10, false)

	r.env.SetBackgroundColor(tcell.ColorDefault)
	r.host_type.SetBackgroundColor(tcell.ColorDefault)
	r.host_name.SetBackgroundColor(tcell.ColorDefault)
	r.ip.SetBackgroundColor(tcell.ColorDefault)
	r.desc.SetBackgroundColor(tcell.ColorDefault)
	//r.score.SetBackgroundColor(tcell.ColorDefault)

	return r
}

// render the current row
func (r *row) render(s *server, selected bool, kws []string) {
	if selected {
		r.flex.SetBackgroundColor(tcell.ColorRoyalBlue)
	} else {
		r.flex.SetBackgroundColor(tcell.ColorDefault)
	}
	var env, host_type, host_name, ip, desc string
	//var score string
	if s != nil {
		env = s.env
		host_type = s.host_type
		host_name = s.host_name
		ip = s.ip
		desc = s.desc

		//score = fmt.Sprintf("%d", s.score)
		desc = s.desc
		if len(s.user) > 0 {
			ip = s.user + "@" + s.ip
		}
		if len(s.port) > 0 {
			ip = ip + ":" + s.port
		}
	}
	r.env.SetText(highlight(env, kws))
	r.host_type.SetText(highlight(host_type, kws))
	r.host_name.SetText(highlight(host_name, kws))
	r.ip.SetText(highlight(ip, kws))
	r.desc.SetText(highlight(desc, kws))
	//r.score.SetText(score)
}

// generate highlight text for the specified string
func highlight(s string, kws []string) (r string) {
	for _, word := range splitKws(s, kws) {
		if word.highlight {
			r += "[red]"
		} else {
			r += "[lawngreen]"
		}
		r += word.txt
	}
	return
}

// split the specified string with kws
func splitKws(s string, kws []string) []word {
	result := []word{{txt: s}}
	for _, kw := range kws {
		tmp := make([]word, 0)
		for _, w := range result {
			if w.highlight {
				tmp = append(tmp, w)
				continue
			}
			parts := strings.Split(w.txt, kw)
			for i, part := range parts {
				if i > 0 {
					tmp = append(tmp, word{txt: kw, highlight: true})
				}
				if len(part) > 0 {
					tmp = append(tmp, word{txt: part, highlight: false})
				}
			}
		}
		result = tmp
	}
	return result
}
