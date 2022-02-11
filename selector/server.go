package selector

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type server struct {
	env       string
	host_type string
	host_name string
	ip        string
	port      string
	user      string
	desc      string
	score     int
}

type serverArray []server

func (a serverArray) Len() int      { return len(a) }
func (a serverArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a serverArray) Less(i, j int) bool {
	if a[i].score != a[j].score {
		// XXX: descending
		return a[i].score > a[j].score
	}

	if a[i].env != a[j].env {
		return a[i].env < a[j].env
	}

	if a[i].host_type != a[j].host_type {
		return a[i].host_type < a[j].host_type
	}

	if a[i].host_name != a[j].host_name {
		return a[i].host_name < a[j].host_name
	}

	return i < j
}

// load servers from config file.
func loadServers(cfg string) (arr []server) {
	arr = make([]server, 0)
	fs, _ := ioutil.ReadFile(cfg)
	if len(fs) == 0 {
		return
	}
	body := string(fs)
	var errs []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' || line[0] == '/' {
			continue
		}
		s := parseServerFull(line)
		if s == nil {
			s = parseServerSimp(line)
		}
		if s != nil {
			arr = append(arr, *s)
		} else {
			errs = append(errs, line)
		}
	}
	if len(errs) > 0 {
		fmt.Printf("> some invalid config in file[%v]: \n", SssFile)
		for i, e := range errs {
			fmt.Printf("> %v: %v \n", i+1, e)
		}
		fmt.Println("> press any key to continue")
		getchar()
	}
	return
}

var fullPtn = regexp.MustCompile("^(\\w+)\\s+([\\w.]+)\\s+(\\d+)\\s+([\\w.]+)\\s+(.*)$")

// parse server by full pattern
func parseServerFull(s string) *server {
	//sm := fullPtn.FindStringSubmatch(s)
	sm := strings.Split(s, " ")
	if len(sm) == 0 || len(sm) != 7 {
		return nil
	}
	return &server{
		env:       sm[0],
		host_type: sm[1],
		host_name: sm[2],
		ip:        sm[3],
		port:      sm[4],
		user:      sm[5],
		desc:      sm[6],
	}
}

var simpPtn = regexp.MustCompile("^(\\w+)\\s+([\\w.]+)\\s+(.*)$")

// parse server by simple pattern
func parseServerSimp(s string) *server {
	//sm := simpPtn.FindStringSubmatch(s)
	sm := strings.Split(s, " ")
	if len(sm) == 0 || len(sm) != 4 {
		return nil
	}
	return &server{
		env:       sm[0],
		host_type: sm[1],
		host_name: sm[2],
		ip:        sm[3],
		port:      "",
		user:      "centos",
		//desc:      sm[4],
	}
}
