package routes

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SchemaVersion identifies the current output schema version.
const SchemaVersion = "v1"

// SchemaConfig controls schema validation behaviour.
type SchemaConfig struct {
	// StrictMode causes Validate to return an error on unknown fields.
	StrictMode bool
	// RequiredFields lists field names that must be non-empty in every Route.
	RequiredFields []string
}

// DefaultSchemaConfig returns sensible defaults.
func DefaultSchemaConfig() SchemaConfig {
	return SchemaConfig{
		StrictMode:     false,
		RequiredFields: []string{"destination", "gateway", "iface"},
	}
}

// SchemaValidator validates Route entries against a SchemaConfig.
type SchemaValidator struct {
	cfg SchemaConfig
}

// NewSchemaValidator creates a SchemaValidator with the given config.
func NewSchemaValidator(cfg SchemaConfig) *SchemaValidator {
	return &SchemaValidator{cfg: cfg}
}

// ValidateRoute checks a single Route against the required fields.
func (v *SchemaValidator) ValidateRoute(r Route) error {
	fields := map[string]string{
		"destination": r.Destination,
		"gateway":     r.Gateway,
		"iface":       r.Iface,
		"protocol":    r.Protocol,
	}
	var missing []string
	for _, req := range v.cfg.RequiredFields {
		if val, ok := fields[req]; !ok || strings.TrimSpace(val) == "" {
			missing = append(missing, req)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("route missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ValidateSnapshot validates every route in a snapshot.
func (v *SchemaValidator) ValidateSnapshot(s Snapshot) []error {
	var errs []error
	for _, r := range s.Routes {
		if err := v.ValidateRoute(r); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// MarshalVersioned wraps a Diff in a versioned envelope and returns JSON.
func MarshalVersioned(d Diff) ([]byte, error) {
	envelope := struct {
		Schema  string `json:"schema"`
		Payload Diff   `json:"payload"`
	}{
		Schema:  SchemaVersion,
		Payload: d,
	}
	return json.Marshal(envelope)
}
