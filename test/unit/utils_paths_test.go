package unit

import (
	"os"
	"path/filepath"
	"testing"

	"ai-gateway-hub/internal/utils"
)

func TestPathManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pathmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	t.Run("NewPathManager", func(t *testing.T) {
		pm, err := utils.NewPathManager()
		if err != nil {
			t.Fatalf("NewPathManager failed: %v", err)
		}
		if pm == nil {
			t.Fatal("PathManager is nil")
		}
		
		wd := pm.GetWorkingDir()
		if wd != tempDir {
			t.Errorf("Expected working dir %s, got %s", tempDir, wd)
		}
	})

	t.Run("EnsureDir", func(t *testing.T) {
		pm, _ := utils.NewPathManager()
		
		testDir := "test/subdir"
		err := pm.EnsureDir(testDir)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}
		
		// Check if directory exists
		expectedPath := filepath.Join(tempDir, testDir)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", expectedPath)
		}
	})

	t.Run("EnsureDirForFile", func(t *testing.T) {
		pm, _ := utils.NewPathManager()
		
		testFile := "logs/app/test.log"
		err := pm.EnsureDirForFile(testFile)
		if err != nil {
			t.Fatalf("EnsureDirForFile failed: %v", err)
		}
		
		// Check if directory for file exists
		expectedDir := filepath.Join(tempDir, "logs/app")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created for file", expectedDir)
		}
	})

	t.Run("ResolvePath", func(t *testing.T) {
		pm, _ := utils.NewPathManager()
		
		// Test relative path
		relativePath := "data/test.db"
		resolved := pm.ResolvePath(relativePath)
		expected := filepath.Join(tempDir, relativePath)
		if resolved != expected {
			t.Errorf("Expected %s, got %s", expected, resolved)
		}
		
		// Test absolute path
		absolutePath := "/tmp/test.db"
		resolved = pm.ResolvePath(absolutePath)
		if resolved != absolutePath {
			t.Errorf("Absolute path should not be modified: expected %s, got %s", absolutePath, resolved)
		}
	})
}

func TestGlobalPathManager(t *testing.T) {
	// Test global instance functions
	tempDir, err := os.MkdirTemp("", "global_pathmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	t.Run("InitPathManager", func(t *testing.T) {
		err := utils.InitPathManager()
		if err != nil {
			t.Fatalf("InitPathManager failed: %v", err)
		}
		
		pm := utils.GetPathManager()
		if pm == nil {
			t.Fatal("Global PathManager is nil after initialization")
		}
	})

	t.Run("GlobalEnsureDir", func(t *testing.T) {
		utils.InitPathManager()
		
		testDir := "global/test/dir"
		err := utils.EnsureDir(testDir)
		if err != nil {
			t.Fatalf("Global EnsureDir failed: %v", err)
		}
		
		expectedPath := filepath.Join(tempDir, testDir)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", expectedPath)
		}
	})

	t.Run("GlobalEnsureDirForFile", func(t *testing.T) {
		utils.InitPathManager()
		
		testFile := "global/logs/test.log"
		err := utils.EnsureDirForFile(testFile)
		if err != nil {
			t.Fatalf("Global EnsureDirForFile failed: %v", err)
		}
		
		expectedDir := filepath.Join(tempDir, "global/logs")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created for file", expectedDir)
		}
	})

	t.Run("GlobalResolvePath", func(t *testing.T) {
		utils.InitPathManager()
		
		relativePath := "data/global.db"
		resolved := utils.ResolvePath(relativePath)
		expected := filepath.Join(tempDir, relativePath)
		if resolved != expected {
			t.Errorf("Expected %s, got %s", expected, resolved)
		}
	})
}