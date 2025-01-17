// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc2

import (
	"context"
	"fmt"
	"os"
)

// Handler is the interface used to hook into the message handling of an rpc
// connection.
type Handler interface {
	// Cancel is invoked for cancelled outgoing requests.
	// It is okay to use the connection to send notifications, but the context will
	// be in the cancelled state, so you must do it with the background context
	// instead.
	// If Cancel returns true all subsequent handlers will be invoked with
	// cancelled set to true, and should not attempt to cancel the message.
	Cancel(ctx context.Context, conn *Conn, id ID, cancelled bool) bool

	// Log is invoked for all messages flowing through a Conn.
	// direction indicates if the message being received or sent
	// id is the message id, if not set it was a notification
	// elapsed is the time between a call being seen and the response, and is
	// negative for anything that is not a response.
	// method is the method name specified in the message
	// payload is the parameters for a call or notification, and the result for a
	// response

	// Request is called near the start of processing any request.
	Request(ctx context.Context, conn *Conn, direction Direction, r *WireRequest) context.Context
	// Response is called near the start of processing any response.
	Response(ctx context.Context, conn *Conn, direction Direction, r *WireResponse) context.Context
	// Done is called when any request is fully processed.
	// For calls, this means the response has also been processed, for notifies
	// this is as soon as the message has been written to the stream.
	// If err is set, it implies the request failed.
	Done(ctx context.Context, err error)
	// Read is called with a count each time some data is read from the stream.
	// The read calls are delayed until after the data has been interpreted so
	// that it can be attributed to a request/response.
	Read(ctx context.Context, bytes int64) context.Context
	// Wrote is called each time some data is written to the stream.
	Wrote(ctx context.Context, bytes int64) context.Context
	// Error is called with errors that cannot be delivered through the normal
	// mechanisms, for instance a failure to process a notify cannot be delivered
	// back to the other party.
	Error(ctx context.Context, err error)
}

// Direction is used to indicate to a logger whether the logged message was being
// sent or received.
type Direction bool

const (
	// Send indicates the message is outgoing.
	Send = Direction(true)
	// Receive indicates the message is incoming.
	Receive = Direction(false)
)

func (d Direction) String() string {
	switch d {
	case Send:
		return "send"
	case Receive:
		return "receive"
	default:
		panic("unreachable")
	}
}

type EmptyHandler struct{}

func (EmptyHandler) Cancel(ctx context.Context, conn *Conn, id ID, cancelled bool) bool {
	return false
}

func (EmptyHandler) Request(ctx context.Context, conn *Conn, direction Direction, r *WireRequest) context.Context {
	return ctx
}

func (EmptyHandler) Response(ctx context.Context, conn *Conn, direction Direction, r *WireResponse) context.Context {
	return ctx
}

func (EmptyHandler) Done(ctx context.Context, err error) {
}

func (EmptyHandler) Read(ctx context.Context, bytes int64) context.Context {
	return ctx
}

func (EmptyHandler) Wrote(ctx context.Context, bytes int64) context.Context {
	return ctx
}

func (EmptyHandler) Error(ctx context.Context, err error) {}

type defaultHandler struct{ EmptyHandler }

// Handler that logs all events to a file. Usually used with os.Stderr or
// os.Stdout
type FileHandler struct {
	File *os.File
}

func (f FileHandler) Cancel(ctx context.Context, conn *Conn, id ID, cancelled bool) bool {
	return false
}

func (f FileHandler) Request(ctx context.Context, conn *Conn, direction Direction, r *WireRequest) context.Context {
	yaml := "jsonrpc: 2.0\n" +
		"method: " + r.Method + "\n" +
		"params: " + string(*r.Params) + "\n" +
		"id: " + fmt.Sprint(r.ID.Number) + "\n"

	fmt.Fprintf(f.File, "conn %p response %s:\n%s\n",
		conn, direction.String(), yaml,
	)

	return ctx
}

func (f FileHandler) Response(ctx context.Context, conn *Conn, direction Direction, r *WireResponse) context.Context {
	yaml := "jsonrpc: 2.0\n" +
		"result: " + string(*r.Result) + "\n" +
		"error: " + fmt.Sprint(r.Error) + "\n" +
		"id: " + fmt.Sprint(r.ID.Number) + "\n"

	fmt.Fprintf(f.File, "conn %p response %s:\n%s\n",
		conn, direction.String(), yaml,
	)

	return ctx
}

func (f FileHandler) Done(ctx context.Context, err error) {}

func (f FileHandler) Read(ctx context.Context, bytes int64) context.Context {
	return ctx
}

func (f FileHandler) Wrote(ctx context.Context, bytes int64) context.Context {
	return ctx
}

func (f FileHandler) Error(ctx context.Context, err error) {}
