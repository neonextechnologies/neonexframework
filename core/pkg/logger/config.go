package logger

import (
	"io"
	"os"
)

// Config holds logger configuration
type Config struct {
	Level        string
	Format       string // "text" or "json"
	Output       string // "console", "file", or "both"
	FilePath     string
	MaxSize      int64 // In MB
	MaxBackups   int
	MaxAge       int // In days
	EnableCaller bool
	EnableColor  bool
	RotateOnDate bool
	PrettyPrint  bool // For JSON format
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:        "info",
		Format:       "text",
		Output:       "console",
		FilePath:     "logs/app.log",
		MaxSize:      100,
		MaxBackups:   7,
		MaxAge:       30,
		EnableCaller: true,
		EnableColor:  true,
		RotateOnDate: false,
		PrettyPrint:  false,
	}
}

// LoadConfig loads logger configuration from environment
func LoadConfig() Config {
	config := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = format
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = output
	}
	if path := os.Getenv("LOG_FILE_PATH"); path != "" {
		config.FilePath = path
	}

	return config
}

// Setup configures the global logger based on config
func Setup(config Config) error {
	// Set level
	level := parseLevel(config.Level)
	SetGlobalLevel(level)

	// Set formatter
	var formatter Formatter
	if config.Format == "json" {
		jsonFormatter := NewJSONFormatter()
		jsonFormatter.PrettyPrint = config.PrettyPrint
		formatter = jsonFormatter
	} else {
		textFormatter := NewTextFormatter()
		textFormatter.DisableColors = !config.EnableColor
		formatter = textFormatter
	}
	SetGlobalFormatter(formatter)

	// Enable/disable caller
	EnableGlobalCaller(config.EnableCaller)
	EnableGlobalColor(config.EnableColor)

	// Setup output
	if config.Output == "file" || config.Output == "both" {
		fileWriter, err := NewFileWriter(FileWriterConfig{
			Filename:     config.FilePath,
			MaxSize:      config.MaxSize,
			MaxBackups:   config.MaxBackups,
			MaxAge:       config.MaxAge,
			RotateOnDate: config.RotateOnDate,
		})
		if err != nil {
			return err
		}

		if config.Output == "file" {
			// Replace console with file
			defaultLogger.writers = []io.Writer{fileWriter}
		} else {
			// Add file to existing console
			AddGlobalWriter(fileWriter)
		}
	}

	return nil
}

// parseLevel parses string level to LogLevel
func parseLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
