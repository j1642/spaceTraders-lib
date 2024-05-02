// Composite functions to run resource extraction, dropoff, and sales; register
// new users; collect info on markets; etc.
package composites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/j1642/spaceTraders-lib/objects"
	"github.com/j1642/spaceTraders-lib/requests"
)

// const fancyImporter string = "X1-BD74-H47" // imports precious stones, gems, Au/Ag BARS (not ore)
const contractDestination string = "X1-BD74-H44" // imports ores: Fe, Al, Cu
const exchangePlace = "X1-BD74-H46"              // moon, exchanges both ices, q sand, Si crystals
const system string = "X1-BD74"

// hq coords were: (16, -20)
const hq string = "X1-BD74-A1"

const shipyard string = "X1-BD74-C35" // seems to always be an orbital station

const engineeredAsteroid string = "X1-BD74-DE5F"
const contractID string = "clt1v2rv100whs60cahif98q1"

var miningShips []string = []string{} //readMiningShipNames()

type Point struct {
	X, Y int
}

// Investigate a directed graph of trade routes between exporters and importers
func build_adj_matrix(locations []Point, markets []objects.Market) {
	if len(locations) != len(markets) {
		panic("lengths should be equal, got={len(locations)}, {len(markets)}")
	}
	imported := make([][]string, len(locations))
	exported := make([][]string, len(locations))

	for i, market := range markets {
		local_imports := make([]string, len(market.Imports))
		for j, imported_good := range market.Imports {
			local_imports[j] = imported_good.Symbol
		}
		imported[i] = local_imports

		local_exports := make([]string, len(market.Exports))
		for j, export := range market.Exports {
			local_exports[j] = export.Symbol
		}
		exported[i] = local_exports
	}

	// Directed graph from an exporter to a matching importer
	adj_matrix := make([][]string, len(markets))
	for i := range markets {
		adj_matrix[i] = make([]string, len(markets))
	}

	for i, market_exports := range exported {
		for _, export_good := range market_exports {
			for k, market_imports := range imported {
				if i == k {
					// Cannot import and export within the same market
					continue
				}
				for _, import_good := range market_imports {
					if export_good == import_good {
						adj_matrix[i][k] = export_good
					}
				}
			}
		}
	}

	for i := range adj_matrix {
		fmt.Println(adj_matrix[i])
	}
}

func investigateMarkets(system string, ticker *time.Ticker) ([]Point, []objects.Market) {
	sites_of_interest := []string{"PLANET", "MOON", "ORBITAL_STATION"}
	//sites_of_interest := []string{"ORBITAL_STATION"}
	real_stdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	locations := make([]Point, 0)

	i := 0
	markets := make([]objects.Market, 0)
	for _, site_type := range sites_of_interest {
		sites := requests.ListWaypointsByType(system, site_type, ticker)

		waypoints := objects.Waypoints{}
		err := json.Unmarshal(sites.Bytes(), &waypoints)
		if err != nil {
			panic(err)
		}

		for _, site := range waypoints.Data {
			for _, trait := range site.Traits {
				if trait["symbol"] == "MARKETPLACE" {
					locations = append(locations, Point{X: site.X, Y: site.Y})
					market := objects.MarketData{}
					marketJSON := requests.ViewMarket(site.Symbol, ticker)
					err := json.Unmarshal(marketJSON.Bytes(), &market)
					if err != nil {
						panic(err)
					}
					markets = append(markets, market.Market)
				}
			}
			i += 1
		}
	}
	os.Stdout = real_stdout
	for i, market := range markets {
		fmt.Println(locations[i], market.Symbol)
	}

	return locations, markets
}

func gather(material string, ticker *time.Ticker) {
	// TODO: add channel for survey target if it contains desirable resources.
	for {
		ship := miningShips[0]
		requests.ExtractOre(ship, 3, ticker)
		shipData := requests.DescribeShip(ship, ticker).Ship
		fuelPercent := float64(shipData.Fuel.Current) / float64(shipData.Fuel.Capacity)
		if fuelPercent < 0.5 {
			requests.RefuelShip(ship, ticker)
		}
		cargo := &shipData.Cargo
		available := cargo.Capacity - cargo.Units
		if available == 0 {
			for _, item := range cargo.Inventory {
				fmt.Println(item.Units, item.Symbol)
			}
			dropOffMaterialAndReturn(ship, material, ticker)
		}
	}
	/*
		wg := &sync.WaitGroup{}
		for _, ship := range miningShips {
			//go collectAndDeliverMaterial(ship, "COPPER_ORE", wg, ticker)
		}
		wg.Wait()
		ticker.Stop()
	*/
}

