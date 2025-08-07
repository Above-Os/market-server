package main

import (
	"log"
	"os"
	"path/filepath"

	"app-store-server/internal/images"
)

func main() {
	// Get the demo/affine directory path
	chartDir := "demo/affine"

	// Check if directory exists
	if _, err := os.Stat(chartDir); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", chartDir)
	}

	log.Printf("Testing DownloadImagesInfo with directory: %s", chartDir)

	// Call DownloadImagesInfo function
	err := images.DownloadImagesInfo(chartDir)
	if err != nil {
		log.Fatalf("DownloadImagesInfo failed: %v", err)
	}

	log.Printf("Successfully processed images in %s", chartDir)

	// List the created images directory
	imagesDir := filepath.Join(chartDir, "images")
	if _, err := os.Stat(imagesDir); err == nil {
		log.Printf("Created images directory: %s", imagesDir)

		// List contents
		entries, err := os.ReadDir(imagesDir)
		if err != nil {
			log.Printf("Warning: failed to read images directory: %v", err)
		} else {
			log.Printf("Images directory contains:")
			for _, entry := range entries {
				if entry.IsDir() {
					log.Printf("  - %s/", entry.Name())
				} else {
					log.Printf("  - %s", entry.Name())
				}
			}
		}
	}
}
