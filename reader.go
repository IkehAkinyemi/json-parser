package jsonparser

import "io"

type byteReader struct {
	data []byte
	offset int
	r io.Reader
	err error
}

func (b *byteReader) window() []byte {
	return b.data[b.offset:]
}

func (b *byteReader) release(n int) {
	b.offset += n
}

func (b *byteReader) compact() {
	copy(b.data, b.data[b.offset:])
	b.offset = 0
}

func (b *byteReader) grow() {
	buf := make([]byte, max(cap(b.data)*2, newBufferSize))
	copy(buf, b.data[b.offset:])
	b.data = buf
	b.offset = 0
}

const (
	newBufferSize = 4096
	minReadSize = newBufferSize >> 2
)

func (b *byteReader) extend() int {
	if b.err != nil {
		return 0
	}

	remaining := len(b.data) - b.offset
	if remaining == 0 {
		b.data = b.data[:0]
		b.offset = 0
	}

	if cap(b.data)-len(b.data) >= minReadSize {
	} else if cap(b.data)-remaining >= minReadSize {
		b.compact()
	} else {
		b.grow()
	}

	remaining += b.offset
	n, err := b.r.Read(b.data[remaining:cap(b.data)])
	b.data = b.data[:remaining+n]
	b.err = err
	return n
}