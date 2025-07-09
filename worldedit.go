package mcnbt

type WorldEditNBT struct {
	BlockData     string           `json:"BlockData"`
	BlockEntities []map[string]any `json:"BlockEntities"`
	DataVersion   int              `json:"DataVersion"`
	Height        int              `json:"Height"`
	Length        int              `json:"Length"`
	Metadata      struct {
		WEOffsetX int `json:"WEOffsetX"`
		WEOffsetY int `json:"WEOffsetY"`
		WEOffsetZ int `json:"WEOffsetZ"`
	} `json:"Metadata"`
	Offset     []int          `json:"Offset"`
	Palette    map[string]int `json:"Palette"`
	PaletteMax int            `json:"PaletteMax"`
	Version    int            `json:"Version"`
	Width      int            `json:"Width"`
}
