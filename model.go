package main

import (
	"sort"
	"strings"
	"time"
)

type Stuff struct {
	Name  string
	Link  string
	Count int
}

type Model struct {
	Time  time.Time
	Stuff map[string]*Stuff
}

type Order struct {
	Column    string
	Ascending bool
}

type Row struct {
	Name  string
	Count int
}

type OrderRows struct {
	Orders []Order
	Rows   []Row
}

func (or OrderRows) Len() int {
	return len(or.Rows)
}

func (or OrderRows) Less(i, j int) bool {
	for _, o := range or.Orders {
		if o.Column == "Name" {
			if or.Rows[i].Name == or.Rows[j].Name {
				continue
			}
			if o.Ascending {
				return or.Rows[i].Name < or.Rows[j].Name
			}
			return or.Rows[i].Name > or.Rows[j].Name
		}
		if o.Column == "Count" {
			if or.Rows[i].Count == or.Rows[j].Count {
				continue
			}
			if o.Ascending {
				return or.Rows[i].Count < or.Rows[j].Count
			}
			return or.Rows[i].Count > or.Rows[j].Count
		}
		return or.Rows[i].Name < or.Rows[j].Name
	}
	return false
}

func (or OrderRows) Swap(i, j int) {
	or.Rows[i], or.Rows[j] = or.Rows[j], or.Rows[i]
}

type Result struct {
	Draw            int    `json:"draw"`
	RecordsTotal    int    `json:"recordsTotal"`
	RecordsFiltered int    `json:"recordsFiltered"`
	Data            []Row  `json:"data"`
	Error           string `json:"error"`
}

func (m Model) Search(draw int, search string, order []Order, start, length int) (e Result) {
	e.Draw = draw
	e.RecordsTotal = len(m.Stuff)
	or := OrderRows{order, nil}
	search = strings.ToUpper(search)
	for n, s := range m.Stuff {
		if search == "" || strings.Contains(strings.ToUpper(n), search) {
			e.RecordsFiltered++
			or.Rows = append(or.Rows, Row{n, s.Count})
		}
	}
	sort.Sort(or)
	i := 0
	for _, r := range or.Rows {
		if i < start {
			i++
			continue
		} else if len(e.Data) < length {
			i++
			e.Data = append(e.Data, r)
			continue
		}
		break
	}
	return
}
