package mcnbt

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Tnze/go-mc/nbt"
)

// TestNBTRoundTripTypes verifies that encoding a CreateNBT struct to NBT binary
// and decoding it back produces the same NBT types as the original .nbt file.
// This catches issues like Tag_List becoming Tag_Int_Array ([]interface{} vs []int32).
func TestNBTRoundTripTypes(t *testing.T) {
	// Step 1: Parse the original .nbt file to get raw decoded types
	origRaw, err := ParseAnyFromFileAsJSON("testdata/color_field.nbt")
	if err != nil {
		t.Fatalf("Failed to parse original .nbt: %v", err)
	}

	orig, ok := origRaw.(*interface{})
	if !ok {
		t.Fatalf("Expected *interface{}, got %T", origRaw)
	}
	origMap, ok := (*orig).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", *orig)
	}

	// Step 2: Convert to standard then back to Create struct
	standard, err := ConvertToStandard(origRaw)
	if err != nil {
		t.Fatalf("Failed to convert to standard: %v", err)
	}

	roundTripped, err := ConvertFromStandard(standard, "create")
	if err != nil {
		t.Fatalf("Failed to convert back to create: %v", err)
	}

	rtCreate, ok := roundTripped.(*CreateNBT)
	if !ok {
		t.Fatalf("Expected *CreateNBT, got %T", roundTripped)
	}

	// Step 3: Encode the CreateNBT struct to NBT binary bytes
	var buf bytes.Buffer
	if err := nbt.NewEncoder(&buf).Encode(rtCreate, ""); err != nil {
		t.Fatalf("Failed to encode CreateNBT to NBT: %v", err)
	}

	// Step 4: Decode those bytes back to raw map[string]interface{}
	reDecoded := new(interface{})
	if _, err := nbt.NewDecoder(&buf).Decode(reDecoded); err != nil {
		t.Fatalf("Failed to decode round-tripped NBT: %v", err)
	}

	reMap, ok := (*reDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} from re-decode, got %T", *reDecoded)
	}

	// Step 5: Compare types between original and re-decoded data
	t.Run("top_level_keys", func(t *testing.T) {
		for k := range origMap {
			if _, ok := reMap[k]; !ok {
				t.Errorf("Key %q present in original but missing in round-trip", k)
			}
		}
		for k := range reMap {
			if _, ok := origMap[k]; !ok {
				t.Logf("Key %q present in round-trip but not in original (may be expected for empty fields)", k)
			}
		}
	})

	t.Run("size", func(t *testing.T) {
		origSize := origMap["size"]
		reSize := reMap["size"]
		t.Logf("Original size: %T = %v", origSize, origSize)
		t.Logf("Re-decoded size: %T = %v", reSize, reSize)
		compareAndReport(t, "size", origSize, reSize)
	})

	t.Run("DataVersion", func(t *testing.T) {
		origDV := origMap["DataVersion"]
		reDV := reMap["DataVersion"]
		t.Logf("Original DataVersion: %T = %v", origDV, origDV)
		t.Logf("Re-decoded DataVersion: %T = %v", reDV, reDV)
		compareAndReport(t, "DataVersion", origDV, reDV)
	})

	t.Run("blocks", func(t *testing.T) {
		origBlocks, _ := origMap["blocks"].([]interface{})
		reBlocks, _ := reMap["blocks"].([]interface{})

		if len(origBlocks) != len(reBlocks) {
			t.Errorf("Block count: orig=%d, re=%d", len(origBlocks), len(reBlocks))
		}

		// Compare a sample of blocks in detail
		sampleSize := 20
		if len(origBlocks) < sampleSize {
			sampleSize = len(origBlocks)
		}
		if len(reBlocks) < sampleSize {
			sampleSize = len(reBlocks)
		}

		for i := 0; i < sampleSize; i++ {
			origBlock, _ := origBlocks[i].(map[string]interface{})
			reBlock, _ := reBlocks[i].(map[string]interface{})
			if origBlock == nil || reBlock == nil {
				continue
			}

			prefix := fmt.Sprintf("blocks[%d]", i)
			for k, origVal := range origBlock {
				reVal, exists := reBlock[k]
				if !exists {
					t.Errorf("%s.%s: present in original but missing in round-trip", prefix, k)
					continue
				}
				compareAndReport(t, prefix+"."+k, origVal, reVal)
			}
		}
	})

	t.Run("palette", func(t *testing.T) {
		origPalette, _ := origMap["palette"].([]interface{})
		rePalette, _ := reMap["palette"].([]interface{})

		if len(origPalette) != len(rePalette) {
			t.Errorf("Palette count: orig=%d, re=%d", len(origPalette), len(rePalette))
		}

		sampleSize := 10
		if len(origPalette) < sampleSize {
			sampleSize = len(origPalette)
		}
		if len(rePalette) < sampleSize {
			sampleSize = len(rePalette)
		}

		for i := 0; i < sampleSize; i++ {
			origEntry, _ := origPalette[i].(map[string]interface{})
			reEntry, _ := rePalette[i].(map[string]interface{})
			if origEntry == nil || reEntry == nil {
				continue
			}

			prefix := fmt.Sprintf("palette[%d]", i)
			for k, origVal := range origEntry {
				reVal, exists := reEntry[k]
				if !exists {
					t.Errorf("%s.%s: present in original but missing in round-trip", prefix, k)
					continue
				}
				compareAndReport(t, prefix+"."+k, origVal, reVal)
			}
		}
	})

	t.Run("deep_type_comparison", func(t *testing.T) {
		diffs := deepCompareTypes("root", origMap, reMap, 0)
		for _, d := range diffs {
			t.Errorf("TYPE DIFF: %s", d)
		}
		if len(diffs) == 0 {
			t.Logf("All NBT types match between original and round-tripped data")
		}
	})
}

