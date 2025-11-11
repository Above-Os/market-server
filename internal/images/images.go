package images

import (
	"app-store-server/internal/constants"
	"encoding/json"
	"fmt"
	"io"
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

// getCacheDir returns the persistent cache directory for image manifests
func getCacheDir() string {
	cacheDir := os.Getenv("IMAGE_MANIFESTS_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = constants.ImageManifestsCacheDir
	}
	return cacheDir
}

func DownloadImagesInfo(chartDir string) error {
	// 1. Extract all images from chart directory
	images, err := extractImagesFromDirectory(chartDir)
	if err != nil {
		return fmt.Errorf("failed to extract images: %w", err)
	}

	if len(images) == 0 {
		return nil
	}

	// 2. Initialize persistent cache directory
	cacheDir := getCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 3. Create images directory in chartDir (for packaging)
	imagesDir := filepath.Join(chartDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	// 4. Process each image: download to cache first, then copy to chartDir
	for _, image := range images {
		// Create safe directory name for image
		safeImageName := createSafeDirectoryName(image)

		// Cache directory for this image
		cacheImageDir := filepath.Join(cacheDir, safeImageName)

		// Chart directory for this image (for packaging)
		chartImageDir := filepath.Join(imagesDir, safeImageName)

		// Download and process manifest to cache with retry
		if err := downloadAndProcessManifestWithRetry(image, cacheImageDir); err != nil {
			return fmt.Errorf("failed to process image %s: %w", image, err)
		}

		// Copy from cache to chartDir
		if err := copyManifestFromCache(cacheImageDir, chartImageDir); err != nil {
			log.Printf("Warning: failed to copy manifest from cache for image %s: %v", image, err)
			// Continue even if copy fails, as cache is the primary storage
		}
	}

	return nil
}

// copyManifestFromCache copies manifest files from cache directory to chart directory
func copyManifestFromCache(cacheDir, chartDir string) error {
	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return fmt.Errorf("cache directory does not exist: %s", cacheDir)
	}

	// Create chart directory
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		return fmt.Errorf("failed to create chart directory: %w", err)
	}

	// Recursively copy all files from cache to chart directory
	return filepath.Walk(cacheDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from cache directory
		relPath, err := filepath.Rel(cacheDir, srcPath)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Destination path
		dstPath := filepath.Join(chartDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			return copyFile(srcPath, dstPath, info.Mode())
		}
	})
}

// copyFile copies a file from src to dst
func copyFile(src, dst string, mode os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
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

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := downloadAndProcessManifest(imageName, imageDir)
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

// downloadAndProcessManifest downloads manifest and processes multi-arch images
func downloadAndProcessManifest(imageName, imageDir string) error {
	// Check if main manifest already exists locally
	manifestPath := filepath.Join(imageDir, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		// Read existing manifest to check if it's a multi-architecture manifest list
		existingManifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			// Continue with download if we can't read the existing file
		} else {
			var manifestList ManifestList
			if err := json.Unmarshal(existingManifestData, &manifestList); err == nil {
				// Check if this is a manifest list (multi-architecture)
				if manifestList.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
					manifestList.MediaType == "application/vnd.oci.image.index.v1+json" {
					// Check and download missing architecture manifests
					for _, manifest := range manifestList.Manifests {
						arch := manifest.Platform.Architecture
						osName := manifest.Platform.OS
						variant := manifest.Platform.Variant

						// Create architecture-specific directory
						archDir := filepath.Join(imageDir, fmt.Sprintf("%s-%s", osName, arch))
						if variant != "" {
							archDir = filepath.Join(imageDir, fmt.Sprintf("%s-%s-%s", osName, arch, variant))
						}

						archManifestPath := filepath.Join(archDir, "manifest.json")
						if _, err := os.Stat(archManifestPath); err != nil {
							// Download missing architecture manifest
							if err := downloadArchitectureManifest(imageName, manifest.Digest, archDir, osName, arch, variant); err != nil {
								log.Printf("Warning: failed to download architecture manifest for %s/%s: %v", osName, arch, err)
							}
						}
					}
					return nil
				} else {
					// Single architecture image, manifest already exists
					return nil
				}
			}
		}
	}

	// Download manifest using docker manifest inspect
	manifestData, err := downloadManifestWithRetry(imageName)
	if err != nil {
		return fmt.Errorf("failed to download manifest: %w", err)
	}

	// Ensure directory exists before writing file
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Save the main manifest
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Parse manifest to check if it's a multi-architecture manifest list
	var manifestList ManifestList
	if err := json.Unmarshal(manifestData, &manifestList); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Check if this is a manifest list (multi-architecture)
	if manifestList.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		manifestList.MediaType == "application/vnd.oci.image.index.v1+json" {
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

			// Check if architecture manifest already exists
			archManifestPath := filepath.Join(archDir, "manifest.json")
			if _, err := os.Stat(archManifestPath); err == nil {
				continue
			}

			// Download the specific architecture manifest with retry
			if err := downloadArchitectureManifest(imageName, manifest.Digest, archDir, osName, arch, variant); err != nil {
				log.Printf("Warning: failed to download manifest for %s/%s: %v", osName, arch, err)
				// Create an error manifest file to indicate this architecture was attempted
				emptyManifest := fmt.Sprintf(`{"error": "Failed to download manifest for %s/%s: %v"}`, osName, arch, err)
				_ = os.WriteFile(archManifestPath, []byte(emptyManifest), 0644)
				continue
			}
		}
	}

	return nil
}

