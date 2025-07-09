package mcnbt

type CreateNBT struct {
	Size     []int `json:"size"`
	Entities []struct {
		Nbt struct {
			Brain struct {
				Memories struct {
				} `json:"memories"`
			} `json:"Brain"`
			HurtByTimestamp int `json:"HurtByTimestamp"`
			ForgeData       struct {
			} `json:"ForgeData"`
			Attributes []struct {
				Base float64 `json:"Base"`
				Name string  `json:"Name"`
			} `json:"Attributes"`
			Invulnerable        int       `json:"Invulnerable"`
			FallFlying          int       `json:"FallFlying"`
			PortalCooldown      int       `json:"PortalCooldown"`
			AbsorptionAmount    int       `json:"AbsorptionAmount"`
			FallDistance        int       `json:"FallDistance"`
			CanUpdate           int       `json:"CanUpdate"`
			DeathTime           int       `json:"DeathTime"`
			HandDropChances     []float64 `json:"HandDropChances"`
			PersistenceRequired int       `json:"PersistenceRequired"`
			ID                  string    `json:"id"`
			BatFlags            int       `json:"BatFlags"`
			UUID                []int     `json:"UUID"`
			Motion              []int     `json:"Motion"`
			Health              int       `json:"Health"`
			LeftHanded          int       `json:"LeftHanded"`
			Air                 int       `json:"Air"`
			OnGround            int       `json:"OnGround"`
			Rotation            []int     `json:"Rotation"`
			HandItems           []struct {
			} `json:"HandItems"`
			ArmorDropChances []float64 `json:"ArmorDropChances"`
			Pos              []float64 `json:"Pos"`
			Fire             int       `json:"Fire"`
			ArmorItems       []struct {
			} `json:"ArmorItems"`
			CanPickUpLoot int `json:"CanPickUpLoot"`
			HurtTime      int `json:"HurtTime"`
		} `json:"nbt"`
		BlockPos []int     `json:"blockPos"`
		Pos      []float64 `json:"pos"`
	} `json:"entities"`
	Blocks  string `json:"blocks"`
	Palette []struct {
		Name       string `json:"Name"`
		Properties struct {
			Axis string `json:"axis"`
		} `json:"Properties,omitempty"`
	} `json:"palette"`
	DataVersion int `json:"DataVersion"`
}
