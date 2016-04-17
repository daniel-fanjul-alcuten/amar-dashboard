package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"sync"
)

var (
	Address                *string
	State                  *string
	RefreshContextPath     *string
	StockHtmlContextPath   *string
	StockJsonContextPath   *string
	MinimumHtmlContextPath *string
	MinimumJsonContextPath *string
)

func init() {
	Address = flag.String("listen", ":8080", "address to listen")
	State = flag.String("state", "state.json", "file with the state")
	RefreshContextPath = flag.String("refresh-context-path", "/refresh", "context path")
	StockHtmlContextPath = flag.String("stock-html-context-path", "/stock", "context path")
	StockJsonContextPath = flag.String("stock-json-context-path", "/stock.json", "context path")
	MinimumHtmlContextPath = flag.String("minimum-html-context-path", "/minimum", "context path")
	MinimumJsonContextPath = flag.String("minimum-json-context-path", "/minimum.json", "context path")
}

type Server struct {
	l     *sync.RWMutex
	m     *Model
	stock *template.Template
	min   *template.Template
}

func NewServer() (s *Server, err error) {
	m, err := Load()
	if err != nil {
		return
	}
	stock, err := template.ParseFiles("stock.html")
	if err != nil {
		return
	}
	min, err := template.ParseFiles("minimum.html")
	if err != nil {
		return
	}
	s = &Server{&sync.RWMutex{}, m, stock, min}
	return
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	v, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc(*RefreshContextPath, v.Refresh)
	http.HandleFunc(*StockHtmlContextPath, v.StockHtml)
	http.HandleFunc(*StockJsonContextPath, v.StockJson)
	http.HandleFunc(*MinimumHtmlContextPath, v.MinimumHtml)
	http.HandleFunc(*MinimumJsonContextPath, v.MinimumJson)
	log.Printf("Listening at %v\n", *Address)
	log.Fatal(http.ListenAndServe(*Address, nil))
}
