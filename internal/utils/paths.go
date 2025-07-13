package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// PathManager handles all path-related operations
type PathManager struct {
	workingDir string
}

// NewPathManager creates a new path manager
func NewPathManager() (*PathManager, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	return &PathManager{workingDir: wd}, nil
}

// EnsureDir creates directory if it doesn't exist
func (pm *PathManager) EnsureDir(path string) error {
	absPath := pm.ResolvePath(path)
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", absPath, err)
	}
	return nil
}

// EnsureDirForFile creates directory for the given file path
func (pm *PathManager) EnsureDirForFile(filePath string) error {
	dir := filepath.Dir(pm.ResolvePath(filePath))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

// ResolvePath resolves relative path to absolute path
func (pm *PathManager) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(pm.workingDir, path)
}

// GetWorkingDir returns the current working directory
func (pm *PathManager) GetWorkingDir() string {
	return pm.workingDir
}

// GetDirForFile returns the directory containing the file
func (pm *PathManager) GetDirForFile(filePath string) string {
	return filepath.Dir(pm.ResolvePath(filePath))
}

// Global path manager instance
var globalPathManager *PathManager

// InitPathManager initializes the global path manager
func InitPathManager() error {
	pm, err := NewPathManager()
	if err != nil {
		return err
	}
	globalPathManager = pm
	return nil
}

// GetPathManager returns the global path manager instance
func GetPathManager() *PathManager {
	return globalPathManager
}

// Convenience functions using global instance
func EnsureDir(path string) error {
	if globalPathManager == nil {
		return fmt.Errorf("path manager not initialized")
	}
	return globalPathManager.EnsureDir(path)
}

func EnsureDirForFile(filePath string) error {
	if globalPathManager == nil {
		return fmt.Errorf("path manager not initialized")
	}
	return globalPathManager.EnsureDirForFile(filePath)
}

func ResolvePath(path string) string {
	if globalPathManager == nil {
		return path
	}
	return globalPathManager.ResolvePath(path)
}