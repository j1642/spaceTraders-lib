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

const hq string = "X1-KS52-07960X"
const asteroidField string = "X1-KS52-51225B"
const shipyard string = "X1-KS52-23717D"

var miningShips []string = readMiningShipNames()

func main() {
	gather()
	/*
	   requests.PurchaseShip("SHIP_MINING_DRONE", shipyard)
	   time.Sleep(1 * time.Second)
	   requests.Orbit("BAP-6")
	   time.Sleep(1 * time.Second)
	   fmt.Println(requests.TravelTo("BAP-6", asteroidField))
	*/
}

func gather() {
	// TODO: add channel for survey target if it contains desirable resources.
	wg := &sync.WaitGroup{}
	for _, ship := range miningShips {
		wg.Add(1)
		go collectAndDeliverMaterial(ship, "ALUMINUM_ORE", wg)
		time.Sleep(10 * time.Second)
	}
	wg.Wait()
}

func collectAndDeliverMaterial(ship, material string, wg *sync.WaitGroup) {
	for i := 0; i < 500; i++ {
		requests.ExtractOre(ship, 3)
		time.Sleep(1 * time.Second)
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		sellCargoBesidesMaterial(ship, material)
		time.Sleep(1 * time.Second)
		requests.Orbit(ship)
		time.Sleep(1 * time.Second)

		shipData := requests.DescribeShip(ship).Ship
		time.Sleep(1 * time.Second)
		cargo := &shipData.Cargo
		available := cargo.Capacity - cargo.Units
		if shipData.Frame.Symbol == "FRAME_DRONE" ||
			shipData.Frame.Symbol == "FRAME_MINER" {
			transferCargoFromDrone(ship, cargo)
			time.Sleep(1 * time.Second)
			if cargo.Units < cargo.Capacity {
				continue
			}
			fmt.Println(ship, "waiting to transfer cargo")
			time.Sleep(110 * time.Second)
			continue
		} else if shipData.Frame.Symbol == "FRAME_FRIGATE" && available < 5 {
			dropOffMaterialAndReturn(ship, material)
		}
	}
	wg.Done()
}

func transferCargoFromDrone(drone string, droneCargo *objects.Cargo) {
	transport := requests.DescribeShip(miningShips[0]).Ship
	time.Sleep(1 * time.Second)
	if transport.Nav.WaypointSymbol != asteroidField {
		if float64(droneCargo.Units)/float64(droneCargo.Capacity) < 0.8 {
			return
		}
		fmt.Println(drone, "waiting for transport (whole trip)")
		time.Sleep(110 * time.Second)
		transferCargoFromDrone(drone, droneCargo)
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

		reply := requests.TransferCargo(drone, miningShips[0], item.Symbol, amount)
		transferMsg := objects.Error{}
		err := json.Unmarshal(reply.Bytes(), &transferMsg)
		if err != nil {
			panic(err)
		}
		// Transport ship approaching but has not arrived.
		if transferMsg.ErrBody.Code == 4214 {
			fmt.Println(drone, "waiting for transport (return trip)")
			time.Sleep(45 * time.Second)
			requests.TransferCargo(drone, miningShips[0], item.Symbol, amount)
		}

		// Bookkeeping instead of making another HTTP request.
		// Lets calling func know to continue or wait to transfer cargo later.
		droneCargo.Units -= amount
		time.Sleep(1 * time.Second)
		if amount == availableSpace {
			return
		}
		availableSpace -= amount
	}
}

func dropOffMaterialAndReturn(ship, material string) {
	// Go to drop off point
	fmt.Println(ship, "moving to the drop-off")
	trip := requests.TravelTo(ship, barrenMoon)
	sleepDuringTravel(trip)
	requests.DockShip(ship)
	time.Sleep(1 * time.Second)

	// Drop off contract material.
	requests.DeliverMaterial(ship, material, "cliqep7yu02vvs60d74wj4eej")
	time.Sleep(1 * time.Second)
	requests.Orbit(ship)
	time.Sleep(1 * time.Second)

	// Sell additional materials.
	cargo := requests.DescribeShip(ship).Ship.Cargo
	time.Sleep(1 * time.Second)
	cargoAmounts := make(map[string]int)
	for _, item := range cargo.Inventory {
		cargoAmounts[item.Symbol] = item.Units
	}

	sellCargoOnMoons(ship, cargoAmounts)

	// Return to mining location.
	time.Sleep(1 * time.Second)
	fmt.Println(ship, "returning from the drop-off")
	trip = requests.TravelTo(ship, asteroidField)
	sleepDuringTravel(trip)
}

