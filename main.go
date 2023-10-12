package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

// Creates the struct for holding the start date and end date of the vivvix report
type DateRange struct {
	StartDate string
	EndDate   string
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

func processFile(dir, filename string) {
	filePath := dir + "/" + filename

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0
	var line5 string
	for scanner.Scan() {
		lineCount++
		if lineCount == 5 {
			line5 = scanner.Text()
			break
		}
	}

	dateRange := parser(line5)
	startDate := dateRange.StartDate
	newName := startDate + ".csv"
	validateDir := dir + "/validated"
	newPath := dir + "/validated/" + newName

	// Close the file explicitly before renaming
	file.Close()

	// Create 'validated' folder if it doesn't exist
	if _, err := os.Stat(validateDir); os.IsNotExist(err) {
		err := os.MkdirAll(validateDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", validateDir, err)
			return
		}
	}

	if err := os.Rename(filePath, newPath); err != nil {
		fmt.Printf("Error renaming file %s to %s: %v\n", filename, newName, err)
		return
	}

	// Log the change
	logChange(dir+"/rename_log.csv", filename, newName, dateRange.StartDate, dateRange.EndDate)
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
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// If the file didn't exist (i.e., we just created it), write the headers
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

func main() {
	reader := bufio.NewReader(os.Stdin)

	// asks user to input working directory
	fmt.Println("Enter the directory of files:")
	dir, _ := reader.ReadString('\n')
	dir = strings.TrimSpace(dir) // Remove newline

	// Count CSV files to be processed
	files, err := ioutil.ReadDir(dir)
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

	fmt.Printf("Found %d CSV files in %s. Do you want to proceed? (y/n): ", csvCount, dir)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice != "y" && choice != "Y" {
		fmt.Println("Exiting program.")
		return
	}

	for _, file := range files {
		if file.Name() == "rename_log.csv" {
			continue
		}
		if strings.HasSuffix(file.Name(), ".csv") {
			processFile(dir, file.Name())
		}
	}
}
