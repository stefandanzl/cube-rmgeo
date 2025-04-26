package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FormatSpec represents a single field's formatting specification
type FormatSpec struct {
	Name   string
	Width  int
	Format string // Optional format string
}

// ParseFormatString parses a format string like "P:14 Y:12 X:12 H:10 MC:6 DT:8"
// into a slice of FormatSpec
func ParseFormatString(formatStr string) []FormatSpec {
	parts := strings.Fields(formatStr)
	specs := make([]FormatSpec, 0, len(parts))

	for _, part := range parts {
		// Split each part by colon
		colonParts := strings.Split(part, ":")
		if len(colonParts) != 2 {
			fmt.Printf("Warning: Invalid format part '%s', skipping\n", part)
			continue
		}

		// Get field name and width
		name := colonParts[0]
		width, err := strconv.Atoi(colonParts[1])
		if err != nil {
			fmt.Printf("Warning: Invalid width in '%s', skipping\n", part)
			continue
		}

		specs = append(specs, FormatSpec{
			Name:   name,
			Width:  width,
			Format: "%s", // Default to string format
		})
	}

	return specs
}

// BuildFormattedLine builds a formatted line based on the format specification
func BuildFormattedLine(row []string, specs []FormatSpec) string {
	var result strings.Builder

	// Make sure we have enough data
	if len(row) < len(specs) {
		fmt.Println("Warning: Not enough data to match format specification")
		// Pad the row with empty strings if needed
		for len(row) < len(specs) {
			row = append(row, "")
		}
	}

	// Build the formatted line
	for i, spec := range specs {
		field := ""
		if i < len(row) {
			field = row[i]
		}

		// Format the field with padding
		paddedField := dataPadding(field, spec.Width)

		// Add to result
		result.WriteString(fmt.Sprintf("%s=%s", spec.Name, paddedField))
	}

	return result.String()
}

func dataPadding(data string, length int) string {
	for len(data) < length {
		data = data + " "
	}
	return data
}

func main() {
	// Define command-line flags
	outputFile := flag.String("o", "", "Output file path (optional)")
	separator := flag.String("s", ",", "Input CSV separator (optional, default is comma)")
	formatString := flag.String("f", "P:14 Y:12 X:12 H:10 MC:6 DT:8", "Format string for output (optional)")
	configFile := flag.String("c", "", "Config file for server mode")
	flag.Parse()

	// Check if server mode is activated
	if *configFile != "" {
		// Server mode - all other flags are ignored
		config, err := LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}

		// Start the server
		StartServer(config)
		return
	}

	// CLI mode
	// Get the input file from the first argument
	args := flag.Args()
	if len(args) < 1 {
		// For Windows drag and drop, show a helpful message and wait before exiting
		fmt.Println("Error: Input file is required")
		fmt.Println("Usage: csv-converter input.csv [-o output.csv] [-s separator] [-f formatString]")
		fmt.Println("       csv-converter -c config.json (server mode)")
		fmt.Println("Example format string: \"P:14 Y:12 X:12 H:10 MC:6 DT:8\"")
		fmt.Println("\nPress Enter to close this window...")

		// Wait for user input before closing (helpful for drag-and-drop)
		fmt.Scanln()
		os.Exit(1)
	}
	inputFile := args[0]

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("Error: Input file '%s' does not exist\n", inputFile)
		fmt.Println("\nPress Enter to close this window...")
		fmt.Scanln()
		os.Exit(1)
	}

	// Generate default output filename if not specified
	if *outputFile == "" {
		ext := filepath.Ext(inputFile)
		baseName := strings.TrimSuffix(inputFile, ext)
		*outputFile = baseName + "-converted" + ext
	}

	// Parse the format string
	formatSpecs := ParseFormatString(*formatString)
	if len(formatSpecs) == 0 {
		fmt.Println("Error: Invalid format string. Using default format.")
		formatSpecs = ParseFormatString("P:14 Y:12 X:12 H:10 MC:6 DT:8")
	}

	// Process the file
	err := ProcessFile(inputFile, *outputFile, *separator, formatSpecs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("\nPress Enter to close this window...")
		fmt.Scanln()
		os.Exit(1)
	}

	fmt.Printf("Conversion complete! Output saved to: %s\n", *outputFile)

	// For Windows drag and drop, always pause so the user can see the result
	fmt.Println("\nPress Enter to close this window...")
	fmt.Scanln()
}