func sellCargoOnMoons(ship string, cargoAmounts map[string]int) {
	// Sell cargo to markets that generally pay the most.
	cu_amount, cu_ok := cargoAmounts["COPPER_ORE"]
	al_amount, al_ok := cargoAmounts["ALUMINUM_ORE"]
	fe_amount, fe_ok := cargoAmounts["IRON_ORE"]
	if cu_ok || al_ok || fe_ok {
		/*trip := requests.TravelTo(ship, barrenMoon)
		        fmt.Println(trip)
				sleepDuringTravel(trip)
		*/
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		//requests.ViewMarket(barrenMoon)
		//time.Sleep(1 * time.Second)
		if cu_ok {
			requests.SellCargo(ship, "COPPER_ORE", cu_amount)
			time.Sleep(1 * time.Second)
		}
		if al_ok {
			requests.SellCargo(ship, "ALUMINUM_ORE", al_amount)
			time.Sleep(1 * time.Second)
		}
		if fe_ok {
			requests.SellCargo(ship, "IRON_ORE", fe_amount)
			time.Sleep(1 * time.Second)
		}
		requests.Orbit(ship)
		time.Sleep(1 * time.Second)
	}

	ag_amount, ag_ok := cargoAmounts["SILVER_ORE"]
	au_amount, au_ok := cargoAmounts["GOLD_ORE"]
	pt_amount, pt_ok := cargoAmounts["PLATINUM_ORE"]
	if ag_ok || au_ok || pt_ok {
		trip := requests.TravelTo(ship, frozenMoon)
		sleepDuringTravel(trip)
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		//requests.ViewMarket(frozenMoon)
		//time.Sleep(1 * time.Second)
		if ag_ok {
			requests.SellCargo(ship, "SILVER_ORE", ag_amount)
			time.Sleep(1 * time.Second)
		}
		if au_ok {
			requests.SellCargo(ship, "GOLD_ORE", au_amount)
			time.Sleep(1 * time.Second)
		}
		if pt_ok {
			requests.SellCargo(ship, "PLATINUM_ORE", pt_amount)
			time.Sleep(1 * time.Second)
		}
		requests.Orbit(ship)
		time.Sleep(1 * time.Second)
	}

	if nh3_amount, ok := cargoAmounts["AMMONIA_ICE"]; ok {
		trip := requests.TravelTo(ship, volcanicMoon)
		sleepDuringTravel(trip)
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		if ok {
			requests.SellCargo(ship, "AMMONIA_ICE", nh3_amount)
			time.Sleep(1 * time.Second)
		}
		requests.Orbit(ship)
		time.Sleep(1 * time.Second)
	}
}

func sellCargoBesidesMaterial(ship, material string) {
	cargo := requests.DescribeShip(ship).Ship.Cargo.Inventory
	for i := len(cargo) - 1; i >= 0; i-- {
		item := cargo[i]
		prefix := item.Symbol[0:4]
		if prefix != "ALUM" {
			requests.SellCargo(ship, item.Symbol, item.Units)
		}
		time.Sleep(1 * time.Second)
	}
}

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

func doNewUserBoilerplate(callsign string) {
	// Creates a new user, saves the auth key, and refreshes the ship
	// names file.
	reply := requests.RegisterNewUser(callsign)
	time.Sleep(1 * time.Second)
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
	//time.Sleep(1 * time.Second)

	err = os.WriteFile("miningDrones.txt", []byte(callsign+"-1"), 664)
	if err != nil {
		fmt.Println(registerMsg.UserData.Token)
		panic(err)
	}
}
