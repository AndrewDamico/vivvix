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
	"strconv"
	"strings"
)

// FilePath for storing the user settings
const FilePath = "user_settings.json"

// UserSettings holds user settings
type UserSettings struct {
	Directory  string `json:"Directory"`
	AutoDelete bool   `json:"AutoDelete"`
	// Add other fields as needed
}

// loadSettings attempts to load user settings from a file.
// If the file does not exist, it returns an error that signals no settings.
func loadSettings() error {
	data, err := os.ReadFile(FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist - we could initiate settings with default values here if needed
			settings = UserSettings{
				// Set other default values as needed
				AutoDelete: false, // Default value set for AutoDelete
			}
			return nil // No error, as it's okay if the file doesn't exist yet
		}
		// Some other error occurred while trying to read the file
		return err
	}

	// Unmarshal the JSON into the global settings variable
	err = json.Unmarshal(data, &settings)
	if err != nil {
		// JSON decode error
		return err
	}
	return nil // No error occurred
}

// saveSettings writes the current state of the 'settings' global variable to a file.
func saveSettings() error {
	data, err := json.MarshalIndent(settings, "", "    ") // Pretty print JSON
	if err != nil {
		return err
	}
	return os.WriteFile(FilePath, data, 0644)
}

// getSettings retrieves the settings and prints them. It returns the settings for further use.
func getSettings() UserSettings {
	// If settings were not previously loaded, load them now
	if settings == (UserSettings{}) { // assuming zero value means not loaded
		err := loadSettings()
		if err != nil {
			fmt.Println("Error loading settings:", err)
		}
	}
	return settings
}

func setSettings(settingType string) {
	reader := bufio.NewReader(os.Stdin)

	switch settingType {
	case "Directory":
		// Get the directory from the user input
		fmt.Print("Enter directory: ")
		dir, _ := reader.ReadString('\n')
		dir = strings.TrimSpace(dir) // Remove the newline character

		// Update the settings with the new directory
		settings.Directory = dir

	case "AutoDelete":
		// Get the remove flag from the user input
		fmt.Print("Enable Auto Delete of files after processing? (true/false): ")
		removeStr, _ := reader.ReadString('\n')
		removeStr = strings.TrimSpace(removeStr)

		// Convert string input to boolean and update settings
		remove, err := strconv.ParseBool(removeStr)
		if err != nil {
			fmt.Println("Invalid input. Please enter 'true' or 'false'.")
			return // exit if invalid input
		}
		settings.AutoDelete = remove

	default:
		fmt.Println("Unknown setting type.")
		return // exit if unknown setting type
	}

	err := saveSettings() // No argument needed because it uses the global variable
	if err != nil {
		fmt.Println("Error saving settings:", err)
		return
	}

	fmt.Println("Settings saved for future use.")
}