// StartServer starts the server mode with the given configuration
func StartServer(config *ServerConfig) {
	log.Printf("Starting server on port %d\n", config.Port)
	log.Printf("Watching directory: %s\n", config.Directory)
	log.Printf("Using delimiter: '%s'\n", config.Delimiter)
	log.Printf("Using format string: '%s'\n", config.FormatString)

	// Parse the format string
	formatSpecs := ParseFormatString(config.FormatString)
	if len(formatSpecs) == 0 {
		log.Println("Error: Invalid format string. Using default format.")
		formatSpecs = ParseFormatString("P:14 Y:12 X:12 H:10 MC:6 DT:8")
	}

	// Setup webhook endpoint
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("Received webhook notification, processing files...")

		// Process files in the directory
		processedFiles, err := ProcessDirectoryFiles(config.Directory, config.OutputPattern,
			config.Delimiter, formatSpecs, config.ProcessedDir)

		if err != nil {
			log.Printf("Error processing files: %v\n", err)
			http.Error(w, fmt.Sprintf("Error processing files: %v", err), http.StatusInternalServerError)
			return
		}

		// Return the list of processed files
		response := map[string]interface{}{
			"status":         "success",
			"processedFiles": processedFiles,
			"timestamp":      time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Setup status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":     "running",
			"directory":  config.Directory,
			"startedAt":  time.Now().Format(time.RFC3339),
			"delimiter":  config.Delimiter,
			"formatSpec": config.FormatString,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Start periodically checking for files
	if config.PollInterval > 0 {
		go func() {
			for {
				log.Println("Polling directory for new files...")
				_, err := ProcessDirectoryFiles(config.Directory, config.OutputPattern,
					config.Delimiter, formatSpecs, config.ProcessedDir)
				if err != nil {
					log.Printf("Error during scheduled processing: %v\n", err)
				}

				// Wait for the configured interval
				time.Sleep(time.Duration(config.PollInterval) * time.Second)
			}
		}()
	}

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%d", config.Port)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

// ProcessDirectoryFiles processes all CSV files in the directory that haven't been converted yet
func ProcessDirectoryFiles(directory, outputPattern, delimiter string,
	formatSpecs []FormatSpec, processedDir string) ([]string, error) {

	// List all files in the directory
	files, err := ioutil.ReadDir(directory)
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
		if !strings.EqualFold(ext, ".csv") {
			continue // Skip non-CSV files
		}

		// Generate output filename
		baseName := strings.TrimSuffix(file.Name(), ext)
		outputName := fmt.Sprintf(outputPattern, baseName) + ext
		outputPath := filepath.Join(directory, outputName)

		// Skip if output file already exists and is newer
		if outputFileInfo, err := os.Stat(outputPath); err == nil {
			if outputFileInfo.ModTime().After(file.ModTime()) {
				log.Printf("Skipping already processed file: %s\n", file.Name())
				continue
			}
		}

		// Process the file
		log.Printf("Processing file: %s\n", file.Name())
		err := ProcessFile(filePath, outputPath, delimiter, formatSpecs)
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
				newPath := filepath.Join(processedDir, file.Name())
				if err := os.Rename(filePath, newPath); err != nil {
					log.Printf("Error moving processed file: %v\n", err)
				}
			}
		}
	}

	log.Printf("Processed %d files\n", len(processedFiles))
	return processedFiles, nil
}

// ProcessFile processes a single CSV file
func ProcessFile(inputFile, outputFile, delimiter string, formatSpecs []FormatSpec) error {
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

	// Process each row
	for _, row := range data {
		if len(row) == 0 {
			continue // Skip empty rows
		}

		// Use the dynamic formatting function
		outputLine := BuildFormattedLine(row, formatSpecs)

		// Write to output file
		outFile.WriteString(outputLine + "\n")
	}

	return nil
} // ServerConfig holds the configuration for server mode
type ServerConfig struct {
	Delimiter     string `json:"delimiter"`     // CSV delimiter
	Port          int    `json:"port"`          // Server port
	Directory     string `json:"directory"`     // Directory to watch for files
	OutputPattern string `json:"outputPattern"` // Pattern for output filenames (e.g., "%s-converted")
	FormatString  string `json:"formatString"`  // Format specification
	ProcessedDir  string `json:"processedDir"`  // Directory to move processed files (optional)
	PollInterval  int    `json:"pollInterval"`  // How often to check for new files (seconds)
}

// LoadConfig loads the server configuration from a JSON file
func LoadConfig(configPath string) (*ServerConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config ServerConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Set defaults for missing values
	if config.Delimiter == "" {
		config.Delimiter = ","
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.OutputPattern == "" {
		config.OutputPattern = "%s-converted"
	}
	if config.FormatString == "" {
		config.FormatString = "P:14 Y:12 X:12 H:10 MC:6 DT:8"
	}
	if config.PollInterval == 0 {
		config.PollInterval = 30 // Default to 30 seconds
	}

	return &config, nil
}
