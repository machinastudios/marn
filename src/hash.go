package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
)

// HashStore stores the hash information for a project
type HashStore struct {
    ProjectPath string            `json:"projectPath"`
    SrcHash     string            `json:"srcHash"`
    Files       map[string]string `json:"files"` // file path -> hash
}

// getHashFilePath returns the path to the hash file for a project
func getHashFilePath(projectPath string) string {
    return filepath.Join(projectPath, ".marn", "src-hash.json")
}

// calculateSrcHash calculates the hash of all source files in src/ directory
func calculateSrcHash(projectPath string) (string, map[string]string, error) {
    srcDirs := []string{
        filepath.Join(projectPath, "src", "main", "java"),
        filepath.Join(projectPath, "src", "main", "resources"),
    }

    hasher := sha256.New()
    fileHashes := make(map[string]string)
    var allFilePaths []string

    // Collect all files and calculate individual hashes
    for _, srcDir := range srcDirs {
        err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return nil // Skip files that can't be accessed
            }

            if info.IsDir() {
                return nil
            }

            // Calculate hash for this file
            fileHash, err := calculateFileHash(path)
            if err != nil {
                return nil // Skip files that can't be read
            }

            // Get relative path from project root
            relPath, err := filepath.Rel(projectPath, path)
            if err != nil {
                relPath = path
            }

            fileHashes[relPath] = fileHash
            allFilePaths = append(allFilePaths, relPath)

            return nil
        })

        if err != nil {
            return "", nil, err
        }
    }

    // Sort file paths for consistent hashing
    sort.Strings(allFilePaths)

    // Calculate combined hash
    for _, relPath := range allFilePaths {
        hasher.Write([]byte(relPath))
        hasher.Write([]byte(fileHashes[relPath]))
    }

    combinedHash := hex.EncodeToString(hasher.Sum(nil))
    return combinedHash, fileHashes, nil
}

// calculateFileHash calculates SHA256 hash of a file
func calculateFileHash(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", err
    }

    return hex.EncodeToString(hasher.Sum(nil)), nil
}

// loadHashStore loads the hash store from disk
func loadHashStore(projectPath string) (*HashStore, error) {
    hashFilePath := getHashFilePath(projectPath)

    // If file doesn't exist, return empty store
    if _, err := os.Stat(hashFilePath); os.IsNotExist(err) {
        return &HashStore{
            ProjectPath: projectPath,
            SrcHash:     "",
            Files:       make(map[string]string),
        }, nil
    }

    data, err := os.ReadFile(hashFilePath)
    if err != nil {
        return nil, err
    }

    var store HashStore
    if err := json.Unmarshal(data, &store); err != nil {
        return nil, err
    }

    return &store, nil
}

// saveHashStore saves the hash store to disk
func saveHashStore(store *HashStore) error {
    hashFilePath := getHashFilePath(store.ProjectPath)

    // Create .marn directory if it doesn't exist
    hashDir := filepath.Dir(hashFilePath)
    if err := os.MkdirAll(hashDir, 0755); err != nil {
        return err
    }

    data, err := json.MarshalIndent(store, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(hashFilePath, data, 0644)
}

// hasSrcChanged checks if the source files have changed since last build
// Returns true if changed or if no previous hash exists
func hasSrcChanged(projectPath string) (bool, error) {
    // Calculate current hash
    currentHash, _, err := calculateSrcHash(projectPath)
    if err != nil {
        return true, err // Assume changed if we can't calculate hash
    }

    // Load stored hash
    store, err := loadHashStore(projectPath)
    if err != nil {
        return true, nil // Assume changed if we can't load store
    }

    // If no previous hash, consider it changed
    if store.SrcHash == "" {
        return true, nil
    }

    // Compare hashes
    return currentHash != store.SrcHash, nil
}

// updateSrcHash updates the stored hash for a project
func updateSrcHash(projectPath string) error {
    currentHash, fileHashes, err := calculateSrcHash(projectPath)
    if err != nil {
        return err
    }

    store := &HashStore{
        ProjectPath: projectPath,
        SrcHash:     currentHash,
        Files:       fileHashes,
    }

    return saveHashStore(store)
}

// getSrcHashDisplay returns a short hash for display purposes
func getSrcHashDisplay(hash string) string {
    if hash == "" {
        return "none"
    }

    if len(hash) > 8 {
        return hash[:8]
    }

    return hash
}

// shouldRebuildDependency checks if a dependency needs to be rebuilt
// Returns (shouldRebuild, currentHash, storedHash, error)
func shouldRebuildDependency(projectPath string) (bool, string, string, error) {
    // Calculate current hash
    currentHash, _, err := calculateSrcHash(projectPath)
    if err != nil {
        return true, "", "", err // Assume rebuild if we can't calculate hash
    }

    // Load stored hash
    store, err := loadHashStore(projectPath)
    if err != nil {
        return true, currentHash, "", nil // Assume rebuild if we can't load store
    }

    storedHash := store.SrcHash

    // If no previous hash, need to rebuild
    if storedHash == "" {
        return true, currentHash, "", nil
    }

    // Compare hashes
    shouldRebuild := currentHash != storedHash
    return shouldRebuild, currentHash, storedHash, nil
}

// printHashComparison prints a comparison of hashes for debugging
func printHashComparison(projectPath string, currentHash, storedHash string) {
    if storedHash == "" {
        fmt.Printf("%s  Hash: %s (new)%s\n", colors.Yellow, getSrcHashDisplay(currentHash), colors.Reset)
    } else if currentHash != storedHash {
        fmt.Printf("%s  Hash changed: %s -> %s%s\n", colors.Yellow, getSrcHashDisplay(storedHash), getSrcHashDisplay(currentHash), colors.Reset)
    } else {
        fmt.Printf("%s  Hash: %s (unchanged)%s\n", colors.Green, getSrcHashDisplay(currentHash), colors.Reset)
    }
}
