package mcnbt

import (
	"testing"
)

// TestConversionBetweenFormats tests conversion between different schematic formats
func TestConversionBetweenFormats(t *testing.T) {
	testCases := []struct {
		name          string
		inputFile     string
		outputFormats []string
	}{
		{
			name:          "Convert Litematica to other formats",
			inputFile:     "testdata/color_field.litematic",
			outputFormats: []string{"litematica", "worldedit", "create"},
		},
		{
			name:          "Convert WorldEdit to other formats",
			inputFile:     "testdata/color_field.schem",
			outputFormats: []string{"litematica", "worldedit", "create"},
		},
		{
			name:          "Convert Create to other formats",
			inputFile:     "testdata/color_field.nbt",
			outputFormats: []string{"litematica", "worldedit", "create"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(tc.inputFile)
			if err != nil {
				t.Fatalf("Failed to parse input file %s: %v", tc.inputFile, err)
			}

			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert to standard format: %v", err)
			}

			if len(standard.Blocks) == 0 {
				t.Errorf("Standard format has no blocks")
			} else {
				t.Logf("Standard format contains %d blocks", len(standard.Blocks))
			}

			for _, format := range tc.outputFormats {
				t.Run(format, func(t *testing.T) {
					result, err := ConvertFromStandard(standard, format)
					if err != nil {
						t.Errorf("Unexpected error converting to %s: %v", format, err)
						return
					}
					if result == nil {
						t.Errorf("Result is nil after converting to %s", format)
						return
					}
					t.Logf("Successfully converted to %s format", format)
				})
			}
		})
	}
}

// TestBlockCounts verifies that all formats extract meaningful block data
func TestBlockCounts(t *testing.T) {
	files := map[string]string{
		"litematica": "testdata/color_field.litematic",
		"worldedit":  "testdata/color_field.schem",
		"create":     "testdata/color_field.nbt",
	}

	results := make(map[string]*StandardFormat)

	for name, path := range files {
		t.Run(name, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(path)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", path, err)
			}

			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert %s to standard: %v", name, err)
			}

			results[name] = standard

			// All formats should have blocks
			if len(standard.Blocks) == 0 {
				t.Fatalf("%s produced 0 blocks", name)
			}

			// Litematica and WorldEdit should produce volume-filling blocks
			if name == "litematica" {
				expectedVolume := standard.Size.X * standard.Size.Y * standard.Size.Z
				if len(standard.Blocks) != expectedVolume {
					t.Errorf("%s: expected %d blocks (full volume), got %d", name, expectedVolume, len(standard.Blocks))
				}
			}

			if name == "worldedit" {
				expectedVolume := standard.Size.X * standard.Size.Y * standard.Size.Z
				if len(standard.Blocks) != expectedVolume {
					t.Errorf("%s: expected %d blocks (full volume), got %d", name, expectedVolume, len(standard.Blocks))
				}
			}

			// Count non-air blocks
			nonAirCount := 0
			for _, block := range standard.Blocks {
				if p, ok := standard.Palette[block.State]; ok {
					if p.Name != "minecraft:air" {
						nonAirCount++
					}
				}
			}
			t.Logf("%s: %d total blocks, %d non-air blocks, %d palette entries",
				name, len(standard.Blocks), nonAirCount, len(standard.Palette))
		})
	}
}

// TestBlockEntityCount verifies that block entities are preserved across formats
func TestBlockEntityCount(t *testing.T) {
	files := map[string]string{
		"litematica": "testdata/color_field.litematic",
		"worldedit":  "testdata/color_field.schem",
	}

	blockEntityCounts := make(map[string]int)

	for name, path := range files {
		t.Run(name, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(path)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", path, err)
			}

			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert %s: %v", name, err)
			}

			count := 0
			for _, block := range standard.Blocks {
				if block.Type == "block_entity" {
					count++
				}
			}
			blockEntityCounts[name] = count
			t.Logf("%s: %d block entities", name, count)
		})
	}

	// Both formats should have the same number of block entities
	if blockEntityCounts["litematica"] != blockEntityCounts["worldedit"] {
		t.Errorf("Block entity count mismatch: litematica=%d, worldedit=%d",
			blockEntityCounts["litematica"], blockEntityCounts["worldedit"])
	}
}

