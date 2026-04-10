package review

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Type represents which review file variant to operate on.
type Type string

const (
	TypeDefault Type = "default" // maps to review.yml
	TypeDocs    Type = "docs"    // maps to review-docs.yml
	TypeDevOps  Type = "devops"  // maps to review-devops.yml
)

// Finding represents a single review comment or issue.
type Finding struct {
	ID       string `yaml:"id"`
	File     string `yaml:"file"`
	Line     int    `yaml:"line"`
	Feedback string `yaml:"feedback"`
	Status   string `yaml:"status"`
}

var kebabCase = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// filename returns the YAML file name for the given review type.
func filename(t Type) (string, error) {
	switch t {
	case TypeDefault:
		return "review.yml", nil
	case TypeDocs:
		return "review-docs.yml", nil
	case TypeDevOps:
		return "review-devops.yml", nil
	default:
		return "", fmt.Errorf("unknown review type: %q", t)
	}
}

func validate(f Finding) error {
	if f.ID == "" {
		return fmt.Errorf("finding ID must not be empty")
	}
	if !kebabCase.MatchString(f.ID) {
		return fmt.Errorf("finding ID %q must be kebab-case (lowercase letters, digits, hyphens)", f.ID)
	}
	if f.Feedback == "" {
		return fmt.Errorf("finding Feedback must not be empty")
	}
	if f.Status != "open" && f.Status != "resolved" {
		return fmt.Errorf("finding Status %q must be \"open\" or \"resolved\"", f.Status)
	}
	return nil
}

func atomicWrite(path string, findings []Finding) error {
	data, err := yaml.Marshal(findings)
	if err != nil {
		return fmt.Errorf("marshaling findings: %w", err)
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".review.tmp.*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()        //nolint:errcheck
		os.Remove(tmpName) //nolint:errcheck
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName) //nolint:errcheck
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// Create creates the review file for the given type if it does not already exist (idempotent).
func Create(featureDir string, t Type) error {
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		return fmt.Errorf("feature directory does not exist: %s", featureDir)
	}
	name, err := filename(t)
	if err != nil {
		return err
	}
	path := filepath.Join(featureDir, name)
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	return atomicWrite(path, []Finding{})
}

// Load reads and validates all findings from the review file for the given type.
// Returns an empty slice if the file does not exist.
func Load(featureDir string, t Type) ([]Finding, error) {
	name, err := filename(t)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(featureDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Finding{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", name, err)
	}
	var findings []Finding
	if err := yaml.Unmarshal(data, &findings); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", name, err)
	}
	for _, f := range findings {
		if err := validate(f); err != nil {
			return nil, fmt.Errorf("invalid finding in %s: %w", name, err)
		}
	}
	return findings, nil
}

// Append validates and atomically appends a finding to the review file.
func Append(featureDir string, t Type, f Finding) error {
	if err := validate(f); err != nil {
		return err
	}
	name, err := filename(t)
	if err != nil {
		return err
	}
	path := filepath.Join(featureDir, name)

	existing, err := Load(featureDir, t)
	if err != nil {
		return fmt.Errorf("loading existing findings: %w", err)
	}
	return atomicWrite(path, append(existing, f))
}

// UpdateStatus updates the status of a single finding by ID (atomic write).
// Returns an error if the ID is not found or the status is invalid.
func UpdateStatus(featureDir string, t Type, id, status string) error {
	if status != "open" && status != "resolved" {
		return fmt.Errorf("status %q must be \"open\" or \"resolved\"", status)
	}
	name, err := filename(t)
	if err != nil {
		return err
	}
	path := filepath.Join(featureDir, name)

	findings, err := Load(featureDir, t)
	if err != nil {
		return fmt.Errorf("loading findings: %w", err)
	}
	found := false
	for i := range findings {
		if findings[i].ID == id {
			findings[i].Status = status
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("finding with ID %q not found in %s", id, name)
	}
	return atomicWrite(path, findings)
}