// downloadManifestWithRetry downloads manifest with retry mechanism
func downloadManifestWithRetry(imageName string) ([]byte, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		output, err := downloadManifest(imageName)
		if err == nil {
			return output, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadManifestByDigestWithRetry downloads a specific manifest by digest with retry mechanism
func downloadManifestByDigestWithRetry(imageName, digest string) ([]byte, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		output, err := downloadManifestByDigest(imageName, digest)
		if err == nil {
			return output, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
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
			return imageName
		}
	}

	// Clean up mirror URL - remove protocol and trailing slash
	cleanMirror := strings.TrimSuffix(mirror, "/")
	cleanMirror = strings.TrimPrefix(cleanMirror, "https://")
	cleanMirror = strings.TrimPrefix(cleanMirror, "http://")

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
	imageSet := make(map[string]bool)

	// More specific regex to match Docker image fields in YAML
	// This matches patterns like:
	// - image: nginx:latest
	// - image: "nginx:latest"
	// - image: 'nginx:latest'
	// - image: gcr.io/google-containers/pause:latest
	imageRegex := regexp.MustCompile(`(?m)^\s*image:\s*["\']?([a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9](?::[a-zA-Z0-9._-]+)?(?:@sha256:[a-fA-F0-9]{64})?)["\']?\s*$`)

	matches := imageRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			image := strings.TrimSpace(match[1])

			if image != "" && isValidImageName(image) {
				// Clean up the image name
				cleanImage := cleanImageName(image)
				if cleanImage != "" {
					imageSet[cleanImage] = true
				}
			}
		}
	}

	images := make([]string, 0, len(imageSet))
	for image := range imageSet {
		images = append(images, image)
	}

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

	// Handle registry prefixes - docker.io is the default registry, can be simplified
	cleaned = strings.TrimPrefix(cleaned, "docker.io/")

	// Remove any trailing slashes
	cleaned = strings.TrimSuffix(cleaned, "/")

	return cleaned
}

// downloadArchitectureManifest downloads and saves a specific architecture manifest
func downloadArchitectureManifest(imageName, digest, archDir, osName, arch, variant string) error {
	// Download the specific architecture manifest with retry
	archManifestData, err := downloadManifestByDigestWithRetry(imageName, digest)
	if err != nil {
		return fmt.Errorf("failed to download architecture manifest: %w", err)
	}

	// Ensure directory exists before writing file
	if err := os.MkdirAll(archDir, 0755); err != nil {
		return fmt.Errorf("failed to create arch directory: %w", err)
	}

	// Save architecture-specific manifest
	archManifestPath := filepath.Join(archDir, "manifest.json")
	if err := os.WriteFile(archManifestPath, archManifestData, 0644); err != nil {
		return fmt.Errorf("failed to save arch manifest: %w", err)
	}

	// osName, arch, and variant are kept for potential future use (e.g., logging or validation)
	_ = osName
	_ = arch
	_ = variant

	return nil
}
