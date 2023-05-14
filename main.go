package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"spacetraders/objects"
)

const barrenMoon string = "X1-ZA40-69371X"
const frozenMoon string = "X1-ZA40-11513D"
const volcanicMoon string = "X1-ZA40-97262C"

const hq string = "X1-ZA40-15970B"
const asteroidField string = "X1-ZA40-99095A"
const shipyard string = "X1-ZA40-68707C"

var miningShips []string = readMiningShipNames()
var auth string = generateAuth()
var client *http.Client = &http.Client{}

func main() {
	//listWaypointsInSystem()
	//viewMarket(shipyard)
	//viewContract() // Iron ore
	//fmt.Println(describeShip(miningShip2).Ship.Fuel)
	//viewShipsForSale("X1-ZA40")
	//viewAgent()
	gather()
	//deliverMaterial(miningShips[1])
	//listMyShips()
	//dockShip(miningShips[1])
	//refuelShip(miningShips[0])
	//sellCargo(miningShips[2], "COPPER_ORE", -1)
	//orbitLocation(miningShips[1])
	//travelTo(miningShips[0], asteroidField)
	//dropOffMaterialAndReturn(miningShips[1])
}

func viewAgent() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/agent", nil)
	fmt.Println(sendRequest(req))
}

func gather() {
	wg := &sync.WaitGroup{}
	for _, ship := range miningShips {
		wg.Add(1)
		go collectAndDeliverAl(ship, wg)
		time.Sleep(3 * time.Second)
	}
	wg.Wait()
}

