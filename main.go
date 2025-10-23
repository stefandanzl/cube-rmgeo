package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	file, _ := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer file.Close()
	log.SetOutput(file)

	// Define command-line flags
	outputFile := flag.String("o", "", "Output file path (optional)")
	separator := flag.String("d", ";", "Input CSV deliminator (optional, default is comma)")
	formatString := flag.String("f", DEFAULT_FORMAT, "Format string for output (optional)")
	configFile := flag.String("c", "", "Config file for server mode")
	serverMode := flag.Bool("s", false, "If given will enable server mode at specified port in config file")
	flag.Parse()

	// Check if config file is specified
	if *configFile != "" {
		// Load config file for settings
		config, err := LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}

		extensionsMap := make(map[string]bool)
		for _, e := range config.Extensions {
			extensionsMap[e] = true
		}
		fmt.Println(extensionsMap)

		// Parse the format string
		formatSpecs := ParseFormatString(config.FormatString)
		if len(formatSpecs) == 0 {
			log.Println("Error: Invalid format string. Using default format.")
			formatSpecs = ParseFormatString(DEFAULT_FORMAT)
		}

		processTask := func() ([]string, error) {
			return ProcessDirectoryFiles(config.Directory, config.OutputPattern,
				config.Delimiter, formatSpecs,
				config.ProcessedDir, config.OriginalFile,
				extensionsMap, config.Replacements)
		}

		if *serverMode {
			// Start the server
			StartServer(config, processTask)
			return
		}

		// Check if files are passed as arguments (drag-and-drop or CLI with files)
		args := flag.Args()

		log.Printf(`flag.Args() has %d elements: `, len(args))

		if len(args) > 0 {
			// Drag-and-drop or CLI mode with specific files - use config settings but process only the specified files
			inputFile := args[0]

			// Check if file exists
			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				fmt.Printf("Error: Input file '%s' does not exist\n", inputFile)
				fmt.Println("\nPress Enter to close this window...")
				fmt.Scanln()
				os.Exit(1)
			}

			// Generate output filename using config pattern
			ext := filepath.Ext(inputFile)
			baseName := strings.TrimSuffix(filepath.Base(inputFile), ext)
			outputName := fmt.Sprintf(config.OutputPattern, baseName)

			// Determine output path
			var outputFile string
			if config.ProcessedDir != "" {
				outputFile = filepath.Join(config.ProcessedDir, outputName)
			} else {
				outputFile = filepath.Join(filepath.Dir(inputFile), outputName)
			}

			// Process the file using config settings
			err := ProcessFile(inputFile, outputFile, config.Delimiter, formatSpecs, config.Replacements)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				fmt.Println("\nPress Enter to close this window...")
				fmt.Scanln()
				os.Exit(1)
			}

			fmt.Printf("Conversion complete! Output saved to: %s\n", outputFile)
			fmt.Println("\nPress Enter to close this window...")
			fmt.Scanln()
			return
		}

		// No files specified - run directory processing mode
		processedFiles, err := processTask()
		if err != nil {
			fmt.Printf("Error occurred processing files: %v", err)
		}
		fmt.Printf("Processed following files: %s \n", processedFiles)

		// Start periodically checking for files
		if config.PollInterval > 0 {
			go func() {
				for {
					log.Println("Polling directory for new files...")
					_, err := processTask()
					if err != nil {
						log.Printf("Error during scheduled processing: %v\n", err)
					}

					// Wait for the configured interval
					time.Sleep(time.Duration(config.PollInterval) * time.Second)
				}
			}()
		}
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
		formatSpecs = ParseFormatString(DEFAULT_FORMAT)
	}

	// Process the file
	err := ProcessFile(inputFile, *outputFile, *separator, formatSpecs, nil)
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
