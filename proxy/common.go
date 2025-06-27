package proxy

import "io"

const (
	UDP = iota
	TCP
)

type dataPipe struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}