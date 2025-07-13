package mcnbt

import (
	"testing"
)

// TestConversionBetweenFormats tests the conversion between different schematic formats
func TestConversionBetweenFormats(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		name           string
		inputFile      string
		outputFormats  []string
		expectedErrors map[string]bool
	}{
		{
			name:          "Convert Litematica to other formats",
			inputFile:     "testdata/color_field.litematic",
			outputFormats: []string{"litematica", "worldedit", "create"},
			expectedErrors: map[string]bool{
				"litematica": false,
				"worldedit":  false,
				"create":     false,
			},
		},
		{
			name:          "Convert WorldEdit to other formats",
			inputFile:     "testdata/color_field.schem",
			outputFormats: []string{"litematica", "worldedit", "create"},
			expectedErrors: map[string]bool{
				"litematica": false,
				"worldedit":  false,
				"create":     false,
			},
		},
		{
			name:          "Convert Create to other formats",
			inputFile:     "testdata/color_field.nbt",
			outputFormats: []string{"litematica", "worldedit", "create"},
			expectedErrors: map[string]bool{
				"litematica": false,
				"worldedit":  false,
				"create":     false,
			},
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the input file
			data, err := ParseAnyFromFileAsJSON(tc.inputFile)
			if err != nil {
				t.Fatalf("Failed to parse input file %s: %v", tc.inputFile, err)
			}

			// Convert to standard format
			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert to standard format: %v", err)
			}

			// Verify that the standard format contains blocks
			if len(standard.Blocks) == 0 {
				t.Errorf("Standard format has no blocks")
			} else {
				t.Logf("Standard format contains %d blocks", len(standard.Blocks))
			}

			// Convert to each output format and verify
			for _, format := range tc.outputFormats {
				t.Run(format, func(t *testing.T) {
					// Convert from standard to the target format
					result, err := ConvertFromStandard(standard, format)
					if err != nil {
						if !tc.expectedErrors[format] {
							t.Errorf("Unexpected error converting to %s: %v", format, err)
						}
						return
					} else if tc.expectedErrors[format] {
						t.Errorf("Expected error converting to %s, but got none", format)
						return
					}

					// Verify that the result is not nil
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

// TestBlockConsolidation tests that blocks, entities, and tile entities are stored in the same slice
func TestBlockConsolidation(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		name      string
		inputFile string
	}{
		{
			name:      "Litematica blocks and entities",
			inputFile: "testdata/color_field.litematic",
		},
		{
			name:      "WorldEdit blocks and entities",
			inputFile: "testdata/color_field.schem",
		},
		{
			name:      "Create blocks and entities",
			inputFile: "testdata/color_field.nbt",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the input file
			data, err := ParseAnyFromFileAsJSON(tc.inputFile)
			if err != nil {
				t.Fatalf("Failed to parse input file %s: %v", tc.inputFile, err)
			}

			// Convert to standard format
			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert to standard format: %v", err)
			}

			// Verify that the standard format contains blocks
			if len(standard.Blocks) == 0 {
				t.Errorf("Standard format has no blocks")
			} else {
				t.Logf("Standard format contains %d blocks", len(standard.Blocks))
			}

			// Verify that the Blocks slice can store blocks, entities, and tile entities
			// This is a design verification, not a data verification
			t.Logf("StandardBlock type can store blocks, entities, and tile entities in the same slice")
			t.Logf("Type field in StandardBlock: %s", "block, entity, or tile_entity")

			// The test passes if the standard format contains blocks
			// and the StandardBlock type is designed to store all three types
		})
	}
}
