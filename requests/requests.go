package requests

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
	"time"

	"github.com/j1642/spaceTraders-lib/objects"
)

var auth string = readAuth()
var client *http.Client = &http.Client{}

func makeRequest(httpMethod, url string, msg []byte) *http.Request {
	var request *http.Request
	var err error
	if len(msg) > 0 {
		request, err = http.NewRequest(httpMethod, url, bytes.NewBuffer(msg))
	} else {
		request, err = http.NewRequest(httpMethod, url, nil)
	}
	if err != nil {
		panic(err)
	}
	request.Header.Add("Authorization", auth)
	return request
}

func sendRequest(request *http.Request, ticker *time.Ticker) *http.Response {
	// The response must be closed, whether by readResponse() or other means.
	select {
	case <-ticker.C:
		resp, err := client.Do(request)
		if err != nil {
			panic(err)
		}
		return resp
	}
}

func readResponse(resp *http.Response) *bytes.Buffer {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	json.Indent(&out, body, "", "    ")
	//out.WriteTo(os.Stdout)
	return &out
}

func ViewShipsForSale(waypoint string, ticker *time.Ticker) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		waypoint[:7], "/waypoints/", waypoint, "/shipyard"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	fmt.Println(readResponse(resp))
}

func TransferCargo(fromShip, toShip, material string, amount int, ticker *time.Ticker) *bytes.Buffer {
	jsonPieces := []string{`{"shipSymbol": "`, toShip, `", "tradeSymbol": "`,
		material, `", "units": "`, strconv.Itoa(amount), `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/",
		fromShip, "/transfer"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	fmt.Println(body)
	fmt.Println(fromShip, "transferring", amount, material)
	return body
}

func DeliverMaterial(ship, material, contractId string, ticker *time.Ticker) *bytes.Buffer {
	var amount string
	for _, item := range DescribeShip(ship, ticker).Ship.Cargo.Inventory {
		if item.Symbol == material {
			amount = strconv.Itoa(item.Units)
		}
	}
	jsonStrs := []string{`{"shipSymbol":"`, ship, `",`,
		`"tradeSymbol": "`, material, `",`,
		`"units": "`, amount, `"}`}
	jsonContent := []byte(strings.Join(jsonStrs, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/contracts/",
		contractId, "/deliver"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	fmt.Println(body)

	return body
}

func ViewJumpGate(waypoint string, ticker *time.Ticker) {
	url := strings.Join([]string{"https://api.spacetraders.io/v2/systems/",
		waypoint[:7], "/waypoints/", waypoint, "/jump-gate"}, "")
	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	fmt.Println(body)
}

func FulfillContract(contractId string, ticker *time.Ticker) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/contracts/",
		contractId, "/fulfill"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	fmt.Println(body)
}

func BuyCargo(ship, item string, amount int, ticker *time.Ticker) {
	jsonPieces := []string{"{\n", `"symbol": "`, item, "\",\n", `"units": "`, strconv.Itoa(amount), "\"\n}"}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/purchase"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	fmt.Println(body)
}

func SellCargo(ship, item string, amount int, ticker *time.Ticker) objects.DataBuySell {
	// Set amount to -1 to sell all of the item.
	if amount == -1 {
		inventory := DescribeShip(ship, ticker).Ship.Cargo.Inventory
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

	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	var sale objects.DataBuySell

	err := json.Unmarshal(body.Bytes(), &sale)
	if err != nil {
		panic(err)
	}
	fmt.Println(ship, "sold", amount, item, "at",
		sale.BuySell.Transaction.PricePerUnit, "for",
		sale.BuySell.Transaction.TotalPrice,
		"credits:", sale.BuySell.Agent.Credits,
	)

	return sale
}

func DescribeShip(ship string, ticker *time.Ticker) objects.ShipData {
	// Lists cargo.
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship}
	url := strings.Join(urlPieces, "")

	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	var data objects.ShipData
	err := json.Unmarshal(body.Bytes(), &data)
	if err != nil {
		fmt.Println("body:", string(body.Bytes()))
		log.Fatal(err)
	}
	return data
}

// To view all contracts, enter "" as the ID argument
func ViewContract(id string, ticker *time.Ticker) *bytes.Buffer {
	url := strings.Join([]string{"https://api.spacetraders.io/v2/my/contracts/", id}, "")
	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	fmt.Println(body)
	return body
}

func ViewServerStatus(ticker *time.Ticker) {
	url := "https://api.spacetraders.io/v2/"
	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	fmt.Println(body)
}

func ViewMarket(waypoint string, ticker *time.Ticker) *bytes.Buffer {
	system := strings.Join(strings.Split(waypoint, "-")[:2], "-")
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		system, "/waypoints/", waypoint, "/market"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	fmt.Println(body)
	return body
}

func ConductSurvey(ship string, ticker *time.Ticker) *bytes.Buffer {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/survey"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	fmt.Println(body)
	body.WriteTo(os.Stdout)
	return body
}

