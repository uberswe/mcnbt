package mcnbt

// EntityItem represents an item held by an entity
type EntityItem struct {
	Count int    `json:"Count"`
	Slot  int    `json:"Slot"`
	ID    string `json:"id"`
}

// EntityAttribute represents an attribute of an entity
type EntityAttribute struct {
	Base float64 `json:"Base"`
	Name string  `json:"Name"`
}

// EntityBrain represents the brain of an entity
type EntityBrain struct {
	Memories map[string]interface{} `json:"memories"`
}

// LitematicaMetadata represents the metadata of a litematica schematic
type LitematicaMetadata struct {
	Author           string     `json:"Author"`
	Description      string     `json:"Description"`
	EnclosingSize    Coordinate `json:"EnclosingSize"`
	Name             string     `json:"Name"`
	PreviewImageData []int      `json:"PreviewImageData"`
	RegionCount      int        `json:"RegionCount"`
	TimeCreated      int64      `json:"TimeCreated"`
	TimeModified     int64      `json:"TimeModified"`
	TotalBlocks      int        `json:"TotalBlocks"`
	TotalVolume      int        `json:"TotalVolume"`
}

// LitematicaBlockStatePalette represents a block state in the palette
type LitematicaBlockStatePalette struct {
	Name       string            `json:"Name"`
	Properties map[string]string `json:"Properties,omitempty"`
}

// LitematicaEntity represents an entity in a litematica schematic
type LitematicaEntity struct {
	AbsorptionAmount    int               `json:"AbsorptionAmount"`
	Air                 int               `json:"Air"`
	ArmorDropChances    []float64         `json:"ArmorDropChances"`
	ArmorItems          []EntityItem      `json:"ArmorItems"`
	Attributes          []EntityAttribute `json:"Attributes"`
	BatFlags            int               `json:"BatFlags"`
	Brain               EntityBrain       `json:"Brain"`
	CanPickUpLoot       int               `json:"CanPickUpLoot"`
	DeathTime           int               `json:"DeathTime"`
	FallDistance        int               `json:"FallDistance"`
	FallFlying          int               `json:"FallFlying"`
	Fire                int               `json:"Fire"`
	HandDropChances     []float64         `json:"HandDropChances"`
	HandItems           []EntityItem      `json:"HandItems"`
	Health              int               `json:"Health"`
	HurtByTimestamp     int               `json:"HurtByTimestamp"`
	HurtTime            int               `json:"HurtTime"`
	Invulnerable        int               `json:"Invulnerable"`
	LeftHanded          int               `json:"LeftHanded"`
	Motion              []float64         `json:"Motion"`
	OnGround            int               `json:"OnGround"`
	PersistenceRequired int               `json:"PersistenceRequired"`
	PortalCooldown      int               `json:"PortalCooldown"`
	Pos                 []float64         `json:"Pos"`
	Rotation            []float64         `json:"Rotation"`
	UUID                []int             `json:"UUID"`
	ID                  string            `json:"id"`
}

// LitematicaTileEntity represents a tile entity in a litematica schematic
type LitematicaTileEntity struct {
	Items             []interface{} `json:"Items,omitempty"`
	Id                string        `json:"Id,omitempty"`
	X                 int           `json:"x"`
	Y                 int           `json:"y"`
	Z                 int           `json:"z"`
	CookingTimes      []int         `json:"CookingTimes,omitempty"`
	CookingTotalTimes []int         `json:"CookingTotalTimes,omitempty"`
	Bees              []interface{} `json:"Bees,omitempty"`
}

// LitematicaRegion represents a region in a litematica schematic
type LitematicaRegion struct {
	BlockStatePalette []LitematicaBlockStatePalette `json:"BlockStatePalette"`
	BlockStates       []interface{}                 `json:"BlockStates"`
	Entities          []LitematicaEntity            `json:"Entities"`
	PendingBlockTicks []interface{}                 `json:"PendingBlockTicks"`
	PendingFluidTicks []interface{}                 `json:"PendingFluidTicks"`
	Position          Coordinate                    `json:"Position"`
	Size              Coordinate                    `json:"Size"`
	TileEntities      []LitematicaTileEntity        `json:"TileEntities"`
}

// LitematicaNBT represents a litematica schematic
type LitematicaNBT struct {
	Metadata             LitematicaMetadata          `json:"Metadata"`
	MinecraftDataVersion int                         `json:"MinecraftDataVersion"`
	Regions              map[string]LitematicaRegion `json:"Regions"`
	Version              int                         `json:"Version"`
}
