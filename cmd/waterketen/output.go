package main

import (
	"io"
)

// commandOutput remembers the first terminal write failure so a command cannot
// report success after its output was truncated or discarded.
type commandOutput struct {
	writer io.Writer
	err    error
}

func newCommandOutput(writer io.Writer) *commandOutput {
	return &commandOutput{writer: writer}
}

func (o *commandOutput) Write(data []byte) (int, error) {
	if o.err != nil {
		return 0, o.err
	}
	n, err := o.writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		o.err = err
	}
	return n, err
}

func (o *commandOutput) Err() error { return o.err }
