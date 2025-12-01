package logger

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TextFormatter formats logs as text
type TextFormatter struct {
	DisableColors   bool
	FullTimestamp   bool
	TimestampFormat string
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
}

// Format formats a log entry as text
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var b strings.Builder

	// Timestamp
	timestamp := entry.Time.Format(f.TimestampFormat)

	// Color
	colorReset := "\033[0m"
	levelColor := entry.Level.Color()
	if f.DisableColors {
		levelColor = ""
		colorReset = ""
	}

	// Level
	level := fmt.Sprintf("%-5s", entry.Level.String())

	// Build message
	b.WriteString(fmt.Sprintf("%s[%s]%s ", levelColor, timestamp, colorReset))
	b.WriteString(fmt.Sprintf("%s%s%s ", levelColor, level, colorReset))

	// Caller info
	if entry.File != "" {
		b.WriteString(fmt.Sprintf("[%s:%d] ", entry.File, entry.Line))
	}

	// Message
	b.WriteString(entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		b.WriteString(" | ")
		first := true
		for k, v := range entry.Fields {
			if !first {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// JSONFormatter formats logs as JSON
type JSONFormatter struct {
	PrettyPrint bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		PrettyPrint: false,
	}
}

// Format formats a log entry as JSON
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	data := make(map[string]interface{})

	data["time"] = entry.Time.Format("2006-01-02T15:04:05.000Z07:00")
	data["level"] = entry.Level.String()
	data["message"] = entry.Message

	if entry.File != "" {
		data["caller"] = fmt.Sprintf("%s:%d", entry.File, entry.Line)
	}

	// Add fields
	for k, v := range entry.Fields {
		data[k] = v
	}

	var result []byte
	var err error

	if f.PrettyPrint {
		result, err = json.MarshalIndent(data, "", "  ")
	} else {
		result, err = json.Marshal(data)
	}

	if err != nil {
		return nil, err
	}

	result = append(result, '\n')
	return result, nil
}