func viewShipsForSale(system string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		system, "/waypoints/", spaceport, "/shipyard"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func collectAndDeliverAl(ship string, wg *sync.WaitGroup) {
	for i := 0; i < 100; i++ {
		extractOre(ship, 7)
		time.Sleep(250 * time.Millisecond)
		dockShip(ship)
		time.Sleep(250 * time.Millisecond)
		sellCargoBesidesAl(ship)
		time.Sleep(250 * time.Millisecond)
		orbitLocation(ship)
		time.Sleep(250 * time.Millisecond)
		cargo := describeShip(ship).Ship.Cargo
		if float64(cargo.Units)/float64(cargo.Capacity) > 0.9 {
			dropOffAlAndReturn(ship)
		}
	}
	wg.Done()
}

func deliverAlum(ship string) {
	// Drop off Al ore.
	var amount string
	for _, material := range describeShip(ship).Ship.Cargo.Inventory {
		if material.Symbol == "ALUMINUM_ORE" {
			amount = strconv.Itoa(material.Units)
		}
	}
	//fmt.Println(describeShip(miningShips[1]).Ship.Cargo.Inventory[0].Symbol)
	//amount = strconv.Itoa(describeShip(ship).Ship.Cargo.Units)
	jsonStrs := []string{`{"shipSymbol":"`, ship, `",`,
		`"tradeSymbol": "ALUMINUM_ORE",
"units": "`, amount, `"}`}
	jsonContent := []byte(strings.Join(jsonStrs, ""))

	url := "https://api.spacetraders.io/v2/my/contracts/clhjbx6q88h4as60djwb2iju7/deliver"
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func dropOffAlAndReturn(ship string) {
	// Go to drop off point
	jsonContent := []byte(
		`{
"waypointSymbol": "X1-DF55-20250Z"
}`)

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/navigate"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))

	// Travel time.
	fmt.Println(ship, "moving to the drop-off")
	time.Sleep(130 * time.Second)
	dockShip(ship)

	// Drop off Al
	deliverAlum(ship)

	// Return to mining location.
	orbitLocation(ship)
	travelTo(ship, asteroidField)
	fmt.Println(ship, "returning from the drop-off")
	//fmt.Println(ship, "standing by at the drop-off")
	time.Sleep(130 * time.Second)
}

// TODO: rename
func sellCargoBesidesAl(ship string) {
	log.Println("entering sellCargoBesidesAl()")
	cargo := describeShip(ship).Ship.Cargo.Inventory
	for i := len(cargo) - 1; i >= 0; i-- {
		item := cargo[i]
		prefix := item.Symbol[0:4]
		if prefix != "ALUM" && prefix != "ANTI" { //&& prefix != "IRON" && prefix != "COPP" {
			sellCargo(ship, item.Symbol, item.Units)
			fmt.Println(ship, "selling", item.Symbol)
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(100 * time.Millisecond)
	log.Println("exiting sellCargoBesidesAl()")
}

func sellCargo(ship, item string, amount int) {
	// Set amount to -1 to sell all of the item.
	if amount == -1 {
		inventory := describeShip(ship).Ship.Cargo.Inventory
		for _, material := range inventory {
			if material.Symbol == item {
				amount = material.Units
			}
		}
	}

	jsonPieces := []string{"{\n", `"symbol": "`, item, "\",\n", `"units": "`, strconv.Itoa(amount), "\"\n}"}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/sell"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
	log.Println("exiting sellCargo()")
}

func describeShip(ship string) objects.DataShip {
	// Lists cargo.
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	out := sendRequest(req)
	//out.WriteTo(os.Stdout)
	var data objects.DataShip
	err := json.Unmarshal(out.Bytes(), &data)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func viewContract() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/contracts/clhjbx6q88h4as60djwb2iju7", nil)
	fmt.Println(sendRequest(req))
}

func viewMarket(waypoint string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/X1-DF55/waypoints/", waypoint, "/market"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func extractOre(ship string, repeat int) {
	// Similar to dockShip(), refuelShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/extract"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	for i := 0; i < repeat; i++ {
		cargo := describeShip(ship).Ship.Cargo
		if cargo.Units > cargo.Capacity-4 {
			fmt.Println("cargo full(ish)")
			break
		}
		sendRequest(req)
		fmt.Println(ship, "extracting...", "cargo", cargo.Units, "/", cargo.Capacity)
		if i != (repeat - 1) {
			time.Sleep(71 * time.Second)
		}
	}
}

func orbitLocation(ship string) {
	// Similar to dockShip(), refuelShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/orbit"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	sendRequest(req)
	fmt.Println(ship, "orbiting...")
}

func refuelShip(ship string) {
	// Similar to dockShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/refuel"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	fmt.Println(sendRequest(req))
	shipDetails := describeShip(ship)
	fmt.Printf("%v refueling... %v/%v\n", ship,
		shipDetails.Ship.Fuel.Current,
		shipDetails.Ship.Fuel.Capacity)
}

func dockShip(ship string) {
	// Similar to refuelShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/dock"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	sendRequest(req)
	fmt.Println(ship, "docking...")
	time.Sleep(100 * time.Millisecond)
	shipDetails := describeShip(ship)
	if shipDetails.Ship.Fuel.Current < shipDetails.Ship.Fuel.Capacity/2 {
		refuelShip(ship)
	}
	log.Println("exiting dockShip()")
}

func travelTo(ship, waypoint string) {
	jsonPieces := []string{`{"waypointSymbol": "`, waypoint, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/navigate"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func listMyShips() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/ships", nil)
	fmt.Println(sendRequest(req))
}

func purchaseShip() {
	jsonContent := []byte(
		`{
"shipType": "SHIP_MINING_DRONE",
"waypointSymbol": "X1-DF55-69207D"
}`)

	req := makeRequest("POST", "https://api.spacetraders.io/v2/my/ships", jsonContent)
	req.Header.Set("Content-Type", "application/json")
	sendRequest(req)
}

func listShipsAvailable() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/systems/X1-DF55/waypoints/X1-DF55-69207D/shipyard", nil)
	sendRequest(req)
}

func listWaypointsInSystem() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/systems/X1-DF55/waypoints", nil)
	fmt.Println(sendRequest(req))
}

func makeRequest(httpMethod, url string, msg []byte) *http.Request {
	var request *http.Request
	var err error
	if len(msg) > 0 {
		request, err = http.NewRequest(httpMethod, url, bytes.NewBuffer(msg))
	} else {
		request, err = http.NewRequest(httpMethod, url, nil)
	}
	if err != nil {
		log.Fatalf("%v", err)
	}
	request.Header.Add("Authorization", auth)
	return request
}

func sendRequest(request *http.Request) *bytes.Buffer {
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var out bytes.Buffer
	json.Indent(&out, body, "", "    ")
	//out.WriteTo(os.Stdout)
	return &out
}

func generateAuth() string {
	key, err := os.ReadFile("secrets.txt")
	if err != nil {
		log.Fatal(err)
	}
	auth := fmt.Sprintf("Bearer %s", key)
	auth = strings.ReplaceAll(auth, "\n", "")
	return auth
}

func readMiningShipNames() []string {
	names, err := os.ReadFile("miningDrones.txt")
	if err != nil {
		log.Fatal(err)
	}
	str := strings.TrimSpace(string(names))
	split := strings.Split(str, "\n")
	for _, name := range split {
		split := string(name)
		split = strings.ReplaceAll(split, "\n", "")
	}
	return split
}
