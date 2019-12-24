package mock

import "time"

type Conn struct {
	PutFunc func(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error)
}

func (c *Conn) Put(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error) {
	return c.PutFunc(body, pri, delay, ttr)
}
