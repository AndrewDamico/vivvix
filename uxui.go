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
	screen.Clear()
	screen.MoveTopLeft()
}

func MainMenu() {
	clearScreen()
	getSettings()
	var choice int64 = -1

	scanner := bufio.NewScanner(os.Stdin)

	menuReset := func() {
		fmt.Println()
		fmt.Println("Press Enter to continue...")
		scanner.Scan()
		clearScreen()
	}

	for {
		fmt.Println("VIVVIX AdSpender Converter")
		fmt.Println("")
		fmt.Println("Main Menu")
		fmt.Println("Please choose one of the following options:")
		fmt.Println("1. Convert Files")
		fmt.Println("2. View Existing Coverage")
		fmt.Println("3. Configuration Menu")
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
		// get coverage
		case 2:
			clearScreen()
			findMissingDates()
			menuReset()
		// set options
		case 3:
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
	clearScreen()
	fmt.Println("Configuration Menu")
	fmt.Println()

	// determine which set to test on

	var choice int64 = -1

	scanner := bufio.NewScanner(os.Stdin)

	menuReset := func() {
		fmt.Println()
		fmt.Println("Press Enter to continue...")
		scanner.Scan()
		clearScreen()
	}

	for {
		fmt.Println("1. View Import Directory")
		fmt.Println("2. Set Import Directory")
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
			fmt.Println("Import Directory")
			printCurrentDirectory()
			menuReset()
		case 2:
			clearScreen()
			fmt.Println("Please set the new working directory")
			setSettings()
			menuReset()
		default:
			clearScreen()
			return
		}

		if choice == 0 {
			clearScreen()
			return
		}
		//fmt.Println("Press Enter to continue...")
		//scanner.Scan() // Wait for user to press Enter

	}
}
