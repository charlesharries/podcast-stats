package forms

// errors holds all of our validation errors.
type errors map[string][]string

// Add adds an error message to a field.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get retrieves all errors from a given field.
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}

	return es[0]
}