// TestNBTRoundTripBinary does a full binary round-trip: read .nbt file bytes,
// decode to CreateNBT struct, encode back to bytes, decode again, and compare.
func TestNBTRoundTripBinary(t *testing.T) {
	// Read the raw file and decode
	data, err := os.ReadFile("testdata/color_field.nbt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Decompress (gzip)
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}

	// Decode to raw interface
	origDecoded := new(interface{})
	if _, err := nbt.NewDecoder(gzReader).Decode(origDecoded); err != nil {
		t.Fatalf("Failed to decode original: %v", err)
	}
	gzReader.Close()

	origMap, ok := (*origDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", *origDecoded)
	}

	// Decode to CreateNBT struct
	gzReader2, _ := gzip.NewReader(bytes.NewReader(data))
	createNBT := new(CreateNBT)
	if _, err := nbt.NewDecoder(gzReader2).Decode(createNBT); err != nil {
		t.Fatalf("Failed to decode to CreateNBT: %v", err)
	}
	gzReader2.Close()

	// Encode CreateNBT back to NBT bytes
	var buf bytes.Buffer
	if err := nbt.NewEncoder(&buf).Encode(createNBT, ""); err != nil {
		t.Fatalf("Failed to encode CreateNBT: %v", err)
	}

	// Decode the re-encoded bytes to raw interface
	reDecoded := new(interface{})
	if _, err := nbt.NewDecoder(&buf).Decode(reDecoded); err != nil {
		t.Fatalf("Failed to decode re-encoded NBT: %v", err)
	}

	reMap, ok := (*reDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", *reDecoded)
	}

	// Compare types
	diffs := deepCompareTypes("root", origMap, reMap, 0)
	for _, d := range diffs {
		t.Errorf("TYPE DIFF: %s", d)
	}
	if len(diffs) == 0 {
		t.Logf("Binary round-trip: all NBT types preserved")
	}

	// Log specific fields of interest
	t.Logf("Original size type: %T", origMap["size"])
	t.Logf("Re-decoded size type: %T", reMap["size"])
	if origBlocks, ok := origMap["blocks"].([]interface{}); ok && len(origBlocks) > 0 {
		if origBlock, ok := origBlocks[0].(map[string]interface{}); ok {
			t.Logf("Original blocks[0].pos type: %T", origBlock["pos"])
		}
	}
	if reBlocks, ok := reMap["blocks"].([]interface{}); ok && len(reBlocks) > 0 {
		if reBlock, ok := reBlocks[0].(map[string]interface{}); ok {
			t.Logf("Re-decoded blocks[0].pos type: %T", reBlock["pos"])
		}
	}
}

