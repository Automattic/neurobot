package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type registeredRoute func(w http.ResponseWriter, val map[string]string)

// Server holds the data for running a http server
type Server struct {
	port   int
	routes map[string]registeredRoute
}

// NewServer returns a new instance of http server
func NewServer(port int) Server {
	return Server{
		port:   port,
		routes: make(map[string]registeredRoute),
	}
}

// Run method starts the http server
func (s Server) Run() {
	log.Printf("Starting webhook listener at port %d", s.port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil); err != nil {
		log.Fatal(err)
	}
}

// RegisterRoute saves the callback func for a particular route
func (s Server) RegisterRoute(route string, fn registeredRoute) error {
	if _, ok := s.routes[route]; ok {
		return fmt.Errorf("route %s already registered", route)
	}

	s.routes[route] = fn

	// handle the actual request
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		values := make(map[string]string)
		switch r.Method {
		case "GET":
			q := r.URL.Query()
			for i, v := range q {
				values[i] = v[0]
			}
		case "POST":
			switch r.Header.Values("Content-Type")[0] {
			case "application/json":
				var jsonResult map[string]interface{}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "400 bad request", http.StatusBadRequest)
				}
				json.Unmarshal(body, &jsonResult)
			case "application/x-www-form-urlencoded":
				err := r.ParseForm()
				if err != nil {
					http.Error(w, "400 bad request", http.StatusBadRequest)
				}
				q := r.Form
				for i, v := range q {
					values[i] = v[0]
				}
			}
		}

		fn(w, values)
	})

	return nil
}
