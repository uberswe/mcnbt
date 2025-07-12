package mcnbt

// WorldEditMetadata represents the metadata of a WorldEdit schematic
type WorldEditMetadata struct {
	WEOffsetX int `json:"WEOffsetX"`
	WEOffsetY int `json:"WEOffsetY"`
	WEOffsetZ int `json:"WEOffsetZ"`
}

// WorldEditNBT represents a WorldEdit schematic
type WorldEditNBT struct {
	BlockData     string            `json:"BlockData"`
	BlockEntities []map[string]any  `json:"BlockEntities"`
	DataVersion   int               `json:"DataVersion"`
	Height        int               `json:"Height"`
	Length        int               `json:"Length"`
	Metadata      WorldEditMetadata `json:"Metadata"`
	Offset        []int             `json:"Offset"`
	Palette       map[string]int    `json:"Palette"`
	PaletteMax    int               `json:"PaletteMax"`
	Version       int               `json:"Version"`
	Width         int               `json:"Width"`
}
