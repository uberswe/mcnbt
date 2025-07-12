# MC-NBT

A Go package for parsing Minecraft Schematics

## Features

- Parse Minecraft Schematics (Litematica, Create, WorldEdit, etc.) as JSON
- Extract block counts and other information from schematics
- Convert between different schematic formats

## Installation

```bash
go get github.com/uberswe/mcnbt
```

## Usage

### Command Line

The package includes a command-line tool for parsing and converting Minecraft schematics and world saves:

```bash
# Parse a schematic file
go run cmd/main.go path/to/schematic.litematic


# Specify output format and file
go run cmd/main.go path/to/schematic.litematic --format=json --output=output.json
```

### Library

To use the library in your Go code:

```go
package main

import (
    "fmt"
    "github.com/uberswe/mcnbt"
)

func main() {
    // Parse a schematic file as JSON
    schematicData, err := mcnbt.ParseAnyFromFileAsJSON("path/to/schematic.litematic")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }


    // Convert to standard format
    standardData, err := mcnbt.ConvertToStandard(schematicData)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Convert from standard format to another format
    litematicaData, err := mcnbt.ConvertFromStandard(standardData, "litematica")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Successfully processed data")
}
```

## Supported Formats

The library supports the following formats:

- Litematica (`.litematic`)
- WorldEdit (`.schem`)
- Create (`.nbt`)

## Data Structures

The library provides clean, non-nested struct definitions for each format:

- `LitematicaNBT` - Represents a Litematica schematic
- `WorldEditNBT` - Represents a WorldEdit schematic
- `CreateNBT` - Represents a Create schematic
- `StandardFormat` - A unified format that can represent any of the above


## License

This project is licensed under the MIT License - see the LICENSE file for details.
