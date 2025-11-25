package images

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/transports/alltransports"
	imagetypes "github.com/containers/image/v5/types"
)

// newSystemContextAmd64 creates a system context for amd64 architecture
func newSystemContextAmd64() *imagetypes.SystemContext {
	return &imagetypes.SystemContext{
		ArchitectureChoice: "amd64",
		OSChoice:           "linux",
	}
}

// newSystemContextArm64 creates a system context for arm64 architecture
func newSystemContextArm64() *imagetypes.SystemContext {
	return &imagetypes.SystemContext{
		ArchitectureChoice: "arm64",
		OSChoice:           "linux",
	}
}

// parseImageSourceV2 parses image name and creates an ImageSource using containers/image/v5
func parseImageSourceV2(ctx context.Context, imageName string) (imagetypes.ImageSource, error) {
	// Apply mirror if configured
	modifiedImageName := modifyImageNameWithMirror(imageName)

	// Add docker:// transport prefix if not present
	srcImageName := modifiedImageName
	if !strings.HasPrefix(modifiedImageName, "docker://") &&
		!strings.HasPrefix(modifiedImageName, "oci://") &&
		!strings.HasPrefix(modifiedImageName, "containers-storage://") {
		srcImageName = "docker://" + modifiedImageName
	}

	ref, err := alltransports.ParseImageName(srcImageName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image name %s: %w", srcImageName, err)
	}
	return ref.NewImageSource(ctx, &imagetypes.SystemContext{})
}

// downloadImagesInfoV2WithRetry downloads image information with retry mechanism
func downloadImagesInfoV2WithRetry(imageName, imageDir string) error {
	maxRetries := 3
	retryDelay := 5 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := downloadImagesInfoV2(imageName, imageDir)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadImagesInfoV2 downloads image information using containers/image/v5 library
// and saves it as JSON to the specified directory
func downloadImagesInfoV2(imageName, imageDir string) error {
	ctx := context.TODO()

	// Ensure directory exists
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Check if file already exists in cache
	outputPath := filepath.Join(imageDir, "image-info.json")
	if _, err := os.Stat(outputPath); err == nil {
		// File already exists, skip download
		return nil
	}

	// Parse image source
	src, err := parseImageSourceV2(ctx, imageName)
	if err != nil {
		return fmt.Errorf("failed to parse image source: %w", err)
	}
	defer src.Close()

	// Get unparsed image instance
	unparsedInstance := image.UnparsedInstance(src, nil)

	// Get manifest
	mb, mt, err := unparsedInstance.Manifest(ctx)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	// System contexts for different architectures
	sysCtx := []*imagetypes.SystemContext{newSystemContextAmd64(), newSystemContextArm64()}
	results := make([]json.RawMessage, 0)

	// Process each architecture
	for _, o := range sysCtx {
		if manifest.MIMETypeIsMultiImage(mt) {
			// Multi-architecture manifest list
			lst, err := manifest.ListFromBlob(mb, mt)
			if err != nil {
				log.Printf("Warning: failed to parse manifest list for %s: %v", imageName, err)
				continue
			}

			// Try to choose instance for this architecture
			_, err = lst.ChooseInstance(o)
			if err != nil {
				// This architecture is not available, skip
				continue
			}

			// Get image for this architecture
			img, err := image.FromUnparsedImage(ctx, o, unparsedInstance)
			if err != nil {
				log.Printf("Warning: failed to get image for %s/%s: %v", o.OSChoice, o.ArchitectureChoice, err)
				continue
			}

			// Inspect image
			imgInspect, err := img.Inspect(ctx)
			if err != nil {
				log.Printf("Warning: failed to inspect image for %s/%s: %v", o.OSChoice, o.ArchitectureChoice, err)
				continue
			}

			// Marshal to JSON
			data, err := json.Marshal(imgInspect)
			if err != nil {
				log.Printf("Warning: failed to marshal image inspect for %s/%s: %v", o.OSChoice, o.ArchitectureChoice, err)
				continue
			}

			results = append(results, json.RawMessage(data))
		} else {
			// Single architecture image
			img, err := image.FromUnparsedImage(ctx, o, unparsedInstance)
			if err != nil {
				log.Printf("Warning: failed to get image for %s/%s: %v", o.OSChoice, o.ArchitectureChoice, err)
				continue
			}

			// Inspect image
			imgInspect, err := img.Inspect(ctx)
			if err != nil {
				log.Printf("Warning: failed to inspect image for %s/%s: %v", o.OSChoice, o.ArchitectureChoice, err)
				continue
			}

			// Check if architecture matches
			if imgInspect.Architecture != o.ArchitectureChoice {
				continue
			}
			if imgInspect.Os != o.OSChoice {
				continue
			}

			// Marshal to JSON
			data, err := json.Marshal(imgInspect)
			if err != nil {
				log.Printf("Warning: failed to marshal image inspect for %s/%s: %v", o.OSChoice, o.ArchitectureChoice, err)
				continue
			}

			results = append(results, json.RawMessage(data))
		}
	}

	// If no results, return error
	if len(results) == 0 {
		return fmt.Errorf("no valid architecture found for image %s", imageName)
	}

	// Marshal results to JSON with indentation
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, b, 0644); err != nil {
		return fmt.Errorf("failed to write image info file: %w", err)
	}

	return nil
}

// copyImageInfoFromCache copies image info file from cache directory to chart directory
func copyImageInfoFromCache(cacheDir, chartDir string) error {
	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return fmt.Errorf("cache directory does not exist: %s", cacheDir)
	}

	// Create chart directory
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		return fmt.Errorf("failed to create chart directory: %w", err)
	}

	// Source file path
	srcPath := filepath.Join(cacheDir, "image-info.json")
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("image info file does not exist in cache: %s", srcPath)
	}

	// Destination file path
	dstPath := filepath.Join(chartDir, "image-info.json")

	// Copy file (copyFile is defined in images.go)
	return copyFile(srcPath, dstPath, 0644)
}
