// VIVVIX AdSpender Conversion App
// Copyright (c) 2023 Northwestern University
// Author: Andrew D'Amico
// Date: 10/18/2023

package main

import (
	"bufio"
	"fmt"
	"github.com/inancgumus/screen"
	"os"
)

func clearScreen() {
	// helper function to clear the terminal window
	screen.Clear()
	screen.MoveTopLeft()
}

func MainMenu() {
	clearScreen()
	var choice int64 = -1

	scanner := bufio.NewScanner(os.Stdin)

	menuReset := func() {
		fmt.Println()
		fmt.Println("Press Enter to continue...")
		scanner.Scan()
		clearScreen()
	}

	for {
		fmt.Println("VIVVIX AdSpender Converter: Main Menu")
		fmt.Println()
		fmt.Println("Please choose one of the following options:")
		fmt.Println("1. Convert Files")
		fmt.Println("2. Combine Files")
		fmt.Println("3. View Existing Coverage")
		fmt.Println("4. Configuration Menu")
		fmt.Println("0. Exit")

		var err error

		_, err = fmt.Scanf("%d\n", &choice)
		if err != nil {
			choice = -1
		}

		switch choice {
		case 0:
			// Exit the program
			fmt.Println("Goodbye...")
		// convert files
		case 1:
			clearScreen()
			converter()
			menuReset()
		case 2:
			clearScreen()
			combiner()
			menuReset()
		// get coverage
		case 3:
			clearScreen()
			findMissingDates()
			menuReset()
		// set options
		case 4:
			clearScreen()
			optionsMenu()
		default:
			fmt.Println("Invalid choice! Please try again.")
		}

		if choice == 0 {
			clearScreen()
			return
		}
	}
}

func optionsMenu() {
	scanner := bufio.NewScanner(os.Stdin)

	menuReset := func() {
		fmt.Println()
		fmt.Println("Press Enter to continue...")
		scanner.Scan()
		clearScreen()
	}

	for {
		// Reload settings each time we display the options, to ensure we show the most recent values
		settings := getSettings()

		clearScreen()
		fmt.Println("VIVVIX AdSpender Converter: Configuration Menu")
		fmt.Println()

		autoDeleteStatus := "Disabled"
		if settings.AutoDelete {
			autoDeleteStatus = "Enabled"
		}

		directoryStatus := "None"
		if settings.Directory != "" {
			directoryStatus = settings.Directory
		}

		var choice int64 = -1

		fmt.Printf("1. Current directory for processing: [%s]\n", directoryStatus)
		fmt.Printf("2. Auto-delete of files after processing: [%s]\n", autoDeleteStatus)
		fmt.Println()
		fmt.Println("Press Enter to Return to Previous Menu")

		var err error

		_, err = fmt.Scanf("%d\n", &choice)
		if err != nil {
			choice = -1
		}

		switch choice {
		case 0:
			clearScreen()
			// Return to Previous Menu
			fmt.Println()
		case 1:
			clearScreen()
			fmt.Println("VIVVIX AdSpender Converter: Configuration Menu")
			fmt.Println("Config: Set Directory")
			fmt.Println()
			fmt.Println("Please set the new working directory")
			setSettings("Directory")
			menuReset()
		case 2:
			clearScreen()
			fmt.Println("VIVVIX AdSpender Converter: Configuration Menu")
			fmt.Println("Config: Auto Delete")
			fmt.Println()
			fmt.Println("Please set auto delete of files after processing")
			setSettings("AutoDelete")
			menuReset()

		default:
			clearScreen()
			return
		}

		if choice == 0 {
			clearScreen()
			return
		}
	}
}
