// SpaceTraders dashboard server
package main

import (
	"bytes"
	"encoding/json"
	//"encoding/json"
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

type mapInfo struct {
	MaxY, MinX, Divisor int
	XRange, YRange      []bool
	System              string
	Waypoints           []objects.Waypoint
}

func main() {
	ticker := time.NewTicker(2001 * time.Millisecond)
	data := dashboardData{}
	if agentName != "" {
		/*var ships objects.AllShips
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
		*/
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

	funcMaps := template.FuncMap{
		"increment": func(i int) int { return i + 1 },
		"decrement": func(i int) int { return i - 1 },
		"and3":      func(a, b, c bool) bool { return a && b && c },
		"div":       func(a, b int) int { return a / b },
	}
	temMap := template.Must(template.New("map.html").Funcs(funcMaps).ParseFiles("html/map.html"))
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		system := "X1-V57"
		contents, err := os.ReadFile(fmt.Sprintf("maps/%s.json", system))
		if err != nil {
			log.Fatal(err)
		}
		contents = bytes.Trim(contents, "\n")
		lines := bytes.Split(contents, []byte("\n"))

		maxX, maxY := -100_000, -100_000
		minX, minY := 100_000, 100_000
		var waypoints []objects.Waypoint
		for _, line := range lines {
			var waypoint objects.Waypoint
			json.Unmarshal(line, &waypoint)
			waypoints = append(waypoints, waypoint)
			if waypoint.X < minX {
				minX = waypoint.X
			}
			if waypoint.X > maxX {
				maxX = waypoint.X
			}
			if waypoint.Y < minY {
				minY = waypoint.Y
			}
			if waypoint.Y > maxY {
				maxY = waypoint.Y
			}
		}

		// TODO - finish map
		// TODO: account overlapping sites, like moons with same coords
		divisor := 10
		systemInfo := mapInfo{Divisor: divisor, MaxY: maxY / divisor, MinX: minX / divisor, XRange: make([]bool, (maxX-minX+1)/divisor), YRange: make([]bool, (maxY-minY+1)/divisor), System: system, Waypoints: waypoints}
		err = temMap.Execute(w, systemInfo)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/map-hello", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<td style="width:1em; border: 1px solid black" hx-put="/map-remove" hx-trigger="mouseleave" hx-swap="outerHTML">Hello!</td>`))
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/map-remove", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<td style="width:1em; border: 1px solid black" hx-put="/map-hello" hx-trigger="click" hx-swap="outerHTML"></td>`))
		if err != nil {
			log.Fatal(err)
		}
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
