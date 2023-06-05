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

	"spacetraders/objects"
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

func ViewShipsForSale(system, waypoint string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		system, "/waypoints/", waypoint, "/shipyard"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func TransferCargo(fromShip, toShip, material string, amount int) *bytes.Buffer {
	jsonPieces := []string{`{"shipSymbol": "`, toShip, `", "tradeSymbol": "`,
		material, `", "units": "`, strconv.Itoa(amount), `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/",
		fromShip, "/transfer"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	reply := sendRequest(req)
	fmt.Println(fromShip, "transferring", amount, material)
	return reply
}

func DeliverMaterial(ship, material, contractId string) *bytes.Buffer {
	var amount string
	for _, item := range DescribeShip(ship).Ship.Cargo.Inventory {
		if item.Symbol == material {
			amount = strconv.Itoa(item.Units)
		}
	}
	time.Sleep(1 * time.Second)
	jsonStrs := []string{`{"shipSymbol":"`, ship, `",`,
		`"tradeSymbol": "`, material, `",`,
		`"units": "`, amount, `"}`}
	jsonContent := []byte(strings.Join(jsonStrs, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/contracts/",
		contractId, "/deliver"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	reply := sendRequest(req)
	fmt.Println(reply)

	return reply
}

func ViewJumpGate(waypoint string) {
	url := strings.Join([]string{"https://api.spacetraders.io/v2/systems/",
		waypoint[:7], "/waypoints/", waypoint, "/jump-gate"}, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func ReceiveContractPayment(contractId string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/contracts/",
		contractId, "/fulfill"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	fmt.Println(sendRequest(req))
}

func BuyCargo(ship, item string, amount int) {
	jsonPieces := []string{"{\n", `"symbol": "`, item, "\",\n", `"units": "`, strconv.Itoa(amount), "\"\n}"}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/purchase"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func SellCargo(ship, item string, amount int) {
	// Set amount to -1 to sell all of the item.
	if amount == -1 {
		inventory := DescribeShip(ship).Ship.Cargo.Inventory
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
}

func DescribeShip(ship string) objects.DataShip {
	// Lists cargo.
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship}
	url := strings.Join(urlPieces, "")

	req := makeRequest("GET", url, nil)
	out := sendRequest(req)
	var data objects.DataShip
	err := json.Unmarshal(out.Bytes(), &data)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func ViewContract(id string) *bytes.Buffer {
	url := strings.Join([]string{"https://api.spacetraders.io/v2/my/contracts/", id}, "")
	req := makeRequest("GET", url, nil)
	reply := sendRequest(req)
	fmt.Println(reply)
	return reply
}

func ViewServerStatus() {
	url := "https://api.spacetraders.io/v2/"
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func ViewMarket(waypoint string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/systems/",
		waypoint[:7], "/waypoints/", waypoint, "/market"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func ConductSurvey(ship string) *bytes.Buffer {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/survey"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	reply := sendRequest(req)
	reply.WriteTo(os.Stdout)
	return reply
}

func ExtractOre(ship string, repeat int) {
	//jsonContent := []byte(`{"survey": "X1-ZA40-99095A-172ABE"}`)

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/extract"}
	url := strings.Join(urlPieces, "")
	req := makeRequest("POST", url, nil)
	//req.Header.Set("Content-Type", "application/json")

	for i := 0; i < repeat; i++ {
		shipData := DescribeShip(ship).Ship
		cargo := &shipData.Cargo
		if cargo.Units == cargo.Capacity {
			fmt.Println(ship, "cargo full")
			return
		}

		reply := sendRequest(req)
		extractMsg := objects.ExtractionData{}
		err := json.Unmarshal(reply.Bytes(), &extractMsg)
		if err != nil {
			log.Fatal(err)
		}

		cargo.Units += extractMsg.ExtractBody.Extraction.Yield.Units
		fmt.Println(ship, "extracting...", "cargo", cargo.Units, "/", cargo.Capacity)
		switch shipData.Frame.Symbol {
		case "FRAME_FRIGATE":
			time.Sleep(70 * time.Second)
		case "FRAME_DRONE":
			time.Sleep(70 * time.Second)
		case "FRAME_MINER":
			time.Sleep(80 * time.Second)
		}
	}
}

func Orbit(ship string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/orbit"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	sendRequest(req)
	fmt.Println(ship, "orbiting...")
}

func RefuelShip(ship string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/refuel"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	fmt.Println(sendRequest(req))
	time.Sleep(1 * time.Second)

	shipDetails := DescribeShip(ship)
	fmt.Printf("%v refueling... %v/%v\n", ship,
		shipDetails.Ship.Fuel.Current,
		shipDetails.Ship.Fuel.Capacity)
}

func DockShip(ship string) {
	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/dock"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, nil)
	sendRequest(req)
	fmt.Println(ship, "docking...")
	time.Sleep(1 * time.Second)

	shipDetails := DescribeShip(ship)
	if shipDetails.Ship.Fuel.Current < shipDetails.Ship.Fuel.Capacity/2 {
		RefuelShip(ship)
	}
}

func TravelTo(ship, waypoint string) *bytes.Buffer {
	jsonPieces := []string{`{"waypointSymbol": "`, waypoint, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	urlPieces := []string{"https://api.spacetraders.io/v2/my/ships/", ship, "/navigate"}
	url := strings.Join(urlPieces, "")

	req := makeRequest("POST", url, jsonContent)
	req.Header.Set("Content-Type", "application/json")
	reply := sendRequest(req)

	return reply
}

func ListMyShips() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/ships", nil)
	fmt.Println(sendRequest(req))
}

func PurchaseShip(shipType, waypoint string) {
	jsonPieces := []string{`{"shipType": "`, shipType,
		`", "waypointSymbol": "`, waypoint, `"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))

	req := makeRequest("POST", "https://api.spacetraders.io/v2/my/ships", jsonContent)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(sendRequest(req))
}

func ListWaypointsInSystem(system string) {
	url := strings.Join(
		[]string{"https://api.spacetraders.io/v2/systems/", system, "/waypoints"}, "")
	req := makeRequest("GET", url, nil)
	fmt.Println(sendRequest(req))
}

func ViewAgent() {
	req := makeRequest("GET", "https://api.spacetraders.io/v2/my/agent", nil)
	fmt.Println(sendRequest(req))
}

func AcceptContract(contractId string) {
	url := strings.Join([]string{"https://api.spacetraders.io/v2/my/contracts/",
		contractId, "/accept"}, "")
	req := makeRequest("POST", url, nil)
	fmt.Println(sendRequest(req))
}

func RegisterNewUser(callSign string) *bytes.Buffer {
	jsonPieces := []string{`{"symbol": "`, callSign, `", "faction": "COSMIC"}`}
	jsonContent := []byte(strings.Join(jsonPieces, ""))
	req := makeRequest("POST", "https://api.spacetraders.io/v2/register", jsonContent)
	req.Header.Set("Content-Type", "application/json")
	// Remove invalidated authorization key from past server reset.
	req.Header.Del("Authorization")
	reply := sendRequest(req)
	fmt.Println(reply)

	return reply
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
