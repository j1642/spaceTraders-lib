package main

import (
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

const barrenMoon string = "X1-ZA40-69371X"   // base metals
const frozenMoon string = "X1-ZA40-11513D"   // precious metals
const volcanicMoon string = "X1-ZA40-97262C" // ammonia ice

const hq string = "X1-ZA40-15970B"
const asteroidField string = "X1-ZA40-99095A"
const shipyard string = "X1-ZA40-68707C"

var miningShips []string = readMiningShipNames()

func main() {
	gather()
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
		requests.ExtractOre(ship, 10)
		time.Sleep(1 * time.Second)
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		sellCargoBesidesMaterial(ship, material)
		time.Sleep(1 * time.Second)
		requests.Orbit(ship)
		time.Sleep(1 * time.Second)

		//		shipData := requests.DescribeShip(ship).Ship
		//		time.Sleep(1 * time.Second)
		//		cargo := &shipData.Cargo
		//        available := cargo.Capacity - cargo.Units
		//		if shipData.Frame.Symbol == "FRAME_DRONE" ||
		//			shipData.Frame.Symbol == "FRAME_MINER" {
		//			transferCargoFromDrone(ship, cargo)
		//			time.Sleep(1 * time.Second)
		//			if cargo.Units < cargo.Capacity {
		//				continue
		//			}
		//			fmt.Println(ship, "waiting to transfer cargo")
		//			time.Sleep(130 * time.Second)
		//			continue
		//		} else if shipData.Frame.Symbol == "FRAME_FRIGATE" &&
		//        available == 0 {
		//			dropOffMaterialAndReturn(ship, material)
		//		}
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
		time.Sleep(130 * time.Second)
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
		fmt.Println(reply)
		transferMsg := objects.Error{}
		err := json.Unmarshal(reply.Bytes(), &transferMsg)
		if err != nil {
			log.Fatal(err)
		}
		// Transport ship approaching but has not arrived.
		if transferMsg.ErrBody.Code == 4214 {
			fmt.Println(drone, "waiting for transport (return trip)")
			time.Sleep(30 * time.Second)
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
	//requests.TravelTo(ship, hq)
	//fmt.Println(ship, "moving to the drop-off")
	//time.Sleep(30 * time.Second)
	//requests.DockShip(ship)
	//time.Sleep(1 * time.Second)

	// Drop off contract material.
	//deliverMaterial(ship, material)
	//time.Sleep(1 * time.Second)

	// Sell additional materials.
	cargo := requests.DescribeShip(ship).Ship.Cargo
	cargoAmounts := make(map[string]int)
	for _, item := range cargo.Inventory {
		cargoAmounts[item.Symbol] = item.Units
	}
	if amount, ok := cargoAmounts["ICE_WATER"]; ok {
		requests.SellCargo(ship, "ICE_WATER", amount)
	}
	requests.Orbit(ship)
	time.Sleep(1 * time.Second)

	sellCargoOnMoons(ship, cargoAmounts)

	// Return to mining location.
	time.Sleep(1 * time.Second)
	requests.TravelTo(ship, asteroidField)
	fmt.Println(ship, "returning from the drop-off")
	time.Sleep(30 * time.Second)
}

func sellCargoOnMoons(ship string, cargoAmounts map[string]int) {
	// Sell cargo to markets that generally pay the most.
	cu_amount, cu_ok := cargoAmounts["COPPER_ORE"]
	al_amount, al_ok := cargoAmounts["ALUMINUM_ORE"]
	if cu_ok || al_ok {
		requests.TravelTo(ship, barrenMoon)
		time.Sleep(15 * time.Second)
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		requests.ViewMarket(barrenMoon)
		time.Sleep(1 * time.Second)
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
		requests.TravelTo(ship, frozenMoon)
		time.Sleep(15 * time.Second)
		requests.DockShip(ship)
		time.Sleep(1 * time.Second)
		requests.ViewMarket(frozenMoon)
		time.Sleep(1 * time.Second)
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
		requests.TravelTo(ship, volcanicMoon)
		time.Sleep(15 * time.Second)
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
		if prefix != "ANTI" && prefix != material[:4] {
			requests.SellCargo(ship, item.Symbol, item.Units)
		}
		time.Sleep(1 * time.Second)
	}
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
