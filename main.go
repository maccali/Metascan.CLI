package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"metascan/pkg"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Manifest struct {
	OutputFile       string         `json:"output_file"`
	OutputFormat     string         `json:"output_format"`
	TotalAttempted   int            `json:"total_attempted"`
	TotalProcessed   int            `json:"total_processed"`
	TotalWithErrors  int            `json:"total_with_errors"`
	OutputFileHashes pkg.FileHashes `json:"output_file_hashes"`
	GeneratedAt      string         `json:"generated_at"`
}

func main() {
	dirPath := flag.String("dir", ".", "Path to the directory to process")
	outputName := flag.String("output", "file_metadata_report", "Base name for the output file (without extension)")
	recursive := flag.Bool("r", false, "Process subdirectories recursively")
	extFilter := flag.String("ext", "", "Filter only files with this extension (e.g., .jpg)")
	format := flag.String("format", "csv", "Output format: csv or json")
	flag.Parse()

	if *dirPath == "" {
		log.Println("Error: Directory path is required.")
		flag.Usage()
		os.Exit(1)
	}

	dirInfo, err := os.Stat(*dirPath)
	if os.IsNotExist(err) {
		log.Fatalf("Error: Directory not found at '%s'", *dirPath)
	}
	if err != nil {
		log.Fatalf("Error getting directory info for '%s': %v", *dirPath, err)
	}
	if !dirInfo.IsDir() {
		log.Fatalf("Error: Path '%s' is not a directory.", *dirPath)
	}

	var results []*pkg.FileInfoData

	log.Printf("Processing directory: %s (Recursive: %t)", *dirPath, *recursive)
	var filesProcessed, filesAttempted, filesWithErrors int

	walkFunc := func(path string, d fs.DirEntry, errWalk error) error {
		if errWalk != nil {
			log.Printf("Error accessing '%s': %v (skipping)", path, errWalk)
			return nil
		}

		if d.IsDir() {
			if path == *dirPath {
				return nil
			}
			if !*recursive {
				return filepath.SkipDir
			}
			return nil
		}

		if *extFilter != "" && !strings.HasSuffix(strings.ToLower(path), strings.ToLower(*extFilter)) {
			return nil
		}

		filesAttempted++

		data, errProc := pkg.ProcessFile(path)
		if errProc != nil {
			log.Printf("Error processing file '%s': %v", path, errProc)
			filesWithErrors++
			return nil
		}

		if data != nil {
			results = append(results, data)
			filesProcessed++
		}

		return nil
	}

	if err := filepath.WalkDir(*dirPath, walkFunc); err != nil {
		log.Printf("Final error walking through directory '%s': %v", *dirPath, err)
	}

	outputFilePath := *outputName + "." + strings.ToLower(*format)

	switch strings.ToLower(*format) {
	case "json":
		jsonFile, err := os.Create(outputFilePath)
		if err != nil {
			log.Fatalf("Error creating JSON file '%s': %v", outputFilePath, err)
		}
		defer jsonFile.Close()

		encoder := json.NewEncoder(jsonFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			log.Fatalf("Error writing JSON: %v", err)
		}
		log.Printf("JSON report generated at: %s", outputFilePath)

	default:
		csvFile, err := os.Create(outputFilePath)
		if err != nil {
			log.Fatalf("Error creating CSV file '%s': %v", outputFilePath, err)
		}
		defer csvFile.Close()

		csvWriter := csv.NewWriter(csvFile)
		defer csvWriter.Flush()

		if err := pkg.WriteCSVHeader(csvWriter); err != nil {
			log.Fatalf("Error writing CSV header: %v", err)
		}

		for _, data := range results {
			if err := pkg.WriteCSVRecord(csvWriter, data); err != nil {
				log.Printf("Error writing record to CSV: %v", err)
			}
		}
		log.Printf("CSV report generated at: %s", outputFilePath)
	}

	manifest := Manifest{
		OutputFile:      outputFilePath,
		OutputFormat:    *format,
		TotalAttempted:  filesAttempted,
		TotalProcessed:  filesProcessed,
		TotalWithErrors: filesWithErrors,
		GeneratedAt:     time.Now().Format(time.RFC3339),
	}

	hashes, err := pkg.CalcFileHashes(outputFilePath)
	if err != nil {
		log.Printf("Warning: could not calculate output file hashes: %v", err)
	} else {
		manifest.OutputFileHashes = hashes
	}

	manifestFilePath := *outputName + "-manifest." + strings.ToLower(*format)
	manifestFile, err := os.Create(manifestFilePath)
	if err != nil {
		log.Fatalf("Error creating manifest: %v", err)
	}
	defer manifestFile.Close()

	switch strings.ToLower(*format) {
	case "json":
		encoder := json.NewEncoder(manifestFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(manifest); err != nil {
			log.Fatalf("Error writing JSON manifest: %v", err)
		}
	case "csv":
		manifestWriter := csv.NewWriter(manifestFile)
		defer manifestWriter.Flush()

		manifestWriter.Write([]string{
			"OutputFile", "OutputFormat", "TotalAttempted", "TotalProcessed", "TotalWithErrors",
			"MD5", "SHA1", "SHA256", "GeneratedAt",
		})

		manifestWriter.Write([]string{
			manifest.OutputFile,
			manifest.OutputFormat,
			strconv.Itoa(manifest.TotalAttempted),
			strconv.Itoa(manifest.TotalProcessed),
			strconv.Itoa(manifest.TotalWithErrors),
			manifest.OutputFileHashes.MD5,
			manifest.OutputFileHashes.SHA1,
			manifest.OutputFileHashes.SHA256,
			manifest.GeneratedAt,
		})
	}

	log.Printf("Manifest generated at: %s", manifestFilePath)
	log.Printf("Processing completed. Files attempted: %d. Files included: %d. Errors: %d.",
		filesAttempted, filesProcessed, filesWithErrors)
}
