package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"3thirty.net/jpeg/internal/jpeg"
)

var getFileName = func(filename string) (string, error) {
	_, err := os.Stat(filename)

	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func main() {
	filename, err := parseArgs(os.Args)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	jpeg, err := jpeg.Open(filename)

	if err != nil {
		fmt.Println("Error opening JPEG file:", err)
		return
	}

	if jpeg.HasSOI() {
		fmt.Println("JPEG file has SOI marker")
	} else {
		fmt.Println("JPEG file does not have SOI marker")
	}

	if jpeg.HasEOI() {
		fmt.Println("JPEG file has EOI marker")
	} else {
		fmt.Println("JPEG file does not have EOI marker")
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
