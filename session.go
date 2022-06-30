package net_link

import (
	"github.com/juju/errors"
	"github.com/obgnail/net-link/codec"
	log "github.com/sirupsen/logrus"
	"io"
	"sync"
	"sync/atomic"
)

var (
	sessionClosedError  = errors.New("session closed")
	sessionBlockedError = errors.New("session blocked")
)

var globalSessionId uint64

// 提供线程安全的读和写
type Session struct {
	id    uint64
	codec codec.Codec

	receiveMutex sync.Mutex

	writeMutex sync.RWMutex
	writeChan  chan interface{}

	closeFlag  bool
	closeMutex sync.Mutex
	closeOnce  sync.Once
}

func NewSession(rw io.ReadWriter, protocol codec.Protocol, sendChanSize int) (*Session, error) {
	cdc, err := protocol.NewCodec(rw)
	if err != nil {
		return nil, errors.Trace(err)
	}
	session := &Session{
		id:        atomic.AddUint64(&globalSessionId, 1),
		codec:     cdc,
		writeChan: make(chan interface{}, sendChanSize),
		closeFlag: false,
	}
	go session.writeLoop()
	return session, nil
}

func (s *Session) HiJack() io.ReadWriter {
	return s.codec.HiJack()
}

func (s *Session) writeLoop() {
	for {
		select {
		case msg, ok := <-s.writeChan:
			if !ok || s.codec.Write(msg) != nil {
				_ = s.Close()
				return
			}
		}
	}
}

func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) IsClosed() bool {
	return s.closeFlag
}

// thread safety
func (s *Session) Read(msg interface{}) (err error) {
	s.receiveMutex.Lock()
	defer s.receiveMutex.Unlock()

	err = s.codec.Read(msg)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		_ = s.Close()
		return errors.Trace(err)
	}
	return
}

// thread safety
func (s *Session) Write(msg interface{}) error {
	if s.writeChan == nil {
		if s.IsClosed() {
			return sessionClosedError
		}
		s.writeMutex.Lock()
		defer s.writeMutex.Unlock()

		s.writeChan <- msg
		return nil
	}

	s.writeMutex.RLock()
	defer s.writeMutex.RUnlock()

	if s.IsClosed() {
		return sessionClosedError
	}

	select {
	case s.writeChan <- msg:
		return nil
	default:
		_ = s.Close() // close session when block
		return sessionBlockedError
	}
}

func (s *Session) WriteSync(msg interface{}) error {
	s.writeMutex.RLock()
	defer s.writeMutex.RUnlock()

	if s.IsClosed() {
		return sessionClosedError
	}
	if err := s.codec.Write(msg); err != nil {
		_ = s.Close()
		return err
	}
	return nil
}

// 消耗掉剩余的message
func (s *Session) cleanWriteChan() {
	for range s.writeChan {
	}
}

func (s *Session) close() {
	if !s.IsClosed() {
		if s.writeChan != nil {
			s.writeMutex.Lock()
			close(s.writeChan)
			s.cleanWriteChan()
			s.writeMutex.Unlock()
		}
		s.codec.Close()
		s.closeFlag = true
	}
}

func (s *Session) Close() error {
	s.closeMutex.Lock()
	s.closeOnce.Do(s.close)
	s.closeMutex.Unlock()
	log.Debug("---- session was closed ----")
	return nil
}
