// VIVVIX AdSpender Conversion App
// Copyright (c) 2023 Northwestern University
// Author: Andrew D'Amico
// Date: 10/18/2023

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Creates the struct for holding the start date and end date of the vivvix report
type DateRange struct {
	StartDate string
	EndDate   string
}

type Metadata struct {
	FileName  string `json:"FileName"`
	StartDate string `json:"StartDate"`
	EndDate   string `json:"EndDate"`
}

// parser reads a string and extracts the dates
func parser(line string) DateRange {
	// Define an empty DateRange struct to return in case of errors
	var empty DateRange

	// Use regex to find all date patterns
	re := regexp.MustCompile(`(\d{1,2}/\d{1,2}/\d{4})`)
	matches := re.FindAllStringSubmatch(line, -1)

	if len(matches) < 2 {
		fmt.Println("Couldn't extract both dates from string")
		return empty
	}

	dateStr1 := matches[0][1]
	dateStr2 := matches[1][1]

	// Parse the first date
	t1, err1 := time.Parse("1/2/2006", dateStr1)
	// Parse the second date
	t2, err2 := time.Parse("1/2/2006", dateStr2)

	if err1 != nil || err2 != nil {
		fmt.Println("Error parsing dates")
		return empty
	}

	// Return the populated DateRange struct
	return DateRange{
		StartDate: t1.Format("01022006"),
		EndDate:   t2.Format("01022006"),
	}
}

func safeClose(file *os.File) {
	if file == nil {
		return // If the file reference is nil, we don't need to do anything.
	}

	// Try to close the file and check for the specific error indicating the file is already closed.
	err := file.Close()
	if err != nil {
		// We can check for a specific error like "file already closed" and avoid printing or handling it if that's the case.
		// Otherwise, we handle unexpected errors (like a failing disk or a bug causing premature closing).
		if !strings.Contains(err.Error(), "file already closed") {
			fmt.Printf("Encountered an error while closing file: %v\n", err)
		}
	}
}

func processFile(dir, filename string) bool {
	filePath := dir + "/" + filename

	// Open the file for reading.
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return false
	}

	// Create a temporary file.
	tempFilePath := filePath + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return false
	}

	defer safeClose(tempFile)

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(tempFile)

	lineCount := 0
	var line5 string
	skippedLines := 5 // Number of lines to skip.

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line contains "GRAND TOTAL", stop processing if it does.
		if strings.Contains(line, "GRAND TOTAL") {
			break
		}

		lineCount++

		// Save line 5 for other processing.
		if lineCount == 5 {
			line5 = line
			// Don't break; continue reading to skip the lines.
		}

		// Skip the first 'skippedLines' number of lines.
		if lineCount > skippedLines {
			_, err = writer.WriteString(line + "\n")
			if err != nil {
				fmt.Printf("Error writing to temporary file: %v\n", err)
				return false
			}
		}
	}

	// Ensure all writes are actually written to disk.
	if err = writer.Flush(); err != nil {
		fmt.Printf("Error flushing writer: %v\n", err)
		return false
	}

	// Close both files.
	if err = file.Close(); err != nil {
		fmt.Printf("Error closing original file: %v\n", err)
		return false
	}
	if err = tempFile.Close(); err != nil {
		fmt.Printf("Error closing temporary file: %v\n", err)
		return false
	}

	if settings.AutoDelete {

		// Delete the original file.
		if err = os.Remove(filePath); err != nil {
			fmt.Printf("Error removing original file: %v\n", err)
			return false
		}
	}

	// Extract the dates from the line.
	dateRange := parser(line5)

	// Check if the dates were successfully extracted.
	if dateRange.StartDate == "" || dateRange.EndDate == "" {
		fmt.Printf("Error: Couldn't extract dates from file %s\n", filename)

		// Before returning false, we should clean up by removing the temporary file.
		// This ensures no incorrect files are left behind due to the error.
		cleanupErr := os.Remove(tempFilePath)
		if cleanupErr != nil {
			fmt.Printf("Error cleaning up temporary file: %v\n", cleanupErr)
			// Even if cleanup fails, we still return false as the main operation was not successful.
		}

		return false // Return false because the process failed at an important step.
	}

	startDate := dateRange.StartDate
	newName := startDate + ".csv"
	validateDir := dir + "/validated"
	newPath := dir + "/validated/" + newName

	// Close the file explicitly before renaming
	defer safeClose(file)

	// Create 'validated' folder if it doesn't exist
	if _, err := os.Stat(validateDir); os.IsNotExist(err) {
		err := os.MkdirAll(validateDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", validateDir, err)
			return false

		}
	}

	// Create 'metadata' folder if it doesn't exist
	metaDataDir := dir + "/metadata"
	if _, err := os.Stat(metaDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(metaDataDir, 0755)
		if err != nil {
			fmt.Printf("Error creating metadata directory %s: %v\n", metaDataDir, err)
			return false
		}
	}

	if err := os.Rename(tempFilePath, newPath); err != nil {
		fmt.Printf("Error renaming file: %v\n", filename, newName, err)
		return false

	}

	// After file is renamed, create metadata
	metaData := Metadata{
		FileName:  newName,
		StartDate: dateRange.StartDate,
		EndDate:   dateRange.EndDate,
	}

	// Log the change
	logChange(dir+"/rename_log.csv", filename, newName, dateRange.StartDate, dateRange.EndDate)

	// Write the metadata to a new file in the 'metadata' folder
	metaDataPath := metaDataDir + "/" + strings.TrimSuffix(newName, ".csv") + "_metadata.json"
	err2 := writeMetaData(metaData, metaDataPath)
	if err2 != nil {
		// handle error
		fmt.Println("Error writing metadata:", err)
	}

	return true
}

