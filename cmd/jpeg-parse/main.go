package main

import (
	"errors"
	"fmt"
	"os"
)

var getFileName = func(filename string) (string, error) {
	res, err := os.Stat(filename)

	if err != nil {
		return "", err
	}

	return res.Name(), nil
}

func main() {
	_, error := parseArgs(os.Args)

	if error != nil {
		fmt.Println("Error:", error)
		return
	}
}

// get commandline args, ensure they are valid
func parseArgs(argv []string) (string, error) {
	if len(argv) < 2 {
		return "", errors.New("No arguments provided")
	}

	filename, err := getFileName(argv[1])

	if err != nil {
		return "", fmt.Errorf("Error accessing file: %v", err)
	}

	return filename, nil
}
