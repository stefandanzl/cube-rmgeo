package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ProcessFile processes a single CSV file
func ProcessFile(inputFile, outputFile, delimiter string, formatSpecs []FormatSpec, replacements map[string]string) error {
	// Open the input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Set the field delimiter
	sepChar := delimiter
	if sepChar == "\\t" || sepChar == "tab" {
		sepChar = "\t"
	}
	reader.Comma = []rune(sepChar)[0]

	// Read all data into memory
	data, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV data: %v", err)
	}

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	// Process each row and collect output
	var allOutput strings.Builder

	for r, row := range data {
		if len(row) == 0 {
			continue // Skip empty rows
		}

		// Use the dynamic formatting function
		outputLine, err := BuildFormattedLine(row, formatSpecs, r)
		if err {
			continue
		}

		allOutput.WriteString(outputLine + "\r\n")
	}

	// Apply post-processing replacements
	finalOutput := ApplyReplacements(allOutput.String(), replacements)

	// Write to output file
	outFile.WriteString(finalOutput)

	return nil
}

// ProcessDirectoryFiles processes all CSV files in the directory that haven't been converted yet
func ProcessDirectoryFiles(directory, outputPattern, delimiter string,
	formatSpecs []FormatSpec, processedDir string, originalFile string, extensionsMap map[string]bool, replacements map[string]string) ([]string, error) {

	// List all files in the directory
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	processedFiles := []string{}

	// Process each file
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(directory, file.Name())

		// Check if it's a CSV file and hasn't been processed yet
		ext := filepath.Ext(filePath)
		if _, exists := extensionsMap[strings.Trim(ext, ".")]; !exists {
			// fmt.Println(ext + " exists not")
			continue
		}
		// fmt.Println(ext + " exists")
		// if !strings.EqualFold(ext, ".csv") {
		// 	continue // Skip non-CSV files
		// }

		// Generate output filename
		baseName := strings.TrimSuffix(file.Name(), ext)
		outputName := fmt.Sprintf(outputPattern, baseName)
		outputPath := filepath.Join(processedDir, outputName)

		// // Skip if output file already exists and is newer
		// if outputFileInfo, err := os.Stat(outputPath); err == nil {
		// 	inputFileInfo, err := os.Stat(filePath)
		// 	if err == nil && outputFileInfo.ModTime().After(inputFileInfo.ModTime()) {
		// 		log.Printf("Skipping already processed file: %s\n", file.Name())
		// 		continue
		// 	}
		// }

		// Process the file
		log.Printf("Processing file: %s\n", file.Name())
		err := ProcessFile(filePath, outputPath, delimiter, formatSpecs, replacements)
		if err != nil {
			log.Printf("Error processing file %s: %v\n", file.Name(), err)
			continue
		}

		processedFiles = append(processedFiles, file.Name())

		// Move to processed directory if specified
		if processedDir != "" {
			if err := os.MkdirAll(processedDir, 0755); err != nil {
				log.Printf("Error creating processed directory: %v\n", err)
			} else {
				if originalFile == "move" {
					newPath := filepath.Join(processedDir, file.Name())
					if err := os.Rename(filePath, newPath); err != nil {
						log.Printf("Error moving original file: %v\n", err)
					}
				} else if originalFile == "delete" {
					// TODO
					if err := os.Remove(filepath.Join(directory, file.Name())); err != nil {
						log.Printf("Error deleting original file: %v\n", err)
					}
				}
			}
		}
	}

	log.Printf("Processed %d files\n", len(processedFiles))
	return processedFiles, nil
}