// TestPaletteConsistency verifies that palettes are consistent across formats
func TestPaletteConsistency(t *testing.T) {
	files := map[string]string{
		"litematica": "testdata/color_field.litematic",
		"worldedit":  "testdata/color_field.schem",
	}

	paletteSizes := make(map[string]int)

	for name, path := range files {
		t.Run(name, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(path)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", path, err)
			}

			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert %s: %v", name, err)
			}

			paletteSizes[name] = len(standard.Palette)
			t.Logf("%s: %d palette entries", name, len(standard.Palette))

			// Check that all palette entries have a name
			for i, p := range standard.Palette {
				if p.Name == "" {
					t.Errorf("%s: palette entry %d has empty name", name, i)
				}
			}
		})
	}

	// Litematica and WorldEdit should have the same palette size for the same schematic
	if paletteSizes["litematica"] != paletteSizes["worldedit"] {
		t.Errorf("Palette size mismatch: litematica=%d, worldedit=%d",
			paletteSizes["litematica"], paletteSizes["worldedit"])
	}
}

// TestTypeFieldConsistency verifies that the Type field is set on all blocks
func TestTypeFieldConsistency(t *testing.T) {
	files := map[string]string{
		"litematica": "testdata/color_field.litematic",
		"worldedit":  "testdata/color_field.schem",
		"create":     "testdata/color_field.nbt",
	}

	for name, path := range files {
		t.Run(name, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(path)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", path, err)
			}

			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert %s: %v", name, err)
			}

			emptyTypeCount := 0
			typeCounts := map[string]int{}
			for _, block := range standard.Blocks {
				if block.Type == "" {
					emptyTypeCount++
				}
				typeCounts[block.Type]++
			}

			if emptyTypeCount > 0 {
				t.Errorf("%s: %d blocks have empty Type field", name, emptyTypeCount)
			}

			t.Logf("%s: type distribution: %v", name, typeCounts)

			// Verify only valid types are used
			for typ := range typeCounts {
				switch typ {
				case "block", "block_entity", "entity":
					// valid
				default:
					t.Errorf("%s: unexpected block type %q", name, typ)
				}
			}
		})
	}
}

// TestRoundTrip tests that converting to standard and back preserves data
func TestRoundTrip(t *testing.T) {
	files := map[string]string{
		"litematica": "testdata/color_field.litematic",
		"worldedit":  "testdata/color_field.schem",
		"create":     "testdata/color_field.nbt",
	}

	for name, path := range files {
		t.Run(name, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(path)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", path, err)
			}

			// Convert to standard
			standard1, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert %s to standard: %v", name, err)
			}

			// Convert standard to the same format
			converted, err := ConvertFromStandard(standard1, name)
			if err != nil {
				t.Fatalf("Failed to convert standard to %s: %v", name, err)
			}

			// Convert back to standard
			standard2, err := ConvertToStandard(converted)
			if err != nil {
				t.Fatalf("Failed to convert %s back to standard: %v", name, err)
			}

			// Compare the two standard formats
			if len(standard1.Blocks) != len(standard2.Blocks) {
				t.Errorf("Round-trip block count mismatch: %d vs %d",
					len(standard1.Blocks), len(standard2.Blocks))
			}

			if len(standard1.Palette) != len(standard2.Palette) {
				t.Errorf("Round-trip palette size mismatch: %d vs %d",
					len(standard1.Palette), len(standard2.Palette))
			}

			if standard1.Size != standard2.Size {
				t.Errorf("Round-trip size mismatch: %v vs %v",
					standard1.Size, standard2.Size)
			}

			t.Logf("Round-trip %s: blocks %d→%d, palette %d→%d, size %v→%v",
				name,
				len(standard1.Blocks), len(standard2.Blocks),
				len(standard1.Palette), len(standard2.Palette),
				standard1.Size, standard2.Size)
		})
	}
}
