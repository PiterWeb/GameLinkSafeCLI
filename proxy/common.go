package proxy

import (
	"context"
	"io"
)

const (
	UDP = iota
	TCP
)

type readerCtx struct {
	ctx	context.Context
	r   io.Reader
}

type writerCtx struct {
	ctx context.Context
	w io.Writer
}

func (r *readerCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func (w *writerCtx) Write(p []byte) (n int, err error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}

	return w.w.Write(p)
}

// NewReader gets a context-aware io.Reader.
func NewReader(ctx context.Context, r io.Reader) io.Reader {
	return &readerCtx{
		ctx: ctx, 
		r: r,
	}
}

func NewWriter(ctx context.Context, w io.Writer) io.Writer {
	return &writerCtx{
		ctx: ctx,
		w: w,
	}
}