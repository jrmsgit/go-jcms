package app

import (
    "io"
    "bytes"
)

type Response struct {
    err *Error
    buf bytes.Buffer
    body io.Writer
    size int
}

func newResponse () *Response {
    r := &Response{}
    r.body = io.MultiWriter (&r.buf)
    r.size = 0
    return r
}

func (r *Response) IsError () bool {
    return r.err != nil
}

func (r *Response) Error () *Error {
    return r.err
}

func (r *Response) Write (s string) error {
    n, err := io.WriteString (r.body, s)
    if err != nil {
        r.size += n
    }
    return err
}

func (r *Response) Body () []byte {
    return r.buf.Bytes ()
}