package mcnbt

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save/region"
	"io"
	"log"
	"os"
)

type NbtSchematic struct {
	DataVersion int           `json:"DataVersion" nbt:"DataVersion"`
	Blocks      []Block       `json:"blocks"`
	Entities    []interface{} `json:"entities"`
	Palette     []Palette     `json:"palette"`
	Size        []int         `json:"size"  nbt:"Size"`
}

type Block struct {
	Pos   []int       `json:"pos"`
	State int         `json:"state"`
	Nbt   interface{} `json:"nbt,omitempty" nbt:"nbt,omitempty"`
}

type Nbt struct {
	Item     Item        `json:"Item"`
	Material NbtMaterial `json:"Material"`
	ID       string      `json:"id"`
}

type Item struct {
	Count int    `json:"Count"`
	ID    string `json:"id"`
}

type NbtMaterial struct {
	Name string `json:"Name"`
}

type Palette struct {
	Name       string     `json:"Name"`
	Properties Properties `json:"Properties,omitempty"`
}

type Properties struct {
	Facing      string `json:"facing"`
	Half        string `json:"half"`
	Waterlogged string `json:"waterlogged"`
}

type NbtBlock struct {
	Name  string
	Count int
}

func ParseAnyFromFileAsJSON(f string) (interface{}, error) {
	// Check if the path is a directory
	fileInfo, err := os.Stat(f)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", f, err)
	}

	// If it's a directory, we don't support it
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("directories are not supported: %s", f)
	}

	// Read the file
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", f, err)
	}

	// Decode the data
	res, err := decodeAny(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file %s: %w", f, err)
	}

	return res, nil
}

func Decode(r io.Reader) ([]NbtBlock, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	schematic, err := decodeSchematic(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode schematic: %w", err)
	}

	// Pre-allocate slices with capacity to reduce reallocations
	nbtSlice := make([]NbtBlock, 0, len(schematic.Blocks)/4) // Estimate that ~25% of blocks have NBT data

	// Initialize palette slice with names
	paletteSlice := make([]NbtBlock, len(schematic.Palette))
	for i, tag := range schematic.Palette {
		paletteSlice[i] = NbtBlock{Name: tag.Name}
	}

	// Process blocks
	for _, tag := range schematic.Blocks {
		// Increment block count in palette
		if tag.State >= 0 && tag.State < len(paletteSlice) {
			paletteSlice[tag.State].Count++
		}

		// Process NBT data if present
		if tag.Nbt != nil {
			n, err := decodeNbt(tag.Nbt)
			if err != nil {
				return nil, fmt.Errorf("failed to decode NBT data: %w", err)
			}
			if n != nil && n.Item.ID != "" {
				nbtSlice = append(nbtSlice, NbtBlock{
					Name:  n.Item.ID,
					Count: n.Item.Count,
				})
			}
		}
	}

	// Use a map to aggregate blocks by name for better performance
	blockCounts := make(map[string]int)

	// Add palette blocks to the map
	for _, ps := range paletteSlice {
		if ps.Name != "" {
			blockCounts[ps.Name] += ps.Count
		}
	}

	// Add NBT blocks to the map
	for _, ns := range nbtSlice {
		if ns.Name != "" {
			blockCounts[ns.Name] += ns.Count
		}
	}

	// Convert map to slice
	aggr := make([]NbtBlock, 0, len(blockCounts))
	for name, count := range blockCounts {
		aggr = append(aggr, NbtBlock{
			Name:  name,
			Count: count,
		})
	}

	return aggr, nil
}

func decodeSchematic(data []byte) (*NbtSchematic, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer r.Close()

	schematic := new(NbtSchematic)
	if _, err = nbt.NewDecoder(r).Decode(schematic); err != nil {
		return nil, fmt.Errorf("failed to decode schematic: %w", err)
	}
	return schematic, nil
}

func decodeRegion(f string) error {
	r, err := region.Open(f)
	if err != nil {
		return fmt.Errorf("failed to open region file %s: %w", f, err)
	}
	defer r.Close()

	// Process all sectors in the region file
	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			// Skip non-existent sectors
			if !r.ExistSector(i, j) {
				continue
			}

			// Read sector data
			data, err := r.ReadSector(i, j)
			if err != nil {
				log.Printf("Warning: Failed to read sector (%d,%d) in %s: %v", i, j, f, err)
				continue
			}

			// Skip empty data
			if len(data) == 0 {
				continue
			}

			// Decode the sector data
			dd, err := decodeAny(data)
			if err != nil {
				log.Printf("Warning: Failed to decode sector (%d,%d) in %s: %v", i, j, f, err)
				continue
			}

			// Marshal the decoded data to JSON
			marshal, err := json.MarshalIndent(dd, "", "  ")
			if err != nil {
				log.Printf("Warning: Failed to marshal JSON for sector (%d,%d) in %s: %v", i, j, f, err)
				continue
			}

			// Print the JSON data
			log.Println(string(marshal))
		}
	}
	return nil
}

func decodeAny(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var r io.Reader
	var err error

	// Try different decompression methods based on magic numbers or format indicators
	if len(data) > 1 {
		if data[0] == 1 {
			// GZIP compression with format indicator
			r, err = gzip.NewReader(bytes.NewReader(data[1:]))
		} else if data[0] == 2 {
			// ZLIB compression with format indicator
			r, err = zlib.NewReader(bytes.NewReader(data[1:]))
		} else if data[0] == 0x1f && data[1] == 0x8b {
			// GZIP magic number
			r, err = gzip.NewReader(bytes.NewReader(data))
		} else if data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x9c || data[1] == 0xda) {
			// ZLIB magic number
			r, err = zlib.NewReader(bytes.NewReader(data))
		} else {
			// Assume uncompressed
			r = bytes.NewReader(data)
		}
	} else {
		// Single byte data, assume uncompressed
		r = bytes.NewReader(data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	if r == nil {
		return nil, fmt.Errorf("failed to create reader")
	}

	schematic := new(interface{})
	if _, err = nbt.NewDecoder(r).Decode(schematic); err != nil {
		return nil, fmt.Errorf("failed to decode NBT: %w", err)
	}
	return schematic, nil
}

func decodeNbt(val interface{}) (*Nbt, error) {
	switch data := val.(type) {
	case []byte:
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer r.Close()

		n := new(Nbt)
		if _, err = nbt.NewDecoder(r).Decode(n); err != nil {
			return nil, fmt.Errorf("failed to decode NBT data: %w", err)
		}
		return n, nil

	case map[string]interface{}:
		marshal, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
		}

		n := new(Nbt)
		if err = json.Unmarshal(marshal, n); err != nil {
			// Ignore error because we get weird nbt formats for inventories for example
			log.Println("Skipping invalid NBT:", string(marshal))
			return nil, nil
		}
		return n, nil
	}

	return nil, errors.New("unknown type of nbt")
}

// must is a helper function that panics if err is not nil
// It's useful for operations that should never fail in normal circumstances
func must[T any](v T, err error) T {
	if err != nil {
		// Print the error and exit with a non-zero status code
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
	return v
}
