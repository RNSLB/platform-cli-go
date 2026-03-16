package namespace

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// ============================================================
// TYPE DEFINITIONS
// ============================================================

// Namespace represents a Kubernetes namespace configuration
type Namespace struct {
	Name        string
	Team        string
	CPUQuota    int
	MemoryQuota string
	Environment string
}

// NamespaceConfig wraps the YAML structure
type NamespaceConfig struct {
	Namespaces []YAMLNamespace `yaml:"namespaces"`
}

// YAMLNamespace represents the YAML structure with tags for parsing
type YAMLNamespace struct {
	Name        string `yaml:"name"`
	Team        string `yaml:"team"`
	CPUQuota    int    `yaml:"cpu_quota"`
	MemoryQuota string `yaml:"memory_quota"`
	Environment string `yaml:"environment"`
}

// ============================================================
// CONVERSION METHODS
// ============================================================

// toNamespace converts YAMLNamespace to Namespace
func (yn YAMLNamespace) toNamespace() Namespace {
	return Namespace{
		Name:        yn.Name,
		Team:        yn.Team,
		CPUQuota:    yn.CPUQuota,
		MemoryQuota: yn.MemoryQuota,
		Environment: yn.Environment,
	}
}

// ============================================================
// FILE I/O FUNCTIONS
// ============================================================

// LoadFromYAML reads and parses namespace configurations from a YAML file
//
// Parameters:
//   - filename: path to the YAML file
//
// Returns:
//   - []Namespace: slice of parsed namespaces
//   - error: any error that occurred during reading or parsing
func LoadFromYAML(filename string) ([]Namespace, error) {
	// Read file using os.ReadFile (replaces deprecated ioutil.ReadFile)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML into config structure
	var config NamespaceConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert YAMLNamespace slice to Namespace slice
	namespaces := make([]Namespace, len(config.Namespaces))
	for i, yn := range config.Namespaces {
		namespaces[i] = yn.toNamespace()
	}

	return namespaces, nil
}

// ============================================================
// VALIDATION FUNCTIONS
// ============================================================

// Validate checks if a namespace name follows Kubernetes naming rules
//
// Validation rules:
//   - Length must be 1-63 characters
//   - Only lowercase letters (a-z), numbers (0-9), and hyphens (-) allowed
//   - Cannot start with a hyphen
//   - Cannot end with a hyphen
//
// Parameters:
//   - name: the namespace name to validate
//
// Returns:
//   - bool: true if valid, false if invalid
//   - string: empty string if valid, error message if invalid
func Validate(name string) (bool, string) {
	// Check if empty
	if len(name) == 0 {
		return false, "namespace name cannot be empty"
	}

	// Check maximum length
	if len(name) > 63 {
		return false, fmt.Sprintf("namespace name too long (%d characters, max 63)", len(name))
	}

	// Cannot start with hyphen
	if name[0] == '-' {
		return false, "namespace name cannot start with a hyphen"
	}

	// Cannot end with hyphen
	if name[len(name)-1] == '-' {
		return false, "namespace name cannot end with a hyphen"
	}

	// Must be all lowercase
	if name != strings.ToLower(name) {
		return false, "namespace name must be lowercase (no uppercase letters allowed)"
	}

	// Check each character is valid
	for i, char := range name {
		if !isValidChar(char) {
			return false, fmt.Sprintf("invalid character '%c' at position %d (only a-z, 0-9, and - allowed)", char, i)
		}
	}

	// All validations passed
	return true, ""
}

// isValidChar checks if a character is allowed in a namespace name
//
// Valid characters:
//   - Lowercase letters: a-z
//   - Numbers: 0-9
//   - Hyphen: -
//
// Parameters:
//   - char: the character to check (as a rune)
//
// Returns:
//   - bool: true if character is valid, false otherwise
func isValidChar(char rune) bool {
	isLowercaseLetter := char >= 'a' && char <= 'z'
	isDigit := char >= '0' && char <= '9'
	isHyphen := char == '-'

	return isLowercaseLetter || isDigit || isHyphen
}

