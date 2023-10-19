// VIVVIX AdSpender Conversion App
// Copyright (c) 2023 Northwestern University
// Author: Andrew D'Amico
// Date: 10/18/2023

package main

import (
	"fmt"
)

// set global settings variable
var settings UserSettings

func main() {
	err := loadSettings() // Initialize your settings
	if err != nil {
		fmt.Println("Error loading settings:", err)
		return
	}
	MainMenu()
}
