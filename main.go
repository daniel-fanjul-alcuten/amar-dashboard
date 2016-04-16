package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	Address              *string
	State                *string
	RefreshContextPath   *string
	StockHtmlContextPath *string
	StockJsonContextPath *string
)

func init() {
	Address = flag.String("listen", ":8080", "address to listen")
	State = flag.String("state", "state.json", "file with the state")
	RefreshContextPath = flag.String("refresh-context-path", "/refresh", "context path")
	StockHtmlContextPath = flag.String("stock-html-context-path", "/stock", "context path")
	StockJsonContextPath = flag.String("stock-json-context-path", "/stock.json", "context path")
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	l := &sync.RWMutex{}
	m, err := Load()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := template.ParseFiles("stock.html"); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc(*RefreshContextPath, func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		str, err := Fetch()
		if err != nil {
			http.Error(w, "Fetch my stuff: "+err.Error(), http.StatusInternalServerError)
			return
		}
		p, err := Parse(now, str)
		if err != nil {
			http.Error(w, "Parse my_stuff: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := func() error {
			l.Lock()
			defer l.Unlock()
			if m.Stuff == nil {
				m.Stuff = make(map[string]*Stuff)
			}
			for n := range m.Stuff {
				m.Stuff[n].Count = 0
			}
			m.Time = p.Time
			for n, s := range p.Stuff {
				if m.Stuff[n] == nil {
					m.Stuff[n] = &Stuff{s.Name, s.Link, s.Guild}
					continue
				}
				*m.Stuff[n] = Stuff{s.Name, s.Link, s.Guild}
			}
			return Save(m)
		}(); err != nil {
			log.Println(err)
			http.Error(w, "Save my_stuff: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, *StockHtmlContextPath, http.StatusSeeOther)
		return
	})
	http.HandleFunc(*StockHtmlContextPath, func(w http.ResponseWriter, r *http.Request) {
		var templ *template.Template
		if templ, err = template.ParseFiles("stock.html"); err != nil {
			log.Println(err)
			http.Error(w, "stock.html: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var data struct {
			FetchedAgo           time.Duration
			StockJsonContextPath string
		}
		d := time.Now().Sub(func() time.Time {
			l.Lock()
			defer l.Unlock()
			return m.Time
		}())
		data.FetchedAgo = d - d%time.Second
		data.StockJsonContextPath = *StockJsonContextPath
		if err := templ.Execute(w, data); err != nil {
			log.Println(err)
		}
	})
	http.HandleFunc(*StockJsonContextPath, func(w http.ResponseWriter, r *http.Request) {
		drawString := r.FormValue("draw")
		draw := 0
		if drawString != "" {
			draw, _ = strconv.Atoi(drawString)
		}
		search := r.FormValue("search[value]")
		var order []Order
		{
			i := 0
			for {
				colIdString := r.FormValue(fmt.Sprintf("order[%v][column]", i))
				if colIdString == "" {
					break
				}
				colId, err := strconv.Atoi(colIdString)
				if err != nil {
					http.Error(w, "Invalid column: "+err.Error(), http.StatusBadRequest)
					return
				}
				col := r.FormValue(fmt.Sprintf("columns[%v][name]", colId))
				if col != "Name" && col != "Count" {
					http.Error(w, "Invalid column name: "+col, http.StatusBadRequest)
					return
				}
				dir := r.FormValue(fmt.Sprintf("order[%v][dir]", i))
				if dir == "asc" {
					order = append(order, Order{col, true})
				} else if dir == "desc" {
					order = append(order, Order{col, false})
				} else {
					http.Error(w, "Invalid dir: "+err.Error(), http.StatusBadRequest)
					return
				}
				i++
			}
		}
		startString := r.FormValue("start")
		start := 0
		if startString != "" {
			start, _ = strconv.Atoi(startString)
		}
		lengthString := r.FormValue("length")
		length := math.MaxInt32
		if lengthString != "" {
			length, _ = strconv.Atoi(lengthString)
		}
		e := json.NewEncoder(w)
		if err := e.Encode(func() Result {
			l.RLock()
			defer l.RUnlock()
			return m.Search(draw, search, order, start, length)
		}()); err != nil {
			log.Println(err)
		}
	})
	log.Printf("Listening at %v\n", *Address)
	log.Fatal(http.ListenAndServe(*Address, nil))
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
