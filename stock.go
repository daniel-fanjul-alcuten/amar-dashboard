package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type StockRow struct {
	Name    string
	Count   int
	Minimum int
	Surplus int
}

type StockRows struct {
	Orders []Order
	Rows   []StockRow
}

func (sr StockRows) Len() int {
	return len(sr.Rows)
}

func (sr StockRows) Less(i, j int) bool {
	for _, o := range sr.Orders {
		if o.Column == "Name" {
			if sr.Rows[i].Name == sr.Rows[j].Name {
				continue
			}
			if o.Ascending {
				return sr.Rows[i].Name < sr.Rows[j].Name
			}
			return sr.Rows[i].Name > sr.Rows[j].Name
		}
		if o.Column == "Count" {
			if sr.Rows[i].Count == sr.Rows[j].Count {
				continue
			}
			if o.Ascending {
				return sr.Rows[i].Count < sr.Rows[j].Count
			}
			return sr.Rows[i].Count > sr.Rows[j].Count
		}
		if o.Column == "Minimum" {
			if sr.Rows[i].Minimum == sr.Rows[j].Minimum {
				continue
			}
			if o.Ascending {
				return sr.Rows[i].Minimum < sr.Rows[j].Minimum
			}
			return sr.Rows[i].Minimum > sr.Rows[j].Minimum
		}
		if o.Column == "Surplus" {
			if sr.Rows[i].Surplus == sr.Rows[j].Surplus {
				continue
			}
			if o.Ascending {
				return sr.Rows[i].Surplus < sr.Rows[j].Surplus
			}
			return sr.Rows[i].Surplus > sr.Rows[j].Surplus
		}
		return sr.Rows[i].Name < sr.Rows[j].Name
	}
	return false
}

func (sr StockRows) Swap(i, j int) {
	sr.Rows[i], sr.Rows[j] = sr.Rows[j], sr.Rows[i]
}

type StockResult struct {
	Draw            int        `json:"draw"`
	RecordsTotal    int        `json:"recordsTotal"`
	RecordsFiltered int        `json:"recordsFiltered"`
	Data            []StockRow `json:"data"`
	Error           string     `json:"error"`
}

func (m Model) StockSearch(q ServerSideRequest) (e StockResult) {
	e.Draw = q.draw
	e.RecordsTotal = len(m.Stuff)
	sr := StockRows{q.orders, nil}
	search := strings.ToUpper(q.search)
	for n, s := range m.Stuff {
		if search == "" || strings.Contains(strings.ToUpper(n), search) {
			e.RecordsFiltered++
			min := 0
			for _, ii := range m.Minimum {
				if ii[n] > 0 {
					min += ii[n]
				}
			}
			link := fmt.Sprintf("<a href=\"http://amar.bornofsnails.net%v\">%v</a>", s.Link, s.Name)
			sr.Rows = append(sr.Rows, StockRow{link, s.Count, min, s.Count - min})
		}
	}
	sort.Sort(sr)
	i := 0
	for _, r := range sr.Rows {
		if i < q.start {
			i++
			continue
		} else if len(e.Data) < q.length {
			i++
			e.Data = append(e.Data, r)
			continue
		}
		break
	}
	return
}

func (v *Server) StockJson(w http.ResponseWriter, r *http.Request) {
	var q ServerSideRequest
	if err := q.Parse(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, o := range q.orders {
		if o.Column != "Name" && o.Column != "Count" && o.Column != "Minimum" && o.Column != "Surplus" {
			http.Error(w, "Invalid column name: "+o.Column, http.StatusBadRequest)
			return
		}
	}
	e := json.NewEncoder(w)
	if err := e.Encode(func() StockResult {
		v.l.RLock()
		defer v.l.RUnlock()
		return v.m.StockSearch(q)
	}()); err != nil {
		log.Println(err)
	}
}

func (v *Server) StockHtml(w http.ResponseWriter, r *http.Request) {
	var data struct {
		FetchedAgo           time.Duration
		StockJsonContextPath string
	}
	d := time.Now().Sub(func() time.Time {
		v.l.Lock()
		defer v.l.Unlock()
		return v.m.Time
	}())
	data.FetchedAgo = d - d%time.Second
	data.StockJsonContextPath = *StockJsonContextPath
	if err := v.stock.Execute(w, data); err != nil {
		log.Println(err)
	}
}
