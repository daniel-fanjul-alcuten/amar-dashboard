package main

import (
	"log"
	"net/http"
	"time"
)

func (v *Server) Refresh(w http.ResponseWriter, r *http.Request) {
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
		v.l.Lock()
		defer v.l.Unlock()
		if v.m.Stuff == nil {
			v.m.Stuff = make(map[string]*Stuff)
		}
		for n := range v.m.Stuff {
			v.m.Stuff[n].Count = 0
		}
		v.m.Time = p.Time
		for n, s := range p.Stuff {
			if v.m.Stuff[n] == nil {
				v.m.Stuff[n] = &Stuff{s.Name, s.Link, s.Guild}
				continue
			}
			*v.m.Stuff[n] = Stuff{s.Name, s.Link, s.Guild}
		}
		return Save(v.m)
	}(); err != nil {
		log.Println(err)
		http.Error(w, "Save my_stuff: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, *StockHtmlContextPath, http.StatusSeeOther)
	return
}
