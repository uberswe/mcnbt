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
		return nil, err
	}

	// If it's a directory, check if it's a Minecraft world save
	if fileInfo.IsDir() {
		// Check if it has level.dat and region directory
		levelDatPath := f + "/level.dat"
		regionDirPath := f + "/region"

		_, levelDatErr := os.Stat(levelDatPath)
		regionDirInfo, regionDirErr := os.Stat(regionDirPath)

		if levelDatErr == nil && regionDirErr == nil && regionDirInfo.IsDir() {
			// It's a Minecraft world save, process it
			return decodeWorldSave(f)
		}
	}

	// If it's not a directory or not a Minecraft world save, process it as a file
	r, err := os.OpenFile(f, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	res, err := decodeAny(data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Decode(r io.Reader) ([]NbtBlock, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	schematic, err := decodeSchematic(data)
	if err != nil {
		return nil, err
	}

	nbtSlice := make([]NbtBlock, 0)

	paletteSlice := make([]NbtBlock, 0, len(schematic.Palette))
	for _, tag := range schematic.Palette {
		paletteSlice = append(paletteSlice, NbtBlock{Name: tag.Name})
	}

	for _, tag := range schematic.Blocks {

		if tag.Nbt != nil {
			n, err := decodeNbt(tag.Nbt)
			if err != nil {
				return nil, err
			}
			if n != nil {
				nbtSlice = append(nbtSlice, NbtBlock{
					Name:  n.Item.ID,
					Count: n.Item.Count,
				})
			}
		}

		// Blockstate refers to palette to get type of block
		paletteSlice[tag.State].Count += 1
	}

	aggr := make([]NbtBlock, 0, len(paletteSlice))
	for _, ps := range paletteSlice {
		found := false
		for i, a := range aggr {
			if a.Name == ps.Name {
				found = true
				aggr[i].Count += ps.Count
			}
		}
		if !found && ps.Name != "" {
			aggr = append(aggr, ps)
		}
	}

	for _, ns := range nbtSlice {
		found := false
		for i, a := range aggr {
			if a.Name == ns.Name {
				found = true
				aggr[i].Count += ns.Count
			}
		}
		if !found && ns.Name != "" {
			aggr = append(aggr, ns)
		}
	}

	return aggr, nil
}

func decodeSchematic(data []byte) (*NbtSchematic, error) {
	r, err := gzip.NewReader(bytes.NewReader(data[:]))
	if err != nil {
		return nil, err
	}

	schematic := new(NbtSchematic)
	if _, err = nbt.NewDecoder(r).Decode(schematic); err != nil {
		return nil, err
	}
	return schematic, nil
}

func decodeRegion(f string) error {
	r, err := region.Open(f)
	if err != nil {
		return fmt.Errorf("failed to open region file: %w", err)
	}
	defer r.Close()

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if !r.ExistSector(i, j) {
				continue
			}

			data, err := r.ReadSector(i, j)
			if err != nil {
				log.Printf("Warning: Failed to read sector (%d,%d): %v", i, j, err)
				continue
			}

			// Skip empty data
			if len(data) == 0 {
				continue
			}

			dd, err := decodeAny(data)
			if err != nil {
				log.Printf("Warning: Failed to decode sector (%d,%d): %v", i, j, err)
				continue
			}

			marshal, err := json.MarshalIndent(dd, "", "	")
			if err != nil {
				log.Printf("Warning: Failed to marshal JSON for sector (%d,%d): %v", i, j, err)
				continue
			}

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

	// Try different decompression methods
	if data[0] == 1 {
		// GZIP compression
		r, err = gzip.NewReader(bytes.NewReader(data[1:]))
	} else if data[0] == 2 {
		// ZLIB compression
		r, err = zlib.NewReader(bytes.NewReader(data[1:]))
	} else if data[0] == 0x1f && len(data) > 1 && data[1] == 0x8b {
		// GZIP magic number
		r, err = gzip.NewReader(bytes.NewReader(data))
	} else {
		// Try to detect compression type
		if len(data) > 1 && data[0] == 0x1f && data[1] == 0x8b {
			// GZIP magic number
			r, err = gzip.NewReader(bytes.NewReader(data))
		} else if len(data) > 1 && data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x9c || data[1] == 0xda) {
			// ZLIB magic number
			r, err = zlib.NewReader(bytes.NewReader(data))
		} else {
			// Assume uncompressed
			r = bytes.NewReader(data)
		}
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
	if data, ok := val.([]byte); ok {
		r, err := gzip.NewReader(bytes.NewReader(data[:]))
		if err != nil {
			return nil, err
		}

		n := new(Nbt)
		if _, err = nbt.NewDecoder(r).Decode(n); err != nil {
			return nil, err
		}
		return n, nil
	} else if data, ok := val.(map[string]interface{}); ok {
		marshal, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		n := new(Nbt)
		err = json.Unmarshal(marshal, n)
		if err != nil {
			// Ignore error because we get weird nbt formats for inventories for example
			log.Println("Skipping invalid NBT", string(marshal))
			return nil, nil
		} else {
			return n, nil
		}
	}
	return nil, errors.New("unknown type of nbt")
}

func must[T any](v T, err error) T {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return v
}

// decodeWorldSave processes a Minecraft world save directory and returns the data as JSON
func decodeWorldSave(worldDir string) (interface{}, error) {
	result := make(map[string]interface{})

	// Process level.dat
	levelDatPath := worldDir + "/level.dat"
	levelDatFile, err := os.OpenFile(levelDatPath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer levelDatFile.Close()

	// level.dat is a gzipped NBT file
	r, err := gzip.NewReader(levelDatFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Decode level.dat
	levelDat := new(interface{})
	if _, err = nbt.NewDecoder(r).Decode(levelDat); err != nil {
		return nil, err
	}
	result["level.dat"] = levelDat

	// Process region files
	regionDir := worldDir + "/region"
	regionFiles, err := os.ReadDir(regionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read region directory: %w", err)
	}

	regions := make(map[string]interface{})
	for _, regionFile := range regionFiles {
		if !regionFile.IsDir() && len(regionFile.Name()) > 4 && regionFile.Name()[0] == 'r' && regionFile.Name()[len(regionFile.Name())-4:] == ".mca" {
			regionPath := regionDir + "/" + regionFile.Name()
			r, err := region.Open(regionPath)
			if err != nil {
				log.Printf("Warning: Failed to open region file %s: %v", regionPath, err)
				continue
			}
			defer r.Close()

			chunks := make(map[string]interface{})
			for i := 0; i < 32; i++ {
				for j := 0; j < 32; j++ {
					if !r.ExistSector(i, j) {
						continue
					}

					data, err := r.ReadSector(i, j)
					if err != nil {
						log.Printf("Warning: Failed to read sector (%d,%d) in region file %s: %v", i, j, regionPath, err)
						continue
					}

					// Skip empty data
					if len(data) == 0 {
						continue
					}

					chunk, err := decodeAny(data)
					if err != nil {
						log.Printf("Warning: Failed to decode sector (%d,%d) in region file %s: %v", i, j, regionPath, err)
						continue
					}

					chunkKey := fmt.Sprintf("chunk_%d_%d", i, j)
					chunks[chunkKey] = chunk
				}
			}
			regions[regionFile.Name()] = chunks
		}
	}
	result["regions"] = regions

	// Process entities directory if it exists
	entitiesDir := worldDir + "/entities"
	if entitiesInfo, err := os.Stat(entitiesDir); err == nil && entitiesInfo.IsDir() {
		entitiesFiles, err := os.ReadDir(entitiesDir)
		if err != nil {
			log.Printf("Warning: Failed to read entities directory: %v", err)
		} else {
			entities := make(map[string]interface{})
			for _, entityFile := range entitiesFiles {
				if !entityFile.IsDir() && len(entityFile.Name()) > 4 && entityFile.Name()[len(entityFile.Name())-4:] == ".mca" {
					entityPath := entitiesDir + "/" + entityFile.Name()
					r, err := region.Open(entityPath)
					if err != nil {
						log.Printf("Warning: Failed to open entity file %s: %v", entityPath, err)
						continue
					}
					defer r.Close()

					entityChunks := make(map[string]interface{})
					for i := 0; i < 32; i++ {
						for j := 0; j < 32; j++ {
							if !r.ExistSector(i, j) {
								continue
							}

							data, err := r.ReadSector(i, j)
							if err != nil {
								log.Printf("Warning: Failed to read sector (%d,%d) in entity file %s: %v", i, j, entityPath, err)
								continue
							}

							// Skip empty data
							if len(data) == 0 {
								continue
							}

							entityChunk, err := decodeAny(data)
							if err != nil {
								log.Printf("Warning: Failed to decode sector (%d,%d) in entity file %s: %v", i, j, entityPath, err)
								continue
							}

							entityChunkKey := fmt.Sprintf("chunk_%d_%d", i, j)
							entityChunks[entityChunkKey] = entityChunk
						}
					}
					entities[entityFile.Name()] = entityChunks
				}
			}
			result["entities"] = entities
		}
	}

	return result, nil
}
