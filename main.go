package main

import (
	"net/http"
	"os"

	anniversary "github.com/k-p5w/go-AnniversaryDay-SVG/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	http.HandleFunc("/", anniversary.Handler)
	http.ListenAndServe(":"+port, nil)
	// debugのときはこれでファイアウォールの設定がでなくなるらしい
	// http.ListenAndServe("localhost:"+port, nil)
}
