package net_link

import (
	"github.com/juju/errors"
	"github.com/obgnail/net-link/codec"
	log "github.com/sirupsen/logrus"
	"net"
)

const (
	DefaultSendChanSize = 2 << 10
)

func ListenAndAccept(addr string, protocol codec.Protocol, handler Handler) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return errors.Trace(err)
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return errors.Trace(err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go func() {
			if err := SyncServeConn(conn, protocol, handler); err != nil {
				log.Error(errors.ErrorStack(err))
			}
		}()
	}
}

func AsyncDial(laddr, raddr string, protocol codec.Protocol, handler Handler) error {
	return newLink().dial(laddr, raddr, protocol, handler).err
}

func SyncDial(laddr, raddr string, protocol codec.Protocol, handler Handler) error {
	lnk := <-newLink().dial(laddr, raddr, protocol, handler).done
	return lnk.err
}

func AsyncServeConn(conn net.Conn, protocol codec.Protocol, handler Handler) error {
	return newLink().serveConn(conn, protocol, handler).err
}

func SyncServeConn(conn net.Conn, protocol codec.Protocol, handler Handler) error {
	lnk := <-newLink().serveConn(conn, protocol, handler).done
	return lnk.err
}

type link struct {
	err  error
	done chan *link
}

func newLink() *link {
	return &link{done: make(chan *link, 1)}
}

func (l *link) saveErr(err error) *link {
	l.err = err
	l.done <- l
	return l
}

func (l *link) dial(laddr, raddr string, protocol codec.Protocol, handler Handler) *link {
	ld, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		return l.saveErr(errors.Trace(err))
	}
	rd, err := net.ResolveTCPAddr("tcp", raddr)
	if err != nil {
		return l.saveErr(errors.Trace(err))
	}
	conn, err := net.DialTCP("tcp", ld, rd)
	if err != nil {
		return l.saveErr(errors.Trace(err))
	}

	go l.serveConn(conn, protocol, handler)
	return l
}

func (l *link) serveConn(conn net.Conn, protocol codec.Protocol, handler Handler) *link {
	if l.done == nil {
		l.done = make(chan *link, 1)
	}

	sess, err := NewSession(conn, protocol, DefaultSendChanSize)
	if err != nil {
		return l.saveErr(errors.Trace(err))
	}

	go func() {
		defer sess.Close()
		if err := handler.handler(sess); err != nil {
			l.saveErr(errors.Trace(err))
			return
		}
		l.done <- l // 一切ok
	}()
	return l
}
