package main

import (
	"fmt"
	"net/http"
	"os"

	anniversary "github.com/k-p5w/go-AnniversaryDay-SVG/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	fmt.Println("main start.")
	http.HandleFunc("/", anniversary.Handler)
	http.ListenAndServe(":"+port, nil)
}
