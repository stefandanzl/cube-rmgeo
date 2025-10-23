package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

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
func BuildFormattedLine(row []string, specs []FormatSpec, rownum int) (rowString string, err bool) {
	var result strings.Builder

	// Make sure we have enough data
	if len(row) < len(specs) {
		fmt.Printf("Warning: Not enough data to match format specification: %v \tSoll: %v Ist: %v\n", row, len(row), len(specs))
		log.Printf("Warning: Not enough data to match format specification: %v \tSoll: %v Ist: %v\n", row, len(row), len(specs))
		// Pad the row with empty strings if needed
		for len(row) < len(specs) {
			row = append(row, "")
		}
	}

	// Header needs separate formatting
	if rownum == 0 {
		for i, spec := range specs {
			field := row[i]

			if i >= len(row) {
				continue
			}

			for f := len(field); f <= spec.Width+1; f++ {
				field = field + " "
			}
			if len(field) > spec.Width {
				field = field[:spec.Width+1+len(spec.Name)]
			}

			paddedField := field //[:spec.Width+2]
			// paddedField := dataPadding(field, spec.Width+len(spec.Name)+1)
			fmt.Println(field, "   ", paddedField, len(field), spec.Width+2)
			result.WriteString(fmt.Sprintf("%s ", paddedField))
		}
		return result.String(), false
	}

	row = Process_P_column(row, specs)
	if row == nil {
		return "", true
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
		result.WriteString(fmt.Sprintf("%s=%s ", spec.Name, paddedField))
	}
	return result.String(), false
}

// P column is the first column (index 0), there are a couple of conditions
// that lead to formatting
func Process_P_column(row []string, specs []FormatSpec) []string {
	if row[3] == "0.000" {
		return nil
	}
	strColumnP := row[0]
	strColumnP = strings.TrimPrefix(strColumnP, "_")

	gps := strings.HasSuffix(strColumnP, "_GPS")
	if gps {
		strColumnP = strings.TrimSuffix(strColumnP, "_GPS")
	}
	loc := strings.HasSuffix(strColumnP, "_LOC")
	if loc {
		strColumnP = strings.TrimSuffix(strColumnP, "_LOC")
	}

	strColumnP = strings.Replace(strColumnP, "_", "-", 1)

	// Stabilisierungscharacter in P column
	charStabil := string(strColumnP[len(strColumnP)-2])
	// MC=
	row[4] = Stabilisierung[charStabil]

	if gps {
		strColumnP = strColumnP + "_GPS"
	}

	row[0] = strColumnP
	return row
}
