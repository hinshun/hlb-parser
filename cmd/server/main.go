package main

import (
	"log"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":9090", http.FileServer(http.Dir("public")))
	if err != nil {
		log.Fatalf("err: %s", err)
	}
}
