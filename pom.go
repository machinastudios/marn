package main

import (
    "encoding/xml"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

// POM represents the Maven pom.xml structure
type POM struct {
    XMLName      xml.Name   `xml:"project"`
    ArtifactID   string     `xml:"artifactId"`
    Properties   Properties `xml:"properties"`
    Dependencies struct {
        Dependency []Dependency `xml:"dependency"`
    } `xml:"dependencies"`
    Build struct {
        Plugins struct {
            Plugin []struct {
                Configuration struct {
                    MainClass string `xml:"mainClass"`
                } `xml:"configuration"`
            } `xml:"plugin"`
        } `xml:"plugins"`
    } `xml:"build"`
}

// Properties holds the pom.xml properties including scripts
type Properties struct {
    Raw []byte `xml:",innerxml"`
}

// Dependency represents a Maven dependency
type Dependency struct {
    GroupID    string `xml:"groupId"`
    ArtifactID string `xml:"artifactId"`
    Version    string `xml:"version"`
}

// getScriptsFromPom extracts scripts from pom.xml properties
func getScriptsFromPom() map[string]string {
    scripts := make(map[string]string)

    content, err := os.ReadFile(pomFile)
    if err != nil {
        return scripts
    }

    // Use regex to find script.* properties
    re := regexp.MustCompile(`<script\.([^>]+)>([^<]*)</script\.[^>]+>`)
    matches := re.FindAllStringSubmatch(string(content), -1)

    for _, match := range matches {

        if len(match) >= 3 {
            scripts[match[1]] = strings.TrimSpace(match[2])
        }
    }

    return scripts
}

// getProperty extracts a property from pom.xml
func getProperty(propName string) string {
    content, err := os.ReadFile(pomFile)
    if err != nil {
        return ""
    }

    // Try to find the property with regex
    re := regexp.MustCompile(fmt.Sprintf(`<%s>([^<]*)</%s>`, propName, propName))
    match := re.FindStringSubmatch(string(content))

    if len(match) >= 2 {
        return strings.TrimSpace(match[1])
    }

    return ""
}

// getArtifactID gets the artifact ID from pom.xml
func getArtifactID() string {
    content, err := os.ReadFile(pomFile)
    if err != nil {
        return ""
    }

    var pom POM
    if err := xml.Unmarshal(content, &pom); err != nil {
        return ""
    }

    return pom.ArtifactID
}

// getMainClass gets the main class from pom.xml
func getMainClass() string {
    return getProperty("mainClass")
}

// getLocalDependencies finds local SNAPSHOT dependencies
func getLocalDependencies() []string {
    var deps []string

    // Check for configured deps
    configuredDeps := getProperty("watch.localDeps")
    if configuredDeps != "" {

        for _, dep := range strings.Fields(configuredDeps) {
            deps = append(deps, dep)
        }
    }

    // Parse pom.xml to find SNAPSHOT dependencies
    content, err := os.ReadFile(pomFile)
    if err != nil {
        return deps
    }

    // Parse the POM
    var pom POM
    if err := xml.Unmarshal(content, &pom); err != nil {
        return deps
    }

    // Find SNAPSHOT dependencies and check for local directories
    for _, dep := range pom.Dependencies.Dependency {

        if strings.Contains(dep.Version, "SNAPSHOT") {

            // Try sibling directory
            siblingPath := filepath.Join(currentDir, "..", dep.ArtifactID)
            siblingPom := filepath.Join(siblingPath, "pom.xml")

            if _, err := os.Stat(siblingPom); err == nil {
                absPath, _ := filepath.Abs(siblingPath)
                deps = append(deps, absPath)
            }
        }
    }

    // Remove duplicates
    seen := make(map[string]bool)
    var uniqueDeps []string

    for _, dep := range deps {

        if !seen[dep] {
            seen[dep] = true
            uniqueDeps = append(uniqueDeps, dep)
        }
    }

    return uniqueDeps
}
