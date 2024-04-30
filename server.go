// SpaceTraders dashboard server
package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/j1642/spaceTraders-lib/composites"
)

var agentName string = getAgentName()

func main() {
	ticker := time.NewTicker(2001 * time.Millisecond)
	runServer(ticker)
}

func runServer(ticker *time.Ticker) {
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

	temRegister := template.Must(template.New("register-form.html").ParseFiles("html/register-form.html"))
	http.HandleFunc("/register-form", func(w http.ResponseWriter, r *http.Request) {
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
					log.Println("agentName set to", agentName)
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

	http.HandleFunc("/edit-agent", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/edit-agent.html")
	})

	// Send registration request to SpaceTraders
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && agentName != "" {
			err := composites.DoNewUserBoilerplate(agentName, ticker)
			if err != nil {
				agentName = err.Error()
			}
		}
		log.Println("possibly registered new agent:", agentName)
		err := temIndex.Execute(w, agentName)
		if err != nil {
			log.Fatal(err)
		}
	})

	fmt.Println("Server listening on 8080")
	http.ListenAndServe(":8080", nil)
}

func getAgentName() string {
	agent := ""
	contents, err := os.ReadFile("miningDrones.txt")
	if err != nil {
		return ""
	}
	lines := strings.Split(string(contents), "\n")
	agent = strings.Split(lines[0], "-")[0]
	return agent
}
