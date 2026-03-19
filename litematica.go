package mcnbt

// EntityItem represents an item held by an entity
type EntityItem struct {
	Count int8   `json:"Count" nbt:"Count"`
	Slot  int8   `json:"Slot" nbt:"Slot"`
	ID    string `json:"id" nbt:"id"`
}

// EntityAttribute represents an attribute of an entity
type EntityAttribute struct {
	Base float64 `json:"Base" nbt:"Base"`
	Name string  `json:"Name" nbt:"Name"`
}

// EntityBrain represents the brain of an entity
type EntityBrain struct {
	Memories map[string]interface{} `json:"memories" nbt:"memories"`
}

// LitematicaMetadata represents the metadata of a litematica schematic
type LitematicaMetadata struct {
	Author           string     `json:"Author" nbt:"Author"`
	Description      string     `json:"Description" nbt:"Description"`
	EnclosingSize    Coordinate `json:"EnclosingSize" nbt:"EnclosingSize"`
	Name             string     `json:"Name" nbt:"Name"`
	PreviewImageData []int32    `json:"PreviewImageData" nbt:"PreviewImageData"`
	RegionCount      int32      `json:"RegionCount" nbt:"RegionCount"`
	TimeCreated      int64      `json:"TimeCreated" nbt:"TimeCreated"`
	TimeModified     int64      `json:"TimeModified" nbt:"TimeModified"`
	TotalBlocks      int32      `json:"TotalBlocks" nbt:"TotalBlocks"`
	TotalVolume      int32      `json:"TotalVolume" nbt:"TotalVolume"`
}

// LitematicaBlockStatePalette represents a block state in the palette
type LitematicaBlockStatePalette struct {
	Name       string            `json:"Name" nbt:"Name"`
	Properties map[string]string `json:"Properties,omitempty" nbt:"Properties,omitempty"`
}

// LitematicaEntity represents an entity in a litematica schematic
type LitematicaEntity struct {
	AbsorptionAmount    int32             `json:"AbsorptionAmount" nbt:"AbsorptionAmount"`
	Air                 int32             `json:"Air" nbt:"Air"`
	ArmorDropChances    []float32         `json:"ArmorDropChances" nbt:"ArmorDropChances"`
	ArmorItems          []EntityItem      `json:"ArmorItems" nbt:"ArmorItems"`
	Attributes          []EntityAttribute `json:"Attributes" nbt:"Attributes"`
	BatFlags            int32             `json:"BatFlags" nbt:"BatFlags"`
	Brain               EntityBrain       `json:"Brain" nbt:"Brain"`
	CanPickUpLoot       int32             `json:"CanPickUpLoot" nbt:"CanPickUpLoot"`
	DeathTime           int32             `json:"DeathTime" nbt:"DeathTime"`
	FallDistance        int32             `json:"FallDistance" nbt:"FallDistance"`
	FallFlying          int32             `json:"FallFlying" nbt:"FallFlying"`
	Fire                int32             `json:"Fire" nbt:"Fire"`
	HandDropChances     []float32         `json:"HandDropChances" nbt:"HandDropChances"`
	HandItems           []EntityItem      `json:"HandItems" nbt:"HandItems"`
	Health              int32             `json:"Health" nbt:"Health"`
	HurtByTimestamp     int32             `json:"HurtByTimestamp" nbt:"HurtByTimestamp"`
	HurtTime            int32             `json:"HurtTime" nbt:"HurtTime"`
	Invulnerable        int32             `json:"Invulnerable" nbt:"Invulnerable"`
	LeftHanded          int32             `json:"LeftHanded" nbt:"LeftHanded"`
	Motion              []float64         `json:"Motion" nbt:"Motion"`
	OnGround            int32             `json:"OnGround" nbt:"OnGround"`
	PersistenceRequired int32             `json:"PersistenceRequired" nbt:"PersistenceRequired"`
	PortalCooldown      int32             `json:"PortalCooldown" nbt:"PortalCooldown"`
	Pos                 []float64         `json:"Pos" nbt:"Pos"`
	Rotation            []float32         `json:"Rotation" nbt:"Rotation"`
	UUID                []int32           `json:"UUID" nbt:"UUID"`
	ID                  string            `json:"id" nbt:"id"`
}

// LitematicaTileEntity represents a tile entity in a litematica schematic
type LitematicaTileEntity struct {
	Items             []interface{} `json:"Items,omitempty" nbt:"Items,omitempty"`
	Id                string        `json:"Id,omitempty" nbt:"Id,omitempty"`
	X                 int32         `json:"x" nbt:"x"`
	Y                 int32         `json:"y" nbt:"y"`
	Z                 int32         `json:"z" nbt:"z"`
	CookingTimes      []int32       `json:"CookingTimes,omitempty" nbt:"CookingTimes,omitempty"`
	CookingTotalTimes []int32       `json:"CookingTotalTimes,omitempty" nbt:"CookingTotalTimes,omitempty"`
	Bees              []interface{} `json:"Bees,omitempty" nbt:"Bees,omitempty"`
}

// LitematicaRegion represents a region in a litematica schematic
type LitematicaRegion struct {
	BlockStatePalette []LitematicaBlockStatePalette `json:"BlockStatePalette" nbt:"BlockStatePalette"`
	BlockStates       []int64                       `json:"BlockStates" nbt:"BlockStates"`
	Entities          []LitematicaEntity            `json:"Entities" nbt:"Entities"`
	PendingBlockTicks []interface{}                 `json:"PendingBlockTicks" nbt:"PendingBlockTicks"`
	PendingFluidTicks []interface{}                 `json:"PendingFluidTicks" nbt:"PendingFluidTicks"`
	Position          Coordinate                    `json:"Position" nbt:"Position"`
	Size              Coordinate                    `json:"Size" nbt:"Size"`
	TileEntities      []LitematicaTileEntity        `json:"TileEntities" nbt:"TileEntities"`
}

// LitematicaNBT represents a litematica schematic
type LitematicaNBT struct {
	Metadata             LitematicaMetadata          `json:"Metadata" nbt:"Metadata"`
	MinecraftDataVersion int32                       `json:"MinecraftDataVersion" nbt:"MinecraftDataVersion"`
	Regions              map[string]LitematicaRegion `json:"Regions" nbt:"Regions"`
	SubVersion           int32                       `json:"SubVersion" nbt:"SubVersion"`
	Version              int32                       `json:"Version" nbt:"Version"`
}
