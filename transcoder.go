package multiplexer

import (
	"errors"
	"io"
)

type Transcoder struct {
	input Stream

	encoder Encoder
	io.Reader
}

func NewTranscoder() *Transcoder {
	return &Transcoder{}
}

func (tc *Transcoder) AddSource(id string, stream Stream) error {
	if tc.input != nil {
		return errors.New("source is already present")
	}
	tc.input = stream
	return nil
}

func (tc *Transcoder) AddEncoder(id string, enc Encoder) error {
	if tc.encoder != nil {
		return errors.New("encoder is already present")
	}
	tc.encoder = enc

	return nil
}

func (tc *Transcoder) Read(dst []byte) (int, error) {
	if tc.input == nil {
		return 0, errors.New("input stream is not binded")
	}
	pcm := make([]int16, tc.encoder.SampleSize())
	n, err := tc.input.ReadPCM(pcm)
	if err != nil {
		return 0, err
	}
	return tc.encoder.Encode(pcm[:n], dst)
}

func (tc *Transcoder) ReadPCM(dst []int16) (int, error) {
	if tc.input == nil {
		return 0, errors.New("input stream is not binded")
	}
	return tc.input.ReadPCM(dst)
}

func (tc *Transcoder) WritePCM([]int16) (int, error) {
	return 0, errors.New("transcoder stream doesn't support write method")
}
