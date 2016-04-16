package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var Uid, Pid *string

func init() {
	Uid = flag.String("u", "", "uid")
	Pid = flag.String("p", "", "pid")
}

type MyStuff struct {
	Name      string
	Total     int
	Inventory int
	House     int
	Shared    int
	Guild     int
	Link      string
}

type MyStuffPage struct {
	Time  time.Time
	Stuff map[string]MyStuff
}

var stockRegexp *regexp.Regexp = regexp.MustCompile("<tr><td><a class='link' href='" +
	"([^']+)'>([^<]+)</a><td>([\\d,]*)<td>([\\d,]*)<td>([\\d,]*)<td>([\\d,]*)<td>([\\d,]*)")

func Fetch() (s string, err error) {
	var req *http.Request
	if req, err = http.NewRequest("GET", "http://amar.bornofsnails.net/man/my_stuff", nil); err != nil {
		return
	}
	req.AddCookie(&http.Cookie{Name: "uid", Value: *Uid})
	req.AddCookie(&http.Cookie{Name: "pid", Value: *Pid})
	c := &http.Client{}
	var resp *http.Response
	if resp, err = c.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Get http://amar.bornofsnails.net/man/my_stuff: %v", resp.Status)
		return
	}
	b := &bytes.Buffer{}
	if _, err = io.Copy(b, resp.Body); err != nil {
		return
	}
	s = b.String()
	return
}

func Parse(t time.Time, s string) (page MyStuffPage, err error) {
	page.Time = t
	page.Stuff = make(map[string]MyStuff)
	for _, m := range stockRegexp.FindAllStringSubmatch(s, -1) {
		item := MyStuff{Link: m[1], Name: m[2]}
		if item.Total, err = atoi(m[3]); err != nil {
			return
		}
		if item.Inventory, err = atoi(m[4]); err != nil {
			return
		}
		if item.House, err = atoi(m[5]); err != nil {
			return
		}
		if item.Shared, err = atoi(m[6]); err != nil {
			return
		}
		if item.Guild, err = atoi(m[7]); err != nil {
			return
		}
		page.Stuff[item.Name] = item
	}
	return
}

func atoi(s string) (int, error) {
	s = strings.Replace(s, ",", "", 1)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
