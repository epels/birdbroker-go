package mock

import "time"

type ConsumerConn struct {
	BuryFunc    func(id uint64, pri uint32) error
	DeleteFunc  func(id uint64) error
	ReleaseFunc func(id uint64, pri uint32, delay time.Duration) error
	ReserveFunc func(timeout time.Duration) (id uint64, body []byte, err error)
}

func (c *ConsumerConn) Bury(id uint64, pri uint32) error {
	return c.BuryFunc(id, pri)
}

func (c *ConsumerConn) Delete(id uint64) error {
	return c.DeleteFunc(id)
}

func (c *ConsumerConn) Release(id uint64, pri uint32, delay time.Duration) error {
	return c.ReleaseFunc(id, pri, delay)
}

func (c *ConsumerConn) Reserve(timeout time.Duration) (id uint64, body []byte, err error) {
	return c.ReserveFunc(timeout)
}
