package mcnbt

import (
	"encoding/json"
	"fmt"
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
	Palette map[int]StandardPalette `json:"Palette"`

	// Original format type
	OriginalFormat string `json:"originalFormat"`
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
	// Handle *interface{} type which comes from decodeAny in decoder.go
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
		// Create format requires special handling to ensure blocks are preserved
		create, err := convertStandardToCreate(standard)
		if err != nil {
			return nil, err
		}

		// Ensure the blocks field is not empty
		if len(create.Blocks) == 0 && len(standard.Blocks) > 0 {
			// If blocks field is empty but there should be blocks, create them
			create.Blocks = make([]interface{}, len(standard.Blocks))
			for i, block := range standard.Blocks {
				// Create a map for each block
				blockMap := make(map[string]interface{})

				// Set position, preserving the original position
				blockMap["pos"] = []int{
					int(block.Position.X) - standard.Position.X, // Adjust X position
					int(block.Position.Y) - standard.Position.Y, // Adjust Y position
					int(block.Position.Z) - standard.Position.Z, // Adjust Z position
				}

				// Set state (palette index)
				blockMap["state"] = block.State

				// Add NBT data if available
				if block.NBT != nil {
					blockMap["nbt"] = block.NBT
				}

				// Add the block to the list
				create.Blocks[i] = blockMap
			}
		}

		return create, nil
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
		// Set original format
		OriginalFormat: "litematica",

		// Set version information
		DataVersion: litematica.MinecraftDataVersion,
		Version:     litematica.Version,

		// Initialize slices
		Blocks: make([]StandardBlock, 0),
	}

	// Set metadata
	sf.Metadata.Name = litematica.Metadata.Name
	sf.Metadata.Author = litematica.Metadata.Author
	sf.Metadata.Description = litematica.Metadata.Description
	sf.Metadata.TimeCreated = litematica.Metadata.TimeCreated
	sf.Metadata.TimeModified = litematica.Metadata.TimeModified
	sf.Metadata.TotalBlocks = litematica.Metadata.TotalBlocks
	sf.Metadata.TotalVolume = litematica.Metadata.TotalVolume
	sf.Metadata.PreviewImageData = litematica.Metadata.PreviewImageData

	// Get the first region from the Regions map
	if len(litematica.Regions) == 0 {
		return nil, fmt.Errorf("no regions found in litematica file")
	}

	// Extract the first region
	var region LitematicaRegion
	for _, r := range litematica.Regions {
		region = r
		break
	}

	// Set size and position
	sf.Size.X = region.Size.X
	// Handle negative Y size in Litematica format
	sf.Size.Y = abs(region.Size.Y) // Use abs function to handle negative Y size
	sf.Size.Z = region.Size.Z

	sf.Position.X = region.Position.X
	sf.Position.Y = region.Position.Y
	sf.Position.Z = region.Position.Z

	// Convert palette
	sf.Palette = make(map[int]StandardPalette, len(region.BlockStatePalette))
	for i, palette := range region.BlockStatePalette {
		sf.Palette[i] = StandardPalette{
			Name:       palette.Name,
			Properties: make(map[string]string),
		}
		// Add properties if they exist
		if palette.Properties.Snowy != "" {
			sf.Palette[i].Properties["snowy"] = palette.Properties.Snowy
		}
	}

	// Create a map to store block positions and states for efficient lookup
	blockMap := make(map[string]int)

	// Process blocks if BlockStates array is not empty
	if len(region.BlockStates) > 0 {
		// Calculate a safe capacity for the blocks slice
		// Ensure all dimensions are positive
		sizeX, sizeY, sizeZ := abs(region.Size.X), abs(region.Size.Y), abs(region.Size.Z)

		// Calculate total volume (safely)
		totalVolume := region.Size.X * region.Size.Y * region.Size.Z

		// Use a reasonable default capacity if dimensions are too large
		var capacity int
		if sizeX > 0 && sizeY > 0 && sizeZ > 0 &&
			// Check if multiplication would overflow
			sizeX <= 1000 && sizeY <= 1000 && sizeZ <= 1000 {
			safeVolume := sizeX * sizeY * sizeZ
			// Limit the capacity to a reasonable value
			if safeVolume > 1000000 {
				capacity = 1000000 // Cap at 1 million blocks
			} else {
				capacity = safeVolume / 2 // Estimate that ~50% of blocks are non-air
			}
		} else {
			// Use a modest default capacity
			capacity = 10000
		}

		sf.Blocks = make([]StandardBlock, 0, capacity)

		// Process BlockStates array
		for i := 0; i < totalVolume && i < len(region.BlockStates); i++ {
			// Calculate the 3D position from the 1D index
			x := i % region.Size.X
			y := (i / region.Size.X) % region.Size.Y
			z := i / (region.Size.X * region.Size.Y)

			// Get the palette index for this position
			paletteIndex, ok := getPaletteIndex(region.BlockStates[i])
			if !ok {
				continue // Skip if we can't determine the palette index
			}

			// Skip air blocks (usually palette index 0)
			if paletteIndex == 0 {
				continue
			}

			// Create and add a StandardBlock
			block := StandardBlock{
				Position: StandardBlockPosition{
					X: float64(x),
					Y: float64(y),
					Z: float64(z),
				},
				State: paletteIndex,
			}

			sf.Blocks = append(sf.Blocks, block)

			// Store the block in the map for tile entity lookup
			blockMap[fmt.Sprintf("%d,%d,%d", x, y, z)] = len(sf.Blocks) - 1
		}
	}

	// If no blocks were found in BlockStates, create blocks from tile entities as a fallback
	if len(sf.Blocks) == 0 && len(region.TileEntities) > 0 {
		// Use the first palette entry for all blocks (usually not air)
		paletteIndex := 1
		if len(sf.Palette) <= 1 {
			// If the palette is empty or only has air, add a default block
			sf.Palette[len(sf.Palette)] = StandardPalette{
				Name:       "minecraft:stone",
				Properties: make(map[string]string),
			}
			paletteIndex = 1
		}

		// Pre-allocate blocks slice
		sf.Blocks = make([]StandardBlock, 0, len(region.TileEntities))

		// Create blocks for each tile entity
		for _, tileEntity := range region.TileEntities {
			// Create and add a StandardBlock
			block := StandardBlock{
				Position: struct {
					X float64 `json:"x"`
					Y float64 `json:"y"`
					Z float64 `json:"z"`
				}{
					X: float64(tileEntity.X),
					Y: float64(tileEntity.Y),
					Z: float64(tileEntity.Z),
				},
				State: paletteIndex,
			}

			sf.Blocks = append(sf.Blocks, block)

			// Store the block in the map for tile entity lookup
			blockMap[fmt.Sprintf("%d,%d,%d", tileEntity.X, tileEntity.Y, tileEntity.Z)] = len(sf.Blocks) - 1
		}
	}

	// Associate tile entities with blocks
	for _, tileEntity := range region.TileEntities {
		key := fmt.Sprintf("%d,%d,%d", tileEntity.X, tileEntity.Y, tileEntity.Z)
		if blockIndex, ok := blockMap[key]; ok && blockIndex < len(sf.Blocks) {
			// Create a map for the tile entity data
			teData := make(map[string]interface{})
			teData["x"] = tileEntity.X
			teData["y"] = tileEntity.Y
			teData["z"] = tileEntity.Z

			// Add any other tile entity data
			if len(tileEntity.Items) > 0 {
				teData["Items"] = tileEntity.Items
			}
			if len(tileEntity.CookingTimes) > 0 {
				teData["CookingTimes"] = tileEntity.CookingTimes
			}
			if len(tileEntity.CookingTotalTimes) > 0 {
				teData["CookingTotalTimes"] = tileEntity.CookingTotalTimes
			}
			if len(tileEntity.Bees) > 0 {
				teData["Bees"] = tileEntity.Bees
			}

			// Set the NBT data for the block
			sf.Blocks[blockIndex].NBT = teData
		}
	}

	// Convert entities
	for _, entity := range region.Entities {
		// Skip entities with invalid position data
		if len(entity.Pos) < 3 || len(entity.Rotation) < 2 || len(entity.Motion) < 3 {
			continue
		}

		// Create a StandardBlock for the entity
		entityBlock := StandardBlock{
			Type: "entity",
			ID:   entity.ID,
			Position: StandardBlockPosition{
				X: entity.Pos[0],
				Y: entity.Pos[1],
				Z: entity.Pos[2],
			},
			Rotation: struct {
				Yaw   float64 `json:"yaw,omitempty"`
				Pitch float64 `json:"pitch,omitempty"`
			}{
				Yaw:   entity.Rotation[0],
				Pitch: entity.Rotation[1],
			},
			Motion: struct {
				X float64 `json:"x,omitempty"`
				Y float64 `json:"y,omitempty"`
				Z float64 `json:"z,omitempty"`
			}{
				X: entity.Motion[0],
				Y: entity.Motion[1],
				Z: entity.Motion[2],
			},
		}
		sf.Blocks = append(sf.Blocks, entityBlock)
	}

	// Convert tile entities that don't have associated blocks
	for _, tileEntity := range region.TileEntities {
		// Check if this tile entity is already associated with a block
		key := fmt.Sprintf("%d,%d,%d", tileEntity.X, tileEntity.Y, tileEntity.Z)
		if _, ok := blockMap[key]; !ok {
			// Create a StandardBlock for the tile entity
			tileEntityBlock := StandardBlock{
				Type: "tile_entity",
				ID:   "unknown", // The ID is not provided in the struct
				Position: StandardBlockPosition{
					X: float64(tileEntity.X),
					Y: float64(tileEntity.Y),
					Z: float64(tileEntity.Z),
				},
			}
			sf.Blocks = append(sf.Blocks, tileEntityBlock)
		}
	}

	return sf, nil
}

