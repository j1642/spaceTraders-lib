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

// ship prices: ore hound 206k, drone 91k
const barrenMoon string = "X1-ZA40-69371X"   // base metals
const frozenMoon string = "X1-ZA40-11513D"   // precious metals
const volcanicMoon string = "X1-ZA40-97262C" // ammonia ice

const hq string = "X1-ZA40-15970B"
const asteroidField string = "X1-ZA40-99095A"
const shipyard string = "X1-ZA40-68707C"

var miningShips []string = readMiningShipNames()
var auth string = readAuth()
var client *http.Client = &http.Client{}

func main() {
	//listWaypointsInSystem()
	//conductSurvey(miningShips[0])
	gather()
	//listMyShips()
	//dropOffMaterialAndReturn(miningShips[0], "IRON_ORE")
}

func gather() {
	wg := &sync.WaitGroup{}
	for _, ship := range miningShips {
		wg.Add(1)
		go collectAndDeliverMaterial(ship, "IRON_ORE", wg)
		time.Sleep(4 * time.Second)
	}
	wg.Wait()
}

func viewShipsForSale(system, waypoint string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		system, "/waypoints/", shipyard, "/shipyard"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func transferCargo(fromShip, toShip, material string, amount int) {
	jsonPieces := []string{`{"shipSymbol": "`, toShip, `", "tradeSymbol": "`,
		material, `", "units": "`, strconv.Itoa(amount), `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/",
		fromShip, "/transfer"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func collectAndDeliverMaterial(ship, material string, wg *sync.WaitGroup) {
	for i := 0; i < 500; i++ {
		extractOre(ship, 3)
		time.Sleep(1 * time.Second)
		dockShip(ship)
		time.Sleep(1 * time.Second)
		sellCargoBesidesMaterial(ship, material)
		time.Sleep(1 * time.Second)
		orbitLocation(ship)
		time.Sleep(1 * time.Second)

		shipData := describeShip(ship).Ship
		time.Sleep(1 * time.Second)
		cargo := &shipData.Cargo
		if shipData.Frame.Symbol == "FRAME_DRONE" {
			transferCargoFromDrone(ship, cargo)
			time.Sleep(1 * time.Second)
			if cargo.Units < cargo.Capacity {
				continue
			}
			fmt.Println(ship, "waiting to transfer cargo")
			time.Sleep(140 * time.Second)
			continue
		} else if cargo.Units == 60 {
			dropOffMaterialAndReturn(ship, material)
		}
	}
	wg.Done()
}

func transferCargoFromDrone(drone string, droneCargo *objects.Cargo) {
	// TODO: stop if error in response
	transport := describeShip(miningShips[0]).Ship
	time.Sleep(1 * time.Second)
	if transport.Nav.WaypointSymbol != asteroidField {
		time.Sleep(140 * time.Second)
	}
	availableSpace := transport.Cargo.Capacity - transport.Cargo.Units
	for _, item := range droneCargo.Inventory {
		if availableSpace == 0 {
			return
		}
		var amount int
		if item.Units > availableSpace {
			amount = availableSpace
		} else {
			amount = item.Units
		}
		transferCargo(drone, miningShips[0], item.Symbol, amount)
		fmt.Println("transfering", amount, item.Symbol)
		// Bookkeeping instead of making another http request.
		// Lets calling func know to continue or wait to transfer cargo later.
		droneCargo.Units -= amount
		time.Sleep(1 * time.Second)
		if amount == availableSpace {
			return
		}
		availableSpace -= amount
	}
}

func deliverMaterial(ship, material string) {
	var amount string
	for _, item := range describeShip(ship).Ship.Cargo.Inventory {
		if item.Symbol == material {
			amount = strconv.Itoa(item.Units)
		}
	}
	jsonStrs := []string{`{"shipSymbol":"`, ship, `",`,
		`"tradeSymbol": "`, material, `",`,
		`"units": "`, amount, `"}`}
	jsonContent := []byte(strings.Join(jsonStrs, ""))

	url := "https://api.spacetraders.io/v2/my/contracts/clhmm0r8d0of5s60dn7otx0lc/deliver"
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func dropOffMaterialAndReturn(ship, material string) {
	// Go to drop off point
	jsonPieces := []string{`{"waypointSymbol": "`, hq, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/navigate"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))

	// Travel time.
	fmt.Println(ship, "moving to the drop-off")
	time.Sleep(40 * time.Second)
	dockShip(ship)

	// Drop off contract material.
	deliverMaterial(ship, material)
	time.Sleep(1 * time.Second)
	// Sell additional materials.
	cargo := describeShip(ship).Ship.Cargo
	cargoAmounts := make(map[string]int)
	for _, item := range cargo.Inventory {
		cargoAmounts[item.Symbol] = item.Units
	}
	if amount, ok := cargoAmounts["ICE_WATER"]; ok {
		sellCargo(ship, "ICE_WATER", amount)
	}
	orbitLocation(ship)

	sellCargoOnMoons(ship, cargoAmounts)

	// Return to mining location.
	time.Sleep(1 * time.Second)
	travelTo(ship, asteroidField)
	fmt.Println(ship, "returning from the drop-off")
	time.Sleep(40 * time.Second)
}

func sellCargoOnMoons(ship string, cargoAmounts map[string]int) {
	// Sell cargo to markets that generally pay the most.
	cu_amount, cu_ok := cargoAmounts["COPPER_ORE"]
	al_amount, al_ok := cargoAmounts["ALUMINUM_ORE"]
	if cu_ok || al_ok {
		travelTo(ship, barrenMoon)
		time.Sleep(15 * time.Second)
		dockShip(ship)
		time.Sleep(1 * time.Second)
		if cu_ok {
			sellCargo(ship, "COPPER_ORE", cu_amount)
			time.Sleep(1 * time.Second)
		}
		if al_ok {
			sellCargo(ship, "ALUMINUM_ORE", al_amount)
			time.Sleep(1 * time.Second)
		}
		orbitLocation(ship)
	}

	ag_amount, ag_ok := cargoAmounts["SILVER_ORE"]
	au_amount, au_ok := cargoAmounts["GOLD_ORE"]
	pt_amount, pt_ok := cargoAmounts["PLATINUM_ORE"]
	if ag_ok || au_ok || pt_ok {
		travelTo(ship, frozenMoon)
		time.Sleep(15 * time.Second)
		dockShip(ship)
		time.Sleep(1 * time.Second)
		if ag_ok {
			sellCargo(ship, "SILVER_ORE", ag_amount)
			time.Sleep(1 * time.Second)
		}
		if au_ok {
			sellCargo(ship, "GOLD_ORE", au_amount)
			time.Sleep(1 * time.Second)
		}
		if pt_ok {
			sellCargo(ship, "PLATINUM_ORE", pt_amount)
			time.Sleep(1 * time.Second)
		}
		orbitLocation(ship)
	}

	if nh3_amount, ok := cargoAmounts["AMMONIA_ICE"]; ok {
		travelTo(ship, volcanicMoon)
		time.Sleep(15 * time.Second)
		dockShip(ship)
		time.Sleep(1 * time.Second)
		if ok {
			sellCargo(ship, "AMMONIA_ICE", nh3_amount)
			time.Sleep(1 * time.Second)
		}
		orbitLocation(ship)
	}
}

func sellCargoBesidesMaterial(ship, material string) {
	//log.Println("entering sellCargoBesidesMaterial()")
	cargo := describeShip(ship).Ship.Cargo.Inventory
	for i := len(cargo) - 1; i >= 0; i-- {
		item := cargo[i]
		prefix := item.Symbol[0:4]
		if prefix == "QUAR" || prefix == "SILI" || prefix == "DIAM" || prefix == "ICE_" {
			sellCargo(ship, item.Symbol, item.Units)
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(100 * time.Millisecond)
	//log.Println("exiting sellCargoBesidesMaterial()")
}

func buyCargo(ship, item string, amount int) {
	jsonPieces := []string{"{\n", `"symbol": "`, item, "\",\n", `"units": "`, strconv.Itoa(amount), "\"\n}"}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/purchase"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
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

	reply := sendRequest(req)
	var sale objects.DataBuySell
	err := json.Unmarshal(reply.Bytes(), &sale)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ship, "selling...", amount, item, "at",
		sale.BuySell.Transaction.PricePerUnit, "for",
		sale.BuySell.Transaction.TotalPrice,
		"credits:", sale.BuySell.Agent.Credits)

	//log.Println("exiting sellCargo()")
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
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/contracts/clhmm0r8d0of5s60dn7otx0lc", nil)
	fmt.Println(sendRequest(req))
}

func viewMarket(waypoint string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		hq[:7], "/waypoints/", waypoint, "/market"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func conductSurvey(ship string) *bytes.Buffer {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/survey"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	reply := sendRequest(req)
	reply.WriteTo(os.Stdout)
	return reply
}

func extractOre(ship string, repeat int) {
	// Similar to dockShip(), refuelShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/extract"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	for i := 0; i < repeat; i++ {
		cargo := describeShip(ship).Ship.Cargo
		if cargo.Units == cargo.Capacity {
			fmt.Println("cargo full")
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
	//log.Println("exiting dockShip()")
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

func purchaseShip(shipType, waypoint string) {
	jsonPieces := []string{`{"shipType": "`, shipType,
		`", "waypointSymbol": "`, waypoint, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	req := makeRequest("POST", "https://api.spacetraders.io/v2/my/ships", jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func listWaypointsInSystem() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/systems/X1-ZA40/waypoints", nil)
	fmt.Println(sendRequest(req))
}

func viewAgent() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/agent", nil)
	fmt.Println(sendRequest(req))
}

func register(callSign string) {
	jsonPieces := []string{`{"symbol": "`, callSign, `", "faction": "COSMIC"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))
	req := makeRequest("POST", "https://api.spacetraders.io/v2/register", jsonContent)
	req.Header.Set("Content-Type", "application/json")
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

func readAuth() string {
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
