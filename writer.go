// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"net"
)

type Writer interface {
	Close() error
	Write([]byte) (int, error)
	WriteMessage(*Message) error
	WriteRaw([]byte) error
}

// Writer implements io.Writer and is used to send both discrete
// messages to a graylog2 server, or data from a stream-oriented
// interface (like the functions in log).
type GelfWriter struct {
	addr     string
	conn     net.Conn
	hostname string
	Facility string // defaults to current process name
	proto    string
}

// Close connection and interrupt blocked Read or Write operations
func (w *GelfWriter) Close() error {
	if w.conn == nil {
		return nil
	}
	return w.conn.Close()
}

func ProcessLog(log []byte) (message []byte, err error) {
	file, line := getCallerIgnoringLogMulti(1)
	m := constructMessage(log, "", "", file, line)

	return ProcessMessage(m)
}

func ProcessMessage(msg *Message) (message []byte, err error) {
	buf := newBuffer()
	defer bufPool.Put(buf)
	messageBytes, err := msg.toBytes(buf)
	if err != nil {
		return nil, err
	}

	return append(messageBytes, 0), nil
}
