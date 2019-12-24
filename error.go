package birdbroker

type ClientError struct {
	Reason string
}

func (c ClientError) Error() string {
	return c.Reason
}
