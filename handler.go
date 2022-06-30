package net_link

type Handler interface {
	handler(s *Session) error
}

type HandlerFunc func(*Session) error

func (f HandlerFunc) handler(s *Session) error {
	return f(s)
}