// Helper function to get palette index from various types
func getPaletteIndex(value interface{}) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
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

	sf := &StandardFormat{
		// Set original format
		OriginalFormat: "worldedit",

		// Set version information
		DataVersion: worldEdit.DataVersion,
		Version:     worldEdit.Version,

		// Initialize slices
		Blocks: make([]StandardBlock, 0),
	}

	// Set size
	sf.Size.X = worldEdit.Width
	sf.Size.Y = worldEdit.Height
	sf.Size.Z = worldEdit.Length

	// Set position/offset
	if len(worldEdit.Offset) >= 3 {
		sf.Position.X = worldEdit.Offset[0]
		sf.Position.Y = worldEdit.Offset[1]
		sf.Position.Z = worldEdit.Offset[2]
	}

	// Convert palette
	sf.Palette = make(map[int]StandardPalette, len(worldEdit.Palette))
	i := 0
	for name := range worldEdit.Palette {
		// Parse the name and properties
		// In WorldEdit, the block name might include properties in the format "minecraft:block[property1=value1,property2=value2]"
		nameAndProps := strings.SplitN(name, "[", 2)
		blockName := nameAndProps[0]
		properties := make(map[string]string)

		if len(nameAndProps) > 1 {
			// Remove the closing bracket
			propsStr := strings.TrimSuffix(nameAndProps[1], "]")
			// Split by comma to get individual properties
			props := strings.Split(propsStr, ",")
			for _, prop := range props {
				// Split by equal sign to get property name and value
				kv := strings.SplitN(prop, "=", 2)
				if len(kv) == 2 {
					properties[kv[0]] = kv[1]
				}
			}
		}

		sf.Palette[i] = StandardPalette{
			Name:       blockName,
			Properties: properties,
		}
		i++
	}

	// Create a map to store block positions and states for efficient lookup
	blockMap := make(map[string]int)

	// Process blocks if BlockData is not empty
	if len(worldEdit.BlockData) > 0 {
		// Get the total volume of the schematic
		totalVolume := worldEdit.Width * worldEdit.Height * worldEdit.Length

		// Pre-allocate blocks slice with estimated capacity
		estimatedNonAirBlocks := totalVolume / 2 // Estimate that ~50% of blocks are non-air
		sf.Blocks = make([]StandardBlock, 0, estimatedNonAirBlocks)

		// Decode the BlockData
		// BlockData is typically a base64-encoded byte array where each byte represents a palette index
		// For simplicity, we'll assume it's already decoded and just iterate through the characters
		for i := 0; i < len(worldEdit.BlockData) && i < totalVolume; i++ {
			// Calculate the 3D position from the 1D index
			// WorldEdit uses YZX order
			x := i % worldEdit.Width
			z := (i / worldEdit.Width) % worldEdit.Length
			y := i / (worldEdit.Width * worldEdit.Length)

			// Get the palette index for this position
			paletteIndex := int(worldEdit.BlockData[i])

			// Skip air blocks (usually palette index 0)
			if paletteIndex == 0 {
				continue
			}

			// Create and add a StandardBlock
			block := StandardBlock{
				Position: StandardBlockPosition{
					X: float64(x),
					Y: float64(y),
					Z: float64(z),
				},
				State: paletteIndex,
			}

			sf.Blocks = append(sf.Blocks, block)

			// Store the block in the map for block entity lookup
			blockMap[fmt.Sprintf("%d,%d,%d", x, y, z)] = len(sf.Blocks) - 1
		}
	}

	// If no blocks were found in BlockData but there are block entities, create blocks from them
	if len(sf.Blocks) == 0 && len(worldEdit.BlockEntities) > 0 {
		// Use the first palette entry for all blocks (usually not air)
		paletteIndex := 1
		if len(sf.Palette) <= 1 {
			// If the palette is empty or only has air, add a default block
			sf.Palette[len(sf.Palette)] = StandardPalette{
				Name:       "minecraft:stone",
				Properties: make(map[string]string),
			}
			paletteIndex = 1
		}

		// Pre-allocate blocks slice
		sf.Blocks = make([]StandardBlock, 0, len(worldEdit.BlockEntities))

		// Create blocks for each block entity
		for _, blockEntity := range worldEdit.BlockEntities {
			// Extract position
			x, y, z := extractBlockEntityPosition(blockEntity)

			// Create and add a StandardBlock
			block := StandardBlock{
				Position: StandardBlockPosition{
					X: float64(x),
					Y: float64(y),
					Z: float64(z),
				},
				State: paletteIndex,
			}

			sf.Blocks = append(sf.Blocks, block)

			// Store the block in the map for block entity lookup
			blockMap[fmt.Sprintf("%d,%d,%d", x, y, z)] = len(sf.Blocks) - 1
		}
	}

	// Associate block entities with blocks
	for _, blockEntity := range worldEdit.BlockEntities {
		// Extract position
		x, y, z := extractBlockEntityPosition(blockEntity)

		key := fmt.Sprintf("%d,%d,%d", x, y, z)
		if blockIndex, ok := blockMap[key]; ok && blockIndex < len(sf.Blocks) {
			// Set the NBT data for the block
			sf.Blocks[blockIndex].NBT = blockEntity
		}
	}

	// Convert block entities to tile entities that don't have associated blocks
	for _, blockEntity := range worldEdit.BlockEntities {
		// Extract position
		x, y, z := extractBlockEntityPosition(blockEntity)

		// Check if this tile entity is already associated with a block
		key := fmt.Sprintf("%d,%d,%d", x, y, z)
		if _, ok := blockMap[key]; !ok {
			// Extract ID
			id := "unknown"
			if idVal, ok := blockEntity["id"].(string); ok {
				id = idVal
			}

			// Create a StandardBlock for the tile entity
			tileEntityBlock := StandardBlock{
				Type: "tile_entity",
				ID:   id,
				Position: StandardBlockPosition{
					X: float64(x),
					Y: float64(y),
					Z: float64(z),
				},
				NBT: blockEntity,
			}
			sf.Blocks = append(sf.Blocks, tileEntityBlock)
		}
	}

	return sf, nil
}

