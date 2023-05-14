package objects

// Ship description.
type DataShip struct {
	Ship Ship `json:"data,omitempty"`
}

func (d *DataShip) String() string {
	return d.Ship.Symbol + "2"
}

type Ship struct {
	Symbol       string             `json:"symbol,omitempty"`
	Nav          Nav                `json:"nav,omitmpty"`
	Crew         Crew               `json:"crew,omitempty"`
	Fuel         OneIndentedField   `json:"fuel,omitempty"`
	Frame        OneIndentedField   `json:"frame,omitempty"`
	Reactor      OneIndentedField   `json:"reactor,omitempty"`
	Engine       OneIndentedField   `json:"engine,omitempty"`
	Modules      []OneIndentedField `json:"modules,omitempty"`
	Mounts       []OneIndentedField `json:"mounts,omitempty"`
	Registration map[string]string  `json:"registration,omitempty"`
	Cargo        Cargo              `json:"cargo,omitempty"`
}

type Nav struct {
	SystemSymbol   string   `json:"systemSymbol,omitempty"`
	WaypointSymbol string   `json:"waypointSymbol,omitempty"`
	Route          NavRoute `json:"route,omitempty"`
	Arrival        string   `json:"arrival,omitempty"`
	DepartureTime  string   `json:"departureTime,omitempty"`
}

type OneIndentedField struct {
	Symbol         string         `json:"symbol,omitempty"`
	Name           string         `json:"name,omitempty"`
	Description    string         `json:"description,omitempty"`
	Strength       int            `json:"strength,omitempty"`
	Capacity       int            `json:"capacity,omitempty"`
	Current        int            `json:"current,omitempty"`
	Speed          int            `json:"speed,omitempty"`
	Condition      int            `json:"condition,omitempty"`
	PowerOutput    int            `json:"powerOutput,omitempty"`
	ModuleSlots    int            `json:"moduleSlots,omitempty"`
	MountingPoints int            `json:"mountingPoints,omitempty"`
	FuelCapacity   int            `json:"fuelCapacity,omitempty"`
	Requirements   map[string]int `json:"requirements,omitempty"`
	Consumed       FuelConsumed   `json:"consumed,omitempty"`
}

type NavRoute struct {
	Departure   NavRouteLocation `json:"departure,omitempty"`
	Destination NavRouteLocation `json:"destination,omitempty"`
}

type NavRouteLocation struct {
	Symbol       string `json:"symbol,omitempty"`
	Type         string `json:"type,omitempty"`
	SystemSymbol string `json:"systemSymbol,omitempty"`
	X            int    `json:"x,omitempty"`
	Y            int    `json:"y,omitempty"`
}

type Crew struct {
	Current  int    `json:"current,omitempty"`
	Capacity int    `json:"capacity,omitempty"`
	Required int    `json:"required,omitempty"`
	Rotation string `json:"rotation,omitempty"`
	Morale   int    `json:"morale,omitempty"`
	Wages    int    `json:"wages,omitempty"`
}

type FuelConsumed struct {
	Amount    int    `json:"amount,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

type Cargo struct {
	Capacity  int         `json:"capacity,omitempty"`
	Units     int         `json:"units,omitempty"`
	Inventory []CargoItem `json:"inventory,omitempty"`
}

type CargoItem struct {
	Symbol      string `json:"symbol,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Units       int    `json:"units,omitempty"`
}

// End ship description.