func ExtractOre(ship string, ticker *time.Ticker) objects.ExtractionData {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/extract"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)

	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	extractMsg := objects.ExtractionData{}
	err := json.Unmarshal(body.Bytes(), &extractMsg)
	if err != nil {
		panic(err)
	}

	return extractMsg
	// Error code 4000: cooldownConflictError
	// Error code 4236: shipNotInOrbitError
}

func SiphonGas(ship string, ticker *time.Ticker) objects.ExtractionData {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/siphon"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)

	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	siphonMsg := objects.ExtractionData{}
	err := json.Unmarshal(body.Bytes(), &siphonMsg)
	if err != nil {
		panic(err)
	}

	return siphonMsg
}

func Orbit(ship string, ticker *time.Ticker) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/orbit"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	resp := sendRequest(req, ticker)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Println(ship, "orbiting...")
}

func RefuelShip(ship string, ticker *time.Ticker) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/refuel"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	resp := sendRequest(req, ticker)
	fmt.Println(readResponse(resp))

	shipDetails := DescribeShip(ship, ticker)
	fmt.Printf("%v refueling... %v/%v\n", ship,
		shipDetails.Ship.Fuel.Current,
		shipDetails.Ship.Fuel.Capacity)
}

func DockShip(ship string, ticker *time.Ticker) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/dock"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	resp := sendRequest(req, ticker)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Println(ship, "docking...")

	shipDetails := DescribeShip(ship, ticker)
	if shipDetails.Ship.Fuel.Current < shipDetails.Ship.Fuel.Capacity/2 {
		RefuelShip(ship, ticker)
	}
}

func TravelTo(ship, waypoint string, ticker *time.Ticker) *bytes.Buffer {
	jsonPieces := []string{`{"waypointSymbol": "`, waypoint, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/navigate"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	resp := sendRequest(req, ticker)

	return readResponse(resp)
}

func ListMyShips(ticker *time.Ticker) *bytes.Buffer {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/ships", nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	//fmt.Println(body)
	return body
}

func PurchaseShip(shipType, waypoint string, ticker *time.Ticker) {
	jsonPieces := []string{`{"shipType": "`, shipType,
		`", "waypointSymbol": "`, waypoint, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	req := makeRequest("POST", "https://api.spacetraders.io/v2/my/ships", jsonContent)
	req.Header.Set("Content-Type", "application/json")
	resp := sendRequest(req, ticker)
	fmt.Println(readResponse(resp))
}

func ListWaypointsInSystem(system string, ticker *time.Ticker, page int) *bytes.Buffer {
	if page < 1 {
		log.Fatal("page arg must be >= 1")
	}
	pg := strconv.Itoa(page)
	url := strings.Join(
		[]string{"https://api.spacetraders.io/v2/systems/", system, "/waypoints/?limit=20&page=", pg}, "")
	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)
	return body
}

// types: PLANET, MOON, ORBITAL_STATION, ASTEROID_FIELD, ENGINEERED_ASTEROID,
// ASTEROID, ASTEROID_BASE, GAS_GIANT, JUMP_GATE, NEBULA, DEBRIS_FIELD,
// GRAVITY_WELL, ARTIFICIAL_GRAVITY_WELL, FUEL_STATION
func ListWaypointsByType(system string, typ string, ticker *time.Ticker) *bytes.Buffer {
	url := strings.Join(
		[]string{"https://api.spacetraders.io/v2/systems/", system, "/waypoints?type=", typ, "&limit=20"}, "")
	req := makeRequest("GET", url, nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	fmt.Println(body)
	return body
}

func ViewAgent(ticker *time.Ticker) *bytes.Buffer {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/agent", nil)
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	fmt.Println(body)
	return body
}

func AcceptContract(contractId string, ticker *time.Ticker) {
	url := strings.Join([]string{"https://api.spacetraders.io/v2/my/contracts/",
		contractId, "/accept"}, "")
	req := makeRequest("POST", url, nil)
	resp := sendRequest(req, ticker)
	fmt.Println(readResponse(resp))
}

func RegisterNewUser(callSign string, ticker *time.Ticker) *bytes.Buffer {
	jsonPieces := []string{`{"symbol": "`, callSign, `", "faction": "COSMIC"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))
	req := makeRequest("POST", "https://api.spacetraders.io/v2/register", jsonContent)
	req.Header.Set("Content-Type", "application/json")
	// Remove invalidated authorization key from past server reset.
	req.Header.Del("Authorization")
	resp := sendRequest(req, ticker)
	body := readResponse(resp)

	fmt.Println(body)
	return body
}

func readAuth() string {
	key, err := os.ReadFile("secrets.txt")
	if err != nil {
		log.Println("requests.readAuth: secrets.txt does not exist")
	}
	auth := fmt.Sprintf("Bearer %s", key)
	auth = strings.ReplaceAll(auth, "\n", "")
	return auth
}
