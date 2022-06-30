package main

import (
	"bytes"
	"fmt"
	net_link "github.com/obgnail/net-link"
	"github.com/obgnail/net-link/codec"
	"github.com/obgnail/net-link/demo"
	"github.com/pingcap/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
)

func main() {
	byteServer()
	//jsonServer()
}

func byteServer() {
	log.Info("---- server start ----")
	err := net_link.ListenAndAccept("127.0.0.1:10086", codec.Byte(), net_link.HandlerFunc(
		func(s *net_link.Session) error {
			receiver := make([]byte, 9)
			if err := s.Read(receiver); err != nil {
				return errors.Trace(err)
			}
			if !bytes.Equal(receiver, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
				panic("read err")
			}
			tcpConn := s.HiJack()
			fmt.Println("remote addr:", tcpConn.(net.Conn).RemoteAddr())
			if _, err := tcpConn.Write([]byte{2, 3, 4}); err != nil {
				return errors.Trace(err)
			}
			return nil
		},
	))
	if err != nil {
		panic(err)
	}
}

func jsonServer() {
	msg := &demo.Student{
		Name: "foobar",
		Age:  999,
	}

	log.Info("---- server start ----")
	err := net_link.ListenAndAccept("127.0.0.1:10086", codec.Json(), net_link.HandlerFunc(
		func(s *net_link.Session) error {
			receiver := new(demo.Student)
			if err := s.Read(receiver); err != nil {
				return errors.Trace(err)
			}
			if !reflect.DeepEqual(receiver, msg) {
				panic("read err")
			}
			if err := s.Write(receiver); err != nil {
				return errors.Trace(err)
			}
			log.Info("--- ok ---")
			return nil
		},
	))
	if err != nil {
		panic(err)
	}
}
