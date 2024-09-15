package avmuxer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/zaf/g711"
)

type G711Type int

const (
	G711Type_Alaw G711Type = iota + 1
	G711Type_Ulaw
)

// commonly known as PCM-A and PCM-U
type G711Stream struct {
	id string

	decoder     *g711.Decoder
	inputBuffer *bytes.Buffer
}

func NewG711Stream(id string, stype G711Type) (Stream, error) {
	buf := bytes.NewBuffer([]byte{})
	var decoder *g711.Decoder
	var err error
	switch stype {
	case G711Type_Alaw:
		decoder, err = g711.NewAlawDecoder(buf)
	case G711Type_Ulaw:
		decoder, err = g711.NewUlawDecoder(buf)
	default:
		return nil, fmt.Errorf("unknown g711 stream type: %v", stype)
	}
	if err != nil {
		return nil, err
	}
	return &G711Stream{
		id:          id,
		decoder:     decoder,
		inputBuffer: buf,
	}, nil
}

func (gs *G711Stream) Write(pkt []byte) (int, error) {
	return gs.inputBuffer.Write(pkt)
}

func (gs *G711Stream) Read(buf []byte) (int, error) {
	return gs.decoder.Read(buf)
}

func (gs *G711Stream) ReadPCM(dst []int16) (int, error) {
	bb := make([]byte, len(dst)*2)
	n, err := gs.Read(bb)
	if err != nil {
		return 0, err
	}
	pcm := byteSliceToInt16(bb[:n])
	if len(dst) < len(pcm) {
		return 0, io.ErrShortBuffer
	}
	n = copy(dst[:len(pcm)], pcm)
	return n, nil
}

func (gs *G711Stream) WritePCM(data []int16) (int, error) {
	return 0, errors.New("g711 stream doesn't support write pcm")
}
