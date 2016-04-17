package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type MinimumRow struct {
	User  string
	Item  string
	Count int
}

type MinimumRows struct {
	Orders []Order
	Rows   []MinimumRow
}

func (mr MinimumRows) Len() int {
	return len(mr.Rows)
}

func (mr MinimumRows) Less(i, j int) bool {
	for _, o := range mr.Orders {
		if o.Column == "User" {
			if mr.Rows[i].User == mr.Rows[j].User {
				continue
			}
			if o.Ascending {
				return mr.Rows[i].User < mr.Rows[j].User
			}
			return mr.Rows[i].User > mr.Rows[j].User
		}
		if o.Column == "Item" {
			if mr.Rows[i].Item == mr.Rows[j].Item {
				continue
			}
			if o.Ascending {
				return mr.Rows[i].Item < mr.Rows[j].Item
			}
			return mr.Rows[i].Item > mr.Rows[j].Item
		}
		if o.Column == "Count" {
			if mr.Rows[i].Count == mr.Rows[j].Count {
				continue
			}
			if o.Ascending {
				return mr.Rows[i].Count < mr.Rows[j].Count
			}
			return mr.Rows[i].Count > mr.Rows[j].Count
		}
		return mr.Rows[i].User < mr.Rows[j].User
	}
	return false
}

func (mr MinimumRows) Swap(i, j int) {
	mr.Rows[i], mr.Rows[j] = mr.Rows[j], mr.Rows[i]
}

type MinimumResult struct {
	Draw            int          `json:"draw"`
	RecordsTotal    int          `json:"recordsTotal"`
	RecordsFiltered int          `json:"recordsFiltered"`
	Data            []MinimumRow `json:"data"`
	Error           string       `json:"error"`
}

func (m Model) MinimumSearch(q ServerSideRequest) (e MinimumResult) {
	e.Draw = q.draw
	e.RecordsTotal = len(m.Minimum)
	sr := MinimumRows{q.orders, nil}
	search := strings.ToUpper(q.search)
	for n, ii := range m.Minimum {
		for i, c := range ii {
			if search == "" || strings.Contains(strings.ToUpper(n), search) || strings.Contains(strings.ToUpper(i), search) {
				e.RecordsFiltered++
				sr.Rows = append(sr.Rows, MinimumRow{n, i, c})
			}
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

func (v *Server) MinimumJson(w http.ResponseWriter, r *http.Request) {
	var q ServerSideRequest
	if err := q.Parse(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, o := range q.orders {
		if o.Column != "User" && o.Column != "Item" && o.Column != "Count" {
			http.Error(w, "Invalid column name: "+o.Column, http.StatusBadRequest)
			return
		}
	}
	e := json.NewEncoder(w)
	if err := e.Encode(func() MinimumResult {
		v.l.RLock()
		defer v.l.RUnlock()
		return v.m.MinimumSearch(q)
	}()); err != nil {
		log.Println(err)
	}
}

func (v *Server) MinimumHtml(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := r.FormValue("user")
		if user == "" {
			http.Redirect(w, r, *MinimumHtmlContextPath, http.StatusSeeOther)
			return
		}
		user = template.HTMLEscapeString(user)
		item := r.FormValue("item")
		if item == "" {
			http.Redirect(w, r, *MinimumHtmlContextPath, http.StatusSeeOther)
			return
		}
		item = template.HTMLEscapeString(item)
		countString := r.FormValue("count")
		count, err := strconv.Atoi(countString)
		if err != nil {
			http.Redirect(w, r, *MinimumHtmlContextPath, http.StatusSeeOther)
			return
		}
		if err = func() error {
			v.l.Lock()
			defer v.l.Unlock()
			u := v.m.Minimum[user]
			if u == nil {
				u = make(map[string]int)
				if v.m.Minimum == nil {
					v.m.Minimum = make(map[string]map[string]int)
				}
				v.m.Minimum[user] = u
			}
			if count > 0 {
				u[item] = count
			} else {
				delete(u, item)
			}
			if len(u) == 0 {
				delete(v.m.Minimum, user)
			}
			return Save(v.m)
		}(); err != nil {
			log.Println(err)
			http.Error(w, "Save my_stuff: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, *MinimumHtmlContextPath, http.StatusSeeOther)
		return
	}
	var data struct {
		MinimumHtmlContextPath string
		MinimumJsonContextPath string
	}
	data.MinimumHtmlContextPath = *MinimumHtmlContextPath
	data.MinimumJsonContextPath = *MinimumJsonContextPath
	if err := v.min.Execute(w, data); err != nil {
		log.Println(err)
	}
	return
}
