package mcnbt

// CreateMemories represents the memories of an entity in a Create schematic
type CreateMemories struct {
}

// CreateBrain represents the brain of an entity in a Create schematic
type CreateBrain struct {
	Memories CreateMemories `json:"memories" nbt:"memories"`
}

// CreateForgeData represents the forge data of an entity in a Create schematic
type CreateForgeData struct {
}

// CreateAttribute represents an attribute of an entity in a Create schematic
type CreateAttribute struct {
	Base float64 `json:"Base" nbt:"Base"`
	Name string  `json:"Name" nbt:"Name"`
}

// CreateItem represents an item held by an entity in a Create schematic
type CreateItem struct {
}

// CreateEntityNbt represents the NBT data of an entity in a Create schematic
type CreateEntityNbt struct {
	Brain               CreateBrain       `json:"Brain" nbt:"Brain"`
	HurtByTimestamp     int32             `json:"HurtByTimestamp" nbt:"HurtByTimestamp"`
	ForgeData           CreateForgeData   `json:"ForgeData" nbt:"ForgeData"`
	Attributes          []CreateAttribute `json:"Attributes" nbt:"Attributes"`
	Invulnerable        int32             `json:"Invulnerable" nbt:"Invulnerable"`
	FallFlying          int32             `json:"FallFlying" nbt:"FallFlying"`
	PortalCooldown      int32             `json:"PortalCooldown" nbt:"PortalCooldown"`
	AbsorptionAmount    int32             `json:"AbsorptionAmount" nbt:"AbsorptionAmount"`
	FallDistance        int32             `json:"FallDistance" nbt:"FallDistance"`
	CanUpdate           int32             `json:"CanUpdate" nbt:"CanUpdate"`
	DeathTime           int32             `json:"DeathTime" nbt:"DeathTime"`
	HandDropChances     []float32         `json:"HandDropChances" nbt:"HandDropChances"`
	PersistenceRequired int32             `json:"PersistenceRequired" nbt:"PersistenceRequired"`
	ID                  string            `json:"id" nbt:"id"`
	BatFlags            int32             `json:"BatFlags" nbt:"BatFlags"`
	UUID                []int32           `json:"UUID" nbt:"UUID"`
	Motion              []float64         `json:"Motion" nbt:"Motion"`
	Health              int32             `json:"Health" nbt:"Health"`
	LeftHanded          int32             `json:"LeftHanded" nbt:"LeftHanded"`
	Air                 int32             `json:"Air" nbt:"Air"`
	OnGround            int32             `json:"OnGround" nbt:"OnGround"`
	Rotation            []float32         `json:"Rotation" nbt:"Rotation"`
	HandItems           []CreateItem      `json:"HandItems" nbt:"HandItems"`
	ArmorDropChances    []float32         `json:"ArmorDropChances" nbt:"ArmorDropChances"`
	Pos                 []float64         `json:"Pos" nbt:"Pos"`
	Fire                int32             `json:"Fire" nbt:"Fire"`
	ArmorItems          []CreateItem      `json:"ArmorItems" nbt:"ArmorItems"`
	CanPickUpLoot       int32             `json:"CanPickUpLoot" nbt:"CanPickUpLoot"`
	HurtTime            int32             `json:"HurtTime" nbt:"HurtTime"`
}

// CreateEntity represents an entity in a Create schematic
type CreateEntity struct {
	Nbt      CreateEntityNbt `json:"nbt" nbt:"nbt"`
	BlockPos []int32         `json:"blockPos" nbt:"blockPos,list"`
	Pos      []float64       `json:"pos" nbt:"pos"`
}

// CreatePalette represents a block in the palette of a Create/Vanilla structure
type CreatePalette struct {
	Name       string            `json:"Name" nbt:"Name"`
	Properties map[string]string `json:"Properties,omitempty" nbt:"Properties,omitempty"`
}

// CreateTileEntity represents a tile entity in a Create schematic
type CreateTileEntity struct {
	Pos []int32                `json:"pos" nbt:"pos,list"`
	NBT map[string]interface{} `json:"nbt" nbt:"nbt"`
}

// CreateNBT represents a Create/Vanilla structure NBT
type CreateNBT struct {
	Size                []int32            `json:"size" nbt:"size,list"`
	Entities            []CreateEntity     `json:"entities" nbt:"entities"`
	Blocks              []CreateBlock      `json:"blocks" nbt:"blocks"`
	TileEntities        []CreateTileEntity `json:"tileEntities,omitempty" nbt:"tileEntities,omitempty"`
	Palette             []CreatePalette    `json:"palette" nbt:"palette"`
	DataVersion         int32              `json:"DataVersion" nbt:"DataVersion"`
	RailwaysDataVersion int32              `json:"Railways_DataVersion,omitempty" nbt:"Railways_DataVersion,omitempty"`
}

// CreateBlock represents a single block in a Create/Vanilla structure
type CreateBlock struct {
	Nbt   interface{} `json:"nbt" nbt:"nbt,omitempty"`
	Pos   []int32     `json:"pos" nbt:"pos,list"`
	State int32       `json:"state" nbt:"state"`
}
