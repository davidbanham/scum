package util

type ClientSafeError struct {
	Message string
}

func (err ClientSafeError) ClientSafeMessage() string {
	return err.Message
}

func (err ClientSafeError) Error() string {
	return err.Message
}
