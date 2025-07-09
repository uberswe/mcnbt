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
	Metadata struct {
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
	} `json:"metadata"`

	// Minecraft version information
	DataVersion int `json:"dataVersion"`
	Version     int `json:"version"`

	// Dimensions of the structure
	Size struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"size"`

	// Position/offset information
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"position"`

	// Block data
	Blocks []StandardBlock `json:"blocks"`

	// Block palette (mapping block names to IDs)
	Palette []StandardPalette `json:"palette"`

	// Entities in the structure
	Entities []Entity `json:"entities"`

	// Tile entities (block entities) in the structure
	TileEntities []TileEntity `json:"tileEntities"`

	// Original format type
	OriginalFormat string `json:"originalFormat"`

	// Raw data from the original format (for any format-specific data)
	RawData interface{} `json:"rawData,omitempty"`
}

// Entity represents a Minecraft entity
type Entity struct {
	// Entity type ID
	ID string `json:"id"`

	// Entity position
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"position"`

	// Entity rotation
	Rotation struct {
		Yaw   float64 `json:"yaw"`
		Pitch float64 `json:"pitch"`
	} `json:"rotation"`

	// Entity motion/velocity
	Motion struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"motion"`

	// Entity NBT data (for format-specific entity data)
	NBT interface{} `json:"nbt,omitempty"`
}

// TileEntity represents a Minecraft tile entity (block entity)
type TileEntity struct {
	// Tile entity type ID
	ID string `json:"id"`

	// Tile entity position
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"position"`

	// Tile entity NBT data (for format-specific tile entity data)
	NBT interface{} `json:"nbt,omitempty"`
}

// StandardBlock represents a block in the standard format
type StandardBlock struct {
	// Position of the block
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"position"`

	// State/ID of the block in the palette
	State int `json:"state"`

	// NBT data for the block (if any)
	NBT interface{} `json:"nbt,omitempty"`
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
	// Try to identify the format based on the structure of the data

	// Handle *interface{} type which comes from decodeAny in decoder.go
	if ptr, ok := data.(*interface{}); ok {
		// Dereference the pointer to get the actual value
		return ConvertToStandard(*ptr)
	}

	switch v := data.(type) {
	case *LitematicaNBT:
		return convertLitematicaToStandard(v)
	case *WorldEditNBT:
		return convertWorldEditToStandard(v)
	case *WorldSave:
		return convertWorldSaveToStandard(v)
	case *CreateNBT:
		return convertCreateToStandard(v)
	case *StandardFormat:
		// Already in standard format
		return v, nil
	case map[string]interface{}:
		// Try to identify the format based on the keys in the map
		if _, ok := v["Metadata"]; ok {
			if _, ok := v["Regions"]; ok {
				// It's likely a Litematica format
				litematica := &LitematicaNBT{}
				jsonData, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
				}
				if err := json.Unmarshal(jsonData, litematica); err != nil {
					return nil, fmt.Errorf("failed to unmarshal data to LitematicaNBT: %w", err)
				}
				return convertLitematicaToStandard(litematica)
			}
		}
		if _, ok := v["BlockData"]; ok {
			if _, ok := v["Palette"]; ok {
				// It's likely a WorldEdit format
				worldEdit := &WorldEditNBT{}
				jsonData, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
				}
				if err := json.Unmarshal(jsonData, worldEdit); err != nil {
					return nil, fmt.Errorf("failed to unmarshal data to WorldEditNBT: %w", err)
				}
				return convertWorldEditToStandard(worldEdit)
			}
		}
		if _, ok := v["level.dat"]; ok {
			if _, ok := v["regions"]; ok {
				// It's likely a Minecraft world save
				worldSave := &WorldSave{}
				jsonData, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
				}
				if err := json.Unmarshal(jsonData, worldSave); err != nil {
					return nil, fmt.Errorf("failed to unmarshal data to WorldSave: %w", err)
				}
				return convertWorldSaveToStandard(worldSave)
			}
		}
		if _, ok := v["blocks"]; ok {
			if _, ok := v["palette"]; ok {
				// It's likely a Create format
				create := &CreateNBT{}
				jsonData, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
				}
				if err := json.Unmarshal(jsonData, create); err != nil {
					return nil, fmt.Errorf("failed to unmarshal data to CreateNBT: %w", err)
				}
				return convertCreateToStandard(create)
			}
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
	case "worldsave":
		return convertStandardToWorldSave(standard)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// convertLitematicaToStandard converts a LitematicaNBT to StandardFormat
