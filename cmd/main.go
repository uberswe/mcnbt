package main

import (
	"encoding/json"
	"fmt"
	"github.com/uberswe/mcnbt"
	"log"
	"os"
	"strings"
)

func main() {
	// Ensure a file path argument is provided
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	path := os.Args[1]
	outputFormat := "json"        // Default output format
	outputPath := "./output.json" // Default output path

	// Parse command line arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--format=") {
			outputFormat = strings.TrimPrefix(arg, "--format=")
		} else if strings.HasPrefix(arg, "--output=") {
			outputPath = strings.TrimPrefix(arg, "--output=")
		} else if arg == "--help" {
			printUsage()
			os.Exit(0)
		}
	}

	// Parse the input file
	data, err := mcnbt.ParseAnyFromFileAsJSON(path)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", path, err)
	}

	// Debug: Print the type of data
	log.Printf("Data type: %T", data)

	// Convert to the requested format
	var outputData interface{}

	if outputFormat == "json" {
		// Keep the original format
		outputData = data
	} else {
		// First convert to standard format
		standardData, err := mcnbt.ConvertToStandard(data)
		if err != nil {
			log.Fatalf("Failed to convert to standard format: %v", err)
		}

		// Then convert from standard to the requested format
		if outputFormat == "standard" {
			outputData = standardData
		} else {
			outputData, err = mcnbt.ConvertFromStandard(standardData, outputFormat)
			if err != nil {
				log.Fatalf("Failed to convert to %s format: %v", outputFormat, err)
			}
		}
	}

	// Marshal the output data to JSON
	b, err := json.Marshal(outputData)
	if err != nil {
		log.Fatalf("Failed to marshal output data to JSON: %v", err)
	}

	// Output the result
	if len(b) < 20000 && outputPath == "./output.json" {
		// Small output and default output path, print to console
		prettyPrint(outputData)
	} else {
		// Large output or custom output path, save to file
		log.Printf("Output is %d bytes, saving to file: %s", len(b), outputPath)

		// Create the output file
		outputFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer outputFile.Close()

		// Use MarshalIndent for pretty formatting in the file
		prettyJSON, err := json.MarshalIndent(outputData, "", "	")
		if err != nil {
			log.Fatalf("Failed to marshal JSON with indentation: %v", err)
		}

		// Write the pretty JSON to the file
		_, err = outputFile.Write(prettyJSON)
		if err != nil {
			log.Fatalf("Failed to write to output file: %v", err)
		}

		log.Printf("Successfully saved JSON to %s", outputPath)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <file_path> [options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --format=<format>   Output format (json, standard, litematica, worldedit, create, worldsave)\n")
	fmt.Fprintf(os.Stderr, "  --output=<path>     Output file path\n")
	fmt.Fprintf(os.Stderr, "  --help              Show this help message\n")
}

func prettyPrint(o interface{}) {
	b, err := json.MarshalIndent(o, "", "	")
	if err != nil {
		log.Fatalf("Failed to marshal JSON with indentation: %v", err)
	}
	fmt.Println(string(b))
}
