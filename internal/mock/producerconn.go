package mock

import "time"

type ProducerConn struct {
	PutFunc func(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error)
}

func (c *ProducerConn) Put(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error) {
	return c.PutFunc(body, pri, delay, ttr)
}
