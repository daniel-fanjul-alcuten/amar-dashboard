package main

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

type Stuff struct {
	Name  string
	Link  string
	Count int
}

type Model struct {
	Time    time.Time
	Stuff   map[string]*Stuff
	Minimum map[string]map[string]int
}

func Load() (m *Model, err error) {
	var f *os.File
	if f, err = os.Open(*State); err != nil {
		if os.IsNotExist(err) {
			m, err = &Model{}, nil
		}
		return
	}
	r := bufio.NewReader(f)
	d := json.NewDecoder(r)
	err = d.Decode(&m)
	return
}

func Save(m *Model) (err error) {
	var f *os.File
	if f, err = os.Create(*State + ".tmp"); err != nil {
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	e := json.NewEncoder(w)
	if err = e.Encode(m); err != nil {
		return
	}
	if err = w.Flush(); err != nil {
		return
	}
	if err = f.Close(); err != nil {
		return
	}
	if err = os.Rename(f.Name(), *State); err != nil {
		return
	}
	return
}