func convertLitematicaToStandard(litematica *LitematicaNBT) (*StandardFormat, error) {
	sf := &StandardFormat{}

	// Set original format
	sf.OriginalFormat = "litematica"

	// Set metadata
	sf.Metadata.Name = litematica.Metadata.Name
	sf.Metadata.Author = litematica.Metadata.Author
	sf.Metadata.Description = litematica.Metadata.Description
	sf.Metadata.TimeCreated = litematica.Metadata.TimeCreated
	sf.Metadata.TimeModified = litematica.Metadata.TimeModified
	sf.Metadata.TotalBlocks = litematica.Metadata.TotalBlocks
	sf.Metadata.TotalVolume = litematica.Metadata.TotalVolume
	sf.Metadata.PreviewImageData = litematica.Metadata.PreviewImageData

	// Set version information
	sf.DataVersion = litematica.MinecraftDataVersion
	sf.Version = litematica.Version

	// Set size and position
	// Note: In Litematica, there's a single region with a specific name
	// We need to access the first region, regardless of its name

	// Get the first region from the Regions map
	var region LitematicaRegion
	if len(litematica.Regions) > 0 {
		// Get the first region from the map
		for _, r := range litematica.Regions {
			region = r
			break
		}
	} else {
		return nil, fmt.Errorf("no regions found in litematica file")
	}

	sf.Size.X = region.Size.X
	sf.Size.Y = region.Size.Y
	sf.Size.Z = region.Size.Z

	sf.Position.X = region.Position.X
	sf.Position.Y = region.Position.Y
	sf.Position.Z = region.Position.Z

	// Convert palette
	sf.Palette = make([]StandardPalette, len(region.BlockStatePalette))
	for i, palette := range region.BlockStatePalette {
		sf.Palette[i] = StandardPalette{
			Name:       palette.Name,
			Properties: make(map[string]string),
		}
		// Add properties if they exist
		// This is a simplified example; in a real implementation,
		// you would need to handle all possible properties
		if palette.Properties.Snowy != "" {
			sf.Palette[i].Properties["snowy"] = palette.Properties.Snowy
		}
	}

	// Convert blocks
	// This is a simplified example; in a real implementation,
	// you would need to decode the BlockStates array to get the actual blocks
	sf.Blocks = []StandardBlock{}

	// Convert entities
	sf.Entities = []Entity{}
	for _, entity := range region.Entities {
		e := Entity{
			ID: entity.ID,
		}
		e.Position.X = entity.Pos[0]
		e.Position.Y = entity.Pos[1]
		e.Position.Z = entity.Pos[2]
		e.Rotation.Yaw = entity.Rotation[0]
		e.Rotation.Pitch = entity.Rotation[1]
		e.Motion.X = entity.Motion[0]
		e.Motion.Y = entity.Motion[1]
		e.Motion.Z = entity.Motion[2]
		sf.Entities = append(sf.Entities, e)
	}

	// Convert tile entities
	sf.TileEntities = []TileEntity{}
	for _, tileEntity := range region.TileEntities {
		te := TileEntity{
			ID: "unknown", // The ID is not provided in the struct
		}
		te.Position.X = tileEntity.X
		te.Position.Y = tileEntity.Y
		te.Position.Z = tileEntity.Z
		sf.TileEntities = append(sf.TileEntities, te)
	}

	// Store the original data
	sf.RawData = litematica

	return sf, nil
}

