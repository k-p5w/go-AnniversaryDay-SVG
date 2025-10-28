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

	// APIとして公開モードで稼働させるとき
	http.ListenAndServe(":"+port, nil)

	// ローカルホストで動かすとき　// debugのときはこれでファイアウォールの設定がでなくなるらしい
	// http.ListenAndServe("localhost:"+port, nil)
}
