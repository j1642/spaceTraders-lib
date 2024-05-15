// SpaceTraders dashboard server
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
type travelInfo struct {
	Ship      objects.Ship
	Waypoints []objects.Waypoint
}

func main() {
	ticker := time.NewTicker(2001 * time.Millisecond)
	data := dashboardData{}
	if agentName != "" {
		var ships objects.AllShips
		json.Unmarshal(requests.ListMyShips(ticker).Bytes(), &ships)
		data.Ships = ships.Ships
		// Remove date from time stamps
		for i := range data.Ships {
			_, hms, _ := strings.Cut(data.Ships[i].Nav.Route.Arrival, "T")
			hms, _, _ = strings.Cut(hms, ".")
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
	travelDurations := make([]int, len(data.Ships))
	extractCooldowns := make([]int, len(data.Ships))

	temIndex := template.Must(template.New("root.gohtml").ParseFiles("gohtml/root.gohtml"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := temIndex.Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "htmx.min.js")
	})
	http.HandleFunc("/htmx.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "htmx.js")
	})
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gohtml/about.gohtml")
	})
	http.HandleFunc("/edit-agent", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gohtml/edit-agent.gohtml")
	})

	funcMaps := template.FuncMap{
		"increment": func(i int) int { return i + 1 },
		"decrement": func(i int) int { return i - 1 },
		"div":       func(a, b int) int { return a / b },
		"jsonVals": func(cssClass string, idx int) template.HTMLAttr {
			return template.HTMLAttr(strings.Join([]string{
				`{"parent":[{"type":"`, cssClass, `-cell"},{"waypointIdx":"`, fmt.Sprint(idx), `"}]}`}, ""))
		},
	}
	temMap := template.Must(template.New("map.gohtml").Funcs(funcMaps).ParseFiles("gohtml/map.gohtml"))
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		system := data.Ships[0].Nav.SystemSymbol
		contents, err := os.ReadFile(fmt.Sprintf("maps/%s.json", system))
		if err != nil {
			composites.StoreSystemWaypoints(system, ticker)
			contents, err = os.ReadFile(fmt.Sprintf("maps/%s.json", system))
			if err != nil {
				log.Fatal(err)
			}
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
		//   - orbital station do not appear on the map b/c of this
		divisor := 10
		systemInfo := mapInfo{Divisor: divisor, MaxY: maxY / divisor, MinX: minX / divisor, XRange: make([]bool, (maxX-minX+1)/divisor), YRange: make([]bool, (maxY-minY+1)/divisor), System: system, Waypoints: waypoints}
		//systemInfo := mapInfo{Divisor: divisor, MaxY: maxY / divisor, MinX: minX / divisor, XRange: make([]bool, (maxX+1)/divisor), YRange: make([]bool, (maxY+1)/divisor), System: system, Waypoints: waypoints}
		err = temMap.Execute(w, systemInfo)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/map-remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			split := strings.Split(string(body), "%22") // %22 is probably escaped quotation mark
			celestialType := split[3]
			waypointIdx := split[7]

			_, err = w.Write([]byte(
				strings.Join([]string{
					`<td class="map-width bordered `, celestialType, `" hx-put="/map-describe" hx-trigger="click" hx-swap="outerHTML" hx-vals='{"parent":[{"type":"`, celestialType, `"},{"waypointSymbol":"`, waypointIdx, `"}]}'></td>`}, ""),
			))
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		log.Fatal("map-remove: not a PUT request")
	})

	http.HandleFunc("/map-describe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			split := strings.Split(string(body), "%22") // %22 is probably escaped quotation mark
			celestialType := split[3]
			waypointIdx := split[7]

			system := "X1-V57"
			contents, err := os.ReadFile(fmt.Sprintf("maps/%s.json", system))
			if err != nil {
				composites.StoreSystemWaypoints(system, ticker)
				contents, err = os.ReadFile(fmt.Sprintf("maps/%s.json", system))
				if err != nil {
					log.Fatal(err)
				}
			}
			contents = bytes.Trim(contents, "\n")
			lines := bytes.Split(contents, []byte("\n"))

			var waypointDescription string
			idx, err := strconv.Atoi(waypointIdx)
			if err != nil {
				log.Fatal(err)
			}
			line := lines[idx]
			var waypoint objects.Waypoint
			json.Unmarshal(line, &waypoint)
			waypointDescription = fmt.Sprintf("%v", waypoint)

			_, err = w.Write([]byte(
				strings.Join([]string{
					`<td class="bordered" hx-put="/map-remove" hx-swap="outerHTML" hx-trigger="mouseleave" hx-vals='{"parent":[{"type":"`, celestialType, `"},{"waypointSymbol":"`, waypointIdx, `"}]}'>`, waypointDescription, `</td>`}, ""),
			))
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		log.Fatal("map-describe: not a PUT")
	})

	temRegister := template.Must(template.New("register-form.gohtml").ParseFiles("gohtml/register-form.gohtml"))
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
					break
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
			for i, ship := range data.Ships {
				if ship.Symbol == shipName {
					data.Ships[i].Nav.Status = flipTo
					break
				}
			}

			log.Println(shipName, flipTo)
		}
	})

	temTravel := template.Must(template.New("travel.gohtml").ParseFiles("gohtml/travel.gohtml"))
	// Provide travel menu after Travel button is clicked
	http.HandleFunc("/travel", func(w http.ResponseWriter, r *http.Request) {
		/* TODO:
				   - show selected destination description
				   - add button to exit travel menu
		           - add button to travel while traveling, to correct a wrong route
		*/
		shipName := r.URL.Query().Get("ship")

		// Isolate ship of interest
		shipIdx := -1
		for i, ship := range data.Ships {
			if ship.Symbol == shipName {
				shipIdx = i
				break
			}
		}

		waypoints := readSystemWaypointsFromFile(
			data.Ships[shipIdx].Nav.SystemSymbol,
			ticker,
		)

		travelInfo := travelInfo{
			Ship:      data.Ships[shipIdx],
			Waypoints: waypoints,
		}
		err := temTravel.Execute(w, travelInfo)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Return all system waypoints of a certain Type as <options>
	http.HandleFunc("/travel-filter-dests", func(w http.ResponseWriter, r *http.Request) {
		destType := r.URL.Query().Get("dest-type")
		system := r.URL.Query().Get("system")

		waypoints := readSystemWaypointsFromFile(system, ticker)

		message := make([]byte, 0)
		for _, waypoint := range waypoints {
			if waypoint.Type == destType {
				s := strings.Join([]string{
					`<option value="`, waypoint.Symbol, `">`, waypoint.Symbol, `</option>`,
				}, "")
				message = append(message, []byte(s)...)
			}
		}

		_, err := w.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Send the travel HTTP request to the Spacetraders API and housekeeping
	http.HandleFunc("/execute-trip", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		requestData := strings.Split(string(body), "&")
		var shipName string
		var destID string
		for i := range requestData {
			attrValue := strings.Split(requestData[i], "=")
			if attrValue[0] == "ship" {
				shipName = attrValue[1]
			} else if attrValue[0] == "dest-id" {
				destID = attrValue[1]
			}
		}

		// Check that there is enough info to travel
		if shipName == "" || destID == "" {
			log.Printf("travel error: ship=%s, dest=%s", shipName, destID)
			return
		}

		reply := requests.TravelTo(shipName, destID, ticker)
		travelMsg := objects.TravelData{}
		err = json.Unmarshal(reply.Bytes(), &travelMsg)
		if err != nil {
			panic(err)
		}

		// Log trip
		_, arrival, _ := strings.Cut(travelMsg.Travel.Nav.Route.Arrival, "T")
		arrival, _, _ = strings.Cut(arrival, ".")
		log.Printf("%s travels to %s (%s), arriving %s. Fuel %d/%d\n",
			shipName, destID, travelMsg.Travel.Nav.Route.Destination.Type, arrival,
			travelMsg.Travel.Fuel.Current, travelMsg.Travel.Fuel.Capacity,
		)

		// Update root data, displays when the page is refreshed
		shipIdx := -1
		for i, ship := range data.Ships {
			if ship.Symbol == shipName {
				shipIdx = i
				data.Ships[i].Fuel = travelMsg.Travel.Fuel
				data.Ships[i].Nav = travelMsg.Travel.Nav
				data.Ships[i].Nav.Route.Arrival = arrival
				break
			}
		}

		// Get time info for travel countdown
		format := "2006-01-02T15:04:05.000Z"
		start, err := time.Parse(format, travelMsg.Travel.Nav.Route.DepartureTime)
		if err != nil {
			log.Println("Failed to parse time: likely trying to travel from/to the same place")
			return
		}
		end, err := time.Parse(format, travelMsg.Travel.Nav.Route.Arrival)
		if err != nil {
			panic(err)
		}

		travelDuration := end.Sub(start).Round(time.Second).Seconds()
		travelDurations[shipIdx] = int(travelDuration)

		_, err = w.Write([]byte(
			`<div hx-trigger="end-countdown" hx-get="/fresh-travel-button"
            hx-swap="innerHTML"
            hx-target="#td-travel-` + fmt.Sprint(shipIdx) + `"
            hx-vals={"ship":"` + shipName + `"}>
              <p role="status" id="pblabel" tabindex="-1" autofocus>Traveling</p>

              <div
                hx-get="/countdown-progress"
                hx-trigger="every 2s"
                hx-target="this"
                hx-swap="innerHTML"
                hx-vals='{"shipIdx":"` + fmt.Sprint(shipIdx) + `"}'>
              </div>
            </div>`,
		))
	})

	// Decrement the travel countdown of a particular ship
	http.HandleFunc("/countdown-progress", func(w http.ResponseWriter, r *http.Request) {
		shipIdxStr := r.URL.Query().Get("shipIdx")
		shipIdx, err := strconv.Atoi(shipIdxStr)
		if err != nil {
			log.Fatal(err)
		}

		travelDurations[shipIdx] -= 2
		if travelDurations[shipIdx] <= 0 {
			travelDurations[shipIdx] = 0
			w.Header().Add("HX-Trigger", "end-countdown")
		}
		progress := fmt.Sprint(travelDurations[shipIdx])
		_, err = w.Write([]byte(strings.Join([]string{"<p>", progress, "s</p>"}, "")))

		if err != nil {
			log.Fatal(err)
		}
	})

	// Return a travel <button> for a given ship
	http.HandleFunc("/fresh-travel-button", func(w http.ResponseWriter, r *http.Request) {
		shipName := r.URL.Query().Get("ship")
		if shipName == "" {
			log.Fatal("shipName is empty")
		}

		_, err := w.Write([]byte(`
        <td><button hx-get="/travel" hx-swap="outerHTML"
            hx-vals='{"ship":"` + shipName + `"}'>
            Travel</button>
        </td>`,
		))
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/extract", func(w http.ResponseWriter, r *http.Request) {
		// TODO: finish
		shipName := r.URL.Query().Get("ship")

		// Isolate ship of interest
		shipIdx := -1
		for i, ship := range data.Ships {
			if ship.Symbol == shipName {
				shipIdx = i
				break
			}
		}
		if shipIdx == -1 {
			log.Println("/extract failed, shipIdx=-1")
			return
		}

		extractMsg := requests.ExtractOre(data.Ships[shipIdx].Symbol, ticker)
		extractCooldowns[shipIdx] = extractMsg.ExtractBody.Cooldown.RemainingSeconds
		// TODO: Can HTMX update the Cargo <td> and the Extract <td>?
		data.Ships[shipIdx].Cargo = extractMsg.ExtractBody.Cargo
		fmt.Printf("%+v\n", extractMsg)

		_, err := w.Write([]byte(
			`<div hx-trigger="extract-end-cooldown" hx-get="/extract-new-button"
            hx-swap="innerHTML"
            hx-target="#td-extract-` + fmt.Sprint(shipIdx) + `"
            hx-vals={"ship":"` + shipName + `"}>
              <p>Cooldown</p>
              <div
                hx-get="/extract-cooldown-decrement"
                hx-trigger="every 2s"
                hx-target="this"
                hx-swap="innerHTML"
                hx-vals='{"shipIdx":"` + fmt.Sprint(shipIdx) + `"}'>
              </div>
            </div>`,
		))
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/extract-cooldown-decrement", func(w http.ResponseWriter, r *http.Request) {
		shipIdx, err := strconv.Atoi(r.URL.Query().Get("shipIdx"))
		if err != nil {
			log.Fatal(err)
		}

		extractCooldowns[shipIdx] -= 2
		if extractCooldowns[shipIdx] <= 0 {
			w.Header().Add("HX-Trigger", "extract-end-cooldown")
			extractCooldowns[shipIdx] = 0
		}

		_, err = w.Write([]byte(
			"<p>" + fmt.Sprint(extractCooldowns[shipIdx]) + "s</p>",
		))
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/extract-new-button", func(w http.ResponseWriter, r *http.Request) {
		shipName := r.URL.Query().Get("ship")
		_, err := w.Write([]byte(`<button
            hx-get="/extract"
            hx-swap="outerHTML"
            hx-target="this"
            hx-vals='{"ship":"` + shipName + `"}'>
            Extract</button>`,
		))
		if err != nil {
			log.Fatal(err)
		}
	})

	temSell := template.Must(template.New("sell.gohtml").ParseFiles("gohtml/sell.gohtml"))
	http.HandleFunc("/sell", func(w http.ResponseWriter, r *http.Request) {
		shipName := r.URL.Query().Get("ship")
		// Isolate ship of interest
		shipIdx := -1
		for i, ship := range data.Ships {
			if ship.Symbol == shipName {
				shipIdx = i
				break
			}
		}

		sellInfo := struct {
			Ship    objects.Ship
			ShipIdx int
		}{
			Ship:    data.Ships[shipIdx],
			ShipIdx: shipIdx,
		}

		err := temSell.Execute(w, sellInfo)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/sell-execute", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		var shipName string
		var sellItem string
		var sellAmount int
		requestData := strings.Split(string(body), "&")
		for i := range requestData {
			attrValue := strings.Split(requestData[i], "=")
			if attrValue[0] == "ship" {
				shipName = attrValue[1]
			} else if attrValue[0] == "sell-cargo-type" {
				sellItem = attrValue[1]
			} else if attrValue[0] == "sell-amount" {
				sellAmount, err = strconv.Atoi(attrValue[1])
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		shipIdx := -1
		for i, ship := range data.Ships {
			if ship.Symbol == shipName {
				shipIdx = i
				break
			}
		}

		saleResult := requests.SellCargo(shipName, sellItem, sellAmount, ticker)
		// Check that the sale actually happened and saleResult is not empty
		if saleResult.BuySell.Transaction.TotalPrice != 0 {
			data.Ships[shipIdx].Cargo = saleResult.BuySell.Cargo
			data.Agent.Credits = saleResult.BuySell.Agent.Credits
		}

		// New sell button
		_, err = w.Write([]byte(
			`<td id="td-sell-` + fmt.Sprint(shipIdx) + `">
                <button
                hx-get="/sell"
                hx-swap="outerHTML"
                hx-target="this"
                hx-vals='{"ship":"` + shipName + `"}'>
                Sell</button>
            </td>`,
		))
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/siphon", func(w http.ResponseWriter, r *http.Request) {
		shipName := r.URL.Query().Get("ship")
		if shipName == "" {
			log.Println("siphon error: empty shipName")
			return
		}

		siphonMsg := requests.SiphonGas(shipName, ticker)
		log.Printf("%s siphoned %d %s, cargo %d/%d\n", shipName,
			siphonMsg.ExtractBody.Siphon.Yield.Units,
			siphonMsg.ExtractBody.Siphon.Yield.Item,
			siphonMsg.ExtractBody.Cargo.Units,
			siphonMsg.ExtractBody.Cargo.Capacity,
		)

		shipIdx := -1
		for i, ship := range data.Ships {
			if ship.Symbol == shipName {
				shipIdx = i
				break
			}
		}
		if siphonMsg.ExtractBody.Siphon.Yield.Units != 0 {
			data.Ships[shipIdx].Cargo = siphonMsg.ExtractBody.Cargo
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

// Return slice of system waypoints
func readSystemWaypointsFromFile(system string, ticker *time.Ticker) []objects.Waypoint {
	contents, err := os.ReadFile(fmt.Sprintf("maps/%s.json", system))
	if err != nil {
		composites.StoreSystemWaypoints(system, ticker)
		contents, err = os.ReadFile(fmt.Sprintf("maps/%s.json", system))
		if err != nil {
			log.Fatal(err)
		}
	}
	contents = bytes.Trim(contents, "\n")
	lines := bytes.Split(contents, []byte("\n"))
	waypoints := make([]objects.Waypoint, 0)
	for _, line := range lines {
		var waypoint objects.Waypoint
		json.Unmarshal(line, &waypoint)
		waypoints = append(waypoints, waypoint)
	}

	return waypoints
}
