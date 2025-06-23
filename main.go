package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
)

type ShortifyData struct {
	URL      string
	ShortURL string
	Error    string
}

var (
	urlMap   = make(map[string]string)
	mapMutex = sync.Mutex{}
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Starting Shortify server on http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := ShortifyData{}

	if r.Method == http.MethodPost {
		r.ParseForm()
		url := r.FormValue("url")
		if url == "" {
			data.Error = "Please enter a URL"
		} else {
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}

			hash := md5.Sum([]byte(url))
			shortKey := hex.EncodeToString(hash[:3]) // Use first 6 chars for demo

			mapMutex.Lock()
			urlMap[shortKey] = url
			mapMutex.Unlock()

			data.URL = url
			data.ShortURL = fmt.Sprintf("http://shortify/%s", shortKey)
		}
	}

	key := strings.TrimPrefix(r.URL.Path, "/")
	if key != "" && key != "favicon.ico" {
		mapMutex.Lock()
		originalURL, exists := urlMap[key]
		mapMutex.Unlock()
		if exists {
			http.Redirect(w, r, originalURL, http.StatusSeeOther)
			return
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, data)
}
