// Copyright 2018 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ioutil

import (
	"fmt"
	"io"
)

type Peeker interface {
	Peek(int) ([]byte, error)
}

func ShouldPeekByte(src Peeker) (byte, error) {
	bs, err := ShouldPeek(src, 1)
	if err != nil {
		return 0, err
	}
	return bs[0], nil
}

func ShouldPeek(src Peeker, num int) ([]byte, error) {
	bs, err := src.Peek(num)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(bs) < num {
		return nil, fmt.Errorf("ioutil: unexpected EOF")
	}
	return bs, nil
}

func ShouldReadByte(src io.ByteReader) (byte, error) {
	b, err := src.ReadByte()
	if err != nil {
		if err == io.EOF {
			return 0, fmt.Errorf("ioutil: unexpected EOF")
		}
		return 0, err
	}
	return b, nil
}

func ShouldRead(src io.ByteReader, expected byte) error {
	b, err := src.ReadByte()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("ioutil: unexpected EOF")
		}
		return err
	}
	if b != expected {
		return fmt.Errorf("ioutil: expected %q but %q", expected, b)
	}

	return nil
}

type BackslashNewLineStripper struct {
	r      io.Reader
	buf    []byte
	lastbs bool
	eof    bool
}

func NewBackslashNewLineStripper(r io.Reader) *BackslashNewLineStripper {
	return &BackslashNewLineStripper{
		r: r,
	}
}

func (s *BackslashNewLineStripper) Read(bs []byte) (int, error) {
	var err error
	for len(s.buf) < len(bs) && !s.eof {
		buf := make([]byte, len(bs)-len(s.buf))
		n := 0
		n, err = s.r.Read(buf)
		s.buf = append(s.buf, buf[:n]...)
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			s.eof = true
		}
	}

	dstI := 0
	for dstI < len(bs) && 0 < len(s.buf) {
		switch s.buf[0] {
		case '\\':
			if s.lastbs {
				bs[dstI] = '\\'
				dstI++
			}
		case '\n':
			if !s.lastbs {
				bs[dstI] = s.buf[0]
				dstI++
			}
		default:
			if s.lastbs {
				bs[dstI] = '\\'
				dstI++
				if dstI >= len(bs) {
					s.lastbs = s.buf[0] == '\\'
					s.buf = s.buf[1:]
					break
				}
			}
			bs[dstI] = s.buf[0]
			dstI++
		}
		s.lastbs = s.buf[0] == '\\'
		s.buf = s.buf[1:]

		// Special tretment for the last backslash
		if s.eof && len(s.buf) == 0 && s.lastbs && dstI < len(bs) {
			bs[dstI] = '\\'
			dstI++
		}
	}

	if dstI == 0 && s.eof {
		return 0, io.EOF
	}
	return dstI, nil
}
