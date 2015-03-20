package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var err error
	port := int64(8081)
	if len(os.Args) > 1 {
		port, err = strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Serving files in the current directory on port %d\n", port)
	http.Handle("/", http.FileServer(http.Dir(".")))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
