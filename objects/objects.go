package objects

type AgentData struct {
	Agent Agent `json:"data"`
}
type Agent struct {
	AccountId      string `json:"accountId"`
	Symbol         string `json:"symbol"`
	Headquarters   string `json:"headquarters"`
	Credits        int    `json:"credits"`
	InitialFaction string `json:"startingFaction"`
	ShipCount      int    `json:"shipCount"`
}

// Buy/sell
type DataBuySell struct {
	BuySell BuySell `json:"data,omitempty"`
}
type BuySell struct {
	Agent       Agent       `json:"agent,omitempty"`
	Cargo       Cargo       `json:"cargo,omitempty"`
	Transaction Transaction `json:"transaction,omitempty"`
}

type Cargo struct {
	Capacity  int         `json:"capacity,omitempty"`
	Units     int         `json:"units,omitempty"`
	Inventory []CargoItem `json:"inventory,omitempty"`
}
type CargoItem struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Units       int    `json:"units,omitempty"`
}

type Contract struct {
	Id             string `json:"id"`
	Faction        string `json:"factionSymbol"`
	Type           string `json:"type"`
	Terms          Terms  `json:"terms"`
	Accepted       bool   `json:"accepted"`
	Fulfilled      bool   `json:"funfilled"`
	Expires        string `json:"expiration"`
	AcceptDeadline string `json:"deadlineToAccept"`
}
type Terms struct {
	Deadline string         `json:"deadline"`
	Payment  map[string]int `json:"payment"`
	Deliver  []Delivery     `json:"deliver"`
}
type ContractData struct {
	Contract Contract `json:"data"`
}
type AllContracts struct {
	Contracts []Contract     `json:"data"`
	Meta      map[string]int `json:"meta"`
}

type Cooldown struct {
	ShipSymbol       string `json:"shipSymbol"`
	TotalSeconds     int    `json:"totalSeconds"`
	RemainingSeconds int    `json:"remainingSeconds"`
	Expiration       string `json:"expiration"`
}

type Crew struct {
	Current  int    `json:"current,omitempty"`
	Capacity int    `json:"capacity,omitempty"`
	Required int    `json:"required,omitempty"`
	Rotation string `json:"rotation,omitempty"`
	Morale   int    `json:"morale,omitempty"`
	Wages    int    `json:"wages,omitempty"`
}

type Delivery struct {
	// Desired contract material and progress.
	Item           string `json:"tradeSymbol"`
	Destination    string `json:"destinationSymbol"`
	UnitsRequired  int    `json:"unitsRequired"`
	UnitsFulfilled int    `json:"unitsFulfilled"`
}

type Error struct {
	ErrBody ErrBody `json:"error"`
}
type ErrBody struct {
	Message string  `json:"message"`
	Code    int     `json:"code"`
	Data    ErrData `json:"data"`
}

// TODO: why is this general error struct so navigational in nature?
type ErrData struct {
	DepartureSymbol   string   `json:"departureSymbol,omitempty"`
	DestinationSymbol string   `json:"destinationSymbol,omitempty"`
	SecondsToArrival  int      `json:"secondsToArrival,omitempty"`
	Cooldown          Cooldown `json:"cooldown,omitempty"`
}

// Extraction and mining
type ExtractionData struct {
	ExtractBody ExtractBody `json:"data,omitempty"`
	ErrBody     ErrBody     `json:"error,omitempty"`
}
type ExtractBody struct {
	Extraction Extraction `json:"extraction,omitempty"`
	Siphon     Extraction `json:"siphon,omitempty"`
	Cooldown   Cooldown   `json:"cooldown"`
	Cargo      Cargo      `json:"cargo"`
}
type Extraction struct {
	ShipSymbol string `json:"shipSymbol"`
	Yield      Yield  `json:"yield"`
}
type Yield struct {
	Item  string `json:"symbol"`
	Units int    `json:"units"`
}

