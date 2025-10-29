// Package fileutil provides secure file operation utilities.
package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Validate path (prevent path traversal)
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
	}

	if err := os.MkdirAll(cleanPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// WriteFile writes content to a file securely.
func WriteFile(path string, content []byte) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Validate path
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
	}

	// Ensure directory exists
	dir := filepath.Dir(cleanPath)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	// Write to temp file first
	tmpFile, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Write content
	if _, err = tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Close temp file
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err = os.Rename(tmpPath, cleanPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// ReadFile reads a file securely.
func ReadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Validate path
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path: path traversal detected")
	}

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", cleanPath)
	}

	// Read file
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}

// CopyFile copies a file from src to dst securely.
func CopyFile(src, dst string) error {
	if src == "" || dst == "" {
		return fmt.Errorf("source and destination paths cannot be empty")
	}

	// Validate paths
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	if strings.Contains(cleanSrc, "..") || strings.Contains(cleanDst, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
	}

	// Open source file
	srcFile, err := os.Open(cleanSrc)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	if err := EnsureDir(filepath.Dir(cleanDst)); err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.Create(cleanDst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Sync to disk
	if err = dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	if path == "" {
		return false
	}

	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	if path == "" {
		return false
	}

	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

// ListFiles lists all files in a directory matching a pattern.
func ListFiles(dir, pattern string) ([]string, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	cleanDir := filepath.Clean(dir)
	if strings.Contains(cleanDir, "..") {
		return nil, fmt.Errorf("invalid path: path traversal detected")
	}

	if !DirExists(cleanDir) {
		return nil, fmt.Errorf("directory does not exist: %s", cleanDir)
	}

	var matches []string
	err := filepath.Walk(cleanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if pattern == "" || pattern == "*" {
			matches = append(matches, path)
			return nil
		}

		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return matches, nil
}

// RemoveFile removes a file securely.
func RemoveFile(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
	}

	if !FileExists(cleanPath) {
		return nil // File doesn't exist, nothing to do
	}

	if err := os.Remove(cleanPath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

// GetFileSize returns the size of a file in bytes.
func GetFileSize(path string) (int64, error) {
	if path == "" {
		return 0, fmt.Errorf("path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return 0, fmt.Errorf("invalid path: path traversal detected")
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	return info.Size(), nil
}
