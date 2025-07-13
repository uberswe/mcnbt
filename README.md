# MC-NBT

MC-NBT is a Go library for working with Minecraft NBT (Named Binary Tag) data, specifically focused on schematic formats used in Minecraft.

## Features

- Parse and decode NBT data from various Minecraft schematic formats:
  - Litematica (.litematic)
  - WorldEdit (.schem)
  - Create (.nbt)
- Convert between different schematic formats
- Unified standard format that consolidates blocks, entities, and tile entities
- Encode and save schematics in any supported format

## Standard Format

The library uses a standard format that can represent any of the supported schematic formats. This standard format consolidates blocks, entities, and tile entities into a single data structure, making it easier to work with and convert between formats.

The `StandardBlock` type represents blocks, entities, and tile entities with a `Type` field to distinguish between them:
- `"block"` for regular blocks
- `"entity"` for entities
- `"tile_entity"` for tile entities (block entities)

## Usage

### Parsing a Schematic

```go
// Parse a schematic file
data, err := mcnbt.ParseAnyFromFileAsJSON("path/to/schematic.litematic")
if err != nil {
    // Handle error
}

// Convert to standard format
standard, err := mcnbt.ConvertToStandard(data)
if err != nil {
    // Handle error
}

// Access blocks, entities, and tile entities
for _, block := range standard.Blocks {
    switch block.Type {
    case "block":
        // Handle block
    case "entity":
        // Handle entity
    case "tile_entity":
        // Handle tile entity
    }
}
```

### Converting Between Formats

```go
// Parse a schematic file
data, err := mcnbt.ParseAnyFromFileAsJSON("path/to/schematic.litematic")
if err != nil {
    // Handle error
}

// Convert to standard format
standard, err := mcnbt.ConvertToStandard(data)
if err != nil {
    // Handle error
}

// Convert to another format
worldEdit, err := mcnbt.ConvertFromStandard(standard, "worldedit")
if err != nil {
    // Handle error
}

// Save to file
err = mcnbt.EncodeToFile(worldEdit, "worldedit", "path/to/output.schem")
if err != nil {
    // Handle error
}
```

## Supported Formats

### Litematica (.litematic)

Litematica is a mod for Minecraft that allows players to create and place schematics. The library supports parsing and creating Litematica schematics.

### WorldEdit (.schem)

WorldEdit is a popular in-game map editor for Minecraft. The library supports parsing and creating WorldEdit schematics.

### Create (.nbt)

Create is a mod for Minecraft that adds various mechanical blocks and tools. The library supports parsing and creating Create schematics.

## Notes

- When converting between formats, some data loss may occur, especially for entities and tile entities, as different formats support different features.
- The library focuses on preserving block data during conversion, while entity and tile entity data may be simplified or lost.