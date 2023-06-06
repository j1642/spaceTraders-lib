package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"spacetraders/objects"
	"spacetraders/requests"
)

const barrenMoon string = "X1-HQ18-93722X"   // base metals
const frozenMoon string = "X1-HQ18-53964F"   // precious metals
const volcanicMoon string = "X1-HQ18-89363Z" // ammonia ice

const hq string = "X1-HQ18-11700D"
const asteroidField string = "X1-HQ18-98695F"
const shipyard string = "X1-HQ18-60817D"

var miningShips []string = readMiningShipNames()

func main() {
	//requests.ViewMarket(asteroidField)
	gather()
	//requests.TravelTo("BAP-1", asteroidField)
}

func gather() {
	// TODO: add channel for survey target if it contains desirable resources.
	wg := &sync.WaitGroup{}
	for _, ship := range miningShips {
		wg.Add(1)
		go collectAndDeliverMaterial(ship, "IRON_ORE", wg)
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
		time.Sleep(100 * time.Second)
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
			log.Fatal(err)
		}
		// Transport ship approaching but has not arrived.
		if transferMsg.ErrBody.Code == 4214 {
			fmt.Println(drone, "waiting for transport (return trip)")
			time.Sleep(28 * time.Second)
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
	requests.DeliverMaterial(ship, material, "clihsz0802xehs60dtjzkwetd")
	time.Sleep(1 * time.Second)
	requests.Orbit(ship)
	time.Sleep(1 * time.Second)

	// Sell additional materials.
	cargo := requests.DescribeShip(ship).Ship.Cargo
	cargoAmounts := make(map[string]int)
	for _, item := range cargo.Inventory {
		cargoAmounts[item.Symbol] = item.Units
	}
	if amount, ok := cargoAmounts["ICE_WATER"]; ok {
		requests.SellCargo(ship, "ICE_WATER", amount)
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
	if cu_ok || al_ok {
		//requests.TravelTo(ship, barrenMoon)
		//time.Sleep(15 * time.Second)
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
		if prefix == "ICE_" || prefix == "SILI" || prefix == "QUAR" || prefix == "DIAM" {
			requests.SellCargo(ship, item.Symbol, item.Units)
		}
		time.Sleep(1 * time.Second)
	}
}

func sleepDuringTravel(reply *bytes.Buffer) {
	travelMsg := objects.TravelData{}
	err := json.Unmarshal(reply.Bytes(), &travelMsg)
	if err != nil {
		log.Fatal(err)
	}

	format := "2006-01-02T15:04:05.000Z"
	start, err := time.Parse(format, travelMsg.Travel.Nav.Route.DepartureTime)
	if err != nil {
		log.Fatal(err)
	}

	end, err := time.Parse(format, travelMsg.Travel.Nav.Route.Arrival)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(end.Sub(start))
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