// Helper function to extract position from a block entity
func extractBlockEntityPosition(blockEntity map[string]any) (x, y, z int) {
	if xVal, ok := blockEntity["x"].(float64); ok {
		x = int(xVal)
	}
	if yVal, ok := blockEntity["y"].(float64); ok {
		y = int(yVal)
	}
	if zVal, ok := blockEntity["z"].(float64); ok {
		z = int(zVal)
	}
	return
}

// convertCreateToStandard converts a CreateNBT to StandardFormat
func convertCreateToStandard(create *CreateNBT) (*StandardFormat, error) {
	if create == nil {
		return nil, fmt.Errorf("create data is nil")
	}

	sf := &StandardFormat{
		// Set original format
		OriginalFormat: "create",

		// Set version information
		DataVersion: create.DataVersion,
		Version:     0, // Create format doesn't have a version field

		// Initialize slices
		Blocks: make([]StandardBlock, 0),
	}

	// Set size
	if len(create.Size) >= 3 {
		sf.Size.X = create.Size[0]
		sf.Size.Y = create.Size[1]
		sf.Size.Z = create.Size[2]
	} else {
		// Default size if not provided
		sf.Size.X = 1
		sf.Size.Y = 1
		sf.Size.Z = 1
	}

	// Find minimum position from blocks (if available)
	minX, minY, minZ := findMinPosition(create.Blocks)

	// Set position
	sf.Position.X = minX
	sf.Position.Y = minY
	sf.Position.Z = minZ

	// Convert palette
	sf.Palette = make(map[int]StandardPalette, len(create.Palette))
	for i, palette := range create.Palette {
		sf.Palette[i] = StandardPalette{
			Name:       palette.Name,
			Properties: make(map[string]string),
		}
		// Add properties if they exist
		if palette.Properties.Axis != "" {
			sf.Palette[i].Properties["axis"] = palette.Properties.Axis
		}
	}

	// Pre-allocate blocks slice with estimated capacity
	estimatedBlocks := len(create.Blocks)
	if estimatedBlocks == 0 && len(create.TileEntities) > 0 {
		estimatedBlocks = len(create.TileEntities)
	}
	sf.Blocks = make([]StandardBlock, 0, estimatedBlocks)

	// Process blocks if available
	if len(create.Blocks) > 0 {
		// Process each block
		for _, block := range create.Blocks {
			// Skip nil blocks
			if block == nil {
				continue
			}

			// Get block as map
			blockMap := getBlockAsMap(block)
			if blockMap == nil {
				continue
			}

			// Extract position and state
			posX, posY, posZ := extractBlockPosition(blockMap)
			state := extractBlockState(blockMap)

			// Create and add a StandardBlock
			sb := StandardBlock{
				Position: StandardBlockPosition{
					X: float64(posX + sf.Position.X),
					Y: float64(posY + sf.Position.Y),
					Z: float64(posZ + sf.Position.Z),
				},
				State: state,
				NBT:   extractBlockNBT(blockMap),
			}

			sf.Blocks = append(sf.Blocks, sb)
		}
	}

	// If no blocks were found but there are tile entities, create blocks from them
	if len(sf.Blocks) == 0 && len(create.TileEntities) > 0 {
		// Use the first palette entry for all blocks (usually not air)
		paletteIndex := 1
		if len(sf.Palette) <= 1 {
			// If the palette is empty or only has air, add a default block
			sf.Palette[len(sf.Palette)] = StandardPalette{
				Name:       "minecraft:stone",
				Properties: make(map[string]string),
			}
			paletteIndex = 1
		}

		// Create blocks for each tile entity
		for _, tileEntity := range create.TileEntities {
			// Skip tile entities with invalid position
			if len(tileEntity.Pos) < 3 {
				continue
			}

			// Create and add a StandardBlock
			block := StandardBlock{
				Position: StandardBlockPosition{
					X: float64(tileEntity.Pos[0] + sf.Position.X),
					Y: float64(tileEntity.Pos[1] + sf.Position.Y),
					Z: float64(tileEntity.Pos[2] + sf.Position.Z),
				},
				State: paletteIndex,
			}

			// Add NBT data if available
			if len(tileEntity.NBT) > 0 {
				block.NBT = tileEntity.NBT
			}

			sf.Blocks = append(sf.Blocks, block)
		}
	}

	// Convert entities
	for _, entity := range create.Entities {
		// Skip entities with invalid data
		if len(entity.Pos) < 3 || len(entity.Nbt.Rotation) < 2 || len(entity.Nbt.Motion) < 3 {
			continue
		}

		// Create a StandardBlock for the entity
		entityBlock := StandardBlock{
			Type: "entity",
			ID:   entity.Nbt.ID,
			Position: StandardBlockPosition{
				X: entity.Pos[0],
				Y: entity.Pos[1],
				Z: entity.Pos[2],
			},
			Rotation: struct {
				Yaw   float64 `json:"yaw,omitempty"`
				Pitch float64 `json:"pitch,omitempty"`
			}{
				Yaw:   float64(entity.Nbt.Rotation[0]),
				Pitch: float64(entity.Nbt.Rotation[1]),
			},
			Motion: struct {
				X float64 `json:"x,omitempty"`
				Y float64 `json:"y,omitempty"`
				Z float64 `json:"z,omitempty"`
			}{
				X: float64(entity.Nbt.Motion[0]),
				Y: float64(entity.Nbt.Motion[1]),
				Z: float64(entity.Nbt.Motion[2]),
			},
		}
		sf.Blocks = append(sf.Blocks, entityBlock)
	}

	// Convert tile entities that don't already have associated blocks
	for _, tileEntity := range create.TileEntities {
		// Skip tile entities with invalid position
		if len(tileEntity.Pos) < 3 {
			continue
		}

		// Extract ID from NBT data
		id := "unknown"
		if idVal, ok := tileEntity.NBT["id"].(string); ok {
			id = idVal
		}

		// Check if this position already has a block
		posX := tileEntity.Pos[0] + sf.Position.X
		posY := tileEntity.Pos[1] + sf.Position.Y
		posZ := tileEntity.Pos[2] + sf.Position.Z

		// Check if we need to create a new block for this tile entity
		needNewBlock := true
		for i, block := range sf.Blocks {
			if int(block.Position.X) == posX && int(block.Position.Y) == posY && int(block.Position.Z) == posZ {
				// This position already has a block, just update its NBT data
				sf.Blocks[i].Type = "block_with_tile_entity"
				sf.Blocks[i].ID = id
				sf.Blocks[i].NBT = tileEntity.NBT
				needNewBlock = false
				break
			}
		}

		if needNewBlock {
			// Create a new block for this tile entity
			tileEntityBlock := StandardBlock{
				Type: "tile_entity",
				ID:   id,
				Position: StandardBlockPosition{
					X: float64(posX),
					Y: float64(posY),
					Z: float64(posZ),
				},
				NBT: tileEntity.NBT,
			}
			sf.Blocks = append(sf.Blocks, tileEntityBlock)
		}
	}

	return sf, nil
}

