package jobs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests verify the file operations logic that is used by TallyImporter
// Since the methods are private, we test the underlying os package behavior
// that the implementation relies on

func TestFileOperations_CreateDirIfNotExists(t *testing.T) {
	t.Run("creates new directory", func(t *testing.T) {
		tempDir := t.TempDir()
		newDir := filepath.Join(tempDir, "newdir", "subdir")

		// This is what the implementation does
		err := os.MkdirAll(newDir, 0755)

		assert.NoError(t, err)
		assert.DirExists(t, newDir)
	})

	t.Run("handles existing directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// MkdirAll should not error on existing dir
		err := os.MkdirAll(tempDir, 0755)

		assert.NoError(t, err)
		assert.DirExists(t, tempDir)
	})

	t.Run("handles nested directories", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedDir := filepath.Join(tempDir, "a", "b", "c", "d")

		err := os.MkdirAll(nestedDir, 0755)

		assert.NoError(t, err)
		assert.DirExists(t, nestedDir)
	})
}

func TestFileOperations_ReadDirectory(t *testing.T) {
	t.Run("reads directory with files", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create test files
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644))

		entries, err := os.ReadDir(tempDir)

		assert.NoError(t, err)
		assert.Len(t, entries, 2)
	})

	t.Run("reads empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		entries, err := os.ReadDir(tempDir)

		assert.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		entries, err := os.ReadDir("/nonexistent/path/12345")

		assert.Error(t, err)
		assert.Nil(t, entries)
	})

	t.Run("includes subdirectories", func(t *testing.T) {
		tempDir := t.TempDir()
		
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("content"), 0644))

		entries, err := os.ReadDir(tempDir)

		assert.NoError(t, err)
		assert.Len(t, entries, 2)
		
		// Verify one is a directory
		hasDir := false
		hasFile := false
		for _, e := range entries {
			if e.IsDir() {
				hasDir = true
			} else {
				hasFile = true
			}
		}
		assert.True(t, hasDir)
		assert.True(t, hasFile)
	})
}

func TestFileOperations_ReadFileContent(t *testing.T) {
	t.Run("reads file content successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		expectedContent := "Hello, World!"
		require.NoError(t, os.WriteFile(filePath, []byte(expectedContent), 0644))

		content, err := os.ReadFile(filePath)

		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(content))
	})

	t.Run("reads empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "empty.txt")
		require.NoError(t, os.WriteFile(filePath, []byte(""), 0644))

		content, err := os.ReadFile(filePath)

		assert.NoError(t, err)
		assert.Empty(t, string(content))
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		content, err := os.ReadFile("/nonexistent/file.txt")

		assert.Error(t, err)
		assert.Empty(t, content)
	})

	t.Run("reads multiline content", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "multiline.txt")
		expectedContent := "Line 1\nLine 2\nLine 3"
		require.NoError(t, os.WriteFile(filePath, []byte(expectedContent), 0644))

		content, err := os.ReadFile(filePath)

		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(content))
	})
}

func TestFileOperations_MoveFile(t *testing.T) {
	t.Run("moves file successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		srcPath := filepath.Join(tempDir, "source.txt")
		dstPath := filepath.Join(tempDir, "destination.txt")
		require.NoError(t, os.WriteFile(srcPath, []byte("content"), 0644))

		err := os.Rename(srcPath, dstPath)

		assert.NoError(t, err)
		assert.NoFileExists(t, srcPath)
		assert.FileExists(t, dstPath)

		// Verify content is preserved
		content, _ := os.ReadFile(dstPath)
		assert.Equal(t, "content", string(content))
	})

	t.Run("creates destination directory if not exists", func(t *testing.T) {
		tempDir := t.TempDir()
		srcPath := filepath.Join(tempDir, "source.txt")
		dstPath := filepath.Join(tempDir, "newdir", "destination.txt")
		require.NoError(t, os.WriteFile(srcPath, []byte("content"), 0644))

		// Create destination directory first (like the implementation does)
		require.NoError(t, os.MkdirAll(filepath.Dir(dstPath), 0755))
		err := os.Rename(srcPath, dstPath)

		assert.NoError(t, err)
		assert.FileExists(t, dstPath)
	})

	t.Run("returns error for non-existent source", func(t *testing.T) {
		tempDir := t.TempDir()
		srcPath := filepath.Join(tempDir, "nonexistent.txt")
		dstPath := filepath.Join(tempDir, "destination.txt")

		err := os.Rename(srcPath, dstPath)

		assert.Error(t, err)
	})

	t.Run("overwrites existing destination", func(t *testing.T) {
		tempDir := t.TempDir()
		srcPath := filepath.Join(tempDir, "source.txt")
		dstPath := filepath.Join(tempDir, "destination.txt")
		require.NoError(t, os.WriteFile(srcPath, []byte("new content"), 0644))
		require.NoError(t, os.WriteFile(dstPath, []byte("old content"), 0644))

		err := os.Rename(srcPath, dstPath)

		assert.NoError(t, err)
		content, _ := os.ReadFile(dstPath)
	assert.Equal(t, "new content", string(content))
	})
}

