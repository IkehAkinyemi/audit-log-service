package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
)

func TestWriteJSON(t *testing.T) {
	// Test successful JSON marshalling and writing
	data := Envelope{
		"Data": map[string]string{
			"name": "John",
			"age":  "30",
		},
	}
	header := http.Header{
		"X-Test": []string{"test-value"},
	}

	w := httptest.NewRecorder()
	if err := WriteJSON(w, http.StatusOK, data, header); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("unexpected content-type: %s", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("X-Test") != "test-value" {
		t.Errorf("unexpected header value: %s", w.Header().Get("X-Test"))
	}

	var result = Envelope{"Data": map[string]string{}}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test error during JSON marshalling
	badData := Envelope{
		"Data": make(chan int),
	}

	w = httptest.NewRecorder()
	if err := WriteJSON(w, http.StatusOK, badData, header); err == nil {
		t.Error("expected error but got nil")
	}
}

func TestReadJSON(t *testing.T) {
	// Test successful JSON unmarshalling
	data := Envelope{
		"Data": map[string]string{
			"name": "John",
			"age":  "30",
		},
	}
	jsonData, _ := json.Marshal(data)
	r := httptest.NewRequest("POST", "/", bytes.NewReader(jsonData))
	r.Header.Set("Content-Type", "application/json")

	var result Envelope
	if err := ReadJSON(nil, r, &result); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test error during JSON unmarshalling
	r = httptest.NewRequest("POST", "/", bytes.NewReader([]byte("{")))
	r.Header.Set("Content-Type", "application/json")

	if err := ReadJSON(nil, r, &result); err == nil {
		t.Error("expected error but got nil")
	}

	// Test error when body is too large
	r = httptest.NewRequest("POST", "/", bytes.NewReader(make([]byte, 1048579)))
	r.Header.Set("Content-Type", "application/json")

	if err := ReadJSON(nil, r, &result); err == nil {
		t.Error("expected error but got nil")
	}
}

func TestReadStr(t *testing.T) {
	queryStr := url.Values{}
	queryStr.Add("key", "value")

	// Test with key present in query string
	result := ReadStr(queryStr, "key", "default")
	if result != "value" {
		t.Errorf("Expected 'value', got '%s'", result)
	}

	// Test with key not present in query string
	result = ReadStr(queryStr, "missing", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

func TestParseTime(t *testing.T) {
	queryStr := url.Values{}
	queryStr.Add("key", "2022-01-01T00:00:00Z")

	// Test with valid time string in query string
	result := ParseTime(queryStr, "key")
	expected, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-01-01T00:00:00Z")
	if result != expected {
		t.Errorf("Expected '%v', got '%v'", expected, result)
	}

	// Test with invalid time string in query string
	queryStr.Set("key", "invalid-time")
	result = ParseTime(queryStr, "key")
	if !result.IsZero() {
		t.Errorf("Expected zero time, got '%v'", result)
	}

	// Test with missing key in query string
	result = ParseTime(queryStr, "missing")
	if !result.IsZero() {
		t.Errorf("Expected zero time, got '%v'", result)
	}
}

func TestSortValues(t *testing.T) {
	queryStr := url.Values{}
	queryStr.Add("sort", "field")

	// Test with valid sort field in query string
	sortField, sortDesc := SortValues(queryStr, "sort")
	if sortField != "field" || sortDesc != false {
		t.Errorf("Expected 'field' and false, got '%s' and %v", sortField, sortDesc)
	}

	// Test with valid sort field and sort direction in query string
	queryStr.Set("sort", "-field")
	sortField, sortDesc = SortValues(queryStr, "sort")
	if sortField != "field" || sortDesc != true {
		t.Errorf("Expected 'field' and true, got '%s' and %v", sortField, sortDesc)
	}

	// Test with missing key in query string
	sortField, sortDesc = SortValues(queryStr, "missing")
	if sortField != "" || sortDesc != false {
		t.Errorf("Expected '' and false, got '%s' and %v", sortField, sortDesc)
	}
}

func TestReadInt(t *testing.T) {
	queryStr := url.Values{}
	validator := NewValidator()

	// Test with valid int in query string
	queryStr.Add("key", "5")
	result := ReadInt(queryStr, "key", 0, validator)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}

	// Test with invalid int in query string
	queryStr.Set("key", "not-a-number")
	result = ReadInt(queryStr, "key", 0, validator)
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
	if _, ok := validator.Errors["key"]; !ok {
		t.Errorf("Expected error message for key")
	}

	// Test with missing key in query string
	result = ReadInt(queryStr, "missing", 0, validator)
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

func TestValidator(t *testing.T) {
	validator := NewValidator()

	// Test AddError
	validator.AddError("key", "message")
	if _, ok := validator.Errors["key"]; !ok {
		t.Errorf("Expected error message for key")
	}
	validator.AddError("key", "new message")
	if validator.Errors["key"] != "message" {
		t.Errorf("Expected 'message', got %s", validator.Errors["key"])
	}

	// Test Check
	validator = NewValidator()
	validator.Check(false, "key", "message")
	if _, ok := validator.Errors["key"]; !ok {
		t.Errorf("Expected error message for key")
	}
	validator.Check(true, "key-x", "message")
	if _, ok := validator.Errors["key-x"]; ok {
		t.Errorf("Expected no error message for key--")
	}

	// Test Valid
	validator = NewValidator()
	if validator.Valid() != true {
		t.Errorf("Expected true")
	}
	validator.AddError("key", "message")
	if validator.Valid() != false {
		t.Errorf("Expected false")
	}

	// Test ValidateTokenPlaintext
	validator = NewValidator()
	ValidateTokenPlaintext(validator, "")
	if _, ok := validator.Errors["key"]; !ok {
		t.Errorf("Expected error message for key")
	}
	validator = NewValidator()
	ValidateTokenPlaintext(validator, "abcdefghijklmnopqrstuvwxyz")
	if _, ok := validator.Errors["key"]; ok {
		t.Errorf("Expected no error message for key")
	}
	validator = NewValidator()
	ValidateTokenPlaintext(validator, "abcdefghijklmnopqrstuvwxy")
	if _, ok := validator.Errors["key"]; !ok {
		t.Errorf("Expected error message for key")
	}
}

func TestValidateLog(t *testing.T) {
	validator := NewValidator()

	// Test with valid log
	log := &model.Log{
		Timestamp: time.Now(),
		Action:    "create",
		Actor: model.Actor{
			ID:   "1",
			Type: "user",
		},
		Entity: model.Entity{
			Type: "document",
		},
		Context: model.Context{
			IPAddr:   "127.0.0.1",
			Location: "us",
		},
	}
	ValidateLog(validator, log)
	if !validator.Valid() {
		t.Errorf("Expected valid log")
	}

	// Test with missing timestamp
	validator = NewValidator()
	log.Timestamp = time.Time{}
	ValidateLog(validator, log)
	if _, ok := validator.Errors["timestamp"]; !ok {
		t.Errorf("Expected error message for timestamp")
	}

	// Test with missing action
	validator = NewValidator()
	log.Timestamp = time.Now()
	log.Action = ""
	ValidateLog(validator, log)
	if _, ok := validator.Errors["action"]; !ok {
		t.Errorf("Expected error message for action")
	}

	// Test with invalid IP address
	validator = NewValidator()
	log.Timestamp = time.Now()
	log.Action = "create"
	log.Context.IPAddr = "not-an-ip-address"
	ValidateLog(validator, log)
	if _, ok := validator.Errors["context.ip_address"]; !ok {
		t.Errorf("Expected error message for context.ip_address")
	}
}

func TestValidateFilters(t *testing.T) {
	validator := NewValidator()

	// Test with valid filters
	filters := Filters{
		Page:     1,
		PageSize: 50,
	}
	ValidateFilters(validator, filters)
	if !validator.Valid() {
		t.Errorf("Expected valid filters")
	}

	// Test with page less than 1
	validator = NewValidator()
	filters.Page = 0
	ValidateFilters(validator, filters)
	if _, ok := validator.Errors["page"]; !ok {
		t.Errorf("Expected error message for page")
	}

	// Test with page greater than 10 million
	validator = NewValidator()
	filters.Page = 10_000_001
	ValidateFilters(validator, filters)
	if _, ok := validator.Errors["page"]; !ok {
		t.Errorf("Expected error message for page")
	}

	// Test with page size less than 1
	validator = NewValidator()
	filters.Page = 1
	filters.PageSize = 0
	ValidateFilters(validator, filters)
	if _, ok := validator.Errors["page_size"]; !ok {
		t.Errorf("Expected error message for page_size")
	}

	// Test with page size greater than 100
	validator = NewValidator()
	filters.PageSize = 101
	ValidateFilters(validator, filters)
	if _, ok := validator.Errors["page_size"]; !ok {
		t.Errorf("Expected error message for page_size")
	}
}
