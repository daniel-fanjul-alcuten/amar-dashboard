package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type Order struct {
	Column    string
	Ascending bool
}

type ServerSideRequest struct {
	draw          int
	search        string
	orders        []Order
	start, length int
}

func (q *ServerSideRequest) Parse(r *http.Request) (err error) {
	drawString := r.FormValue("draw")
	if q.draw = 0; drawString != "" {
		q.draw, _ = strconv.Atoi(drawString)
	}
	q.search = r.FormValue("search[value]")
	i := 0
	for {
		colIdString := r.FormValue(fmt.Sprintf("order[%v][column]", i))
		if colIdString == "" {
			break
		}
		var colId int
		colId, err = strconv.Atoi(colIdString)
		if err != nil {
			err = fmt.Errorf("Invalid column: %v", err.Error())
			return
		}
		col := r.FormValue(fmt.Sprintf("columns[%v][name]", colId))
		dir := r.FormValue(fmt.Sprintf("order[%v][dir]", i))
		if dir == "asc" {
			q.orders = append(q.orders, Order{col, true})
		} else if dir == "desc" {
			q.orders = append(q.orders, Order{col, false})
		} else {
			err = fmt.Errorf("Invalid dir: %v", dir)
			return
		}
		i++
	}
	startString := r.FormValue("start")
	if q.start = 0; startString != "" {
		q.start, _ = strconv.Atoi(startString)
	}
	lengthString := r.FormValue("length")
	if q.length = math.MaxInt32; lengthString != "" {
		q.length, _ = strconv.Atoi(lengthString)
	}
	return
}
