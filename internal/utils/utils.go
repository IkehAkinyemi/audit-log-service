package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
)

// Config defines the requirement for server configuration.
type Config struct {
	Env           string
	Port          string
	DBConnURI     string
	AMQP_CONN_URI string
}

// parseConfig retrieves the environment variables.
func ParseConfig() (*Config, error) {
	amqpURI := os.Getenv("AMQP_CONN_URI")
	if amqpURI == "" {
		return nil, fmt.Errorf("no provided value for AMQP_CONN_URI")
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	dbURI := os.Getenv("MONGODB_CONN_URI")
	if dbURI == "" {
		return nil, fmt.Errorf("no provided value for MONGODB_CONN_URI")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	return &Config{
		Env:           env,
		Port:          httpPort,
		DBConnURI:     dbURI,
		AMQP_CONN_URI: amqpURI,
	}, nil
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

// ReadStr reads/parses the query string for a key's value
func ReadStr(queryStr url.Values, key, defaultValue string) string {
	str := queryStr.Get(key)
	if str == "" {
		return defaultValue
	}
	return str
}

// ParseTime parse the query string for timestamps values.
func ParseTime(queryStr url.Values, key string) time.Time {
	value := queryStr.Get(key)
	if value == "" {
		return time.Time{}
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", value)
	if err != nil {
		return time.Time{}
	}

	return parsedTime
}

// SortValues retrieve the sort and sort direction values.
func SortValues(queryStr url.Values, key string) (string, bool) {
	sortField := queryStr.Get(key)
	sortDesc := false

	if sortField == "" {
		return sortField, sortDesc
	}

	if sortField[0] == '-' {
		sortDesc = true
		sortField = sortField[1:]
	}

	return sortField, sortDesc
}

// ReadInt parses integer values provided through the query string
func ReadInt(queryStr url.Values, key string, defaultValue int, v *Validator) int {
	str := queryStr.Get(key)
	if str == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(str)
	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}

	return intValue
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

// ValidateLog validates each log event submitted.
func ValidateLog(v *Validator, log *model.Log) {
	v.Check(!log.Timestamp.IsZero(), "timestamp", "must be provided")
	v.Check(log.Action != "", "action", "must be provided")
	v.Check(log.Actor.ID != "", "actor.id", "must be provided")
	v.Check(log.Actor.Type != "", "action.type", "must be provided")
	v.Check(log.Entity.Type != "", "entity.type", "must be provided")
	v.Check(net.ParseIP(log.Context.IPAddr) != nil, "context.ip_address", "not a valid IP address")
	v.Check(log.Context.Location != "", "context.location", "must be provided")
}

// ValidateFilters validates the query_string for abnormalities
func ValidateFilters(v *Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")

	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
}
