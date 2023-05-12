package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
    "strconv"
    //"time"
)

var auth string = generateAuth()
var client *http.Client = &http.Client{}

func main() {
    //listWaypointsInSystem()
    //travelToAsteroidField()
    //listMyShips()
    //orbitLocation("shipName")
    extractOre("shipName")
    //localMarket()
    //viewContract() // Aluminum ore
    //dockShip("shipName")
    //listCargo("shipName")
    //sellCargo("shipName", "SILICON_CRYSTALS", 2)
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

func listCargo(ship string) {
    urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship}
    url := strings.Join(urlPieces, "")
    req := makeRequest("GET", url, nil)
    sendRequest(req)
}

func viewContract() {
    req := makeRequest("GET", "https://api.spacetraders.io/v2/my/contracts/clhjbx6q88h4as60djwb2iju7", nil)
    sendRequest(req) 
}

func localMarket() {
    asteroids := "X1-DF55-17335A"
    urlPieces := []string{"https://api.spacetraders.io/v2/systems/X1-DF55/waypoints/", asteroids, "/market"}
    url := strings.Join(urlPieces, "")
    req := makeRequest("GET", url, nil)
    sendRequest(req)
}

func extractOre(ship string) {
    // Similar to dockShip(), refuelShip()
    urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/extract"}
    url := strings.Join(urlPieces, "")
    req := makeRequest("POST", url, nil)
    sendRequest(req)
}

func orbitLocation(ship string) {
    // Similar to dockShip(), refuelShip()
    urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/orbit"}
    url := strings.Join(urlPieces, "")
    req := makeRequest("POST", url, nil)
    sendRequest(req)
}

func refuelShip(ship string) {
    // Similar to dockShip()
    urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/refuel"}
    url := strings.Join(urlPieces, "")
    req := makeRequest("POST", url, nil)
    sendRequest(req)
}

func dockShip(ship string) {
    // Similar to refuelShip()
    urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/dock"}
    url := strings.Join(urlPieces, "")
    req := makeRequest("POST", url, nil)
    sendRequest(req)
}

func travelToAsteroidField() {
    jsonContent := []byte(
`{
"waypointSymbol": "X1-DF55-17335A"
}`)
    
    req := makeRequest("POST", "https://api.spacetraders.io/v2/my/ships/shipName/navigate", jsonContent)
    req.Header.Set("Content-Type", "application/json")
    sendRequest(req)
}

func listMyShips() {
    req := makeRequest("GET", "https://api.spacetraders.io/v2/my/ships", nil)
    sendRequest(req)
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
	sendRequest(req)
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

func sendRequest(request *http.Request) {
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var out bytes.Buffer
	json.Indent(&out, body, "=", "    ")
	out.WriteTo(os.Stdout)
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
