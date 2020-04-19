// Package server serves the content.
package server

import (
	"html/template"
	"log"
	"net/http"

	"github.com/studieren-ohne-grenzen/sog-podcast/generator"
)

var indexTemplate *template.Template

var episodes []generator.EpisodeConfigType

func init() {
	temp, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	indexTemplate = temp
}

// ErrorClosure calculates the error closure of a http function
func ErrorClosure(inputFunc func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := inputFunc(writer, request)
		if err != nil {
			log.Println(err)
			writer.Write([]byte("Fehler :-("))
		}
	}
}

// Index handles the index (/) of the server
var Index = ErrorClosure(index)

func index(writer http.ResponseWriter, request *http.Request) error {
	err := indexTemplate.Execute(writer, generator.CurrentConfig.Episode)
	if err != nil {
		return err
	}
	return nil
}

// RSSFeed handles the rss feed
var RSSFeed = ErrorClosure(rssFeed)

func rssFeed(writer http.ResponseWriter, request *http.Request) error {
	writer.Header().Add("Content-Type", "application/rss+xml")
	writer.Write(generator.CurrentFeedBytes)
	return nil
}
