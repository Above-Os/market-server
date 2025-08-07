package images

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ImageManifest represents the structure of a Docker image manifest
type ImageManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

// ManifestList represents a multi-architecture manifest list
type ManifestList struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Manifests     []struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
		Platform  struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Variant      string `json:"variant,omitempty"`
		} `json:"platform"`
	} `json:"manifests"`
}

func DownloadImagesInfo(chartDir string) error {
	// 1. Extract all images from chart directory
	images, err := extractImagesFromDirectory(chartDir)
	if err != nil {
		return fmt.Errorf("failed to extract images: %w", err)
	}

	log.Printf("Found %d unique images in chart directory", len(images))

	// 2. Create images directory
	imagesDir := filepath.Join(chartDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	// 3. Process each image with retry mechanism
	for _, image := range images {
		log.Printf("Processing image: %s", image)

		// Create safe directory name for image
		safeImageName := createSafeDirectoryName(image)
		imageDir := filepath.Join(imagesDir, safeImageName)

		if err := os.MkdirAll(imageDir, 0755); err != nil {
			log.Printf("Warning: failed to create directory for image %s: %v", image, err)
			continue
		}

		// Download and process manifest with retry
		if err := downloadAndProcessManifestWithRetry(image, imageDir); err != nil {
			log.Printf("Error: failed to process manifest for image %s after all retries: %v", image, err)
			return fmt.Errorf("failed to process image %s: %w", image, err)
		}
	}

	return nil
}

// createSafeDirectoryName creates a safe directory name from image name
func createSafeDirectoryName(imageName string) string {
	// Replace invalid characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	safeName := re.ReplaceAllString(imageName, "_")

	// Remove leading/trailing underscores
	safeName = strings.Trim(safeName, "_")

	// Ensure it's not empty
	if safeName == "" {
		safeName = "unknown_image"
	}

	return safeName
}

// downloadAndProcessManifestWithRetry downloads manifest with retry mechanism
func downloadAndProcessManifestWithRetry(imageName, imageDir string) error {
	maxRetries := 3
	retryDelay := 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempt %d/%d for image %s", attempt, maxRetries, imageName)

		err := downloadAndProcessManifest(imageName, imageDir)
		if err == nil {
			log.Printf("Successfully processed image %s on attempt %d", imageName, attempt)
			return nil
		}

		if attempt < maxRetries {
			log.Printf("Attempt %d failed for image %s: %v. Retrying in %v...", attempt, imageName, err, retryDelay)
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		} else {
			return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}
	}

	return fmt.Errorf("unexpected error in retry loop")
}

// downloadAndProcessManifest downloads manifest and processes multi-arch images
func downloadAndProcessManifest(imageName, imageDir string) error {
	// Download manifest using docker manifest inspect
	manifestData, err := downloadManifestWithRetry(imageName)
	if err != nil {
		return fmt.Errorf("failed to download manifest: %w", err)
	}

	// Save the main manifest
	manifestPath := filepath.Join(imageDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Parse manifest to check if it's a multi-architecture manifest list
	var manifestList ManifestList
	if err := json.Unmarshal(manifestData, &manifestList); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Check if this is a manifest list (multi-architecture)
	if manifestList.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" {
		log.Printf("Image %s is a multi-architecture manifest list with %d architectures",
			imageName, len(manifestList.Manifests))

		// Download each architecture's manifest with retry
		for _, manifest := range manifestList.Manifests {
			arch := manifest.Platform.Architecture
			osName := manifest.Platform.OS
			variant := manifest.Platform.Variant

			// Create architecture-specific directory
			archDir := filepath.Join(imageDir, fmt.Sprintf("%s-%s", osName, arch))
			if variant != "" {
				archDir = filepath.Join(imageDir, fmt.Sprintf("%s-%s-%s", osName, arch, variant))
			}

			if err := os.MkdirAll(archDir, 0755); err != nil {
				log.Printf("Warning: failed to create arch directory %s: %v", archDir, err)
				continue
			}

			// Download the specific architecture manifest with retry
			archManifestData, err := downloadManifestByDigestWithRetry(imageName, manifest.Digest)
			if err != nil {
				log.Printf("Warning: failed to download manifest for %s/%s: %v", osName, arch, err)
				// Create an error manifest file to indicate this architecture was attempted
				archManifestPath := filepath.Join(archDir, "manifest.json")
				emptyManifest := fmt.Sprintf(`{"error": "Failed to download manifest for %s/%s: %v"}`, osName, arch, err)
				if writeErr := os.WriteFile(archManifestPath, []byte(emptyManifest), 0644); writeErr != nil {
					log.Printf("Warning: failed to write error manifest: %v", writeErr)
				}
				continue
			}

			// Save architecture-specific manifest
			archManifestPath := filepath.Join(archDir, "manifest.json")
			if err := os.WriteFile(archManifestPath, archManifestData, 0644); err != nil {
				log.Printf("Warning: failed to save arch manifest: %v", err)
				continue
			}

			log.Printf("Downloaded manifest for %s/%s (variant: %s)", osName, arch, variant)
		}
	} else {
		log.Printf("Image %s is a single architecture image", imageName)
	}

	return nil
}

