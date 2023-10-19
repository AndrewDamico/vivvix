// VIVVIX AdSpender Conversion App
// Copyright (c) 2023 Northwestern University
// Author: Andrew D'Amico
// Date: 10/18/2023

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func findMissingDates() {
	fmt.Println("VIVVIX AdSpender Converter: View Existing Coverage")
	fmt.Println("")
	// determines if any dates are missing in the range provided

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
		} else {
			fmt.Println("Operation cancelled by the user.")
			return // Exit the function since the user did not agree.
		}
	} else {
		fmt.Println("Current directory in settings:", settings.Directory)
	}

	// Prompt user for start and end date
	fmt.Print("Enter start date (MM-DD-YYYY): ")
	startDateStr, _ := reader.ReadString('\n')
	startDateStr = strings.TrimSpace(startDateStr)

	fmt.Print("Enter end date (MM-DD-YYYY): ")
	endDateStr, _ := reader.ReadString('\n')
	endDateStr = strings.TrimSpace(endDateStr)

	// Convert string inputs to time.Time types
	startDate, err := time.Parse("01-02-2006", startDateStr)
	if err != nil {
		fmt.Println("Invalid start date format. Please use MM-DD-YYYY.")
		return
	}

	endDate, err := time.Parse("01-02-2006", endDateStr)
	if err != nil {
		fmt.Println("Invalid end date format. Please use MM-DD-YYYY.")
		return
	}

	// Ensure the 'metadata' directory exists
	metaDataDir := settings.Directory + "/metadata" // or specify the full path if necessary
	if _, err := os.Stat(metaDataDir); os.IsNotExist(err) {
		fmt.Println("No metadata directory found.")
		return
	}

	// Read all metadata files and process only .json files
	files, err := os.ReadDir(metaDataDir)
	if err != nil {
		fmt.Println("Error reading metadata directory:", err)
		return
	}

	// Create a map to track the days for which we have data
	dateMap := make(map[string][]string)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := metaDataDir + "/" + file.Name()
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file.Name(), err)
			continue
		}

		var metaData Metadata
		err = json.Unmarshal(content, &metaData)
		if err != nil {
			fmt.Printf("Error parsing metadata JSON from file %s: %v\n", file.Name(), err)
			continue
		}

		startDateParsed, err := time.Parse("01022006", metaData.StartDate)
		if err != nil {
			fmt.Printf("Error parsing start date in file %s: %v\n", file.Name(), err)
			continue
		}

		endDateParsed, err := time.Parse("01022006", metaData.EndDate)
		if err != nil {
			fmt.Printf("Error parsing end date in file %s: %v\n", file.Name(), err)
			continue
		}

		currentDay := startDateParsed
		for currentDay.Before(endDateParsed.AddDate(0, 0, 1)) {
			// Add the filename to the slice for this date
			dateStr := currentDay.Format("01-02-2006")
			dateMap[dateStr] = append(dateMap[dateStr], file.Name())
			currentDay = currentDay.AddDate(0, 0, 1)
		}
	}

	// Check each day in the range to see if it's missing
	day := startDate
	missingDates := []string{}

	for day.Before(endDate.AddDate(0, 0, 1)) {
		if _, exists := dateMap[day.Format("01-02-2006")]; !exists {
			missingDates = append(missingDates, day.Format("January 2, 2006"))
		}
		day = day.AddDate(0, 0, 1)
	}

	if len(missingDates) == 0 {
		fmt.Println("There are no missing dates in the range.")
	} else {
		fmt.Println("Missing dates:")
		for _, date := range missingDates {
			fmt.Println(date)
		}
	}

	multiFileDates := make(map[string][]string)
	for date, filenames := range dateMap {
		if len(filenames) > 1 {
			multiFileDates[date] = filenames
		}
	}

	if len(multiFileDates) > 0 {
		fmt.Println("\nDates covered by multiple files:")
		for date, filenames := range multiFileDates {
			fmt.Printf("%s: %s\n", date, strings.Join(filenames, ", "))
		}
	} else {
		fmt.Println("\nThere are no dates covered by multiple files.")
	}
}
