package jsonlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestLevelString(t *testing.T) {
	// Test the String() method of the Level type
	if LevelInfo.String() != "INFO" {
		t.Error("LevelInfo.String() should return 'INFO'")
	}
	if LevelError.String() != "ERROR" {
		t.Error("LevelError.String() should return 'ERROR'")
	}
	if LevelFatal.String() != "FATAL" {
		t.Error("LevelFatal.String() should return 'FATAL'")
	}
	if LevelDebug.String() != "DEBUG" {
		t.Error("LevelDebug.String() should return 'DEBUG'")
	}
}
func TestLogger(t *testing.T) {
	// Test the Logger struct
	var b bytes.Buffer
	l := New(&b, LevelInfo)

	// Test the PrintInfo method
	l.PrintInfo("This is an info message", map[string]string{"foo": "bar"})
	var log map[string]interface{}
	json.Unmarshal(b.Bytes(), &log)
	if log["level"] != "INFO" {
		t.Error("Log level should be 'INFO'")
	}
	if log["message"] != "This is an info message" {
		t.Error("Log message should be 'This is an info message'")
	}
	if log["properties"].(map[string]interface{})["foo"] != "bar" {
		t.Error("Log properties should include {'foo': 'bar'}")
	}
	if log["trace"] != nil {
		t.Error("Log should not include a trace for level 'INFO'")
	}

	b.Reset()

	// Test the PrintDebug method
	log = make(map[string]interface{})
	l.PrintDebug("This is a debug message", map[string]string{"foo": "bar"})
	json.Unmarshal(b.Bytes(), &log)
	if log["level"] != "DEBUG" {
		t.Error("Log level should be 'DEBUG'")
	}
	if log["message"] != "This is a debug message" {
		t.Error("Log message should be 'This is a debug message'")
	}
	if log["properties"].(map[string]interface{})["foo"] != "bar" {
		t.Error("Log properties should include {'foo': 'bar'}")
	}
	if log["trace"] != nil {
		fmt.Println(log)
		t.Error("Log should not include a trace for level 'DEBUG'")
	}

	b.Reset()

	// Test the PrintError method
	log = make(map[string]interface{})
	l.PrintError(errors.New("This is an error"), map[string]string{"foo": "bar"})
	json.Unmarshal(b.Bytes(), &log)
	if log["level"] != "ERROR" {
		t.Error("Log level should be 'ERROR'")
	}
	if log["message"] != "This is an error" {
		t.Error("Log message should be 'This is an error'")
	}
	if log["properties"].(map[string]interface{})["foo"] != "bar" {
		t.Error("Log properties should include {'foo': 'bar'}")
	}
	if log["trace"] == nil {
		t.Error("Log should include a trace for level 'ERROR'")
	}

	b.Reset()

	// Test the Write method
	log = make(map[string]interface{})
	l.Write([]byte("This is a message"))
	json.Unmarshal(b.Bytes(), &log)
	if log["level"] != "ERROR" {
		t.Error("Log level should be 'ERROR'")
	}
	if log["message"] != "This is a message" {
		t.Error("Log message should be 'This is a message'")
	}
	if log["properties"] != nil {
		t.Error("Log properties should be empty")
	}
	if log["trace"] == nil {
		t.Error("Log should include a trace for level 'ERROR'")
	}
}