// Helper function to find the minimum position from blocks
func findMinPosition(blocks []interface{}) (minX, minY, minZ int) {
	// Default to 0,0,0
	minX, minY, minZ = 0, 0, 0

	if len(blocks) == 0 {
		return
	}

	// Initialize with maximum values
	minX, minY, minZ = 1000000, 1000000, 1000000
	foundValidPosition := false

	// Find the minimum position
	for _, block := range blocks {
		// Skip nil blocks or non-map blocks
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract position
		pos, ok := blockMap["pos"].([]interface{})
		if !ok || len(pos) < 3 {
			continue
		}

		// Update minimum position
		if x, ok := pos[0].(float64); ok && int(x) < minX {
			minX = int(x)
			foundValidPosition = true
		}
		if y, ok := pos[1].(float64); ok && int(y) < minY {
			minY = int(y)
			foundValidPosition = true
		}
		if z, ok := pos[2].(float64); ok && int(z) < minZ {
			minZ = int(z)
			foundValidPosition = true
		}
	}

	// If no valid positions were found, reset to 0
	if !foundValidPosition {
		minX, minY, minZ = 0, 0, 0
	}

	return
}

// Helper function to convert a block to a map
func getBlockAsMap(block interface{}) map[string]interface{} {
	// Try direct type assertion
	if blockMap, ok := block.(map[string]interface{}); ok {
		return blockMap
	}

	// Try JSON conversion
	jsonData, err := json.Marshal(block)
	if err != nil {
		return nil
	}

	var tempMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &tempMap); err != nil {
		return nil
	}

	return tempMap
}

