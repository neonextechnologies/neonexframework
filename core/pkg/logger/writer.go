package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileWriter writes logs to a file with rotation support
type FileWriter struct {
	mu           sync.Mutex
	filename     string
	file         *os.File
	maxSize      int64 // Maximum size in bytes
	maxBackups   int   // Maximum number of backup files
	maxAge       int   // Maximum age in days
	size         int64
	rotateOnDate bool
	currentDate  string
}

// FileWriterConfig holds configuration for file writer
type FileWriterConfig struct {
	Filename     string
	MaxSize      int64 // In MB
	MaxBackups   int
	MaxAge       int // In days
	RotateOnDate bool
}

// NewFileWriter creates a new file writer
func NewFileWriter(config FileWriterConfig) (*FileWriter, error) {
	if config.MaxSize == 0 {
		config.MaxSize = 100 // Default 100MB
	}

	fw := &FileWriter{
		filename:     config.Filename,
		maxSize:      config.MaxSize * 1024 * 1024, // Convert MB to bytes
		maxBackups:   config.MaxBackups,
		maxAge:       config.MaxAge,
		rotateOnDate: config.RotateOnDate,
		currentDate:  time.Now().Format("2006-01-02"),
	}

	if err := fw.openFile(); err != nil {
		return nil, err
	}

	return fw, nil
}

// Write writes data to the file
func (fw *FileWriter) Write(p []byte) (n int, err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Check if we need to rotate
	if fw.shouldRotate() {
		if err := fw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = fw.file.Write(p)
	fw.size += int64(n)
	return n, err
}

// Close closes the file
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}

// openFile opens the log file
func (fw *FileWriter) openFile() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(fw.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open or create file
	file, err := os.OpenFile(fw.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fw.file = file

	// Get current size
	info, err := file.Stat()
	if err != nil {
		return err
	}
	fw.size = info.Size()

	return nil
}

// shouldRotate checks if the file should be rotated
func (fw *FileWriter) shouldRotate() bool {
	// Check size
	if fw.size >= fw.maxSize {
		return true
	}

	// Check date
	if fw.rotateOnDate {
		currentDate := time.Now().Format("2006-01-02")
		if currentDate != fw.currentDate {
			fw.currentDate = currentDate
			return true
		}
	}

	return false
}

// rotate rotates the log file
func (fw *FileWriter) rotate() error {
	// Close current file
	if fw.file != nil {
		fw.file.Close()
	}

	// Rename current file
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s", fw.filename, timestamp)

	if err := os.Rename(fw.filename, backupName); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Open new file
	if err := fw.openFile(); err != nil {
		return err
	}

	// Clean old backups
	go fw.cleanOldBackups()

	return nil
}

// cleanOldBackups removes old backup files
func (fw *FileWriter) cleanOldBackups() {
	dir := filepath.Dir(fw.filename)
	base := filepath.Base(fw.filename)

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []os.DirEntry
	for _, f := range files {
		if !f.IsDir() && len(f.Name()) > len(base) && f.Name()[:len(base)] == base {
			backups = append(backups, f)
		}
	}

	// Remove by count
	if fw.maxBackups > 0 && len(backups) > fw.maxBackups {
		for i := 0; i < len(backups)-fw.maxBackups; i++ {
			os.Remove(filepath.Join(dir, backups[i].Name()))
		}
	}

	// Remove by age
	if fw.maxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -fw.maxAge)
		for _, backup := range backups {
			info, err := backup.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				os.Remove(filepath.Join(dir, backup.Name()))
			}
		}
	}
}

// MultiWriter writes to multiple writers
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter creates a new multi writer
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write writes to all writers
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}

// Add adds a writer
func (mw *MultiWriter) Add(writer io.Writer) {
	mw.writers = append(mw.writers, writer)
}
