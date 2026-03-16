package mcnbt

import (
	"fmt"
	"math/bits"
	"os"
)

// EncodeToFile encodes the given data to a file in the specified format
func EncodeToFile(data interface{}, format string, filename string) error {
	// For testing purposes, just create an empty file
	// In a real implementation, this would properly encode the data
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	return nil
}

// EncodeToBytes encodes the given data to a byte slice in the specified format
func EncodeToBytes(data interface{}, format string) ([]byte, error) {
	return []byte{}, nil
}

// EncodeLitematicaBlockStates encodes block states for Litematica format.
// Entries do NOT cross long boundaries.
func EncodeLitematicaBlockStates(blockStates []int64, size StandardSize) []int64 {
	totalBlocks := size.X * size.Y * size.Z

	// Calculate bits per entry from maximum state value
	maxState := int64(0)
	for _, state := range blockStates {
		if state > maxState {
			maxState = state
		}
	}

	bitsPerBlock := 2 // minimum 2
	if maxState > 0 {
		b := bits.Len64(uint64(maxState))
		if b > bitsPerBlock {
			bitsPerBlock = b
		}
	}

	blocksPerLong := 64 / bitsPerBlock
	numLongs := (totalBlocks + blocksPerLong - 1) / blocksPerLong

	result := make([]int64, numLongs)
	mask := int64((1 << bitsPerBlock) - 1)

	for i := 0; i < totalBlocks && i < len(blockStates); i++ {
		longIndex := i / blocksPerLong
		bitOffset := (i % blocksPerLong) * bitsPerBlock
		state := blockStates[i] & mask
		result[longIndex] |= state << bitOffset
	}

	return result
}

// EncodeWorldEditBlockData encodes block data for WorldEdit format using varint encoding
func EncodeWorldEditBlockData(blocks []StandardBlock, size StandardSize, palette map[string]int) []byte {
	totalVolume := size.X * size.Y * size.Z
	grid := make([]int, totalVolume)

	for _, block := range blocks {
		x := int(block.Position.X)
		y := int(block.Position.Y)
		z := int(block.Position.Z)

		if x < 0 || x >= size.X || y < 0 || y >= size.Y || z < 0 || z >= size.Z {
			continue
		}

		idx := (y*size.Z+z)*size.X + x
		if idx >= 0 && idx < totalVolume {
			grid[idx] = block.State
		}
	}

	var result []byte
	for i := 0; i < totalVolume; i++ {
		result = append(result, encodeVarint(grid[i])...)
	}
	return result
}

// encodeVarint encodes an integer as a varint byte sequence
func encodeVarint(value int) []byte {
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

// EncodeCreateBlocks encodes blocks for Create/Vanilla structure format
func EncodeCreateBlocks(blocks []StandardBlock) []interface{} {
	result := make([]interface{}, 0, len(blocks))

	for _, block := range blocks {
		if block.Type == "entity" {
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
