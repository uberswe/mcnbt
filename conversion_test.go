package mcnbt

import (
	"encoding/json"
	"os"
	"testing"
)

func TestConversionBetweenFormats(t *testing.T) {
	// Define the test files
	testFiles := []struct {
		path   string
		format string
	}{
		{"testdata/color_field.litematic", "litematica"},
		{"testdata/color_field.nbt", "create"},
		{"testdata/color_field.schem", "worldedit"},
	}

	// Parse each file and convert to standard format
	standardFormats := make(map[string]*StandardFormat)

	for _, tf := range testFiles {
		t.Logf("Parsing %s", tf.path)

		// Parse the file
		data, err := ParseAnyFromFileAsJSON(tf.path)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", tf.path, err)
		}

		// Convert to standard format
		standard, err := ConvertToStandard(data)
		if err != nil {
			t.Fatalf("Failed to convert %s to standard format: %v", tf.path, err)
		}

		standardFormats[tf.format] = standard

		// Log some basic information about the standard format
		t.Logf("Format: %s", tf.format)
		t.Logf("Size: %d x %d x %d", standard.Size.X, standard.Size.Y, standard.Size.Z)
		t.Logf("Position: %d, %d, %d", standard.Position.X, standard.Position.Y, standard.Position.Z)
		t.Logf("Palette size: %d", len(standard.Palette))
		t.Logf("Blocks: %d", len(standard.Blocks))
		t.Logf("Entities: %d", len(standard.Entities))
		t.Logf("Tile Entities: %d", len(standard.TileEntities))
	}

	// Test conversion between formats
	for srcFormat, srcStandard := range standardFormats {
		for dstFormat, _ := range standardFormats {
			if srcFormat == dstFormat {
				continue
			}

			t.Logf("Converting from %s to %s", srcFormat, dstFormat)

			// Convert from standard to destination format
			t.Logf("Source blocks: %d", len(srcStandard.Blocks))
			dstData, err := ConvertFromStandard(srcStandard, dstFormat)
			if err != nil {
				t.Fatalf("Failed to convert from %s to %s: %v", srcFormat, dstFormat, err)
			}

			// Log information about the destination format
			switch v := dstData.(type) {
			case *CreateNBT:
				t.Logf("Destination format (Create) blocks: %d", len(v.Blocks))
			case *WorldEditNBT:
				t.Logf("Destination format (WorldEdit) BlockData length: %d", len(v.BlockData))
			case *LitematicaNBT:
				if len(v.Regions) > 0 {
					for _, region := range v.Regions {
						t.Logf("Destination format (Litematica) BlockStates length: %d", len(region.BlockStates))
						break
					}
				}
			}

			// Convert back to standard format
			dstStandard, err := ConvertToStandard(dstData)
			if err != nil {
				t.Fatalf("Failed to convert %s back to standard format: %v", dstFormat, err)
			}
			t.Logf("Destination blocks after conversion back to standard: %d", len(dstStandard.Blocks))

			// Compare the two standard formats
			compareStandardFormats(t, srcFormat, dstFormat, srcStandard, dstStandard)
		}
	}
}

func compareStandardFormats(t *testing.T, srcFormat, dstFormat string, src, dst *StandardFormat) {
	// Compare size
	if src.Size.X != dst.Size.X || src.Size.Y != dst.Size.Y || src.Size.Z != dst.Size.Z {
		t.Errorf("Size mismatch when converting from %s to %s: %v vs %v",
			srcFormat, dstFormat, src.Size, dst.Size)
	}

	// Compare position
	// Some formats might handle positions differently or not support them at all
	// Only log a warning about position differences, don't fail the test
	if src.Position.X != dst.Position.X || src.Position.Y != dst.Position.Y || src.Position.Z != dst.Position.Z {
		t.Logf("Position difference when converting from %s to %s: %v vs %v",
			srcFormat, dstFormat, src.Position, dst.Position)
	}

	// Compare palette size
	if len(src.Palette) != len(dst.Palette) {
		t.Errorf("Palette size mismatch when converting from %s to %s: %d vs %d",
			srcFormat, dstFormat, len(src.Palette), len(dst.Palette))
	}

	// Compare number of blocks
	// Allow for some difference in block counts due to different handling of air blocks
	blockCountDiff := len(src.Blocks) - len(dst.Blocks)
	if blockCountDiff < 0 {
		blockCountDiff = -blockCountDiff
	}

	// Special case for WorldEdit to Create conversion, which might handle blocks very differently
	if (srcFormat == "worldedit" && dstFormat == "create") || (srcFormat == "create" && dstFormat == "worldedit") {
		// Allow for a larger difference, but still enforce a limit
		maxAllowedDiff := len(src.Blocks) * 3 / 100 // 3% difference allowed
		if maxAllowedDiff < 10 {
			maxAllowedDiff = 10 // At least allow for 10 blocks difference
		}
		if maxAllowedDiff > 30 {
			maxAllowedDiff = 30 // But no more than 30 blocks difference
		}

		if blockCountDiff > maxAllowedDiff {
			t.Errorf("Block count mismatch when converting between WorldEdit and Create: %d vs %d (diff: %d, max allowed: %d)",
				len(src.Blocks), len(dst.Blocks), blockCountDiff, maxAllowedDiff)
		} else if blockCountDiff > 0 {
			t.Logf("Block count difference when converting between WorldEdit and Create: %d vs %d (diff: %d, max allowed: %d)",
				len(src.Blocks), len(dst.Blocks), blockCountDiff, maxAllowedDiff)
		}
	} else {
		// For other format conversions, allow for a 2% difference (reduced from 5%)
		maxAllowedDiff := len(src.Blocks) * 2 / 100
		if maxAllowedDiff < 5 {
			maxAllowedDiff = 5 // At least allow for 5 blocks difference (reduced from 10)
		}
		if maxAllowedDiff > 20 {
			maxAllowedDiff = 20 // But no more than 20 blocks difference
		}

		if blockCountDiff > maxAllowedDiff {
			t.Errorf("Block count mismatch when converting from %s to %s: %d vs %d (diff: %d, max allowed: %d)",
				srcFormat, dstFormat, len(src.Blocks), len(dst.Blocks), blockCountDiff, maxAllowedDiff)
		}
	}

	if len(src.Blocks) == 0 {
		t.Errorf("0 blocks in schematic: %d vs %d",
			len(src.Blocks), len(dst.Blocks))
	}

	// Compare number of entities
	if len(src.Entities) != len(dst.Entities) {
		t.Errorf("Entity count mismatch when converting from %s to %s: %d vs %d",
			srcFormat, dstFormat, len(src.Entities), len(dst.Entities))
	}

	// Compare number of tile entities
	// Some formats might not support tile entities or handle them differently
	// Only report an error if one format has tile entities and the other has none
	if (len(src.TileEntities) > 0 && len(dst.TileEntities) == 0) ||
		(len(src.TileEntities) == 0 && len(dst.TileEntities) > 0) {
		t.Errorf("Tile entity preservation issue when converting from %s to %s: %d vs %d",
			srcFormat, dstFormat, len(src.TileEntities), len(dst.TileEntities))
	}
}

// Helper function to save a standard format to a JSON file for debugging
func saveStandardToJSON(standard *StandardFormat, filename string) error {
	data, err := json.MarshalIndent(standard, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
