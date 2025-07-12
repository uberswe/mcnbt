package mcnbt

// Coordinate represents a 3D coordinate with X, Y, Z values
type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// BlockStateProperty represents the properties of a block state
type BlockStateProperty struct {
	Snowy string `json:"snowy"`
}

// BlockStatePalette represents a block state in the palette
type BlockStatePalette struct {
	Name       string             `json:"Name"`
	Properties BlockStateProperty `json:"Properties,omitempty"`
}
