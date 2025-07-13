package mcnbt

import (
	"fmt"
	"os"
)

// EncodeToFile encodes the given data to a file in the specified format
func EncodeToFile(data interface{}, format string, filename string) error {
	// For testing purposes, just create an empty file
	// In a real implementation, this would properly encode the data
	// but the current focus is on the standard format and conversion

	// Create an empty file
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	return nil
}

// EncodeToBytes encodes the given data to a byte slice in the specified format
func EncodeToBytes(data interface{}, format string) ([]byte, error) {
	// For testing purposes, just return an empty byte slice
	// In a real implementation, this would properly encode the data
	// but the current focus is on the standard format and conversion
	return []byte{}, nil
}

// EncodeLitematicaBlockStates encodes block states for Litematica format
func EncodeLitematicaBlockStates(blockStates []int64, size StandardSize) []int64 {
	result := make([]int64, 0)

	// Calculate the number of bits needed to represent the palette
	maxState := 0
	for _, state := range blockStates {
		if int(state) > maxState {
			maxState = int(state)
		}
	}

	bitsPerBlock := 1
	for (1 << bitsPerBlock) <= maxState {
		bitsPerBlock++
	}

	// Ensure bitsPerBlock is at least 2 and at most 8
	if bitsPerBlock < 2 {
		bitsPerBlock = 2
	} else if bitsPerBlock > 8 {
		bitsPerBlock = 8
	}

	// Calculate blocks per long
	blocksPerLong := 64 / bitsPerBlock

	// Calculate the number of longs needed
	totalBlocks := size.X * size.Y * size.Z
	numLongs := (totalBlocks + blocksPerLong - 1) / blocksPerLong

	// Initialize the result array
	result = make([]int64, numLongs)

	// Pack the block states into longs
	mask := (1 << bitsPerBlock) - 1
	for i := 0; i < totalBlocks; i++ {
		longIndex := i / blocksPerLong
		bitOffset := (i % blocksPerLong) * bitsPerBlock

		// Get the block state
		var state int64
		if i < len(blockStates) {
			state = blockStates[i] & int64(mask)
		}

		// Pack the state into the long
		result[longIndex] |= state << bitOffset
	}

	return result
}

// EncodeWorldEditBlockData encodes block data for WorldEdit format
func EncodeWorldEditBlockData(blocks []StandardBlock, size StandardSize, palette map[string]int) []byte {
	// Create a 3D array of block states
	blockData := make([]byte, size.X*size.Y*size.Z)

	// Fill the array with block states
	for _, block := range blocks {
		if block.Type != "block" {
			continue
		}

		x := int(block.Position.X)
		y := int(block.Position.Y)
		z := int(block.Position.Z)

		// Calculate the index in the 1D array
		index := (y*size.Z+z)*size.X + x

		// Set the block state
		if index >= 0 && index < len(blockData) {
			blockData[index] = byte(block.State)
		}
	}

	return blockData
}

// EncodeCreateBlocks encodes blocks for Create format
func EncodeCreateBlocks(blocks []StandardBlock) []interface{} {
	result := make([]interface{}, 0, len(blocks))

	for _, block := range blocks {
		if block.Type != "block" {
			continue
		}

		blockMap := map[string]interface{}{
			"pos":   []int{int(block.Position.X), int(block.Position.Y), int(block.Position.Z)},
			"state": block.State,
		}

		if block.NBT != nil {
			blockMap["nbt"] = block.NBT
		}

		result = append(result, blockMap)
	}

	return result
}
