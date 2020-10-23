package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/studieren-ohne-grenzen/sog-podcast/server"
)

func main() {
	portPtr := flag.Int("port", 80, "the port to run on")
	flag.Parse()

	staticFs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFs))
	downloadFs := http.FileServer(http.Dir("static/download/"))
	http.Handle("/download/", http.StripPrefix("/download/", downloadFs))

	http.HandleFunc("/", server.Index)
	http.HandleFunc("/rss", server.RSSFeed)

	for {
		err := http.ListenAndServe(":"+strconv.Itoa(*portPtr), nil)
		log.Println(err)
	}
}
