package mcnbt

// Coordinate represents a 3D coordinate with X, Y, Z values
type Coordinate struct {
	X int32 `json:"x" nbt:"x"`
	Y int32 `json:"y" nbt:"y"`
	Z int32 `json:"z" nbt:"z"`
}
