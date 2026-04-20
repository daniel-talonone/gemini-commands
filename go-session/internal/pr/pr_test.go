package pr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "feature-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Test case 1: Create a new pr.md
	err = Create(tempDir)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	filePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("pr.md was not created")
	}

	// Test case 2: Create when pr.md already exists (should do nothing and not error)
	err = Create(tempDir)
	if err != nil {
		t.Fatalf("Create failed when file already exists: %v", err)
	}
}

func TestWriteAndRead(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "feature-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to clean up temp dir: %v", err)
		}
	}()

	content := "This is a test PR description."

	// Test Write
	err = Write(tempDir, content)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Test Read
	readContent, err := Read(tempDir)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if readContent != content {
		t.Errorf("Read content mismatch. Expected '%s', got '%s'", content, readContent)
	}

	// Test Read with non-existent file
	if err := os.Remove(filepath.Join(tempDir, filename)); err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}
	readContent, err = Read(tempDir)
	if err != nil {
		t.Fatalf("Read failed for non-existent file: %v", err)
	}
	if readContent != "" {
		t.Errorf("Read content mismatch for non-existent file. Expected empty string, got '%s'", readContent)
	}
}
