package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

type ProtectedHandler struct {
	H    http.Handler
	User string
	Pass string
}

func (h *ProtectedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.User != "" && h.Pass != "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		user, pass, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return
		}

		if user != h.User || pass != h.Pass {
			http.Error(w, "Not authorized", 401)
			return
		}
	}

	h.H.ServeHTTP(w, r)
}

func main() {
	var (
		port = flag.Int("port", 8081, "port to serve on")
		user = flag.String("u", "", "user to protect with")
		pass = flag.String("p", "", "password to protect with")
	)
	flag.Parse()

	log.Printf("Serving files in the current directory on port %d", *port)
	http.Handle("/", &ProtectedHandler{http.FileServer(http.Dir(".")), *user, *pass})
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
