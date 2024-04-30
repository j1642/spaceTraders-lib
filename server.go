// SpaceTraders dashboard server
package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

var agentName string

func main() {
	runServer()
}

func runServer() {
	temIndex := template.Must(template.New("root.html").ParseFiles("html/root.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := temIndex.Execute(w, agentName)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "htmx.min.js")
	})
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/about.html")
	})

	temRegister := template.Must(template.New("register.html").ParseFiles("html/register.html"))
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			requestData := strings.Split(string(body), "&")
			for i := range requestData {
				attrValue := strings.Split(requestData[i], "=")
				if attrValue[0] == "agent" {
					agentName = attrValue[1]
				}
			}
			err = temRegister.Execute(w, agentName)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		err := temRegister.Execute(w, agentName)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/register-new", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/register-new.html")
	})

	fmt.Println("Server listening on 8080")
	http.ListenAndServe(":8080", nil)
}
