package mcnbt

// CreateMemories represents the memories of an entity in a Create schematic
type CreateMemories struct {
}

// CreateBrain represents the brain of an entity in a Create schematic
type CreateBrain struct {
	Memories CreateMemories `json:"memories"`
}

// CreateForgeData represents the forge data of an entity in a Create schematic
type CreateForgeData struct {
}

// CreateAttribute represents an attribute of an entity in a Create schematic
type CreateAttribute struct {
	Base float64 `json:"Base"`
	Name string  `json:"Name"`
}

// CreateItem represents an item held by an entity in a Create schematic
type CreateItem struct {
}

// CreateEntityNbt represents the NBT data of an entity in a Create schematic
type CreateEntityNbt struct {
	Brain               CreateBrain       `json:"Brain"`
	HurtByTimestamp     int               `json:"HurtByTimestamp"`
	ForgeData           CreateForgeData   `json:"ForgeData"`
	Attributes          []CreateAttribute `json:"Attributes"`
	Invulnerable        int               `json:"Invulnerable"`
	FallFlying          int               `json:"FallFlying"`
	PortalCooldown      int               `json:"PortalCooldown"`
	AbsorptionAmount    int               `json:"AbsorptionAmount"`
	FallDistance        int               `json:"FallDistance"`
	CanUpdate           int               `json:"CanUpdate"`
	DeathTime           int               `json:"DeathTime"`
	HandDropChances     []float64         `json:"HandDropChances"`
	PersistenceRequired int               `json:"PersistenceRequired"`
	ID                  string            `json:"id"`
	BatFlags            int               `json:"BatFlags"`
	UUID                []int             `json:"UUID"`
	Motion              []int             `json:"Motion"`
	Health              int               `json:"Health"`
	LeftHanded          int               `json:"LeftHanded"`
	Air                 int               `json:"Air"`
	OnGround            int               `json:"OnGround"`
	Rotation            []int             `json:"Rotation"`
	HandItems           []CreateItem      `json:"HandItems"`
	ArmorDropChances    []float64         `json:"ArmorDropChances"`
	Pos                 []float64         `json:"Pos"`
	Fire                int               `json:"Fire"`
	ArmorItems          []CreateItem      `json:"ArmorItems"`
	CanPickUpLoot       int               `json:"CanPickUpLoot"`
	HurtTime            int               `json:"HurtTime"`
}

// CreateEntity represents an entity in a Create schematic
type CreateEntity struct {
	Nbt      CreateEntityNbt `json:"nbt"`
	BlockPos []int           `json:"blockPos"`
	Pos      []float64       `json:"pos"`
}

// CreateBlockProperty represents the properties of a block in a Create schematic
type CreateBlockProperty struct {
	Axis string `json:"axis"`
}

// CreatePalette represents a block in the palette of a Create schematic
type CreatePalette struct {
	Name       string              `json:"Name"`
	Properties CreateBlockProperty `json:"Properties,omitempty"`
}

// CreateTileEntity represents a tile entity in a Create schematic
type CreateTileEntity struct {
	Pos []int                  `json:"pos"`
	NBT map[string]interface{} `json:"nbt"`
}

// CreateNBT represents a Create schematic
type CreateNBT struct {
	Size         []int              `json:"size"`
	Entities     []CreateEntity     `json:"entities"`
	Blocks       []interface{}      `json:"blocks"`
	TileEntities []CreateTileEntity `json:"tileEntities,omitempty"`
	Palette      []CreatePalette    `json:"palette"`
	DataVersion  int                `json:"DataVersion"`
}