// Helper function to extract position from a block
func extractBlockPosition(blockMap map[string]interface{}) (x, y, z int) {
	if pos, ok := blockMap["pos"]; ok {
		// Try as []interface{}
		if posArray, ok := pos.([]interface{}); ok && len(posArray) >= 3 {
			if xVal, ok := posArray[0].(float64); ok {
				x = int(xVal)
			}
			if yVal, ok := posArray[1].(float64); ok {
				y = int(yVal)
			}
			if zVal, ok := posArray[2].(float64); ok {
				z = int(zVal)
			}
		} else if posArray, ok := pos.([]int); ok && len(posArray) >= 3 {
			// Try as []int
			x = posArray[0]
			y = posArray[1]
			z = posArray[2]
		} else if posArray, ok := pos.([]float64); ok && len(posArray) >= 3 {
			// Try as []float64
			x = int(posArray[0])
			y = int(posArray[1])
			z = int(posArray[2])
		}
	}
	return
}

// Helper function to extract state from a block
func extractBlockState(blockMap map[string]interface{}) int {
	var state int
	if stateVal, ok := blockMap["state"]; ok {
		if stateFloat, ok := stateVal.(float64); ok {
			state = int(stateFloat)
		} else if stateInt, ok := stateVal.(int); ok {
			state = stateInt
		}
	}
	return state
}

