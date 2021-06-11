package deploy

type ErrNoResults struct {
	Message string
}

func (e *ErrNoResults) Error() string {
	return e.Message
}