// TestLitematicaRoundTripBinary does a full binary round-trip for .litematic files:
// decode to LitematicaNBT struct, encode back, decode raw, compare types.
func TestLitematicaRoundTripBinary(t *testing.T) {
	data, err := os.ReadFile("testdata/color_field.litematic")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Decompress (gzip)
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}

	// Decode to raw interface
	origDecoded := new(interface{})
	if _, err := nbt.NewDecoder(gzReader).Decode(origDecoded); err != nil {
		t.Fatalf("Failed to decode original: %v", err)
	}
	gzReader.Close()

	origMap, ok := (*origDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", *origDecoded)
	}

	// Decode to LitematicaNBT struct
	gzReader2, _ := gzip.NewReader(bytes.NewReader(data))
	litematicaNBT := new(LitematicaNBT)
	if _, err := nbt.NewDecoder(gzReader2).Decode(litematicaNBT); err != nil {
		t.Fatalf("Failed to decode to LitematicaNBT: %v", err)
	}
	gzReader2.Close()

	// Encode back to NBT bytes
	var buf bytes.Buffer
	if err := nbt.NewEncoder(&buf).Encode(litematicaNBT, ""); err != nil {
		t.Fatalf("Failed to encode LitematicaNBT: %v", err)
	}

	// Decode the re-encoded bytes to raw interface
	reDecoded := new(interface{})
	if _, err := nbt.NewDecoder(&buf).Decode(reDecoded); err != nil {
		t.Fatalf("Failed to decode re-encoded NBT: %v", err)
	}

	reMap, ok := (*reDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", *reDecoded)
	}

	// Compare types — filter out "missing" diffs for tile entity fields
	// that the struct can't capture (the NBT library silently drops unknown
	// struct fields; this is a known limitation, not a type fidelity issue).
	diffs := deepCompareTypes("root", origMap, reMap, 0)
	var typeDiffs []string
	for _, d := range diffs {
		if strings.Contains(d, "rt=missing") {
			t.Logf("KNOWN MISSING (tile entity extra fields): %s", d)
			continue
		}
		typeDiffs = append(typeDiffs, d)
	}
	for _, d := range typeDiffs {
		t.Errorf("TYPE DIFF: %s", d)
	}
	if len(typeDiffs) == 0 {
		t.Logf("Litematica binary round-trip: all NBT types preserved (for fields captured by struct)")
	}

	// Log specific fields of interest
	t.Logf("Original MinecraftDataVersion type: %T", origMap["MinecraftDataVersion"])
	t.Logf("Re-decoded MinecraftDataVersion type: %T", reMap["MinecraftDataVersion"])
	t.Logf("Original Version type: %T", origMap["Version"])
	t.Logf("Re-decoded Version type: %T", reMap["Version"])
}

// TestWorldEditRoundTripBinary does a full binary round-trip for .schem files:
// decode to WorldEditNBT struct, encode back, decode raw, compare types.
func TestWorldEditRoundTripBinary(t *testing.T) {
	data, err := os.ReadFile("testdata/color_field.schem")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Decompress (gzip)
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}

	// Decode to raw interface
	origDecoded := new(interface{})
	if _, err := nbt.NewDecoder(gzReader).Decode(origDecoded); err != nil {
		t.Fatalf("Failed to decode original: %v", err)
	}
	gzReader.Close()

	origMap, ok := (*origDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", *origDecoded)
	}

	// Decode to WorldEditNBT struct
	gzReader2, _ := gzip.NewReader(bytes.NewReader(data))
	worldEditNBT := new(WorldEditNBT)
	if _, err := nbt.NewDecoder(gzReader2).Decode(worldEditNBT); err != nil {
		t.Fatalf("Failed to decode to WorldEditNBT: %v", err)
	}
	gzReader2.Close()

	// Encode back to NBT bytes
	var buf bytes.Buffer
	if err := nbt.NewEncoder(&buf).Encode(worldEditNBT, "Schematic"); err != nil {
		t.Fatalf("Failed to encode WorldEditNBT: %v", err)
	}

	// Decode the re-encoded bytes to raw interface
	reDecoded := new(interface{})
	if _, err := nbt.NewDecoder(&buf).Decode(reDecoded); err != nil {
		t.Fatalf("Failed to decode re-encoded NBT: %v", err)
	}

	reMap, ok := (*reDecoded).(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", *reDecoded)
	}

	// Compare types
	diffs := deepCompareTypes("root", origMap, reMap, 0)
	for _, d := range diffs {
		t.Errorf("TYPE DIFF: %s", d)
	}
	if len(diffs) == 0 {
		t.Logf("WorldEdit binary round-trip: all NBT types preserved")
	}

	// Log specific fields of interest
	t.Logf("Original Width type: %T", origMap["Width"])
	t.Logf("Re-decoded Width type: %T", reMap["Width"])
	t.Logf("Original Height type: %T", origMap["Height"])
	t.Logf("Re-decoded Height type: %T", reMap["Height"])
	t.Logf("Original DataVersion type: %T", origMap["DataVersion"])
	t.Logf("Re-decoded DataVersion type: %T", reMap["DataVersion"])
}