// Helper function to extract NBT data from a block
func extractBlockNBT(blockMap map[string]interface{}) interface{} {
	if nbt, ok := blockMap["nbt"]; ok {
		if nbtMap, ok := nbt.(map[string]interface{}); ok {
			return nbtMap
		}

		// Try JSON conversion
		jsonData, err := json.Marshal(nbt)
		if err == nil {
			var tempMap map[string]interface{}
			if err := json.Unmarshal(jsonData, &tempMap); err == nil {
				return tempMap
			}
		}
	}
	return nil
}

// convertStandardToLitematica converts a StandardFormat to LitematicaNBT
func convertStandardToLitematica(standard *StandardFormat) (*LitematicaNBT, error) {
	litematica := &LitematicaNBT{}

	// Set version information
	litematica.MinecraftDataVersion = standard.DataVersion
	litematica.Version = standard.Version

	// Set metadata
	litematica.Metadata.Name = standard.Metadata.Name
	litematica.Metadata.Author = standard.Metadata.Author
	litematica.Metadata.Description = standard.Metadata.Description
	litematica.Metadata.TimeCreated = standard.Metadata.TimeCreated
	litematica.Metadata.TimeModified = standard.Metadata.TimeModified
	litematica.Metadata.TotalBlocks = standard.Metadata.TotalBlocks
	litematica.Metadata.TotalVolume = standard.Metadata.TotalVolume
	litematica.Metadata.PreviewImageData = standard.Metadata.PreviewImageData
	litematica.Metadata.EnclosingSize.X = standard.Size.X
	litematica.Metadata.EnclosingSize.Y = standard.Size.Y
	litematica.Metadata.EnclosingSize.Z = standard.Size.Z
	litematica.Metadata.RegionCount = 1

	// Create a region
	region := LitematicaRegion{}

	// Set region size and position
	region.Size.X = standard.Size.X
	region.Size.Y = standard.Size.Y
	region.Size.Z = standard.Size.Z
	region.Position.X = standard.Position.X
	region.Position.Y = standard.Position.Y
	region.Position.Z = standard.Position.Z

	// Convert palette
	region.BlockStatePalette = make([]LitematicaBlockStatePalette, len(standard.Palette))
	for i, palette := range standard.Palette {
		region.BlockStatePalette[i].Name = palette.Name

		// Convert properties
		// This is a simplified example; in a real implementation,
		// you would need to handle all possible properties
		if snowy, ok := palette.Properties["snowy"]; ok {
			region.BlockStatePalette[i].Properties.Snowy = snowy
		}
	}

	// Convert blocks to BlockStates array
	// Create a 3D grid to represent the blocks
	grid := make([][][]int, region.Size.X)
	for x := range grid {
		grid[x] = make([][]int, region.Size.Y)
		for y := range grid[x] {
			grid[x][y] = make([]int, region.Size.Z)
			// Initialize with air (palette index 0)
			for z := range grid[x][y] {
				grid[x][y][z] = 0
			}
		}
	}

	// Fill the grid with block data
	for _, block := range standard.Blocks {
		x, y, z := int(block.Position.X), int(block.Position.Y), int(block.Position.Z)

		// Skip blocks outside the region bounds
		if x < 0 || int(x) >= region.Size.X || y < 0 || int(y) >= region.Size.Y || z < 0 || int(z) >= region.Size.Z {
			continue
		}

		// Set the palette index for this position
		grid[x][y][z] = block.State
	}

	// Convert the 3D grid to a 1D array
	// Litematica uses XZY order
	blockStates := make([]interface{}, region.Size.X*region.Size.Y*region.Size.Z)
	index := 0
	for x := 0; x < region.Size.X; x++ {
		for z := 0; z < region.Size.Z; z++ {
			for y := 0; y < region.Size.Y; y++ {
				// Get the palette index for this position
				paletteIndex := grid[x][y][z]

				// Set the value in the BlockStates array
				if index < len(blockStates) {
					blockStates[index] = paletteIndex
				}

				index++
			}
		}
	}

	// Set the BlockStates
	region.BlockStates = blockStates

	// Convert entities
	region.Entities = make([]LitematicaEntity, len(standard.Blocks))
	for i, entity := range standard.Blocks {
		region.Entities[i].ID = entity.ID
		region.Entities[i].Pos = []float64{entity.Position.X, entity.Position.Y, entity.Position.Z}
		region.Entities[i].Rotation = []float64{entity.Rotation.Yaw, entity.Rotation.Pitch}
		region.Entities[i].Motion = []float64{entity.Motion.X, entity.Motion.Y, entity.Motion.Z}
	}

	// We skip litematica tile entities when converting from standard

	// Set the region
	litematica.Regions = make(map[string]LitematicaRegion)
	litematica.Regions["main"] = region

	return litematica, nil
}

