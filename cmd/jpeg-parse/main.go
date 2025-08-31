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

	jpegFile, err := jpeg.Open(filename)

	if err != nil {
		fmt.Println("Error opening JPEG file:", err)
		return
	}

	if jpegFile.HasSOI() {
		fmt.Println("✅ JPEG file has SOI marker")
	} else {
		fmt.Println("❌ JPEG file does not have SOI marker")
	}

	if jpegFile.HasEOI() {
		fmt.Println("✅ JPEG file has EOI marker")
	} else {
		fmt.Println("❌ JPEG file does not have EOI marker")
	}

	fmt.Println("App Data:")
	for _, appData := range jpegFile.AppData() {
		data := jpeg.ParseAppData(appData, jpegFile)
		fmt.Printf("\t%-5s %s (%d bytes)\n", appData.Name, data["identifier"], appData.Length)
	}

	width, err := jpegFile.Width()
	if err != nil {
		fmt.Println("❌ Failed to get width")
		return
	}

	height, err := jpegFile.Height()
	if err != nil {
		fmt.Println("❌ Failed to get height")
		return
	}

	if width > 0 && height > 0 {
		fmt.Printf("✅ Dimensions: %d ˣ %d\n", width, height)
	}

	fmt.Println("")

	fmt.Printf("Fields: %+v\n", jpegFile)

	imageData := jpeg.CompressedData(jpegFile)
	fmt.Printf("\nCompressed Data Length: %d bytes\n", len(imageData))
}

// get commandline args, ensure they are valid
func parseArgs(argv []string) (string, error) {
	if len(argv) < 2 {
		return "", errors.New("no arguments provided")
	}

	filename, err := getFileName(argv[1])

	if err != nil {
		return "", fmt.Errorf("error accessing file: %v", err)
	}

	return filename, nil
}