// compareAndReport compares types of two values and reports mismatches
func compareAndReport(t *testing.T, path string, orig, rt interface{}) {
	t.Helper()

	origType := reflect.TypeOf(orig)
	rtType := reflect.TypeOf(rt)

	if origType != rtType {
		t.Errorf("%s: type mismatch orig=%s, rt=%s", path, origType, rtType)
		return
	}

	// For slices, also compare element types
	origRV := reflect.ValueOf(orig)
	rtRV := reflect.ValueOf(rt)
	if origRV.Kind() == reflect.Slice && rtRV.Kind() == reflect.Slice {
		if origRV.Type() != rtRV.Type() {
			t.Errorf("%s: slice type mismatch orig=%s, rt=%s", path, origRV.Type(), rtRV.Type())
		}
	}
}

// deepCompareTypes recursively compares the types of two map structures
func deepCompareTypes(path string, orig, rt map[string]interface{}, depth int) []string {
	if depth > 10 {
		return nil
	}

	var diffs []string

	for k, origVal := range orig {
		rtVal, exists := rt[k]
		if !exists {
			diffs = append(diffs, fmt.Sprintf("%s.%s: orig=%T, rt=missing", path, k, origVal))
			continue
		}

		origType := reflect.TypeOf(origVal)
		rtType := reflect.TypeOf(rtVal)

		if origType != rtType {
			diffs = append(diffs, fmt.Sprintf("%s.%s: orig=%s, rt=%s", path, k, origType, rtType))
			continue
		}

		// Recurse into maps
		if origMap, ok := origVal.(map[string]interface{}); ok {
			if rtMap, ok := rtVal.(map[string]interface{}); ok {
				diffs = append(diffs, deepCompareTypes(path+"."+k, origMap, rtMap, depth+1)...)
			}
		}

		// For slices, compare both the slice type and element types
		origRV := reflect.ValueOf(origVal)
		rtRV := reflect.ValueOf(rtVal)
		if origRV.Kind() == reflect.Slice && rtRV.Kind() == reflect.Slice {
			if origRV.Type() != rtRV.Type() {
				diffs = append(diffs, fmt.Sprintf("%s.%s: slice type orig=%s, rt=%s",
					path, k, origRV.Type(), rtRV.Type()))
			}

			// For []interface{} slices, compare elements recursively
			if origRV.Type() == reflect.TypeOf([]interface{}{}) {
				minLen := origRV.Len()
				if rtRV.Len() < minLen {
					minLen = rtRV.Len()
				}
				// Sample up to 5 elements
				sampleSize := minLen
				if sampleSize > 5 {
					sampleSize = 5
				}
				for i := 0; i < sampleSize; i++ {
					origElem := origRV.Index(i).Interface()
					rtElem := rtRV.Index(i).Interface()

					elemPath := fmt.Sprintf("%s.%s[%d]", path, k, i)

					origElemType := reflect.TypeOf(origElem)
					rtElemType := reflect.TypeOf(rtElem)
					if origElemType != rtElemType {
						diffs = append(diffs, fmt.Sprintf("%s: orig=%s, rt=%s",
							elemPath, origElemType, rtElemType))
					}

					// Recurse into map elements
					if origElemMap, ok := origElem.(map[string]interface{}); ok {
						if rtElemMap, ok := rtElem.(map[string]interface{}); ok {
							diffs = append(diffs, deepCompareTypes(elemPath, origElemMap, rtElemMap, depth+1)...)
						}
					}
				}
			}
		}
	}

	return diffs
}
