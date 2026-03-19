package mcnbt

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"strings"
)

// StandardFormat represents a unified structure that can hold data from
// different Minecraft schematic formats (Litematica, WorldEdit, etc.)
// and Minecraft world saves.
type StandardFormat struct {
	// Metadata about the schematic or world save
	Metadata StandardMetadata `json:"metadata"`

	// Minecraft version information
	DataVersion int `json:"dataVersion"`
	Version     int `json:"version"`

	// Dimensions of the structure
	Size StandardSize `json:"size"`

	// Position/offset information
	Position StandardPosition `json:"position"`

	// Block data
	Blocks []StandardBlock `json:"blocks"`

	// Palette data
	Palette map[int]StandardPalette `json:"palette"`

	// Original format type
	OriginalFormat string `json:"originalFormat"`

	// Extra format-specific data that should be preserved during round-trips
	Extra map[string]interface{} `json:"extra,omitempty"`
}

type StandardMetadata struct {
	// Basic information
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`

	// Time information
	TimeCreated  int64 `json:"timeCreated"`
	TimeModified int64 `json:"timeModified"`

	// Size and volume information
	TotalBlocks int `json:"totalBlocks"`
	TotalVolume int `json:"totalVolume"`

	// Preview image if available
	PreviewImageData []int `json:"previewImageData,omitempty"`
}

type StandardSize struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type StandardPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// StandardBlock represents a block, entity, or tile entity in the standard format
type StandardBlock struct {
	// Type of the object (block, entity, or tile entity)
	Type string `json:"type,omitempty"`

	// ID for entities and tile entities
	ID string `json:"id,omitempty"`

	// Position of the block/tile entity (integer coordinates)
	Position StandardBlockPosition `json:"position"`

	// Entity rotation
	Rotation StandardRotation `json:"rotation,omitempty"`

	// Entity motion/velocity
	Motion StandardMotion `json:"motion,omitempty"`

	// State/ID of the block in the palette
	State int `json:"state,omitempty"`

	// NBT data for the block/entity/tile entity (if any)
	NBT interface{} `json:"nbt,omitempty"`
}

type StandardBlockPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type StandardRotation struct {
	Yaw   float64 `json:"yaw,omitempty"`
	Pitch float64 `json:"pitch,omitempty"`
}

type StandardMotion struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
	Z float64 `json:"z,omitempty"`
}

// StandardPalette represents a block type in the palette
type StandardPalette struct {
	// Block name (e.g., "minecraft:stone")
	Name string `json:"name"`

	// Block properties (if any)
	Properties map[string]string `json:"properties,omitempty"`
}

// ConvertToStandard converts any supported format to the StandardFormat
func ConvertToStandard(data interface{}) (*StandardFormat, error) {
	// Handle *interface{} type which comes from DecodeAny in decoder.go
	if ptr, ok := data.(*interface{}); ok {
		// Dereference the pointer to get the actual value
		return ConvertToStandard(*ptr)
	}

	// Try to identify the format based on the structure of the data
	switch v := data.(type) {
	case *LitematicaNBT:
		return convertLitematicaToStandard(v)
	case *WorldEditNBT:
		return convertWorldEditToStandard(v)
	case *CreateNBT:
		return convertCreateToStandard(v)
	case *StandardFormat:
		// Already in standard format
		return v, nil
	case map[string]interface{}:
		// Helper function to convert map to a specific format
		convertMapToFormat := func(formatType string, dest interface{}, formatDetector func(map[string]interface{}) bool) (*StandardFormat, error) {
			if formatDetector(v) {
				jsonData, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal data to JSON for %s format: %w", formatType, err)
				}
				if err := json.Unmarshal(jsonData, dest); err != nil {
					return nil, fmt.Errorf("failed to unmarshal data to %s format: %w", formatType, err)
				}

				// Use type switch to call the appropriate conversion function
				switch typedDest := dest.(type) {
				case *LitematicaNBT:
					return convertLitematicaToStandard(typedDest)
				case *WorldEditNBT:
					return convertWorldEditToStandard(typedDest)
				case *CreateNBT:
					return convertCreateToStandard(typedDest)
				default:
					return nil, fmt.Errorf("unexpected destination type for %s format", formatType)
				}
			}
			return nil, nil
		}

		// Define format detectors
		isLitematica := func(m map[string]interface{}) bool {
			_, hasMetadata := m["Metadata"]
			_, hasRegions := m["Regions"]
			return hasMetadata && hasRegions
		}

		isWorldEdit := func(m map[string]interface{}) bool {
			_, hasBlockData := m["BlockData"]
			_, hasPalette := m["Palette"]
			return hasBlockData && hasPalette
		}

		isCreate := func(m map[string]interface{}) bool {
			_, hasBlocks := m["blocks"]
			_, hasPalette := m["palette"]
			return hasBlocks && hasPalette
		}

		// Try each format
		if result, err := convertMapToFormat("Litematica", &LitematicaNBT{}, isLitematica); err != nil {
			return nil, err
		} else if result != nil {
			return result, nil
		}

		if result, err := convertMapToFormat("WorldEdit", &WorldEditNBT{}, isWorldEdit); err != nil {
			return nil, err
		} else if result != nil {
			return result, nil
		}

		if result, err := convertMapToFormat("Create", &CreateNBT{}, isCreate); err != nil {
			return nil, err
		} else if result != nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("unsupported format or unable to identify format")
}

// ConvertFromStandard converts a StandardFormat to the specified format
func ConvertFromStandard(standard *StandardFormat, format string) (interface{}, error) {
	switch format {
	case "standard":
		return standard, nil
	case "json":
		return standard, nil
	case "litematica":
		return convertStandardToLitematica(standard)
	case "worldedit":
		return convertStandardToWorldEdit(standard)
	case "create":
		return convertStandardToCreate(standard)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// convertLitematicaToStandard converts a LitematicaNBT to StandardFormat
func convertLitematicaToStandard(litematica *LitematicaNBT) (*StandardFormat, error) {
	if litematica == nil {
		return nil, fmt.Errorf("litematica data is nil")
	}

	sf := &StandardFormat{
		OriginalFormat: "litematica",
		DataVersion:    int(litematica.MinecraftDataVersion),
		Version:        int(litematica.Version),
	}

	// Set metadata
	sf.Metadata.Name = litematica.Metadata.Name
	sf.Metadata.Author = litematica.Metadata.Author
	sf.Metadata.Description = litematica.Metadata.Description
	sf.Metadata.TimeCreated = litematica.Metadata.TimeCreated
	sf.Metadata.TimeModified = litematica.Metadata.TimeModified
	sf.Metadata.TotalBlocks = int(litematica.Metadata.TotalBlocks)
	sf.Metadata.TotalVolume = int(litematica.Metadata.TotalVolume)

	// Convert []int32 preview image data to []int
	if litematica.Metadata.PreviewImageData != nil {
		sf.Metadata.PreviewImageData = make([]int, len(litematica.Metadata.PreviewImageData))
		for i, v := range litematica.Metadata.PreviewImageData {
			sf.Metadata.PreviewImageData[i] = int(v)
		}
	}

	if len(litematica.Regions) == 0 {
		return nil, fmt.Errorf("no regions found in litematica file")
	}

	// Extract the first region
	var region LitematicaRegion
	for _, r := range litematica.Regions {
		region = r
		break
	}

	// Handle negative sizes (Litematica uses negative sizes to indicate direction)
	sizeX := abs(int(region.Size.X))
	sizeY := abs(int(region.Size.Y))
	sizeZ := abs(int(region.Size.Z))

	sf.Size.X = sizeX
	sf.Size.Y = sizeY
	sf.Size.Z = sizeZ

	sf.Position.X = int(region.Position.X)
	sf.Position.Y = int(region.Position.Y)
	sf.Position.Z = int(region.Position.Z)

	// Convert palette
	sf.Palette = make(map[int]StandardPalette, len(region.BlockStatePalette))
	for i, palette := range region.BlockStatePalette {
		props := palette.Properties
		if props == nil {
			props = make(map[string]string)
		}
		sf.Palette[i] = StandardPalette{
			Name:       palette.Name,
			Properties: props,
		}
	}

	// Decode the packed BlockStates int64 array
	totalVolume := sizeX * sizeY * sizeZ
	paletteSize := len(region.BlockStatePalette)

	// Calculate bits per entry
	bitsPerEntry := 2 // minimum 2
	if paletteSize > 0 {
		b := bits.Len(uint(paletteSize - 1))
		if b > bitsPerEntry {
			bitsPerEntry = b
		}
	}

	// BlockStates is now directly []int64
	longs := region.BlockStates

	// Unpack palette indices from the long array
	// Litematica: entries do NOT cross long boundaries
	entriesPerLong := 64 / bitsPerEntry
	mask := int64((1 << bitsPerEntry) - 1)

	paletteIndices := make([]int, totalVolume)
	for i := 0; i < totalVolume; i++ {
		longIndex := i / entriesPerLong
		bitOffset := (i % entriesPerLong) * bitsPerEntry

		if longIndex < len(longs) {
			paletteIndices[i] = int((longs[longIndex] >> bitOffset) & mask)
		}
	}

	// Build a map of tile entity positions for merging
	tileEntityMap := make(map[[3]int]LitematicaTileEntity)
	for _, te := range region.TileEntities {
		key := [3]int{int(te.X), int(te.Y), int(te.Z)}
		tileEntityMap[key] = te
	}

	// Convert XZY-ordered indices to blocks with positions
	// Litematica order: iterate X, then Z, then Y (innermost)
	sf.Blocks = make([]StandardBlock, 0, totalVolume)
	idx := 0
	for y := 0; y < sizeY; y++ {
		for z := 0; z < sizeZ; z++ {
			for x := 0; x < sizeX; x++ {
				if idx >= len(paletteIndices) {
					break
				}
				paletteIdx := paletteIndices[idx]
				idx++

				block := StandardBlock{
					Type:  "block",
					State: paletteIdx,
					Position: StandardBlockPosition{
						X: float64(x),
						Y: float64(y),
						Z: float64(z),
					},
				}

				// Check if there's a tile entity at this position
				key := [3]int{x, y, z}
				if te, ok := tileEntityMap[key]; ok {
					block.Type = "block_entity"
					block.ID = te.Id
					// Build NBT from tile entity fields
					nbtData := make(map[string]interface{})
					nbtData["id"] = te.Id
					nbtData["x"] = int(te.X)
					nbtData["y"] = int(te.Y)
					nbtData["z"] = int(te.Z)
					if len(te.Items) > 0 {
						nbtData["Items"] = te.Items
					}
					block.NBT = nbtData
				}

				// Set the block ID from palette
				if p, ok := sf.Palette[paletteIdx]; ok {
					if block.ID == "" {
						block.ID = p.Name
					}
				}

				sf.Blocks = append(sf.Blocks, block)
			}
		}
	}

	return sf, nil
}

// Helper function to get absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// convertWorldEditToStandard converts a WorldEditNBT to StandardFormat
func convertWorldEditToStandard(worldEdit *WorldEditNBT) (*StandardFormat, error) {
	if worldEdit == nil {
		return nil, fmt.Errorf("worldEdit data is nil")
	}

	width := int(worldEdit.Width)
	height := int(worldEdit.Height)
	length := int(worldEdit.Length)

	sf := &StandardFormat{
		OriginalFormat: "worldedit",
		DataVersion:    int(worldEdit.DataVersion),
		Version:        int(worldEdit.Version),
	}

	sf.Size.X = width
	sf.Size.Y = height
	sf.Size.Z = length

	if len(worldEdit.Offset) >= 3 {
		sf.Position.X = int(worldEdit.Offset[0])
		sf.Position.Y = int(worldEdit.Offset[1])
		sf.Position.Z = int(worldEdit.Offset[2])
	}

	// Convert palette — WorldEdit palette maps "name[props]" → index
	// The map VALUES are the palette indices, not iteration order
	sf.Palette = make(map[int]StandardPalette, len(worldEdit.Palette))
	for name, paletteIndex := range worldEdit.Palette {
		blockName, properties := parseWorldEditBlockName(name)
		sf.Palette[int(paletteIndex)] = StandardPalette{
			Name:       blockName,
			Properties: properties,
		}
	}

	// Decode the varint-encoded BlockData byte array
	// WorldEdit BlockData is a varint-encoded stream iterated in YZX order
	totalVolume := width * height * length
	paletteIndices := make([]int, 0, totalVolume)

	blockDataBytes := worldEdit.BlockData
	offset := 0
	for offset < len(blockDataBytes) && len(paletteIndices) < totalVolume {
		value, bytesRead := readVarint(blockDataBytes, offset)
		if bytesRead == 0 {
			break
		}
		offset += bytesRead
		paletteIndices = append(paletteIndices, value)
	}

	// Build a map of block entity positions for merging
	blockEntityMap := make(map[[3]int]map[string]interface{})
	for _, be := range worldEdit.BlockEntities {
		x, y, z := extractBlockEntityPosition(be)
		key := [3]int{int(x), int(y), int(z)}
		blockEntityMap[key] = be
	}

	// Convert YZX-ordered indices to blocks with positions
	sf.Blocks = make([]StandardBlock, 0, totalVolume)
	idx := 0
	for y := 0; y < height; y++ {
		for z := 0; z < length; z++ {
			for x := 0; x < width; x++ {
				if idx >= len(paletteIndices) {
					break
				}
				paletteIdx := paletteIndices[idx]
				idx++

				block := StandardBlock{
					Type:  "block",
					State: paletteIdx,
					Position: StandardBlockPosition{
						X: float64(x),
						Y: float64(y),
						Z: float64(z),
					},
				}

				// Check if there's a block entity at this position
				key := [3]int{x, y, z}
				if be, ok := blockEntityMap[key]; ok {
					block.Type = "block_entity"
					if id, ok := be["Id"].(string); ok {
						block.ID = id
					}
					block.NBT = be
				}

				// Set the block ID from palette
				if p, ok := sf.Palette[paletteIdx]; ok {
					if block.ID == "" {
						block.ID = p.Name
					}
				}

				sf.Blocks = append(sf.Blocks, block)
			}
		}
	}

	return sf, nil
}

// parseWorldEditBlockName parses "minecraft:block[prop1=val1,prop2=val2]" into name and properties
func parseWorldEditBlockName(name string) (string, map[string]string) {
	nameAndProps := strings.SplitN(name, "[", 2)
	blockName := nameAndProps[0]
	properties := make(map[string]string)

	if len(nameAndProps) > 1 {
		propsStr := strings.TrimSuffix(nameAndProps[1], "]")
		props := strings.Split(propsStr, ",")
		for _, prop := range props {
			kv := strings.SplitN(prop, "=", 2)
			if len(kv) == 2 {
				properties[kv[0]] = kv[1]
			}
		}
	}

	return blockName, properties
}

// readVarint reads a varint from a byte slice at the given offset.
// Returns the decoded value and the number of bytes consumed.
func readVarint(data []byte, offset int) (int, int) {
	result := 0
	shift := 0
	bytesRead := 0

	for offset < len(data) {
		b := data[offset]
		offset++
		bytesRead++

		result |= int(b&0x7F) << shift
		shift += 7

		if b&0x80 == 0 {
			break
		}
	}

	return result, bytesRead
}

// Helper function to extract position from a block entity
func extractBlockEntityPosition(blockEntity map[string]any) (x, y, z float64) {
	if vals, ok := blockEntity["Pos"].([]interface{}); ok && len(vals) >= 3 {
		x, _ = toFloat64(vals[0])
		y, _ = toFloat64(vals[1])
		z, _ = toFloat64(vals[2])
		return
	}
	// Try individual x/y/z fields
	if v, ok := blockEntity["x"]; ok {
		x, _ = toFloat64(v)
	}
	if v, ok := blockEntity["y"]; ok {
		y, _ = toFloat64(v)
	}
	if v, ok := blockEntity["z"]; ok {
		z, _ = toFloat64(v)
	}
	return
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	}
	return 0, false
}

// convertCreateToStandard converts a CreateNBT (vanilla structure format) to StandardFormat
func convertCreateToStandard(create *CreateNBT) (*StandardFormat, error) {
	if create == nil {
		return nil, fmt.Errorf("create data is nil")
	}

	sf := &StandardFormat{
		OriginalFormat: "create",
		DataVersion:    int(create.DataVersion),
		Extra:          make(map[string]interface{}),
	}

	// Preserve mod-specific data versions
	if create.RailwaysDataVersion != 0 {
		sf.Extra["Railways_DataVersion"] = create.RailwaysDataVersion
	}

	// Set size
	if len(create.Size) >= 3 {
		sf.Size.X = int(create.Size[0])
		sf.Size.Y = int(create.Size[1])
		sf.Size.Z = int(create.Size[2])
	}

	// Convert palette — Properties is now map[string]string
	sf.Palette = make(map[int]StandardPalette, len(create.Palette))
	for i, palette := range create.Palette {
		props := palette.Properties
		if props == nil {
			props = make(map[string]string)
		}
		sf.Palette[i] = StandardPalette{
			Name:       palette.Name,
			Properties: props,
		}
	}

	// Build a map of tile entity positions for merging
	tileEntityMap := make(map[[3]int32]CreateTileEntity)
	for _, te := range create.TileEntities {
		if len(te.Pos) >= 3 {
			key := [3]int32{te.Pos[0], te.Pos[1], te.Pos[2]}
			tileEntityMap[key] = te
		}
	}

	// Process blocks
	sf.Blocks = make([]StandardBlock, 0, len(create.Blocks))
	for _, block := range create.Blocks {
		if len(block.Pos) < 3 {
			continue
		}

		sb := StandardBlock{
			Type:  "block",
			State: int(block.State),
			Position: StandardBlockPosition{
				X: float64(block.Pos[0]),
				Y: float64(block.Pos[1]),
				Z: float64(block.Pos[2]),
			},
		}

		// Set the block ID from palette
		if p, ok := sf.Palette[int(block.State)]; ok {
			sb.ID = p.Name
		}

		// Handle NBT from the block itself
		if block.Nbt != nil {
			sb.NBT = block.Nbt
		}

		// Check if there's a tile entity at this position
		key := [3]int32{block.Pos[0], block.Pos[1], block.Pos[2]}
		if te, ok := tileEntityMap[key]; ok {
			sb.Type = "block_entity"
			if idVal, ok := te.NBT["id"].(string); ok {
				sb.ID = idVal
			}
			sb.NBT = te.NBT
			delete(tileEntityMap, key) // mark as consumed
		}

		sf.Blocks = append(sf.Blocks, sb)
	}

	// Add any remaining tile entities that weren't matched to blocks
	for _, te := range tileEntityMap {
		if len(te.Pos) < 3 {
			continue
		}
		id := "unknown"
		if idVal, ok := te.NBT["id"].(string); ok {
			id = idVal
		}
		sb := StandardBlock{
			Type: "block_entity",
			ID:   id,
			Position: StandardBlockPosition{
				X: float64(te.Pos[0]),
				Y: float64(te.Pos[1]),
				Z: float64(te.Pos[2]),
			},
			NBT: te.NBT,
		}
		sf.Blocks = append(sf.Blocks, sb)
	}

	// Convert entities
	for _, entity := range create.Entities {
		if len(entity.Pos) < 3 {
			continue
		}

		entityBlock := StandardBlock{
			Type: "entity",
			ID:   entity.Nbt.ID,
			Position: StandardBlockPosition{
				X: entity.Pos[0],
				Y: entity.Pos[1],
				Z: entity.Pos[2],
			},
		}

		if len(entity.Nbt.Rotation) >= 2 {
			entityBlock.Rotation = StandardRotation{
				Yaw:   float64(entity.Nbt.Rotation[0]),
				Pitch: float64(entity.Nbt.Rotation[1]),
			}
		}

		if len(entity.Nbt.Motion) >= 3 {
			entityBlock.Motion = StandardMotion{
				X: float64(entity.Nbt.Motion[0]),
				Y: float64(entity.Nbt.Motion[1]),
				Z: float64(entity.Nbt.Motion[2]),
			}
		}

		sf.Blocks = append(sf.Blocks, entityBlock)
	}

	return sf, nil
}

// convertStandardToLitematica converts a StandardFormat to LitematicaNBT
func convertStandardToLitematica(standard *StandardFormat) (*LitematicaNBT, error) {
	litematica := &LitematicaNBT{}

	litematica.MinecraftDataVersion = int32(standard.DataVersion)
	litematica.Version = int32(standard.Version)

	litematica.Metadata.Name = standard.Metadata.Name
	litematica.Metadata.Author = standard.Metadata.Author
	litematica.Metadata.Description = standard.Metadata.Description
	litematica.Metadata.TimeCreated = standard.Metadata.TimeCreated
	litematica.Metadata.TimeModified = standard.Metadata.TimeModified
	litematica.Metadata.TotalBlocks = int32(standard.Metadata.TotalBlocks)
	litematica.Metadata.TotalVolume = int32(standard.Metadata.TotalVolume)

	// Convert []int preview image data to []int32
	if standard.Metadata.PreviewImageData != nil {
		litematica.Metadata.PreviewImageData = make([]int32, len(standard.Metadata.PreviewImageData))
		for i, v := range standard.Metadata.PreviewImageData {
			litematica.Metadata.PreviewImageData[i] = int32(v)
		}
	}

	litematica.Metadata.EnclosingSize.X = int32(standard.Size.X)
	litematica.Metadata.EnclosingSize.Y = int32(standard.Size.Y)
	litematica.Metadata.EnclosingSize.Z = int32(standard.Size.Z)
	litematica.Metadata.RegionCount = 1

	region := LitematicaRegion{}

	region.Size.X = int32(standard.Size.X)
	region.Size.Y = int32(standard.Size.Y)
	region.Size.Z = int32(standard.Size.Z)
	region.Position.X = int32(standard.Position.X)
	region.Position.Y = int32(standard.Position.Y)
	region.Position.Z = int32(standard.Position.Z)

	// Convert palette
	region.BlockStatePalette = make([]LitematicaBlockStatePalette, len(standard.Palette))
	for i, palette := range standard.Palette {
		region.BlockStatePalette[i] = LitematicaBlockStatePalette{
			Name:       palette.Name,
			Properties: palette.Properties,
		}
	}

	// Build a 3D grid of palette indices from standard blocks
	sizeX := standard.Size.X
	sizeY := standard.Size.Y
	sizeZ := standard.Size.Z
	totalVolume := sizeX * sizeY * sizeZ

	grid := make([]int, totalVolume)
	var tileEntities []LitematicaTileEntity
	var entities []LitematicaEntity

	for _, block := range standard.Blocks {
		if block.Type == "entity" {
			e := LitematicaEntity{
				ID:       block.ID,
				Pos:      []float64{block.Position.X, block.Position.Y, block.Position.Z},
				Rotation: []float32{float32(block.Rotation.Yaw), float32(block.Rotation.Pitch)},
				Motion:   []float64{block.Motion.X, block.Motion.Y, block.Motion.Z},
			}
			entities = append(entities, e)
			continue
		}

		x, y, z := int(block.Position.X), int(block.Position.Y), int(block.Position.Z)
		if x < 0 || x >= sizeX || y < 0 || y >= sizeY || z < 0 || z >= sizeZ {
			continue
		}

		// YZX order for the flat grid
		idx := y*sizeZ*sizeX + z*sizeX + x
		if idx >= 0 && idx < totalVolume {
			grid[idx] = block.State
		}

		// Collect tile entities
		if block.Type == "block_entity" && block.NBT != nil {
			te := LitematicaTileEntity{
				Id: block.ID,
				X:  int32(x),
				Y:  int32(y),
				Z:  int32(z),
			}
			tileEntities = append(tileEntities, te)
		}
	}

	// Pack palette indices into int64 long array
	// Entries do NOT cross long boundaries in Litematica
	paletteSize := len(region.BlockStatePalette)
	bitsPerEntry := 2 // minimum
	if paletteSize > 0 {
		b := bits.Len(uint(paletteSize - 1))
		if b > bitsPerEntry {
			bitsPerEntry = b
		}
	}

	entriesPerLong := 64 / bitsPerEntry
	numLongs := (totalVolume + entriesPerLong - 1) / entriesPerLong
	mask := int64((1 << bitsPerEntry) - 1)

	packedLongs := make([]int64, numLongs)

	// Pack in YZX order (same order as the grid)
	for i := 0; i < totalVolume; i++ {
		longIndex := i / entriesPerLong
		bitOffset := (i % entriesPerLong) * bitsPerEntry
		state := int64(grid[i]) & mask
		packedLongs[longIndex] |= state << bitOffset
	}

	// BlockStates is now directly []int64
	region.BlockStates = packedLongs

	region.TileEntities = tileEntities
	region.Entities = entities

	litematica.Regions = map[string]LitematicaRegion{"main": region}

	return litematica, nil
}

// convertStandardToWorldEdit converts a StandardFormat to WorldEditNBT
func convertStandardToWorldEdit(standard *StandardFormat) (*WorldEditNBT, error) {
	worldEdit := &WorldEditNBT{}

	worldEdit.DataVersion = int32(standard.DataVersion)
	worldEdit.Version = int32(standard.Version)

	worldEdit.Width = int16(standard.Size.X)
	worldEdit.Height = int16(standard.Size.Y)
	worldEdit.Length = int16(standard.Size.Z)

	worldEdit.Offset = []int32{int32(standard.Position.X), int32(standard.Position.Y), int32(standard.Position.Z)}

	worldEdit.Metadata.WEOffsetX = int32(standard.Position.X)
	worldEdit.Metadata.WEOffsetY = int32(standard.Position.Y)
	worldEdit.Metadata.WEOffsetZ = int32(standard.Position.Z)

	width := standard.Size.X
	height := standard.Size.Y
	length := standard.Size.Z

	// Convert palette — WorldEdit uses "name[props]" → index
	worldEdit.Palette = make(map[string]int32)
	for i, palette := range standard.Palette {
		blockName := palette.Name
		if len(palette.Properties) > 0 {
			blockName += "["
			first := true
			for key, value := range palette.Properties {
				if !first {
					blockName += ","
				}
				blockName += key + "=" + value
				first = false
			}
			blockName += "]"
		}
		worldEdit.Palette[blockName] = int32(i)
	}
	worldEdit.PaletteMax = int32(len(standard.Palette))

	// Build a 3D grid of palette indices
	totalVolume := width * height * length
	grid := make([]int, totalVolume)

	var blockEntities []map[string]any

	for _, block := range standard.Blocks {
		if block.Type == "entity" {
			continue
		}

		x, y, z := int(block.Position.X), int(block.Position.Y), int(block.Position.Z)
		if x < 0 || x >= width || y < 0 || y >= height || z < 0 || z >= length {
			continue
		}

		idx := y*length*width + z*width + x
		if idx >= 0 && idx < totalVolume {
			grid[idx] = block.State
		}

		// Collect block entities
		if block.Type == "block_entity" {
			be := map[string]any{
				"Id": block.ID,
				"Pos": []int{
					int(block.Position.X),
					int(block.Position.Y),
					int(block.Position.Z),
				},
			}
			if nbtMap, ok := block.NBT.(map[string]interface{}); ok {
				for key, value := range nbtMap {
					be[key] = value
				}
			}
			blockEntities = append(blockEntities, be)
		}
	}

	// Encode block data as varint byte array in YZX order
	var blockData []byte
	for i := 0; i < totalVolume; i++ {
		blockData = append(blockData, writeVarint(grid[i])...)
	}
	worldEdit.BlockData = blockData
	worldEdit.BlockEntities = blockEntities

	return worldEdit, nil
}

// writeVarint encodes an integer as a varint byte sequence
func writeVarint(value int) []byte {
	var buf []byte
	uval := uint32(value)
	for {
		b := byte(uval & 0x7F)
		uval >>= 7
		if uval != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if uval == 0 {
			break
		}
	}
	return buf
}

// convertStandardToCreate converts a StandardFormat to CreateNBT (vanilla structure format)
func convertStandardToCreate(standard *StandardFormat) (*CreateNBT, error) {
	create := &CreateNBT{}

	create.DataVersion = int32(standard.DataVersion)
	create.Size = []int32{int32(standard.Size.X), int32(standard.Size.Y), int32(standard.Size.Z)}

	// Restore mod-specific data versions
	if v, ok := standard.Extra["Railways_DataVersion"]; ok {
		switch val := v.(type) {
		case int32:
			create.RailwaysDataVersion = val
		case int:
			create.RailwaysDataVersion = int32(val)
		case int64:
			create.RailwaysDataVersion = int32(val)
		}
	}

	// Convert palette — Properties is now map[string]string
	create.Palette = make([]CreatePalette, len(standard.Palette))
	for i, palette := range standard.Palette {
		props := palette.Properties
		if props == nil {
			props = make(map[string]string)
		}
		create.Palette[i] = CreatePalette{
			Name:       palette.Name,
			Properties: props,
		}
	}

	// Convert blocks
	var blocks []CreateBlock
	var entities []CreateEntity
	var tileEntities []CreateTileEntity

	for _, block := range standard.Blocks {
		if block.Type == "entity" {
			e := CreateEntity{
				Pos: []float64{block.Position.X, block.Position.Y, block.Position.Z},
				Nbt: CreateEntityNbt{
					ID: block.ID,
				},
			}
			if block.Rotation.Yaw != 0 || block.Rotation.Pitch != 0 {
				e.Nbt.Rotation = []float32{float32(block.Rotation.Yaw), float32(block.Rotation.Pitch)}
			}
			if block.Motion.X != 0 || block.Motion.Y != 0 || block.Motion.Z != 0 {
				e.Nbt.Motion = []float64{block.Motion.X, block.Motion.Y, block.Motion.Z}
			}
			entities = append(entities, e)
			continue
		}

		cb := CreateBlock{
			Pos:   []int32{int32(block.Position.X), int32(block.Position.Y), int32(block.Position.Z)},
			State: int32(block.State),
			Nbt:   block.NBT,
		}
		blocks = append(blocks, cb)

		// Collect tile entities
		if block.Type == "block_entity" && block.NBT != nil {
			te := CreateTileEntity{
				Pos: []int32{int32(block.Position.X), int32(block.Position.Y), int32(block.Position.Z)},
			}
			if nbtMap, ok := block.NBT.(map[string]interface{}); ok {
				te.NBT = nbtMap
			} else {
				te.NBT = map[string]interface{}{"id": block.ID}
			}
			tileEntities = append(tileEntities, te)
		}
	}

	create.Blocks = blocks
	create.Entities = entities
	create.TileEntities = tileEntities

	return create, nil
}
