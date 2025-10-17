package utils

import (
	"encoding/hex"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// GenerateSimpleID Generate an 8-character ID without error handling
func GenerateSimpleID() string {
	// Set random seed (only needs to be set once)
	// Note: In actual projects, usually set once during program initialization
	seed := time.Now().UnixNano()
	localRand := rand.New(rand.NewSource(seed))

	// Generate 4 bytes of random data (becomes 8 characters when hex encoded)
	bytes := make([]byte, 4)
	localRand.Read(bytes) // Ignore error

	// Return 8-character hex string
	return hex.EncodeToString(bytes)
}

// ModelNameToDeploymentName Convert model name to Kubernetes-compatible deployment name
func ModelNameToDeploymentName(modelName string) string {
	// 1. Convert to lowercase
	name := strings.ToLower(modelName)

	// 2. Replace . with - (or other strategy, like removal)
	// Example: 0.6 → 0-6
	name = strings.ReplaceAll(name, ".", "-")

	// 3. Keep only: letters, numbers, -
	// Use regex to replace illegal characters
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	name = reg.ReplaceAllString(name, "")

	// 4. Ensure starts with letter
	if len(name) == 0 || !isLetter(name[0]) {
		name = "model-" + name
	}

	// 5. Ensure ends with letter or number
	for len(name) > 0 && !isAlnum(name[len(name)-1]) {
		name = name[:len(name)-1]
	}

	// 6. Limit length (63 characters)
	if len(name) > 63 {
		name = name[:63]
	}

	// 7. Prevent all hyphens or empty string
	if name == "" {
		name = "model-default"
	}
	for strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		name = strings.Trim(name, "-")
	}
	if name == "" {
		name = "model"
	}

	return name
}

// Utility functions
func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z')
}

func isAlnum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}