// convertWorldEditToStandard converts a WorldEditNBT to StandardFormat
func convertWorldEditToStandard(worldEdit *WorldEditNBT) (*StandardFormat, error) {
	sf := &StandardFormat{}

	// Set original format
	sf.OriginalFormat = "worldedit"

	// Set version information
	sf.DataVersion = worldEdit.DataVersion
	sf.Version = worldEdit.Version

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
	sf.Palette = make([]StandardPalette, len(worldEdit.Palette))
	i := 0
	for name, _ := range worldEdit.Palette {
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

	// Convert blocks
	// This is a simplified example; in a real implementation,
	// you would need to decode the BlockData string to get the actual blocks
	sf.Blocks = []StandardBlock{}

	// Convert block entities
	sf.TileEntities = []TileEntity{}
	for _, blockEntity := range worldEdit.BlockEntities {
		te := TileEntity{
			ID: "unknown", // The ID might be in the map
		}
		if x, ok := blockEntity["x"].(float64); ok {
			te.Position.X = int(x)
		}
		if y, ok := blockEntity["y"].(float64); ok {
			te.Position.Y = int(y)
		}
		if z, ok := blockEntity["z"].(float64); ok {
			te.Position.Z = int(z)
		}
		if id, ok := blockEntity["id"].(string); ok {
			te.ID = id
		}
		te.NBT = blockEntity
		sf.TileEntities = append(sf.TileEntities, te)
	}

	// Store the original data
	sf.RawData = worldEdit

	return sf, nil
}

// convertWorldSaveToStandard converts a WorldSave to StandardFormat
func convertWorldSaveToStandard(worldSave *WorldSave) (*StandardFormat, error) {
	sf := &StandardFormat{}

	// Set original format
	sf.OriginalFormat = "worldsave"

	// Set metadata
	sf.Metadata.Name = worldSave.Metadata.Name
	sf.Metadata.Author = worldSave.Metadata.Author
	sf.Metadata.Description = worldSave.Metadata.Description
	sf.Metadata.TimeCreated = worldSave.Metadata.TimeCreated
	sf.Metadata.TimeModified = worldSave.Metadata.TimeModified
	sf.Metadata.TotalBlocks = worldSave.Metadata.TotalBlocks
	sf.Metadata.TotalVolume = worldSave.Metadata.TotalVolume
	sf.Metadata.PreviewImageData = worldSave.Metadata.PreviewImageData

	// Set version information
	sf.DataVersion = worldSave.MinecraftDataVersion
	sf.Version = worldSave.Version

	// Set size and position
	// Note: This assumes there's only one region in the WorldSave
	// In a real implementation, you might need to handle multiple regions
	for _, region := range worldSave.Regions {
		sf.Size.X = region.Size.X
		sf.Size.Y = region.Size.Y
		sf.Size.Z = region.Size.Z

		sf.Position.X = region.Position.X
		sf.Position.Y = region.Position.Y
		sf.Position.Z = region.Position.Z

		// Convert palette
		sf.Palette = make([]StandardPalette, len(region.BlockStatePalette))
		for i, palette := range region.BlockStatePalette {
			sf.Palette[i] = StandardPalette{
				Name:       palette.Name,
				Properties: make(map[string]string),
			}
			// Add properties if they exist
			// This is a simplified example; in a real implementation,
			// you would need to handle all possible properties
			if palette.Properties.Snowy != "" {
				sf.Palette[i].Properties["snowy"] = palette.Properties.Snowy
			}
		}

		// Convert blocks
		// This is a simplified example; in a real implementation,
		// you would need to decode the BlockStates array to get the actual blocks
		sf.Blocks = []StandardBlock{}

		// Convert entities
		sf.Entities = []Entity{}
		for _, entity := range region.Entities {
			e := Entity{
				ID: entity.ID,
			}
			e.Position.X = entity.Pos[0]
			e.Position.Y = entity.Pos[1]
			e.Position.Z = entity.Pos[2]
			e.Rotation.Yaw = float64(entity.Rotation[0])
			e.Rotation.Pitch = float64(entity.Rotation[1])
			e.Motion.X = float64(entity.Motion[0])
			e.Motion.Y = float64(entity.Motion[1])
			e.Motion.Z = float64(entity.Motion[2])
			sf.Entities = append(sf.Entities, e)
		}

		// Convert tile entities
		sf.TileEntities = []TileEntity{}
		for _, tileEntity := range region.TileEntities {
			te := TileEntity{
				ID: "unknown", // The ID is not provided in the struct
			}
			te.Position.X = tileEntity.X
			te.Position.Y = tileEntity.Y
			te.Position.Z = tileEntity.Z
			sf.TileEntities = append(sf.TileEntities, te)
		}

		// We only process the first region for simplicity
		break
	}

	// Store the original data
	sf.RawData = worldSave

	return sf, nil
}

// convertCreateToStandard converts a CreateNBT to StandardFormat
func convertCreateToStandard(create *CreateNBT) (*StandardFormat, error) {
	sf := &StandardFormat{}

	// Set original format
	sf.OriginalFormat = "create"

	// Set version information
	sf.DataVersion = create.DataVersion
	sf.Version = 0 // Create format doesn't have a version field

	// Set size
	if len(create.Size) >= 3 {
		sf.Size.X = create.Size[0]
		sf.Size.Y = create.Size[1]
		sf.Size.Z = create.Size[2]
	}

	// Position is not provided in Create format
	sf.Position.X = 0
	sf.Position.Y = 0
	sf.Position.Z = 0

	// Convert palette
	sf.Palette = make([]StandardPalette, len(create.Palette))
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

	// Convert blocks
	// This is a simplified example; in a real implementation,
	// you would need to decode the Blocks string to get the actual blocks
	sf.Blocks = []StandardBlock{}

	// Convert entities
	sf.Entities = []Entity{}
	for _, entity := range create.Entities {
		e := Entity{
			ID: entity.Nbt.ID,
		}
		e.Position.X = entity.Pos[0]
		e.Position.Y = entity.Pos[1]
		e.Position.Z = entity.Pos[2]
		e.Rotation.Yaw = float64(entity.Nbt.Rotation[0])
		e.Rotation.Pitch = float64(entity.Nbt.Rotation[1])
		e.Motion.X = float64(entity.Nbt.Motion[0])
		e.Motion.Y = float64(entity.Nbt.Motion[1])
		e.Motion.Z = float64(entity.Nbt.Motion[2])
		sf.Entities = append(sf.Entities, e)
	}

	// Create format doesn't have tile entities
	sf.TileEntities = []TileEntity{}

	// Store the original data
	sf.RawData = create

	return sf, nil
}

// convertStandardToLitematica converts a StandardFormat to LitematicaNBT
func convertStandardToLitematica(standard *StandardFormat) (*LitematicaNBT, error) {
	// Skip implementation for now - we're focusing on the litematica to worldedit conversion
	return nil, fmt.Errorf("conversion from standard to litematica is not implemented yet")
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

	// Convert blocks
	// This is a simplified example; in a real implementation,
	// you would need to encode the blocks to the BlockData string
	worldEdit.BlockData = ""

	// Convert block entities
	worldEdit.BlockEntities = make([]map[string]any, len(standard.TileEntities))
	for i, tileEntity := range standard.TileEntities {
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
	create.Palette = make([]struct {
		Name       string `json:"Name"`
		Properties struct {
			Axis string `json:"axis"`
		} `json:"Properties,omitempty"`
	}, len(standard.Palette))
	for i, palette := range standard.Palette {
		create.Palette[i].Name = palette.Name
		// This is a simplified example; in a real implementation,
		// you would need to handle all possible properties
		if axis, ok := palette.Properties["axis"]; ok {
			create.Palette[i].Properties.Axis = axis
		}
	}

	// Convert blocks
	// This is a simplified example; in a real implementation,
	// you would need to encode the blocks to the Blocks string
	create.Blocks = ""

	// Convert entities
	create.Entities = make([]struct {
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
	}, len(standard.Entities))
	for i, entity := range standard.Entities {
		create.Entities[i].Nbt.ID = entity.ID
		create.Entities[i].Nbt.Motion = []int{int(entity.Motion.X), int(entity.Motion.Y), int(entity.Motion.Z)}
		create.Entities[i].Nbt.Rotation = []int{int(entity.Rotation.Yaw), int(entity.Rotation.Pitch)}
		create.Entities[i].Pos = []float64{entity.Position.X, entity.Position.Y, entity.Position.Z}
		create.Entities[i].BlockPos = []int{int(entity.Position.X), int(entity.Position.Y), int(entity.Position.Z)}
	}

	return create, nil
}

// convertStandardToWorldSave converts a StandardFormat to WorldSave
func convertStandardToWorldSave(standard *StandardFormat) (*WorldSave, error) {
	worldSave := &WorldSave{}

	// Set metadata
	worldSave.Metadata.Name = standard.Metadata.Name
	worldSave.Metadata.Author = standard.Metadata.Author
	worldSave.Metadata.Description = standard.Metadata.Description
	worldSave.Metadata.TimeCreated = standard.Metadata.TimeCreated
	worldSave.Metadata.TimeModified = standard.Metadata.TimeModified
	worldSave.Metadata.TotalBlocks = standard.Metadata.TotalBlocks
	worldSave.Metadata.TotalVolume = standard.Metadata.TotalVolume
	worldSave.Metadata.PreviewImageData = standard.Metadata.PreviewImageData
	worldSave.Metadata.EnclosingSize.X = standard.Size.X
	worldSave.Metadata.EnclosingSize.Y = standard.Size.Y
	worldSave.Metadata.EnclosingSize.Z = standard.Size.Z
	worldSave.Metadata.RegionCount = 1

	// Set version information
	worldSave.MinecraftDataVersion = standard.DataVersion
	worldSave.Version = standard.Version

	// Create a region
	region := struct {
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
			Motion              []int     `json:"Motion"`
			OnGround            int       `json:"OnGround"`
			PersistenceRequired int       `json:"PersistenceRequired"`
			PortalCooldown      int       `json:"PortalCooldown"`
			Pos                 []float64 `json:"Pos"`
			Rotation            []int     `json:"Rotation"`
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
	}{}

	// Set region size and position
	region.Size.X = standard.Size.X
	region.Size.Y = standard.Size.Y
	region.Size.Z = standard.Size.Z
	region.Position.X = standard.Position.X
	region.Position.Y = standard.Position.Y
	region.Position.Z = standard.Position.Z

	// Convert palette
	region.BlockStatePalette = make([]struct {
		Name       string `json:"Name"`
		Properties struct {
			Snowy string `json:"snowy"`
		} `json:"Properties,omitempty"`
	}, len(standard.Palette))
	for i, palette := range standard.Palette {
		region.BlockStatePalette[i].Name = palette.Name
		// This is a simplified example; in a real implementation,
		// you would need to handle all possible properties
		if snowy, ok := palette.Properties["snowy"]; ok {
			region.BlockStatePalette[i].Properties.Snowy = snowy
		}
	}

	// Convert blocks
	// This is a simplified example; in a real implementation,
	// you would need to encode the blocks to the BlockStates array
	region.BlockStates = []interface{}{}

	// Convert entities
	region.Entities = make([]struct {
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
		Motion              []int     `json:"Motion"`
		OnGround            int       `json:"OnGround"`
		PersistenceRequired int       `json:"PersistenceRequired"`
		PortalCooldown      int       `json:"PortalCooldown"`
		Pos                 []float64 `json:"Pos"`
		Rotation            []int     `json:"Rotation"`
		UUID                []int     `json:"UUID"`
		ID                  string    `json:"id"`
	}, len(standard.Entities))
	for i, entity := range standard.Entities {
		region.Entities[i].ID = entity.ID
		region.Entities[i].Pos = []float64{entity.Position.X, entity.Position.Y, entity.Position.Z}
		region.Entities[i].Rotation = []int{int(entity.Rotation.Yaw), int(entity.Rotation.Pitch)}
		region.Entities[i].Motion = []int{int(entity.Motion.X), int(entity.Motion.Y), int(entity.Motion.Z)}
	}

	// Convert tile entities
	region.TileEntities = make([]struct {
		Items             []interface{} `json:"Items,omitempty"`
		X                 int           `json:"x"`
		Y                 int           `json:"y"`
		Z                 int           `json:"z"`
		CookingTimes      []int         `json:"CookingTimes,omitempty"`
		CookingTotalTimes []int         `json:"CookingTotalTimes,omitempty"`
		Bees              []interface{} `json:"Bees,omitempty"`
	}, len(standard.TileEntities))
	for i, tileEntity := range standard.TileEntities {
		region.TileEntities[i].X = tileEntity.Position.X
		region.TileEntities[i].Y = tileEntity.Position.Y
		region.TileEntities[i].Z = tileEntity.Position.Z
	}

	// Set the region
	worldSave.Regions = make(map[string]struct {
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
			Motion              []int     `json:"Motion"`
			OnGround            int       `json:"OnGround"`
			PersistenceRequired int       `json:"PersistenceRequired"`
			PortalCooldown      int       `json:"PortalCooldown"`
			Pos                 []float64 `json:"Pos"`
			Rotation            []int     `json:"Rotation"`
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
	})
	worldSave.Regions["main"] = region

	return worldSave, nil
}
