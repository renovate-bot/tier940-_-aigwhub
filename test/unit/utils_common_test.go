package unit

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"ai-gateway-hub/internal/utils"
)

// Note: Environment variable helper tests removed - use viper for configuration management

func TestContextHelpers(t *testing.T) {
	t.Run("NewContext", func(t *testing.T) {
		ctx := utils.NewContext()
		if ctx == nil {
			t.Fatal("NewContext returned nil")
		}
		if ctx.Err() != nil {
			t.Errorf("New context should not have error: %v", ctx.Err())
		}
	})

	t.Run("NewContextWithTimeout", func(t *testing.T) {
		duration := 100 * time.Millisecond
		ctx, cancel := utils.NewContextWithTimeout(duration)
		defer cancel()
		
		if ctx == nil {
			t.Fatal("NewContextWithTimeout returned nil context")
		}
		
		// Wait for timeout
		time.Sleep(150 * time.Millisecond)
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
		}
	})
}

func TestJSONHelpers(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("MarshalJSON", func(t *testing.T) {
		data := TestStruct{Name: "test", Value: 42}
		result, err := utils.MarshalJSON(data)
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
		
		expected := `{"name":"test","value":42}`
		if string(result) != expected {
			t.Errorf("Expected %s, got %s", expected, string(result))
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		jsonData := []byte(`{"name":"test","value":42}`)
		var result TestStruct
		
		err := utils.UnmarshalJSON(jsonData, &result)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		
		if result.Name != "test" || result.Value != 42 {
			t.Errorf("Expected {Name: 'test', Value: 42}, got %+v", result)
		}
	})

	t.Run("UnmarshalJSON_InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"name": invalid}`)
		var result TestStruct
		
		err := utils.UnmarshalJSON(invalidJSON, &result)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Run("WrapError", func(t *testing.T) {
		originalErr := utils.NewError("original error")
		wrappedErr := utils.WrapError(originalErr, "wrapped")
		
		if wrappedErr == nil {
			t.Fatal("WrapError returned nil")
		}
		
		expectedMsg := "wrapped: original error"
		if wrappedErr.Error() != expectedMsg {
			t.Errorf("Expected '%s', got '%s'", expectedMsg, wrappedErr.Error())
		}
	})

	t.Run("WrapError_NilError", func(t *testing.T) {
		result := utils.WrapError(nil, "should not wrap")
		if result != nil {
			t.Errorf("Expected nil for wrapping nil error, got %v", result)
		}
	})

	t.Run("NewError", func(t *testing.T) {
		err := utils.NewError("test error with %s", "format")
		if err == nil {
			t.Fatal("NewError returned nil")
		}
		
		expected := "test error with format"
		if err.Error() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, err.Error())
		}
	})
}

func TestMutexHelpers(t *testing.T) {
	t.Run("WithLock", func(t *testing.T) {
		var mu sync.Mutex
		counter := 0
		
		err := utils.WithLock(&mu, func() error {
			counter++
			return nil
		})
		
		if err != nil {
			t.Errorf("WithLock returned error: %v", err)
		}
		if counter != 1 {
			t.Errorf("Expected counter to be 1, got %d", counter)
		}
	})

	t.Run("WithRLock", func(t *testing.T) {
		var mu sync.RWMutex
		counter := 0
		
		err := utils.WithRLock(&mu, func() error {
			counter++
			return nil
		})
		
		if err != nil {
			t.Errorf("WithRLock returned error: %v", err)
		}
		if counter != 1 {
			t.Errorf("Expected counter to be 1, got %d", counter)
		}
	})

	t.Run("WithLock_PropagatesError", func(t *testing.T) {
		var mu sync.Mutex
		expectedErr := utils.NewError("test error")
		
		err := utils.WithLock(&mu, func() error {
			return expectedErr
		})
		
		if err != expectedErr {
			t.Errorf("Expected error to be propagated, got %v", err)
		}
	})
}

func TestFileHelpers(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file_helpers_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize path manager for testing
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	utils.InitPathManager()

	t.Run("WriteToFile_and_ReadFromFile", func(t *testing.T) {
		testFile := "test/data/file.txt"
		testData := []byte("Hello, World!")
		
		// Write file
		err := utils.WriteToFile(testFile, testData)
		if err != nil {
			t.Fatalf("WriteToFile failed: %v", err)
		}
		
		// Read file
		result, err := utils.ReadFromFile(testFile)
		if err != nil {
			t.Fatalf("ReadFromFile failed: %v", err)
		}
		
		if string(result) != string(testData) {
			t.Errorf("Expected '%s', got '%s'", string(testData), string(result))
		}
	})

	t.Run("CreateFile", func(t *testing.T) {
		testFile := "test/logs/app.log"
		
		file, err := utils.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFile failed: %v", err)
		}
		defer file.Close()
		
		// Write to file
		_, err = file.WriteString("Test log entry")
		if err != nil {
			t.Errorf("Failed to write to created file: %v", err)
		}
	})

	t.Run("ReadFromFile_NonExistentFile", func(t *testing.T) {
		_, err := utils.ReadFromFile("non_existent_file.txt")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}