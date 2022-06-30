package main

import (
	net_link "github.com/obgnail/net-link"
	"github.com/obgnail/net-link/codec"
	"github.com/obgnail/net-link/demo"
	"github.com/pingcap/errors"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

func main() {
	byteClient()
	//jsonClient()
}

func byteClient() {
	log.Info("---- server start ----")
	err := net_link.DialSync("127.0.0.1:0", "127.0.0.1:10086", codec.Byte(), net_link.HandlerFunc(
		func(s *net_link.Session) error {
			msg := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
			if err := s.WriteSync(msg); err != nil {
				return errors.Trace(err)
			}
			newMsg := make([]byte, 3)
			if err := s.Read(newMsg); err != nil {
				return errors.Trace(err)
			}
			if !reflect.DeepEqual(newMsg, []byte{2, 3, 4}) {
				panic("read err")
			}
			return nil
		},
	))
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		panic(err)
	}
	log.Info("---- end ----")
}

func jsonClient() {
	msg := &demo.Student{
		Name: "foobar",
		Age:  999,
	}

	log.Info("---- client start ----")
	err := net_link.DialSync("127.0.0.1:0", "127.0.0.1:10086", codec.Json(), net_link.HandlerFunc(
		func(s *net_link.Session) error {
			if err := s.Write(msg); err != nil {
				return errors.Trace(err)
			}
			receiver := new(demo.Student)
			if err := s.Read(receiver); err != nil {
				return errors.Trace(err)
			}
			if !reflect.DeepEqual(receiver, msg) {
				panic("read err")
			}
			return nil
		},
	))
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		panic(err)
	}
	log.Info("---- end ----")
}
