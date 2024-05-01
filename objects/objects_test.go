package objects

import (
	"encoding/json"
	"testing"
)

// These tests do not seem authoritative. They often do not break when a
// field is removed.

func TestShipData(t *testing.T) {
	jsonIn := `{
    "data": {
        "symbol": "A",
        "nav": {
            "systemSymbol": "A",
            "waypointSymbol": "A",
            "route": {
                "origin": {
                    "symbol": "A",
                    "type": "MOON",
                    "systemSymbol": "A",
                    "x": 1,
                    "y": 1
                },
                "destination": {
                    "symbol": "A",
                    "type": "PLANET",
                    "systemSymbol": "A",
                    "x": 1,
                    "y": 1
                },
                "arrival": "A",
                "departureTime": "A"
            },
            "status": "IN_ORBIT",
            "flightMode": "CRUISE"
        },
        "crew": {
            "current": 0,
            "capacity": 0,
            "required": 0,
            "rotation": "STRICT",
            "morale": 100,
            "wages": 0
        },
        "fuel": {
            "current": 100,
            "capacity": 100,
            "consumed": {
                "amount": 32,
                "timestamp": "A"
            }
        },
        "frame": {
            "symbol": "FRAME_DRONE",
            "name": "Frame Drone",
            "description": "A small, unmanned spacecraft used for various tasks, such as surveillance, transportation, or combat.",
            "moduleSlots": 3,
            "mountingPoints": 2,
            "fuelCapacity": 100,
            "condition": 100,
            "requirements": {
                "power": 1,
                "crew": -3
            }
        },
        "reactor": {
            "symbol": "REACTOR_CHEMICAL_I",
            "name": "Chemical Reactor I",
            "description": "A basic chemical power reactor, used to generate electricity from chemical reactions.",
            "condition": 100,
            "powerOutput": 15,
            "requirements": {
                "crew": 3
            }
        },
        "engine": {
            "symbol": "ENGINE_IMPULSE_DRIVE_I",
            "name": "Impulse Drive I",
            "description": "A basic low-energy propulsion system that generates thrust for interplanetary travel.",
            "condition": 100,
            "speed": 2,
            "requirements": {
                "power": 1,
                "crew": 0
            }
        },
        "modules": [
            {
                "symbol": "MODULE_CARGO_HOLD_I",
                "name": "Cargo Hold",
                "description": "A module that increases a ship's cargo capacity.",
                "capacity": 30,
                "requirements": {
                    "crew": 0,
                    "power": 1,
                    "slots": 1
                }
            },
            {
                "symbol": "MODULE_MINERAL_PROCESSOR_I",
                "name": "Mineral Processor",
                "description": "Crushes and processes extracted minerals and ores into their component parts, filters out impurities, and containerizes them into raw storage units.",
                "requirements": {
                    "crew": 0,
                    "power": 1,
                    "slots": 2
                }
            }
        ],
        "mounts": [
            {
                "symbol": "MOUNT_MINING_LASER_I",
                "name": "Mining Laser I",
                "description": "A basic mining laser that can be used to extract valuable minerals from asteroids and other space objects.",
                "strength": 10,
                "requirements": {
                    "crew": 0,
                    "power": 1
                }
            }
        ],
        "registration": {
            "name": "A",
            "factionSymbol": "A",
            "role": "EXCAVATOR"
        },
        "cargo": {
            "capacity": 30,
            "units": 0,
            "inventory": []
        }
    }
}`
	var data ShipData
	err := json.Unmarshal([]byte(jsonIn), &data)
	if err != nil {
		t.Fatal("TestDataShip():", err)
	}
	if data.Ship.Nav.Route.Origin.Type != "MOON" {
		t.Fatalf("Expected 'MOON', got=%v", data.Ship.Nav.Route.Origin.Type)
	}
}

