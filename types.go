package main

// FormatSpec represents a single field's formatting specification
type FormatSpec struct {
	Name   string
	Width  int
	Format string // Optional format string
}

// Config holds the configuration for server mode
type Config struct {
	Delimiter     string            `json:"delimiter"` // CSV delimiter
	Extensions    []string          `json:"extensions"`
	Port          int               `json:"port"`          // Server port
	Directory     string            `json:"directory"`     // Directory to watch for files
	OutputPattern string            `json:"outputPattern"` // Pattern for output filenames (e.g., "%s-converted")
	FormatString  string            `json:"formatString"`  // Format specification
	ProcessedDir  string            `json:"processedDir"`  // Directory to move processed files (optional)
	PollInterval  int               `json:"pollInterval"`  // How often to check for new files (seconds)
	OriginalFile  string            `json:"originalFile"`  // What to do with original file
	CertFile      string            `json:"certFile"`      // What to do with original file
	KeyFile       string            `json:"keyFile"`       // What to do with original file
	Replacements  map[string]string `json:"replacements"`  // Post-processing string replacements
}

var Stabilisierung = map[string]string{
	"A": "14",
	"B": "1",
	"C": "2",
	"D": "rmGEO",
	"E": "13",
	"F": "3",
	"G": "rmGEO",
	"H": "1",
	"J": "rmGEO",
	"K": "20",
	"L": "rmGEO",
	"M": "24",
	"T": "21",
}

const DEFAULT_FORMAT = "P:14 Y:12 X:12 H:10 MC:6 DT:10"