func collectAndDeliverMaterial(ship, material string, wg *sync.WaitGroup, ticker *time.Ticker) {
	wg.Add(1)
	for i := 0; i < 500; i++ {
		requests.ExtractOre(ship, 3, ticker)
		requests.DockShip(ship, ticker)
		sellCargoBesidesMaterial(ship, material, ticker)
		requests.Orbit(ship, ticker)

		shipData := requests.DescribeShip(ship, ticker).Ship
		cargo := &shipData.Cargo
		available := cargo.Capacity - cargo.Units
		if shipData.Frame.Symbol == "FRAME_DRONE" ||
			shipData.Frame.Symbol == "FRAME_MINER" {
			transferCargoFromDrone(ship, cargo, ticker)
			if cargo.Units < cargo.Capacity {
				continue
			}
			fmt.Println(ship, "waiting to transfer cargo")
			time.Sleep(110 * time.Second)
			continue
		} else if shipData.Frame.Symbol == "FRAME_FRIGATE" && available < 5 {
			dropOffMaterialAndReturn(ship, material, ticker)
		}
	}
	wg.Done()
}

// Transfer cargo from a small, slow ship (usually a MINING_DRONE) to a faster,
// larger transport ship
func transferCargoFromDrone(drone string, droneCargo *objects.Cargo, ticker *time.Ticker) {
	transport := requests.DescribeShip(miningShips[0], ticker).Ship
	if transport.Nav.WaypointSymbol != engineeredAsteroid {
		if float64(droneCargo.Units)/float64(droneCargo.Capacity) < 0.8 {
			return
		}
		fmt.Println(drone, "waiting for transport (whole trip)")
		time.Sleep(110 * time.Second)
		transferCargoFromDrone(drone, droneCargo, ticker)
		return
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

		reply := requests.TransferCargo(drone, miningShips[0], item.Symbol, amount, ticker)
		transferMsg := objects.Error{}
		err := json.Unmarshal(reply.Bytes(), &transferMsg)
		if err != nil {
			panic(err)
		}
		// Transport ship approaching but has not arrived.
		if transferMsg.ErrBody.Code == 4214 {
			fmt.Println(drone, "waiting for transport (return trip)")
			time.Sleep(45 * time.Second)
			requests.TransferCargo(drone, miningShips[0], item.Symbol, amount, ticker)
		}

		// Bookkeeping instead of making another HTTP request.
		// Lets the calling func know to continue mining or wait to transfer cargo
		droneCargo.Units -= amount
		if amount == availableSpace {
			return
		}
		availableSpace -= amount
	}
}

// Travel, deliver contract deliverables, sell assorted resources, and return
func dropOffMaterialAndReturn(ship, material string, ticker *time.Ticker) {
	cargo := requests.DescribeShip(ship, ticker).Ship.Cargo
	cargoAmounts := make(map[string]int)
	for _, item := range cargo.Inventory {
		cargoAmounts[item.Symbol] = item.Units
	}
	if cargoAmounts[material] == 0 && cargoAmounts["IRON_ORE"] == 0 &&
		cargoAmounts["ALUMINUM_ORE"] == 0 {
		sellCargoOnMoons(ship, cargoAmounts, ticker)
		trip := requests.TravelTo(ship, engineeredAsteroid, ticker)
		fmt.Println(ship, "returning from exchange place")
		sleepDuringTravel(trip)
		return
	}

	// Go to drop off point
	fmt.Println(ship, "moving to the drop-off")
	trip := requests.TravelTo(ship, contractDestination, ticker)
	sleepDuringTravel(trip)

	// Drop off contract material.
	requests.DockShip(ship, ticker)
	/*
			if cargoAmounts[material] > 0 {
		        delivery := requests.DeliverMaterial(ship, material, contractID, ticker)
		        // TODO: this is probably broken
		        error := objects.Error{}
		        err := json.Unmarshal(delivery.Bytes(), &error)
		        if err != nil {
		            panic("")
		        }
		        if error.ErrBody.Code == 4509 {
		            // Error 4509: Contract terms met, cannot deliver more deliverables
		            requests.FulfillContract(contractID, ticker)
		        }
			}
	*/

	fe_ore_amount, fe_ok := cargoAmounts["IRON_ORE"]
	al_ore_amount, al_ok := cargoAmounts["ALUMINUM_ORE"]
	cu_ore_amount, cu_ok := cargoAmounts["COPPER_ORE"]
	if fe_ok || al_ok || cu_ok {
		if fe_ok {
			requests.SellCargo(ship, "IRON_ORE", fe_ore_amount, ticker)
		}
		if al_ok {
			requests.SellCargo(ship, "ALUMINUM_ORE", al_ore_amount, ticker)
		}
		if cu_ok {
			requests.SellCargo(ship, "COPPER_ORE", cu_ore_amount, ticker)
		}
	}

	requests.Orbit(ship, ticker)
	// Sell additional materials.
	sellCargoOnMoons(ship, cargoAmounts, ticker)

	// Return to mining location.
	trip = requests.TravelTo(ship, engineeredAsteroid, ticker)
	fmt.Println(ship, "returning from the drop-off")
	sleepDuringTravel(trip)
}