type Faction struct {
	Symbol       string              `json:"symbol"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Headquarters string              `json:"headquarters"`
	Traits       []map[string]string `json:"traits"`
	IsRecruiting bool                `json:"isRecruiting"`
}

type FuelConsumed struct {
	Amount    int    `json:"amount"`
	Timestamp string `json:"timestamp"`
}

type MarketData struct {
	Market Market `json:"data"`
}
type Market struct {
	Symbol       string        `json:"symbol"`
	Imports      []CargoItem   `json:"imports"`
	Exports      []CargoItem   `json:"exports"`
	Exchange     []CargoItem   `json:"exchange"`
	Transactions []Transaction `json:"transactions,omitempty"`
	TradeGoods   []TradeGood   `json:"tradeGoods,omitempty"`
}

type Nav struct {
	SystemSymbol   string   `json:"systemSymbol"`
	WaypointSymbol string   `json:"waypointSymbol"`
	Route          NavRoute `json:"route"`
	Status         string   `json:"status"`
	FlightMode     string   `json:"flightMode"`
}
type NavRoute struct {
	Origin        NavRouteLocation `json:"origin"`
	Destination   NavRouteLocation `json:"destination"`
	Arrival       string           `json:"arrival"`
	DepartureTime string           `json:"departureTime"`
}
type NavRouteLocation struct {
	Symbol       string `json:"symbol"`
	Type         string `json:"type"`
	SystemSymbol string `json:"systemSymbol"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
}

// Ship description.
type AllShips struct {
	Ships []Ship `json:"data"`
}
type ShipData struct {
	Ship Ship `json:"data"`
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

type TradeGood struct {
	Symbol        string `json:"symbol"`
	TradeVolume   int    `json:"tradeVolume"`
	Supply        string `json:"supply"`
	PurchasePrice int    `json:"purchasePrice"`
	SellPrice     int    `json:"sellPrice"`
}

type Transaction struct {
	WaypointSymbol string `json:"waypointSymbol,omitempty"`
	ShipSymbol     string `json:"shipSymbol,omitempty"`
	TradeSymbol    string `json:"tradeSymbol,omitempty"`
	Type           string `json:"type,omitempty"`
	Units          int    `json:"units,omitempty"`
	PricePerUnit   int    `json:"pricePerUnit,omitempty"`
	TotalPrice     int    `json:"totalPrice,omitempty"`
	TimeStamp      string `json:"timeStamp,omitempty"`
}

// Travel
type TravelData struct {
	Travel Travel `json:"data"`
}
type Travel struct {
	Nav  Nav              `json:"nav"`
	Fuel OneIndentedField `json:"fuel"`
}

// User Registration
type User struct {
	UserData UserData `json:"data,omitempty"`
	ErrBody  ErrBody  `json:"error,omitempty"`
}
type UserData struct {
	Token    string   `json:"token"`
	Agent    Agent    `json:"agent"`
	Contract Contract `json:"contract"`
	Faction  Faction  `json:"faction"`
	Ship     Ship     `json:"ship"`
}

type Waypoints struct {
	Data []Waypoint     `json:"data"`
	Meta map[string]int `json:"meta,omitempty"`
}
type Waypoint struct {
	System   string              `json:"systemSymbol"`
	Symbol   string              `json:"symbol"`
	Type     string              `json:"type"`
	X        int                 `json:"x"`
	Y        int                 `json:"y"`
	Orbitals []map[string]string `json:"orbitals"`
	// Traits include MARKETPLACE, SHIPYARD
	Traits              []map[string]string `json:"traits"`
	Modifiers           []map[string]string `json:"modifiers"`
	Chart               map[string]string   `json:"chart"`
	Faction             map[string]string   `json:"faction,omitempty"`
	Orbits              string              `json:"orbits,omitempty"`
	IsUnderConstruction bool                `json:"isUnderConstruction"`
}

// Non-alphabetical, general

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
