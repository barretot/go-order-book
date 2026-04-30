package apperrors

type ValidationError struct {
	Status string
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Reason
}
