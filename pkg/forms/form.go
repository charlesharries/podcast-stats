package forms

import (
	"net/url"
	"strings"
)

// Form embeds a url.Values (to hold form data) and an errors
// field to hold validation errors.
type Form struct {
	url.Values
	Errors errors
}

// New generates a new form for us to run our validation over.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks if a field is empty; if so, it adds a message
// to the form errors letting the user know that won't fly.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be empty")
		}
	}
}

// Valid checks if there are any errors in the form.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
