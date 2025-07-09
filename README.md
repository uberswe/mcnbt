# mc nbt

A tool and library for parsing Minecraft Schematics and World Saves

## Features

- Parse Minecraft Schematics (Litematica, Create, WorldEdit, etc.) as JSON
- Parse Minecraft World Saves as JSON
- Extract block counts and other information from schematics

## Usage

### Command Line

```bash
# Parse a schematic file
go run cmd/main.go path/to/schematic.litematic

# Parse a Minecraft world save
go run cmd/main.go path/to/worldsave
```

### Library

To use the library in your Go code:

```
// Parse a schematic file as JSON
schematicData, err := mcnbt.ParseAnyFromFileAsJSON("path/to/schematic.litematic")

// Parse a Minecraft world save as JSON
worldData, err := mcnbt.ParseAnyFromFileAsJSON("path/to/worldsave")
```

## World Save Structure

When parsing a Minecraft world save, the library will:

1. Process the `level.dat` file to extract world metadata
2. Process all region files in the `region` directory to extract chunk data
3. Process all entity files in the `entities` directory (if it exists) to extract entity data

The output is a JSON object with the following structure:

```
{
  "level.dat": { "World metadata": "values" },
  "regions": {
    "r.0.0.mca": { "Chunk data for region (0,0)": "values" }
  },
  "entities": {
    "r.0.0.mca": { "Entity data for region (0,0)": "values" }
  }
}
```