// downloadManifestWithRetry downloads manifest with retry mechanism
func downloadManifestWithRetry(imageName string) ([]byte, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		output, err := downloadManifest(imageName)
		if err == nil {
			return output, nil
		}

		if attempt < maxRetries {
			log.Printf("Attempt %d failed for manifest %s: %v. Retrying in %v...", attempt, imageName, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		} else {
			return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

// downloadManifestByDigestWithRetry downloads a specific manifest by digest with retry mechanism
func downloadManifestByDigestWithRetry(imageName, digest string) ([]byte, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		output, err := downloadManifestByDigest(imageName, digest)
		if err == nil {
			return output, nil
		}

		if attempt < maxRetries {
			log.Printf("Attempt %d failed for digest %s: %v. Retrying in %v...", attempt, digest, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		} else {
			return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

// downloadManifest downloads manifest using docker manifest inspect
// getDockerRegistryMirror returns the docker registry mirror from environment variable
func getDockerRegistryMirror() string {
	return os.Getenv("IMAGES_SOURCE")
}

// modifyImageNameWithMirror modifies image name to use mirror if configured
func modifyImageNameWithMirror(imageName string) string {
	mirror := getDockerRegistryMirror()
	if mirror == "" {
		return imageName
	}

	// Extract registry and repository
	registry, repository := extractRegistryAndRepository(imageName)

	// Skip mirror replacement for specific registries
	skipMirrorRegistries := []string{
		"ghcr.io", "gcr.io", "quay.io", "registry.k8s.io",
		"mcr.microsoft.com", "registry.aliyuncs.com", "registry.cn-hangzhou.aliyuncs.com",
	}

	for _, skipRegistry := range skipMirrorRegistries {
		if registry == skipRegistry {
			log.Printf("Skipping mirror for registry: %s", registry)
			return imageName
		}
	}

	// Clean up mirror URL - remove protocol and trailing slash
	cleanMirror := strings.TrimSuffix(mirror, "/")
	if strings.HasPrefix(cleanMirror, "https://") {
		cleanMirror = strings.TrimPrefix(cleanMirror, "https://")
	} else if strings.HasPrefix(cleanMirror, "http://") {
		cleanMirror = strings.TrimPrefix(cleanMirror, "http://")
	}

	// If it's already using the mirror, return as is
	if registry == cleanMirror {
		return imageName
	}

	// Replace registry with mirror
	return fmt.Sprintf("%s/%s", cleanMirror, repository)
}

func downloadManifest(imageName string) ([]byte, error) {
	// Modify image name to use mirror if configured
	modifiedImageName := modifyImageNameWithMirror(imageName)

	log.Printf("Downloading manifest for %s (original: %s)", modifiedImageName, imageName)

	cmd := exec.Command("docker", "manifest", "inspect", modifiedImageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker manifest inspect failed for %s: %w", modifiedImageName, err)
	}
	return output, nil
}

// downloadManifestByDigest downloads a specific manifest by digest
func downloadManifestByDigest(imageName, digest string) ([]byte, error) {
	// Extract the base image name without tag
	baseImageName := extractBaseImageName(imageName)

	// Modify base image name to use mirror if configured
	modifiedBaseImageName := modifyImageNameWithMirror(baseImageName)

	// Use docker manifest inspect with the base image name and digest
	// Format: docker manifest inspect <base-image>@<digest>
	fullImageName := fmt.Sprintf("%s@%s", modifiedBaseImageName, digest)

	log.Printf("Downloading manifest for %s (original: %s@%s)", fullImageName, baseImageName, digest)

	cmd := exec.Command("docker", "manifest", "inspect", fullImageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to download manifest by digest for %s: %w", fullImageName, err)
	}

	return output, nil
}

// extractBaseImageName extracts the base image name without tag
func extractBaseImageName(imageName string) string {
	// Remove tag if present (everything after the last colon, but not if it's a digest)
	if strings.Contains(imageName, ":") && !strings.Contains(imageName, "@") {
		// Check if the colon is followed by a digest (sha256:...)
		parts := strings.Split(imageName, ":")
		lastPart := parts[len(parts)-1]
		if !strings.HasPrefix(lastPart, "sha256:") {
			// It's a tag, remove it
			return strings.Join(parts[:len(parts)-1], ":")
		}
	}
	return imageName
}

// extractRegistryAndRepository extracts registry and repository from image name
func extractRegistryAndRepository(imageName string) (string, string) {
	parts := strings.Split(imageName, "/")

	if len(parts) == 1 {
		// No registry specified, use docker.io
		return "docker.io", "library/" + parts[0]
	} else if len(parts) == 2 {
		// Check if first part is a registry
		if strings.Contains(parts[0], ".") || parts[0] == "localhost" {
			return parts[0], parts[1]
		} else {
			// No registry, use docker.io
			return "docker.io", strings.Join(parts, "/")
		}
	} else {
		// Multiple parts, first is registry
		return parts[0], strings.Join(parts[1:], "/")
	}
}

// extractImagesFromDirectory extracts all Docker image references from rendered chart files
func extractImagesFromDirectory(chartDir string) ([]string, error) {
	imageSet := make(map[string]bool)

	// Walk through all files in the chart directory
	err := filepath.Walk(chartDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process YAML files
		if !isYAMLFile(path) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read file %s: %v", path, err)
			return nil
		}

		// Extract images from file content
		images := extractImagesFromContent(string(content))
		for _, image := range images {
			imageSet[image] = true
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk chart directory: %w", err)
	}

	// Convert set to slice
	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}

	return images, nil
}

func isYAMLFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".yaml" || ext == ".yml"
}

// extractImagesFromContent extracts Docker image references from file content
func extractImagesFromContent(content string) []string {
	// Add debug logging to help diagnose image extraction issues
	log.Printf("Debug: Extracting images from content (length: %d)", len(content))

	imageSet := make(map[string]bool)

	// More specific regex to match Docker image fields in YAML
	// This matches patterns like:
	// - image: nginx:latest
	// - image: "nginx:latest"
	// - image: 'nginx:latest'
	// - image: gcr.io/google-containers/pause:latest
	imageRegex := regexp.MustCompile(`(?m)^\s*image:\s*["\']?([a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9](?::[a-zA-Z0-9._-]+)?(?:@sha256:[a-fA-F0-9]{64})?)["\']?\s*$`)

	matches := imageRegex.FindAllStringSubmatch(content, -1)
	log.Printf("Debug: Found %d regex matches", len(matches))

	for i, match := range matches {
		log.Printf("Debug: Match %d: %v", i, match)
		if len(match) > 1 {
			image := strings.TrimSpace(match[1])
			log.Printf("Debug: Extracted image candidate: '%s'", image)

			if image != "" && isValidImageName(image) {
				// Clean up the image name
				cleanImage := cleanImageName(image)
				log.Printf("Debug: Cleaned image: '%s'", cleanImage)
				if cleanImage != "" {
					imageSet[cleanImage] = true
					log.Printf("Debug: Added image to set: '%s'", cleanImage)
				}
			} else {
				log.Printf("Debug: Image '%s' failed validation", image)
			}
		}
	}

	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}

	log.Printf("Debug: Final extracted images: %v", images)
	return images
}

// isValidImageName validates if a string is a valid Docker image name
func isValidImageName(imageName string) bool {
	// Basic validation for Docker image names
	if imageName == "" {
		return false
	}

	// Skip obvious template variables and placeholders
	if strings.Contains(imageName, "{{") || strings.Contains(imageName, "}}") {
		return false
	}
	if strings.Contains(imageName, "${") || strings.Contains(imageName, "}") {
		return false
	}

	// Skip common non-image strings
	invalidNames := []string{
		"-", "https", "http", "version", "latest", "stable", "tag", "name",
		"image", "repository", "registry", "docker", "container", "pod",
	}

	lowerImage := strings.ToLower(imageName)
	for _, invalid := range invalidNames {
		if lowerImage == invalid {
			return false
		}
	}

	// Must contain at least one character that's not a template marker
	if strings.HasPrefix(imageName, "$") {
		return false
	}

	// Basic Docker image name format validation
	// Should contain valid characters only
	validImageRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9](?::[a-zA-Z0-9._-]+)?(?:@sha256:[a-fA-F0-9]{64})?$`)
	if !validImageRegex.MatchString(imageName) {
		return false
	}

	// Must have at least one valid component
	parts := strings.Split(strings.Split(imageName, ":")[0], "/")
	if len(parts) == 0 {
		return false
	}

	// Check that components are not too short or invalid
	for _, part := range parts {
		if len(part) < 1 || part == "." || part == ".." {
			return false
		}
	}

	return true
}

// cleanImageName cleans and normalizes an image name
func cleanImageName(imageName string) string {
	// Remove quotes and extra whitespace
	cleaned := strings.Trim(imageName, `"' `)

	// Handle registry prefixes
	if strings.HasPrefix(cleaned, "docker.io/") {
		// docker.io is the default registry, can be simplified
		cleaned = strings.TrimPrefix(cleaned, "docker.io/")
	}

	// Remove any trailing slashes
	cleaned = strings.TrimSuffix(cleaned, "/")

	return cleaned
}
