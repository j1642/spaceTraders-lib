// SpaceTraders dashboard server
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/j1642/spaceTraders-lib/composites"
	"github.com/j1642/spaceTraders-lib/objects"
	"github.com/j1642/spaceTraders-lib/requests"
)

var agentName string = getAgentName()

type dashboardData struct {
	Ships []objects.Ship
	Agent objects.Agent
}

func main() {
	ticker := time.NewTicker(1001 * time.Millisecond)
	data := dashboardData{}
	if agentName != "" {
		var ships objects.AllShips
		json.Unmarshal(requests.ListMyShips(ticker).Bytes(), &ships)
		data.Ships = ships.Ships
		// Remove date from time stamp
		for i := range data.Ships {
			_, hms, isCut := strings.Cut(data.Ships[i].Nav.Route.Arrival, "T")
			if !isCut {
				log.Fatal("Missing T in", data.Ships[i].Nav.Route.Arrival)
			}
			hms, _, isCut = strings.Cut(hms, ".")
			if !isCut {
				log.Fatal("Missing . in", data.Ships[i].Nav.Route.Arrival)
			}
			data.Ships[i].Nav.Route.Arrival = hms
		}
		/*
				var agent objects.AgentData
		        json.Unmarshal(requests.ViewAgent(ticker).Bytes(), &agent)
		        data.Agent = agent.Agent
		*/
	}

	runServer(ticker, data)
}

func runServer(ticker *time.Ticker, data dashboardData) {
	temIndex := template.Must(template.New("root.html").ParseFiles("html/root.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := temIndex.Execute(w, data)
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
	http.HandleFunc("/edit-agent", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/edit-agent.html")
	})
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/map.html")
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

	// Set ships to dock or orbit
	http.HandleFunc("/flip-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			requestData := strings.Split(string(body), "&")
			flipTo := ""
			shipName := ""
			for i := range requestData {
				attrValue := strings.Split(requestData[i], "=")
				if attrValue[0] == "flip-status" {
					flipTo = attrValue[1]
				} else if attrValue[0] == "ship" {
					shipName = attrValue[1]
				}
			}

			_, err = w.Write([]byte(strings.Join([]string{
				`<form hx-put="flip-status" hx-target="this" hx-swap="outerHTML">
                        <div>`, flipTo, `</div>`,
				`<input type="hidden" name="ship" value="`, shipName, `"/>
                        <button name="flip-status" value="IN_ORBIT">Orbit</button>
                        <button name="flip-status" value="DOCKED">Dock</button>
                    </form>`}, ""),
			))
			if err != nil {
				log.Fatal(err)
			}

			if flipTo == "IN_ORBIT" {
				requests.Orbit(shipName, ticker)
			} else if flipTo == "DOCKED" {
				requests.DockShip(shipName, ticker)
			} else {
				log.Fatal("invalid new status", flipTo)
			}

			log.Println(shipName, flipTo)
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
