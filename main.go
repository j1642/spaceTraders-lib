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

const spaceport string = "X1-DF55-69207D"

var miningShips []string = readMiningShipNames()
var auth string = generateAuth()
var client *http.Client = &http.Client{}

func main() {
	//listWaypointsInSystem()
	//travelToAsteroidField()
	//listMyShips()
	//orbitLocation(miningShip2)
	//localMarket()
	//viewContract() // Aluminum ore
	//fmt.Println(describeShip(miningShip2).Ship.Fuel)
	//sellCargo(miningShip2, "ICE_WATER", 7)
	//viewShipsForSale("X1-DF55")
	gather()
	//purchaseShip()
}

func gather() {
	wg := &sync.WaitGroup{}
	for _, ship := range miningShips {
		wg.Add(1)
		go collectAndDeliverAl(ship, wg)
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
		dockShip(ship)
		sellCargoBesidesAl(ship)
		orbitLocation(ship)
		if describeShip(ship).Ship.Cargo.Units > 25 {
			dropOffAlAndReturn(ship)
		}
	}
	wg.Done()
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
	time.Sleep(180 * time.Second)
	dockShip(ship)

	// Drop off Al ore.
	jsonStrs := []string{`{"shipSymbol":"`, ship, `",`,
		`"tradeSymbol": "ALUMINUM_ORE",
"units": "28"}`}
	jsonContent = []byte(strings.Join(jsonStrs, ""))

	url = "https://api.spacetraders.io/v2/my/contracts/clhjbx6q88h4as60djwb2iju7/deliver"
	req = makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))

	// Return to mining location.
	orbitLocation(ship)
	travelToAsteroidField(ship)
	time.Sleep(180 * time.Second)
}

func sellCargoBesidesAl(ship string) {
	for {
		cargo := describeShip(ship).Ship.Cargo.Inventory
		if len(cargo) == 1 {
			fmt.Println(cargo)
			break
		}
		for _, item := range cargo {
			if item.Symbol[0:2] != "AL" {
				fmt.Println(ship, "selling", item.Symbol)
				sellCargo(ship, item.Symbol, item.Units)
			}
		}
	}
}

func sellCargo(ship, item string, amount int) {
	jsonPieces := []string{"{\n", `"symbol": "`, item, "\",\n", `"units": "`, strconv.Itoa(amount), "\"\n}"}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/sell"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	sendRequest(req)
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

func localMarket() {
	asteroids := "X1-DF55-17335A"
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/X1-DF55/waypoints/", asteroids, "/market"}
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
		if cargo.Units == cargo.Capacity {
			fmt.Println("cargo full")
			break
		}
		sendRequest(req)
		fmt.Println(ship, "extraction #", i, " cargo", cargo.Units, "/", cargo.Capacity)
		if i == (repeat - 1) {
			break
		}
		time.Sleep(71 * time.Second)
	}
}

func orbitLocation(ship string) {
	// Similar to dockShip(), refuelShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/orbit"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	sendRequest(req)
	fmt.Println("Orbiting...")
}

func refuelShip(ship string) {
	// Similar to dockShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/refuel"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	fmt.Println(sendRequest(req))
	shipDetails := describeShip(ship)
	fmt.Printf("Refueling... %v/%v\n", shipDetails.Ship.Fuel.Current,
		shipDetails.Ship.Fuel.Capacity)
}

func dockShip(ship string) {
	// Similar to refuelShip()
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/dock"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	sendRequest(req)
	fmt.Println("Docking...")
	shipDetails := describeShip(ship)
	if shipDetails.Ship.Fuel.Current < shipDetails.Ship.Fuel.Capacity {
		refuelShip(ship)
	}
}

func travelToAsteroidField(ship string) {
	jsonContent := []byte(
		`{
"waypointSymbol": "X1-DF55-17335A"
}`)

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