// convertStandardToWorldEdit converts a StandardFormat to WorldEditNBT
func convertStandardToWorldEdit(standard *StandardFormat) (*WorldEditNBT, error) {
	worldEdit := &WorldEditNBT{}

	// Set version information
	worldEdit.DataVersion = standard.DataVersion
	worldEdit.Version = standard.Version

	// Set size
	worldEdit.Width = standard.Size.X
	worldEdit.Height = standard.Size.Y
	worldEdit.Length = standard.Size.Z

	// Set offset
	worldEdit.Offset = []int{standard.Position.X, standard.Position.Y, standard.Position.Z}

	// Set metadata
	worldEdit.Metadata.WEOffsetX = standard.Position.X
	worldEdit.Metadata.WEOffsetY = standard.Position.Y
	worldEdit.Metadata.WEOffsetZ = standard.Position.Z

	// Convert palette
	worldEdit.Palette = make(map[string]int)
	for i, palette := range standard.Palette {
		// In WorldEdit, the block name might include properties in the format "minecraft:block[property1=value1,property2=value2]"
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
		worldEdit.Palette[blockName] = i
	}
	worldEdit.PaletteMax = len(standard.Palette)

	// Convert blocks to BlockData
	// Create a 3D grid to represent the blocks
	grid := make([][][]int, worldEdit.Width)
	for x := range grid {
		grid[x] = make([][]int, worldEdit.Height)
		for y := range grid[x] {
			grid[x][y] = make([]int, worldEdit.Length)
			// Initialize with air (palette index 0)
			for z := range grid[x][y] {
				grid[x][y][z] = 0
			}
		}
	}

	// Fill the grid with block data
	for _, block := range standard.Blocks {
		x, y, z := int(block.Position.X), int(block.Position.Y), int(block.Position.Z)

		// Skip blocks outside the schematic bounds
		if x < 0 || x >= worldEdit.Width || y < 0 || y >= worldEdit.Height || z < 0 || z >= worldEdit.Length {
			continue
		}

		// Set the palette index for this position
		grid[x][y][z] = block.State
	}

	// Convert the 3D grid to a 1D array in YZX order
	blockData := make([]byte, worldEdit.Width*worldEdit.Height*worldEdit.Length)
	index := 0
	for y := 0; y < worldEdit.Height; y++ {
		for z := 0; z < worldEdit.Length; z++ {
			for x := 0; x < worldEdit.Width; x++ {
				// Get the palette index for this position
				paletteIndex := grid[x][y][z]

				// Set the value in the BlockData array
				// In a real implementation, you would need to properly encode the BlockData
				if index < len(blockData) {
					blockData[index] = byte(paletteIndex)
				}

				index++
			}
		}
	}

	// TODO figure out if this is the right way to Set the BlockData
	worldEdit.BlockData = string(blockData)

	// Convert block entities
	worldEdit.BlockEntities = make([]map[string]any, len(standard.Blocks))
	for i, tileEntity := range standard.Blocks {
		worldEdit.BlockEntities[i] = make(map[string]any)
		worldEdit.BlockEntities[i]["id"] = tileEntity.ID
		worldEdit.BlockEntities[i]["x"] = tileEntity.Position.X
		worldEdit.BlockEntities[i]["y"] = tileEntity.Position.Y
		worldEdit.BlockEntities[i]["z"] = tileEntity.Position.Z
		if tileEntity.NBT != nil {
			// Add NBT data if available
			if nbtMap, ok := tileEntity.NBT.(map[string]interface{}); ok {
				for key, value := range nbtMap {
					worldEdit.BlockEntities[i][key] = value
				}
			}
		}
	}

	return worldEdit, nil
}

