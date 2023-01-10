package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"gopkg.in/yaml.v2"
)

// Config defines the requirement for CLI configuration.
type Config struct {
	Env     string `yaml:"env"`
	Port    int    `yaml:"port"`
	MongoDB struct {
		ConnURI string `yaml:"connURI"`
	} `yaml:"mongodb"`
}

// GetConfig read/parse YAML configuration file.
func GetConfig(fileDir string) (*Config, error) {
	file, err := os.Open(fileDir)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// An Envelope wraps the JSON response.
type Envelope map[string]interface{}

// WriteJSON send responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers to include in the response.
func WriteJSON(w http.ResponseWriter, statusCode int, data Envelope, header http.Header) error {
	resp, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	resp = append(resp, '\n')
	for key, value := range header {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(resp)

	return nil
}

// ReadJSON reads/parses request body. Also handles any possible error
func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Restrict r.Body to 1MB
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)

	//Produces error if any unknown json fields is present
	decoder.DisallowUnknownFields()
	err := decoder.Decode(dst)

	if err != nil {
		// types of expected errors
		var syntaxError *json.SyntaxError
		var unmarshaTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("the body contains badly-formatted JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formatted JSON")
		case errors.As(err, &unmarshaTypeError):
			if unmarshaTypeError.Field != "" {
				return fmt.Errorf("body contains badly-formatted JSON type for the field: %q", unmarshaTypeError.Field)
			}
			return fmt.Errorf("body contains badly-formatted JSON type for the field: %d", unmarshaTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("the body contains unknown field %s", field)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		default:
			return err
		}
	}

	// second call to Decode to ensure the request body is just on
	// JSON value
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain only a single JSON")
	}

	return nil
}

// A Validator defines a custom type for validation.
type Validator struct {
	Errors map[string]string
}

// NewValidator creates an instance of the Validator type.
func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// AddError adds error to map no existent error already.
func (v *Validator) AddError(key, message string) {
	if _, exist := v.Errors[key]; !exist {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map only if a validation check is not 'ok'.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Valid returns true if no error occurred, and vice versa.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// ValidateTokenPlaintext checks that the plaintext token was provided and is exactly 26 bytes long.
func ValidateTokenPlaintext(v *Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "key", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "key", "must be 26 bytes long")
}

func ValidateAuditEvent(v *Validator, eventLog *model.AuditEvent) {
	 v.Check(!eventLog.Timestamp.IsZero(), "timestamp", "must be provided")
	 v.Check(eventLog.Action != "", "action", "must be provided")
	 v.Check(eventLog.Actor.ID != "", "actor.id", "must be provided")
	 v.Check(eventLog.Actor.Type != "", "action.type", "must be provided")
	 v.Check(eventLog.Entity.Type != "", "entity.type", "must be provided")
	 v.Check(net.ParseIP(eventLog.Context.IPAddr) != nil, "context.ip_address", "not a valid IP address")
	 v.Check(eventLog.Context.Location != "", "context.location", "must be provided")
}