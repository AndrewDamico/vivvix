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
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Creates the struct for holding the start date and end date of the vivvix report
type DateRange struct {
	StartDate string
	EndDate   string
}

type Metadata struct {
	FileName      string `json:"FileName"`
	OriginalFile  string `json:"OriginalFile"`
	StartDate     string `json:"StartDate"`
	EndDate       string `json:"EndDate"`
	WeekStart     string `json:"WeekStart"`
	DayCount      int    `json:"DayCount"`
	Type          string `json:"Type"`
	NObservations int    `json:"NObservations"`
}

func getWeekStart(date time.Time) time.Time {
	// Function to calculate the start of the week (Monday)
	// we consider Monday as the start of the week
	offset := int(time.Monday - date.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStart := date.AddDate(0, 0, offset)
	return weekStart
}

func getDayCount(startDate, endDate time.Time) int {
	// Function to calculate the number of days between two dates
	return int(endDate.Sub(startDate).Hours()/24) + 1 // +1 because the end date is inclusive
}

func getType(dayCount int, searchIndicator string) string {
	// Determine the "type" based on the searchIndicator and dayCount.
	if searchIndicator == "_S" {
		return "search"
	} else if searchIndicator == "_W" {
		return "no search"
	} else if dayCount < 7 {
		return "partial"
	}
	return "weekly"
}

// intInSlice checks if a number is in the list.
func intInSlice(num int, list []int) bool {
	for _, a := range list {
		if a == num {
			return true
		}
	}
	return false
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

func SafeClose(file *os.File) {
	// safely closes a file if it is not already closed. Avoids unnecessary errors.
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
	// processes a file removing VIVVIX header and footer information
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

	defer SafeClose(tempFile)

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

	processedDir := dir + "/processed"
	processedPath := processedDir + "/" + filename

	// Create 'processed' folder if it doesn't exist
	if _, err := os.Stat(processedDir); os.IsNotExist(err) {
		err := os.MkdirAll(processedDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", processedDir, err)
			return false

		}
	}

	if settings.AutoDelete {

		// Delete the original file.
		if err = os.Remove(filePath); err != nil {
			fmt.Printf("Error removing original file: %v\n", err)
			return false
		}
	} else {
		// Move the file by renaming its path.
		if err = os.Rename(filePath, processedPath); err != nil {
			fmt.Printf("Error moving file to processed folder: %v\n", err)
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

	// Parse the dates to *time.Time, as we need them to calculate WeekStart and DayCount
	tStart, _ := time.Parse("01022006", dateRange.StartDate)
	tEnd, _ := time.Parse("01022006", dateRange.EndDate)

	// Calculate WeekStart and DayCount
	weekStart := getWeekStart(tStart)
	dayCount := getDayCount(tStart, tEnd)

	// Determine the appropriate file name based on whether the week is complete and which day it starts on.
	partialWeekIndicator := ""
	if dayCount < 7 {
		if tStart.Weekday() == time.Monday { // the week is partial, and starts on a Monday
			partialWeekIndicator = "_1"
		} else { // the week is partial, but does not start on a Monday
			partialWeekIndicator = "_2"
		}
	}

	// Determine if this is only a "search" related dataframe
	searchIndicator := ""
	if strings.Contains(filename, "_S") {
		searchIndicator = "_S"
	} else if strings.Contains(filename, "_W") {
		searchIndicator = "_W"
	}

	newName := weekStart.Format("01022006") + partialWeekIndicator + searchIndicator + ".csv"
	validateDir := dir + "/validated"
	partialDir := dir + "/partial"

	// Close the file explicitly before renaming
	defer SafeClose(file)

	// Modify the CSVs

	// Open the temporary CSV file for reading.
	tempFile, err = os.Open(tempFilePath)
	if err != nil {
		fmt.Printf("Error opening temporary file for reading: %v\n", err)
		return false
	}
	defer SafeClose(tempFile) // ensure the file is closed after this function completes

	// Create a new temporary file that will store the final version of the CSV.
	finalTempFilePath := filePath + ".final.tmp"
	finalTempFile, err := os.Create(finalTempFilePath)
	if err != nil {
		fmt.Printf("Error creating final temporary file: %v\n", err)
		return false
	}
	defer SafeClose(finalTempFile)

	reader := csv.NewReader(tempFile)
	rewriter := csv.NewWriter(finalTempFile)

	// Process the header to identify and remove any columns that start with a number and ensure the "TOTAL DIGITAL IMP" column exists.
	header, err := reader.Read()
	if err != nil {
		fmt.Printf("Error reading header: %v\n", err)
		return false
	}

	var newHeader []string
	var indicesToDrop []int
	totalDigitalImpExists := false

	for i, column := range header {
		if strings.HasPrefix(column, "TOTAL DIGITAL IMP") {
			totalDigitalImpExists = true
		}

		// Check if the column starts with a number (indicating a date, likely in a custom format).
		if _, err := strconv.Atoi(string(column[0])); err == nil {
			indicesToDrop = append(indicesToDrop, i) // mark column index for dropping
		} else {
			newHeader = append(newHeader, column) // this column is fine, include it in the new header
		}
	}

	// If "TOTAL DIGITAL IMP" doesn't exist, we add it.
	if !totalDigitalImpExists {
		newHeader = append(newHeader, "TOTAL DIGITAL IMP")
	}

	// Write the new header to the final temporary CSV file.
	if err := rewriter.Write(newHeader); err != nil {
		fmt.Printf("Error writing new header to final temp file: %v\n", err)
		return false
	}

	// Now, we need to process the remaining records in the same manner, dropping the unnecessary columns.
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // end of file
		}
		if err != nil {
			fmt.Printf("Error reading record: %v\n", err)
			return false
		}

		var newRecord []string
		for i, value := range record {
			if !intInSlice(i, indicesToDrop) {
				newRecord = append(newRecord, value)
			}
		}

		// If "TOTAL DIGITAL IMP" column didn't exist, add an empty field in its place for each record.
		if !totalDigitalImpExists {
			newRecord = append(newRecord, "") // add empty value for the "TOTAL DIGITAL IMP" column
		}

		// Write the new record to the final temporary CSV file.
		if err := rewriter.Write(newRecord); err != nil {
			fmt.Printf("Error writing record to final temp file: %v\n", err)
			return false
		}
	}

	// Make sure to flush the writer to ensure all buffered operations have been applied to the file.
	rewriter.Flush()

	if err := rewriter.Error(); err != nil {
		fmt.Printf("Error during writer flush: %v\n", err)
		return false
	}

	SafeClose(tempFile)

	// Now, we don't need the original temporary file anymore. We can delete it.
	if err := os.Remove(tempFilePath); err != nil {
		fmt.Printf("Error removing original temporary file: %v\n", err)
		// Decide how you want to handle the error. If you want to stop processing, you can return false here.
		// Otherwise, you may log the issue and decide on a suitable course of action (like retrying deletion or just continuing).
		// This may depend on how critical it is for your application to ensure that these temporary files are deleted.
	}
	// Create 'partial' folder if it doesn't exist
	if _, err := os.Stat(partialDir); os.IsNotExist(err) {
		err := os.MkdirAll(partialDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", partialDir, err)
			return false

		}
	}

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

	var newPath string

	reportType := getType(dayCount, searchIndicator)

	if reportType == "weekly" {
		newPath = validateDir + "/" + newName
	} else {
		newPath = partialDir + "/" + newName
	}

	SafeClose(finalTempFile)

	if err := os.Rename(finalTempFilePath, newPath); err != nil {
		fmt.Printf("Error renaming file: %v\n", filename, newName, err)
		return false
	}

	metaData := Metadata{
		FileName:      newName,
		OriginalFile:  filename,
		StartDate:     dateRange.StartDate,
		EndDate:       dateRange.EndDate,
		WeekStart:     weekStart.Format("20060102"),
		DayCount:      dayCount,
		Type:          reportType,
		NObservations: lineCount - 6, //to account for header and initial rows removed
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

func writeMetaData(metaData Metadata, metaDataPath string) error {
	// New function to write metadata information
	// Convert struct to JSON
	file, err := os.Create(metaDataPath)
	if err != nil {
		return err
	}

	defer SafeClose(file)

	encoder := json.NewEncoder(file)
	err = encoder.Encode(metaData)
	if err != nil {
		return err
	}

	return nil
}

func logChange(logFile, oldName, newName, startDate, endDate string) {
	// logs changes in the log file
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

	defer SafeClose(file)

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
	// script to identify number of files to be processed and process each
	fmt.Println("VIVVIX AdSpender Converter: Convert VIVVIX reports")
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
