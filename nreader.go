package messagestream

import (
	"encoding/binary"
	"io"
)

// Reads n bytes from a reader
type nReader struct {
	reader    io.Reader
	remainder []byte
}

func newNReader(r io.Reader) (n *nReader) {
	return &nReader{
		reader: r,
	}
}

func readMessage(r *nReader) (result []byte, err error) {
	size, err := binary.ReadUvarint(r)
	if err != nil {
		return
	}
	result, err = r.Read(size)
	return
}

func (n *nReader) ReadByte() (byte, error) {
	b, err := n.Read(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (nreader *nReader) takeRemainder(n uint64) (bytes []byte) {
	rCount := uint64(len(nreader.remainder))

	if rCount == 0 {
		return
	}

	if n >= rCount {
		x := nreader.remainder
		nreader.remainder = nil
		return x
	}

	if n < rCount {
		x := nreader.remainder[:n]
		nreader.remainder = nreader.remainder[n:]
		return x
	}

	return
}

func (nreader *nReader) Read(n uint64) (bytes []byte, err error) {
	bytes = nreader.takeRemainder(n)

	total := uint64(len(bytes))

	for total < n {
		buffer := make([]byte, 1024)

		c, e := nreader.reader.Read(buffer)
		count := uint64(c)

		if e != nil {
			return bytes, e
		}

		total += count

		if total > n {
			needed := n - (total - count)
			bytes = append(bytes, buffer[:needed]...)
			nreader.remainder = buffer[needed:count]
			return
		}

		if count > 0 {
			bytes = append(bytes, buffer[0:count]...)
		}
	}

	return
}
