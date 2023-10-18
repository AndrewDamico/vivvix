// VIVVIX AdSpender Conversion App
// Copyright (c) 2023 Northwestern University
// Author: Andrew D'Amico
// Date: 10/18/2023

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// FilePath for storing the user settings
const FilePath = "user_settings.json"

// UserSettings holds user settings
type UserSettings struct {
	Directory string `json:"directory"`
	// Add other fields as needed
}

// loadSettings attempts to load user settings from a file.
// If the file does not exist, it returns an error that signals no settings.
func loadSettings() (UserSettings, error) {
	data, err := os.ReadFile(FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, meaning no settings have been set
			return UserSettings{}, errors.New("no settings have been set")
		}
		// Some other error occurred while trying to read the file
		return UserSettings{}, err
	}

	var settings UserSettings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		// JSON decode error
		return UserSettings{}, err
	}
	return settings, nil
}

// saveSettings saves the user settings to a file in JSON format.
func saveSettings(settings UserSettings) error {
	data, err := json.MarshalIndent(settings, "", "    ") // Pretty print JSON
	if err != nil {
		return err
	}
	return os.WriteFile(FilePath, data, 0644)
}

// getSettings retrieves the settings and prints them. It returns the settings for further use.
func getSettings() UserSettings {
	// Attempt to load settings
	settings, err := loadSettings()
	if err != nil {
		if err.Error() == "no settings have been set" {
			fmt.Println("No settings have been set. Please configure your settings.")
		} else {
			fmt.Println("Error loading settings:", err)
		}
		return UserSettings{} // returning empty settings
	}

	fmt.Println("Loaded settings:", settings)
	return settings
}

func setSettings() {
	reader := bufio.NewReader(os.Stdin)

	// Load current settings (if any)
	settings := getSettings()

	// Get the directory from the user input
	fmt.Print("Enter directory: ")
	dir, _ := reader.ReadString('\n')
	dir = strings.TrimSpace(dir) // Remove the newline character

	// Update the settings with the new directory
	settings.Directory = dir

	// Save the updated settings
	err := saveSettings(settings)
	if err != nil {
		fmt.Println("Error saving settings:", err)
		return
	}

	fmt.Println("Settings saved for future use.")
}

func printCurrentDirectory() {
	settings, err := loadSettings()
	if err != nil {
		fmt.Println("Error loading settings:", err)
		return
	}

	// Check if the Directory field is set in the settings
	if settings.Directory != "" {
		fmt.Println("Current directory in settings:", settings.Directory)
	} else {
		fmt.Println("No directory set in settings.")
	}
}
