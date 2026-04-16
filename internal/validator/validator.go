package validator

import (
	"regexp"
	"slices"
)

// regular experession for checking email address formats
// taken from https://html.spec.whatwg.org/#valid-e-mail-address
var (
	EmailRX = regexp.MustCompile("/^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/")
)

// validator type with a map of validation errors
type Validator struct {
	Errors map[string]string
}

// a helper to create the Validator instance with an empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if the errors map is empty
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// addError add an entry to the map
// as long as the key doesn't already exist
func (v *Validator) addError(key, message string) {
	if _, exits := v.Errors[key]; !exits {
		v.Errors[key] = message
	}
}

// Check adds an error entry to the map only if a validation check is not okay
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.addError(key, message)
	}
}

// generic function: returns true if a specific  value is in a list of permitted values
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Matches returns True if a string matches the specifed regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// generic function: returns true if all values in a slice are true
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