// ValidateAll validates all namespaces in a slice
//
// Parameters:
//   - namespaces: slice of namespaces to validate
//
// Returns:
//   - validCount: number of valid namespaces
//   - invalidCount: number of invalid namespaces
//   - results: slice of validation results with namespace info
func ValidateAll(namespaces []Namespace) (int, int, []ValidationResult) {
	validCount := 0
	invalidCount := 0
	results := make([]ValidationResult, len(namespaces))

	for i, ns := range namespaces {
		valid, reason := Validate(ns.Name)

		results[i] = ValidationResult{
			Namespace: ns,
			Valid:     valid,
			Reason:    reason,
		}

		if valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	return validCount, invalidCount, results
}

// ============================================================
// HELPER TYPES
// ============================================================

// ValidationResult holds the result of validating a single namespace
type ValidationResult struct {
	Namespace Namespace
	Valid     bool
	Reason    string // Empty if valid, contains error message if invalid
}

// ============================================================
// RESERVED NAMES VALIDATION (OPTIONAL)
// ============================================================

// reservedNames lists Kubernetes reserved namespace names
var reservedNames = []string{
	"default",
	"kube-system",
	"kube-public",
	"kube-node-lease",
}

// IsReserved checks if a namespace name is reserved by Kubernetes
//
// Parameters:
//   - name: the namespace name to check
//
// Returns:
//   - bool: true if the name is reserved, false otherwise
func IsReserved(name string) bool {
	for _, reserved := range reservedNames {
		if name == reserved {
			return true
		}
	}
	return false
}

// ValidateWithReserved performs validation including reserved name check
//
// Parameters:
//   - name: the namespace name to validate
//
// Returns:
//   - bool: true if valid, false if invalid
//   - string: empty string if valid, error message if invalid
func ValidateWithReserved(name string) (bool, string) {
	// First do standard validation
	valid, reason := Validate(name)
	if !valid {
		return false, reason
	}

	// Check if reserved
	if IsReserved(name) {
		return false, fmt.Sprintf("'%s' is a reserved Kubernetes namespace", name)
	}

	return true, ""
}

// ============================================================
// UTILITY FUNCTIONS
// ============================================================

// FilterByTeam returns all namespaces belonging to a specific team
//
// Parameters:
//   - namespaces: slice of namespaces to filter
//   - team: team name to filter by
//
// Returns:
//   - []Namespace: filtered slice containing only namespaces for the specified team
func FilterByTeam(namespaces []Namespace, team string) []Namespace {
	var filtered []Namespace
	for _, ns := range namespaces {
		if ns.Team == team {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

// FilterByEnvironment returns all namespaces for a specific environment
//
// Parameters:
//   - namespaces: slice of namespaces to filter
//   - environment: environment name to filter by (e.g., "dev", "prod")
//
// Returns:
//   - []Namespace: filtered slice containing only namespaces for the specified environment
func FilterByEnvironment(namespaces []Namespace, environment string) []Namespace {
	var filtered []Namespace
	for _, ns := range namespaces {
		if ns.Environment == environment {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

// FindByName searches for a namespace by name
//
// Parameters:
//   - namespaces: slice of namespaces to search
//   - name: namespace name to find
//
// Returns:
//   - *Namespace: pointer to the found namespace, or nil if not found
func FindByName(namespaces []Namespace, name string) *Namespace {
	for i := range namespaces {
		if namespaces[i].Name == name {
			return &namespaces[i]
		}
	}
	return nil
}

// ============================================================
// SUMMARY FUNCTIONS
// ============================================================

// Summary holds aggregate information about a set of namespaces
type Summary struct {
	TotalCount        int
	ValidCount        int
	InvalidCount      int
	TeamCounts        map[string]int
	EnvironmentCounts map[string]int
}

// GetSummary generates a summary of namespace validation results
//
// Parameters:
//   - namespaces: slice of namespaces
//   - results: slice of validation results
//
// Returns:
//   - Summary: aggregate summary information
func GetSummary(namespaces []Namespace, results []ValidationResult) Summary {
	summary := Summary{
		TotalCount:        len(namespaces),
		TeamCounts:        make(map[string]int),
		EnvironmentCounts: make(map[string]int),
	}

	for i, result := range results {
		if result.Valid {
			summary.ValidCount++
		} else {
			summary.InvalidCount++
		}

		// Count by team
		team := namespaces[i].Team
		summary.TeamCounts[team]++

		// Count by environment
		env := namespaces[i].Environment
		summary.EnvironmentCounts[env]++
	}

	return summary
}