func TestDataBuySell(t *testing.T) {
	jsonIn := `{
    "data": {
        "agent": {
            "accountId": "A",
            "symbol": "A",
            "headquarters": "A",
            "credits": 5
        },
        "cargo": {
            "capacity": 30,
            "units": 2,
            "inventory": [
                {
                    "symbol": "SILVER_ORE",
                    "name": "Silver Ore",
                    "description": "A raw, unprocessed form of silver, often found in mines and underground deposits on planets and moons.",
                    "units": 2
                }
            ]
        },
        "transaction": {
            "waypointSymbol": "A",
            "shipSymbol": "A",
            "tradeSymbol": "A",
            "type": "SELL",
            "units": 5,
            "pricePerUnit": 1,
            "totalPrice": 5,
            "timestamp": "A"
        }
    }
}`
	var data DataBuySell
	err := json.Unmarshal([]byte(jsonIn), &data)
	if err != nil {
		t.Fatal("TestDataBuySell():", err)
	}
	if data.BuySell.Transaction.Type != "SELL" {
		t.Fatalf("Expected 'SELL', got=%v", data.BuySell.Transaction.Type)
	}
}

func TestError4214(t *testing.T) {
	jsonReply := `{
    "error": {
        "message": "Ship is currently in-transit from A to A and arrives in 2 seconds.",
        "code": 4214,
        "data": {
            "departureSymbol": "A",
            "destinationSymbol": "A",
            "secondsToArrival": 2
        }
    }
}`
	var data Error
	err := json.Unmarshal([]byte(jsonReply), &data)
	if err != nil {
		t.Fatal("TestDataBuySell():", err)
	}
	if data.ErrBody.Code != 4214 {
		t.Fatalf("Expected 4124, got=%v", data.ErrBody.Code)
	}
	if data.ErrBody.Data.SecondsToArrival != 2 {
		t.Fatalf("Expected 2, got=%v", data.ErrBody.Data.SecondsToArrival)
	}
}

func TestMarketData(t *testing.T) {
	jsonReply := `{
    "data": {
        "symbol": "A",
        "imports": [
            {
                "symbol": "ICE_WATER",
                "name": "Fresh Water",
                "description": "High-quality fresh water, essential for life support and hydroponic agriculture."
            },
            {
                "symbol": "PLASTICS",
                "name": "Plastics",
                "description": "A wide range of plastic materials used in various applications, including packaging, construction, and manufacturing."
            }
        ],
        "exports": [
            {
                "symbol": "GRAVITON_EMITTERS",
                "name": "Graviton Emitters",
                "description": "Advanced devices that generate and manipulate gravitons, used in advanced propulsion and weapons systems."
            }
        ],
        "exchange": [
                    {
                        "symbol": "FUEL",
                        "name": "Fuel",
                        "description": "High-energy fuel used in spacecraft propulsion systems to enable long-distance space travel."
                    }
        ],
        "transactions": [
                    {
                        "waypointSymbol": "A",
                        "shipSymbol": "A",
                        "tradeSymbol": "A",
                        "type": "SELL",
                        "units": 1,
                        "pricePerUnit": 1,
                        "totalPrice": 1,
                        "timestamp": "A"
                    }
        ],
        "tradeGoods": [
                    {
                        "symbol": "A",
                        "tradeVolume": 1000,
                        "supply": "MODERATE",
                        "purchasePrice": 1,
                        "sellPrice": 1
                    }
        ]
    }
}`
	var data MarketData
	err := json.Unmarshal([]byte(jsonReply), &data)
	if err != nil {
		t.Fatal("TestDataBuySell():", err)
	}
	t.Log(data)
	if data.Market.Imports[0].Symbol != "ICE_WATER" {
		t.Fatalf("expected 'ICE_WATER', got=%v", data.Market.Imports[0].Symbol)
	}
}

func TestExtractionData(t *testing.T) {
	jsonReply := `{
    "data": {
        "extraction": {
            "shipSymbol": "A",
            "yield": {
                "symbol": "IRON_ORE",
                "units": 1
            }
        },
        "cooldown": {
            "shipSymbol": "A",
            "totalSeconds": 2,
            "remainingSeconds": 1,
            "expiration": "A"
        },
        "cargo": {
            "capacity": 4,
            "units": 3,
            "inventory": [
                {
                    "symbol": "IRON_ORE",
                    "name": "Iron Ore",
                    "description": "A common and valuable ore used in the production of steel and other alloys.",
                    "units": 3
                }
            ]
        }
    }
}`
	var data ExtractionData
	err := json.Unmarshal([]byte(jsonReply), &data)
	if err != nil {
		t.Fatal("TestDataBuySell():", err)
	}
	if data.ExtractBody.Extraction.Yield.Item != "IRON_ORE" {
		t.Fatalf("expected 'IRON_ORE', got=%v",
			data.ExtractBody.Extraction.Yield.Item)
	}
}
