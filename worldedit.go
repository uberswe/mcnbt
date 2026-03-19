package mcnbt

// WorldEditMetadata represents the metadata of a WorldEdit schematic
type WorldEditMetadata struct {
	WEOffsetX int32 `json:"WEOffsetX" nbt:"WEOffsetX"`
	WEOffsetY int32 `json:"WEOffsetY" nbt:"WEOffsetY"`
	WEOffsetZ int32 `json:"WEOffsetZ" nbt:"WEOffsetZ"`
}

// WorldEditNBT represents a WorldEdit schematic
type WorldEditNBT struct {
	BlockData     []byte            `json:"BlockData" nbt:"BlockData"`
	BlockEntities []map[string]any  `json:"BlockEntities" nbt:"BlockEntities"`
	DataVersion   int32             `json:"DataVersion" nbt:"DataVersion"`
	Height        int16             `json:"Height" nbt:"Height"`
	Length        int16             `json:"Length" nbt:"Length"`
	Metadata      WorldEditMetadata `json:"Metadata" nbt:"Metadata"`
	Offset        []int32           `json:"Offset" nbt:"Offset"`
	Palette       map[string]int32  `json:"Palette" nbt:"Palette"`
	PaletteMax    int32             `json:"PaletteMax" nbt:"PaletteMax"`
	Version       int32             `json:"Version" nbt:"Version"`
	Width         int16             `json:"Width" nbt:"Width"`
}