// Sell cargo to markets that generally pay the most. Locations are currently
// found manually and hardcoded as constants
func sellCargoOnMoons(ship string, cargoAmounts map[string]int, ticker *time.Ticker) {
	nh3_amount, nh3_ok := cargoAmounts["AMMONIA_ICE"]
	h2o_amount, h2o_ok := cargoAmounts["ICE_WATER"]
	si_amount, si_ok := cargoAmounts["SILICON_CRYSTALS"]
	sio2_amount, sio2_ok := cargoAmounts["QUARTZ_SAND"]

	if nh3_ok || h2o_ok || si_ok || sio2_ok {
		trip := requests.TravelTo(ship, exchangePlace, ticker)
		fmt.Println(ship, "travelling to exchange place")
		sleepDuringTravel(trip)
		requests.DockShip(ship, ticker)

		if nh3_ok {
			requests.SellCargo(ship, "AMMONIA_ICE", nh3_amount, ticker)
		}
		if h2o_ok {
			requests.SellCargo(ship, "ICE_WATER", h2o_amount, ticker)
		}
		if si_ok {
			requests.SellCargo(ship, "SILICON_CRYSTALS", si_amount, ticker)
		}
		if sio2_ok {
			requests.SellCargo(ship, "QUARTZ_SAND", sio2_amount, ticker)
		}
		requests.Orbit(ship, ticker)
	}
}

// Sell all cargo in the ship's inventory, besides the hardcoded exceptions
func sellCargoBesidesMaterial(ship, material string, ticker *time.Ticker) {
	cargo := requests.DescribeShip(ship, ticker).Ship.Cargo.Inventory
	for i := len(cargo) - 1; i >= 0; i-- {
		item := cargo[i]
		prefix := item.Symbol[0:4]
		if prefix != "xxxx" {
			requests.SellCargo(ship, item.Symbol, item.Units, ticker)
		}
	}
}

// Tell a goroutine to sleep while travelling between locations
func sleepDuringTravel(reply *bytes.Buffer) {
	travelMsg := objects.TravelData{}
	err := json.Unmarshal(reply.Bytes(), &travelMsg)
	if err != nil {
		panic(err)
	}

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

	time.Sleep(end.Sub(start))
}

// Read ships names from file so they can be assigned to their respective goroutines
func readMiningShipNames() []string {
	names, err := os.ReadFile("miningDrones.txt")
	if err != nil {
		panic(err)
	}
	str := strings.TrimSpace(string(names))
	split := strings.Split(str, "\n")
	for _, name := range split {
		split := string(name)
		split = strings.ReplaceAll(split, "\n", "")
	}
	return split
}

// Creates a new user, saves the auth key, and refreshes the ship names files.
func DoNewUserBoilerplate(callsign string, ticker *time.Ticker) error {
	reply := requests.RegisterNewUser(callsign, ticker)
	fmt.Println(reply)
	registerMsg := objects.User{}
	err := json.Unmarshal(reply.Bytes(), &registerMsg)
	if err != nil {
		panic(err)
	}
	// Check if registration worked
	if registerMsg.ErrBody.Code >= 400 {
		return fmt.Errorf("error: invalid agent, submit a different one")
	}

	err = os.WriteFile("secrets.txt", []byte(registerMsg.UserData.Token), 0600)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}

	// Error b/c auth var in requests is not updated.
	//requests.AcceptContract(registerMsg.UserData.Contract.Id)

	err = os.WriteFile("miningDrones.txt", []byte(callsign+"-1"), 0664)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}
	err = os.WriteFile("probes.txt", []byte(callsign+"-2"), 0664)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}
	return nil
}

// Create or append to a cache of system waypoints, with one JSON waypoint per line
func StoreSystemWaypoints(system string, ticker *time.Ticker) {
	systemFile := fmt.Sprintf("maps/%s.json", system)
	f, err := os.OpenFile(systemFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	contents := make([]byte, 1000000)
	bytesRead, err := f.Read(contents)
	if err != nil {
		// EOF error b/c file is empty. Not a problem
	}
	lines := bytes.Split(contents, []byte("\n"))

	requestPage := 1
	moreRequestsNeeded := true
	var waypoints objects.Waypoints

	for moreRequestsNeeded {
		resp := requests.ListWaypointsInSystem(system, ticker, requestPage)
		json.Unmarshal(resp.Bytes(), &waypoints)

		if waypoints.Meta["page"]*waypoints.Meta["limit"] >= waypoints.Meta["total"] {
			moreRequestsNeeded = false
		}
		requestPage += 1

		for i, place := range waypoints.Data {
			// Strip superfluous fields
			for j := range place.Traits {
				// delete() is idempotent, supposedly
				delete(waypoints.Data[i].Traits[j], "description")
				delete(waypoints.Data[i].Traits[j], "name")
			}

			serialWaypoint, err := json.Marshal(place)
			if err != nil {
				panic(err)
			}
			// Prevent insertion of duplicate waypoints
			// Assumes the API response returns distinct waypoints
			if bytesRead > 0 {
				isDuplicate := false
				for _, line := range lines {
					if bytes.Equal(line, serialWaypoint) {
						isDuplicate = true
						fmt.Println("duplicate:", place.Symbol)
						break
					}
				}
				if isDuplicate {
					continue
				}
			}

			// Write a line-separated unique waypoint
			serialWaypoint = append(serialWaypoint, byte('\n'))
			_, err = f.Write(serialWaypoint)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("wrote to file:", place.Symbol)
		}
	}
}
