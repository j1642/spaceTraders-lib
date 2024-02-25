package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"spacetraders/objects"
	"spacetraders/requests"
)

const barrenMoon string = "X1-KS52-61262Z"   // base metals market
const volcanicMoon string = "X1-KS52-31553B" // ammonia ice market
const frozenMoon string = ""

const system string = "X1-MR34"
const hq string = "X1-MR34-A1"
const asteroidField string = "X1-KS52-51225B"
const shipyard string = "X1-KS52-23717D"

const engineeredAsteroid = "X1-MR34-DX5X"

var miningShips []string = readMiningShipNames()

func main() {
	//requests.ViewServerStatus()
	ticker := time.NewTicker(1100 * time.Millisecond)
	requests.ListWaypointsByType(system, "PLANET", ticker)
	//gather()
	/*
	   requests.PurchaseShip("SHIP_MINING_DRONE", shipyard)
	   requests.Orbit("USER-6")
	   fmt.Println(requests.TravelTo("USER-6", asteroidField))
	*/
}

func gather(ticker *time.Ticker) {
	// TODO: add channel for survey target if it contains desirable resources.
	wg := &sync.WaitGroup{}
	for _, ship := range miningShips {
		go collectAndDeliverMaterial(ship, "ALUMINUM_ORE", wg, ticker)
	}
	wg.Wait()
	ticker.Stop()
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
	if transport.Nav.WaypointSymbol != asteroidField {
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
	// Go to drop off point
	fmt.Println(ship, "moving to the drop-off")
	trip := requests.TravelTo(ship, barrenMoon, ticker)
	sleepDuringTravel(trip)

	requests.DockShip(ship, ticker)
	requests.FulfillContract("cliqep7yu02vvs60d74wj4eej", ticker)

	// Drop off contract material.
	requests.DeliverMaterial(ship, material, "cliqep7yu02vvs60d74wj4eej", ticker)
	requests.Orbit(ship, ticker)

	// Sell additional materials.
	cargo := requests.DescribeShip(ship, ticker).Ship.Cargo
	cargoAmounts := make(map[string]int)
	for _, item := range cargo.Inventory {
		cargoAmounts[item.Symbol] = item.Units
	}

	sellCargoOnMoons(ship, cargoAmounts, ticker)

	// Return to mining location.
	trip = requests.TravelTo(ship, asteroidField, ticker)
	fmt.Println(ship, "returning from the drop-off")
	sleepDuringTravel(trip)
}

// Sell cargo to markets that generally pay the most. Locations are currently
// found manually and hardcoded as constants
func sellCargoOnMoons(ship string, cargoAmounts map[string]int, ticker *time.Ticker) {
	cu_amount, cu_ok := cargoAmounts["COPPER_ORE"]
	al_amount, al_ok := cargoAmounts["ALUMINUM_ORE"]
	fe_amount, fe_ok := cargoAmounts["IRON_ORE"]
	if cu_ok || al_ok || fe_ok {
		/*trip := requests.TravelTo(ship, barrenMoon)
		        fmt.Println(trip)
				sleepDuringTravel(trip)
		*/
		requests.DockShip(ship, ticker)
		//requests.ViewMarket(barrenMoon)
		if cu_ok {
			requests.SellCargo(ship, "COPPER_ORE", cu_amount, ticker)
		}
		if al_ok {
			requests.SellCargo(ship, "ALUMINUM_ORE", al_amount, ticker)
		}
		if fe_ok {
			requests.SellCargo(ship, "IRON_ORE", fe_amount, ticker)
		}
		requests.Orbit(ship, ticker)
	}

	ag_amount, ag_ok := cargoAmounts["SILVER_ORE"]
	au_amount, au_ok := cargoAmounts["GOLD_ORE"]
	pt_amount, pt_ok := cargoAmounts["PLATINUM_ORE"]
	if ag_ok || au_ok || pt_ok {
		trip := requests.TravelTo(ship, frozenMoon, ticker)
		sleepDuringTravel(trip)
		requests.DockShip(ship, ticker)
		//requests.ViewMarket(frozenMoon)
		if ag_ok {
			select {
			case <-ticker.C:
				requests.SellCargo(ship, "SILVER_ORE", ag_amount, ticker)
			}
		}
		if au_ok {
			requests.SellCargo(ship, "GOLD_ORE", au_amount, ticker)
		}
		if pt_ok {
			requests.SellCargo(ship, "PLATINUM_ORE", pt_amount, ticker)
		}
		requests.Orbit(ship, ticker)
	}

	if nh3_amount, ok := cargoAmounts["AMMONIA_ICE"]; ok {
		trip := requests.TravelTo(ship, volcanicMoon, ticker)
		sleepDuringTravel(trip)
		requests.DockShip(ship, ticker)
		if ok {
			requests.SellCargo(ship, "AMMONIA_ICE", nh3_amount, ticker)
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
		if prefix != "ALUM" {
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
		panic(err)
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
func doNewUserBoilerplate(callsign string, ticker *time.Ticker) {
	reply := requests.RegisterNewUser(callsign, ticker)
	fmt.Println(reply)
	registerMsg := objects.User{}
	err := json.Unmarshal(reply.Bytes(), &registerMsg)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("secrets.txt", []byte(registerMsg.UserData.Token), 664)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}

	// Error b/c auth var in requests is not updated.
	//requests.AcceptContract(registerMsg.UserData.Contract.Id)

	err = os.WriteFile("miningDrones.txt", []byte(callsign+"-1"), 664)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}

	err = os.WriteFile("probes.txt", []byte(callsign+"-2"), 664)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}
}
