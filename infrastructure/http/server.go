package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"neurobot/model/payload"
)

type requestHandler func(w http.ResponseWriter, r *http.Request, val payload.Payload)

type httpError struct {
	StatusCode int
	Message    string
	Error      error
}

// Server holds the data for running a http server
type Server struct {
	port   int
	routes map[string]requestHandler
}

// NewServer returns a new instance of http server
func NewServer(port int) *Server {
	return &Server{
		port:   port,
		routes: make(map[string]requestHandler),
	}
}

// Run method starts the http server
func (s *Server) Run() {
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil); err != nil {
		log.Fatal(err)
	}
}

// Port returns port on which server is configured to run
func (s *Server) Port() int {
	return s.port
}

// RegisterRoute saves the callback func for a particular route
func (s *Server) RegisterRoute(route string, fn requestHandler) error {
	if _, ok := s.routes[route]; ok {
		return fmt.Errorf("route %s already registered", route)
	}

	s.routes[route] = fn

	// handle the actual request
	http.HandleFunc(fmt.Sprintf("%s", route), func(w http.ResponseWriter, r *http.Request) {
		requestParameters, err := s.parseRequest(r)
		if err != nil {
			log.Printf("Failed to parse request: %s\n", err.Error)
			http.Error(w, err.Message, err.StatusCode)
			return
		}

		fn(w, r, requestParameters)
	})

	return nil
}

func (s Server) parseRequest(r *http.Request) (p payload.Payload, err *httpError) {
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
				return p, &httpError{
					StatusCode: http.StatusBadRequest,
					Message:    "Failed to parse JSON request body",
					Error:      err,
				}
			}
			if err := json.Unmarshal(body, &jsonResult); err != nil {
				return p, &httpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to parse JSON request body",
					Error:      err,
				}
			}
		case "application/x-www-form-urlencoded":
			err := r.ParseForm()
			if err != nil {
				return p, &httpError{
					StatusCode: http.StatusBadRequest,
					Message:    "Failed to parse request body",
					Error:      err,
				}
			}
			q := r.Form
			for i, v := range q {
				values[i] = v[0]
			}
		}
	}

	// convert map[string]string values to payload type
	if value, ok := values["message"]; ok {
		p.Message = value
	}
	if value, ok := values["room"]; ok {
		p.Room = value
	}

	return
}
