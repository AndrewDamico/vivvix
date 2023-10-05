package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	folderPath := "/path/to/your/csv/files/"
	newNamePrefix := "new_name_prefix_"

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".csv") {
			oldPath := folderPath + file.Name()

			// Open and read the CSV file
			csvfile, err := os.Open(oldPath)
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer csvfile.Close()

			reader := csv.NewReader(csvfile)
			lines, err := reader.ReadAll()
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Check if the specific value exists in the CSV
			specificValue := "desired_value"
			if len(lines) > 0 && len(lines[0]) > 0 && lines[0][0] == specificValue {
				newPath := folderPath + newNamePrefix + file.Name()
				err := os.Rename(oldPath, newPath)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
