package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func combiner() {
	// script to identify number of files to be processed and process each
	fmt.Println("VIVVIX AdSpender Converter: Combine VIVVIX reports")
	fmt.Println()
	reader := bufio.NewReader(os.Stdin)

	// Check if the directory is set in the settings.
	if settings.Directory == "" {
		// If no directory is set, propose the current executable's directory.
		exe, err := os.Executable() // Get the path of the executable.
		if err != nil {
			fmt.Println("Failed to determine the current executable's directory:", err)
			return
		}
		exeDir := filepath.Dir(exe) // Get the directory the executable is located in.

		// Prompt the user to confirm using the current directory.
		fmt.Printf("No directory set in settings. Would you like to use the current directory? (%s) [Y/n]: ", exeDir)
		choice, _ := reader.ReadString('\n')
		// Cleaning the input (removing \n or \r\n depending on the OS)
		choice = strings.TrimSpace(choice)

		if choice == "Y" || choice == "y" || choice == "" { // If user agrees or just hits enter (default yes).
			settings.Directory = exeDir

			err := saveSettings()
			if err != nil {
				fmt.Printf("Failed to save settings: %s\n", err)
				return // Exit if unable to save settings
			}
		} else {
			fmt.Println("Operation cancelled by the user.")
			return // Exit the function since the user did not agree.
		}
	} else {
		fmt.Println("Current directory in settings:", settings.Directory)
	}

	partialDir := settings.Directory + "/partial"   // Directory containing the CSV files
	metaDataDir := settings.Directory + "/metaData" // Directory containing the metadata
	combinedDir := settings.Directory + "/validated"
	processedDir := settings.Directory + "/processed"

	// Count CSV files to be processed
	files, err := os.ReadDir(partialDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	csvCount := 0
	for _, file := range files {
		if file.Name() == "rename_log.csv" {
			continue
		}
		if strings.HasSuffix(file.Name(), ".csv") {
			csvCount++
		}
	}

	fmt.Printf("Found %d CSV files in %s. Do you want to proceed? (y/n): ", csvCount, partialDir)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice != "y" && choice != "Y" {
		fmt.Println("No selection made.")
		return
	}

	err = processCSVFiles(partialDir, metaDataDir, combinedDir, processedDir)
	if err != nil {
		fmt.Println("Error processing CSV files:", err)
		return
	}

	fmt.Println("CSV files processed and combined successfully.")

}

// combineCSVFiles combines multiple CSV files into a single file.
func combineCSVFiles(files []string, combinedFilePath string) error {
	// Create or truncate the combined file
	combinedFile, err := os.Create(combinedFilePath)
	if err != nil {
		return err
	}
	// Do not close here; instead, defer the close operation after confirming the file is open.
	defer combinedFile.Close()

	writer := csv.NewWriter(combinedFile)
	// Ensure the writer buffer is flushed when done.
	defer writer.Flush()

	// Process each file.
	for _, file := range files {
		// Open the file for reading.
		csvFile, err := os.Open(file)
		if err != nil {
			return err
		}

		reader := csv.NewReader(csvFile)

		// Read the current CSV file's records.
		records, err := reader.ReadAll()
		if err != nil {
			csvFile.Close() // Close the file promptly, avoiding defer in the loop.
			return err
		}

		// Skip the header if it's not the first file to avoid duplicate headers.
		if file != files[0] {
			records = records[1:]
		}

		// Write records to the combined CSV file.
		err = writer.WriteAll(records)
		if err != nil {
			csvFile.Close() // Close the file promptly, avoiding defer in the loop.
			return err
		}

		// Close the current CSV file before moving to the next one.
		// This step is crucial to avoid having too many files open simultaneously.
		if errClose := csvFile.Close(); errClose != nil {
			return errClose // Handle the close operation error.
		}
	}

	return nil // No error occurred during the processing.
}

func processCSVFiles(partialDir, metaDataDir, combinedDir, processedDir string) error {
	// Locate all CSV files in the partial directory.
	csvFiles, err := filepath.Glob(filepath.Join(partialDir, "*.csv"))
	if err != nil {
		return fmt.Errorf("error finding CSV files: %v", err)
	}

	// Check if csvFiles contains data
	if len(csvFiles) == 0 {
		fmt.Println("No CSV files found in the directory.")
		return nil // or return an appropriate error
	}

	type dateRange struct {
		start string
		end   string
	}

	// Map to organize files by their date range.
	dateRanges := make(map[dateRange][]string)

	// Process each CSV file.
	for _, file := range csvFiles {
		baseName := filepath.Base(file)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// Construct the metadata file path and read it.
		metaDataPath := filepath.Join(metaDataDir, baseName+"_metadata.json")
		jsonFile, err := os.ReadFile(metaDataPath)
		if err != nil {
			return fmt.Errorf("error reading metadata file: %v", err)
		}

		// Decode the JSON metadata file.
		var metaData Metadata
		err = json.Unmarshal(jsonFile, &metaData)
		if err != nil {
			return fmt.Errorf("error decoding metadata JSON: %v", err)
		}

		// Create a date range based on the StartDate and EndDate.
		dr := dateRange{start: metaData.StartDate, end: metaData.EndDate}

		// Add the file to the list in the map based on the date range.
		dateRanges[dr] = append(dateRanges[dr], file)

		// For debugging: print out the file being processed and its date range
		fmt.Printf("Processing file: %s with date range: %s to %s\n", file, dr.start, dr.end)
	}

	combProcessedDir := processedDir + "/combined"
	combMetaDir := metaDataDir + "/archive"

	// Create 'processed' folder if it doesn't exist
	if _, err := os.Stat(combProcessedDir); os.IsNotExist(err) {
		err := os.MkdirAll(combProcessedDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", processedDir, err)
		}
	}

	// Create 'metadata atchive' folder if it doesn't exist
	if _, err := os.Stat(combMetaDir); os.IsNotExist(err) {
		err := os.MkdirAll(combMetaDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", combMetaDir, err)
		}
	}

	// For each unique date range, combine files and create new metadata.
	for dr, files := range dateRanges {
		if len(files) > 1 {
			combinedFileName := dr.start + ".csv"
			combinedFilePath := filepath.Join(combinedDir, combinedFileName)

			err = combineCSVFiles(files, combinedFilePath)
			if err != nil {
				return fmt.Errorf("error combining CSV files: %v", err)
			}

			for _, file := range files {
				metaDataPath := filepath.Join(metaDataDir, strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))+"_metadata.json")
				jsonFile, err := os.ReadFile(metaDataPath)
				if err != nil {
					return fmt.Errorf("error reading metadata file for summation: %v", err)
				}

				var originalMetaData Metadata
				err = json.Unmarshal(jsonFile, &originalMetaData)
				if err != nil {
					return fmt.Errorf("error decoding original metadata JSON: %v", err)
				}

			}

			// Create a new metadata instance for the combined file.
			newMetaData := Metadata{
				FileName:     combinedFileName,
				OriginalFile: "NA", // As it's a combined file now.
				StartDate:    dr.start,
				EndDate:      dr.end,
				Type:         "combined",
			}

			// Convert the new metadata to JSON.
			newJsonContent, err := json.MarshalIndent(newMetaData, "", "    ")
			if err != nil {
				return fmt.Errorf("error creating JSON content for new metadata: %v", err)
			}

			metaFileNameWithoutExtension := strings.TrimSuffix(combinedFileName, ".csv")

			// Save the new metadata to a file.
			newMetaDataPath := filepath.Join(metaDataDir, metaFileNameWithoutExtension+"_metadata.json")
			if err = os.WriteFile(newMetaDataPath, newJsonContent, 0644); err != nil {
				return fmt.Errorf("error writing new metadata file: %v", err)
			}

			for _, originalFile := range files {
				if settings.AutoDelete {
					err = os.Remove(originalFile)
					if err != nil {
						// Optional: Instead of returning immediately, you might log the error
						// and continue deleting the other files.
						return fmt.Errorf("error deleting original file: %v", err)
					}
					metaDataFilename := strings.TrimSuffix(filepath.Base(originalFile), filepath.Ext(originalFile)) + "_metadata.json"
					originalMetaDataPath := filepath.Join(metaDataDir, metaDataFilename)

					err = os.Remove(originalMetaDataPath)
					if err != nil {
						return fmt.Errorf("error deleting original metadata file: %v", err)
					}
				} else {
					newFilePath := filepath.Join(combProcessedDir, filepath.Base(originalFile))

					// "Move" the file by renaming it's path to the new path.
					err = os.Rename(originalFile, newFilePath)
					if err != nil {
						return fmt.Errorf("error moving original file to archive: %v", err)
					}
					// Construct the metadata file path for each original file.
					metaDataFilename := strings.TrimSuffix(filepath.Base(originalFile), filepath.Ext(originalFile)) + "_metadata.json"
					originalMetaDataPath := filepath.Join(metaDataDir, metaDataFilename)

					// Determine the new path for the original metadata file in the metadata archive directory.
					newMetaDataPath := filepath.Join(combMetaDir, metaDataFilename)

					// "Move" the metadata file by renaming its path to the new path in the metadata archive directory.
					err = os.Rename(originalMetaDataPath, newMetaDataPath)
					if err != nil {
						return fmt.Errorf("error moving original metadata file to archive: %v", err)
					}
				}
			}
		}
	}

	return nil
}
