package mcnbt

import (
	"encoding/json"
	
	"testing"
)

func TestDebugFormats(t *testing.T) {
	files := []string{
		"testdata/color_field.litematic",
		"testdata/color_field.schem",
		"testdata/color_field.nbt",
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			data, err := ParseAnyFromFileAsJSON(f)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			standard, err := ConvertToStandard(data)
			if err != nil {
				t.Fatalf("Failed to convert: %v", err)
			}

			t.Logf("Size: %+v", standard.Size)
			t.Logf("Position: %+v", standard.Position)
			t.Logf("Palette size: %d", len(standard.Palette))
			t.Logf("Block count: %d", len(standard.Blocks))
			t.Logf("DataVersion: %d", standard.DataVersion)
			t.Logf("OriginalFormat: %s", standard.OriginalFormat)

			// Count by type
			typeCounts := map[string]int{}
			for _, b := range standard.Blocks {
				typ := b.Type
				if typ == "" {
					typ = "(empty)"
				}
				typeCounts[typ]++
			}
			t.Logf("Block types: %+v", typeCounts)

			// Show palette
			for i, p := range standard.Palette {
				t.Logf("  Palette[%d]: %s %v", i, p.Name, p.Properties)
			}

			// Show first 5 blocks
			for i, b := range standard.Blocks {
				if i >= 5 {
					break
				}
				j, _ := json.Marshal(b)
				t.Logf("  Block[%d]: %s", i, string(j))
			}
		})
	}
}