// New function to write metadata information
func writeMetaData(metaData Metadata, metaDataPath string) error {
	// Convert struct to JSON
	file, err := os.Create(metaDataPath)
	if err != nil {
		return err
	}

	defer safeClose(file)

	encoder := json.NewEncoder(file)
	err = encoder.Encode(metaData)
	if err != nil {
		return err
	}

	return nil
}

func logChange(logFile, oldName, newName, startDate, endDate string) {
	// Check if log file already exists
	fileExists := true
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fileExists = false
	}

	// Open the log file for appending, or create it if it doesn't exist
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}

	defer safeClose(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// If the file didn't exist write the headers
	if !fileExists {
		if err := writer.Write([]string{"Original Name", "New Name", "Start Date", "End Date"}); err != nil {
			fmt.Println("Error writing headers to log:", err)
			return
		}
	}

	// Write the old and new filenames to the CSV log
	if err := writer.Write([]string{oldName, newName, startDate, endDate}); err != nil {
		fmt.Println("Error writing to log:", err)
	}
}

func converter() {

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

	// Count CSV files to be processed
	files, err := os.ReadDir(settings.Directory)
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

	fmt.Printf("Found %d CSV files in %s. Do you want to proceed? (y/n): ", csvCount, settings.Directory)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice != "y" && choice != "Y" {
		fmt.Println("No selection made.")
		return
	}

	successfulCount := 0
	errorEncountered := false // New variable to track if any file processing failed.

	for _, file := range files {
		if file.Name() == "rename_log.csv" {
			continue
		}
		if strings.HasSuffix(file.Name(), ".csv") {
			success := processFile(settings.Directory, file.Name()) // Process the file and store if it was successful
			if success {
				successfulCount++
			} else {
				errorEncountered = true // Set the error flag if a file fails to process.
			}
		}
	}

	// Provide feedback based on the outcomes of file processing.
	if successfulCount > 0 {
		fmt.Printf("%d files were successfully converted.\n", successfulCount)
	}
	if errorEncountered {
		fmt.Println("Some files were not processed due to errors.")
	}

	if successfulCount == 0 && !errorEncountered {
		fmt.Println("No files were available or matched the criteria for processing.")
	}

}