// convertStandardToCreate converts a StandardFormat to CreateNBT
func convertStandardToCreate(standard *StandardFormat) (*CreateNBT, error) {
	create := &CreateNBT{}

	// Set version information
	create.DataVersion = standard.DataVersion

	// Set size
	create.Size = []int{standard.Size.X, standard.Size.Y, standard.Size.Z}

	// Convert palette
	create.Palette = make([]CreatePalette, len(standard.Palette))
	for i, palette := range standard.Palette {
		create.Palette[i].Name = palette.Name
		// This is a simplified example; in a real implementation,
		// you would need to handle all possible properties
		if axis, ok := palette.Properties["axis"]; ok {
			create.Palette[i].Properties.Axis = axis
		}
	}

	// Convert blocks from standard format to Create format
	create.Blocks = make([]interface{}, len(standard.Blocks))

	// Iterate through the standard blocks and convert them to Create blocks
	for i, block := range standard.Blocks {
		// Create a map for each block
		blockMap := make(map[string]interface{})

		// Set position, preserving the original position
		blockMap["pos"] = []int{
			int(block.Position.X) - standard.Position.X, // Adjust X position
			int(block.Position.Y) - standard.Position.Y, // Adjust Y position
			int(block.Position.Z) - standard.Position.Z, // Adjust Z position
		}

		// Set state (palette index)
		blockMap["state"] = block.State

		// Add NBT data if available
		if block.NBT != nil {
			blockMap["nbt"] = block.NBT
		}

		// Add the block to the list
		create.Blocks[i] = blockMap
	}

	// Skip entities and tile entities when converting to world edit

	return create, nil
}
