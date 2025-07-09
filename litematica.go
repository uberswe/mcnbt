package mcnbt

type LitematicaNBT struct {
	Metadata struct {
		Author        string `json:"Author"`
		Description   string `json:"Description"`
		EnclosingSize struct {
			X int `json:"x"`
			Y int `json:"y"`
			Z int `json:"z"`
		} `json:"EnclosingSize"`
		Name             string `json:"Name"`
		PreviewImageData []int  `json:"PreviewImageData"`
		RegionCount      int    `json:"RegionCount"`
		TimeCreated      int64  `json:"TimeCreated"`
		TimeModified     int64  `json:"TimeModified"`
		TotalBlocks      int    `json:"TotalBlocks"`
		TotalVolume      int    `json:"TotalVolume"`
	} `json:"Metadata"`
	MinecraftDataVersion int                         `json:"MinecraftDataVersion"`
	Regions              map[string]LitematicaRegion `json:"Regions"`
	Version              int                         `json:"Version"`
}

// LitematicaRegion represents a region in a litematica schematic
type LitematicaRegion struct {
	BlockStatePalette []struct {
		Name       string `json:"Name"`
		Properties struct {
			Snowy string `json:"snowy"`
		} `json:"Properties,omitempty"`
	} `json:"BlockStatePalette"`
	BlockStates []interface{} `json:"BlockStates"`
	Entities    []struct {
		AbsorptionAmount int       `json:"AbsorptionAmount"`
		Air              int       `json:"Air"`
		ArmorDropChances []float64 `json:"ArmorDropChances"`
		ArmorItems       []struct {
		} `json:"ArmorItems"`
		Attributes []struct {
			Base      float64 `json:"Base"`
			Name      string  `json:"Name"`
			Modifiers []struct {
				Amount    float64 `json:"Amount"`
				Name      string  `json:"Name"`
				Operation int     `json:"Operation"`
				UUID      []int   `json:"UUID"`
			} `json:"Modifiers,omitempty"`
		} `json:"Attributes"`
		BatFlags int `json:"BatFlags"`
		Brain    struct {
			Memories struct {
			} `json:"memories"`
		} `json:"Brain"`
		CanPickUpLoot   int       `json:"CanPickUpLoot"`
		DeathTime       int       `json:"DeathTime"`
		FallDistance    int       `json:"FallDistance"`
		FallFlying      int       `json:"FallFlying"`
		Fire            int       `json:"Fire"`
		HandDropChances []float64 `json:"HandDropChances"`
		HandItems       []struct {
		} `json:"HandItems"`
		Health              int       `json:"Health"`
		HurtByTimestamp     int       `json:"HurtByTimestamp"`
		HurtTime            int       `json:"HurtTime"`
		Invulnerable        int       `json:"Invulnerable"`
		LeftHanded          int       `json:"LeftHanded"`
		Motion              []float64 `json:"Motion"`
		OnGround            int       `json:"OnGround"`
		PersistenceRequired int       `json:"PersistenceRequired"`
		PortalCooldown      int       `json:"PortalCooldown"`
		Pos                 []float64 `json:"Pos"`
		Rotation            []float64 `json:"Rotation"`
		UUID                []int     `json:"UUID"`
		ID                  string    `json:"id"`
	} `json:"Entities"`
	PendingBlockTicks []interface{} `json:"PendingBlockTicks"`
	PendingFluidTicks []interface{} `json:"PendingFluidTicks"`
	Position          struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"Position"`
	Size struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"Size"`
	TileEntities []struct {
		Items             []interface{} `json:"Items,omitempty"`
		X                 int           `json:"x"`
		Y                 int           `json:"y"`
		Z                 int           `json:"z"`
		CookingTimes      []int         `json:"CookingTimes,omitempty"`
		CookingTotalTimes []int         `json:"CookingTotalTimes,omitempty"`
		Bees              []interface{} `json:"Bees,omitempty"`
	} `json:"TileEntities"`
}
