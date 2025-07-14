package mcnbt

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/nbt"
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
	res, err := DecodeAny(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file %s: %w", f, err)
	}

	return res, nil
}

func DecodeAny(data []byte) (interface{}, error) {
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
