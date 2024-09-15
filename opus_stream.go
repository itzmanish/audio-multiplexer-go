package avmuxer

import (
	"errors"
	"io"
	"log"
)

// OpusStream interface defines the methods for both encoding and decoding Opus streams
type OpusStream interface {
	Stream

	EncodingStream
	DecodingStream

	io.ReadWriter
	SampleRate() int
	ChannelCount() int
	SampleDurationMs() int
	SampleCount() int
	ID() string
}

// opusEncodingStream implements OpusStream for encoding
type opusEncodingStream struct {
	id               string
	sampleRate       int
	channel          int
	sampleDurationMs int
	size             int

	sink io.Writer

	encoder *OpusEncoder
}

// opusDecodingStream implements OpusStream for decoding
type opusDecodingStream struct {
	id               string
	sampleRate       int
	channel          int
	sampleDurationMs int
	size             int

	sink io.Writer

	decoder *OpusDecoder
}

// NewDecodingOpusStream creates a new OpusStream for decoding
func NewDecodingOpusStream(id string, sampleRate, sampleDuration, channel int) (OpusStream, error) {
	// Calculate sample size based on duration and rate
	sampleSize := sampleDuration * sampleRate / 1000

	// Create a new OpusDecoder
	dec, err := NewOpusDecoder(sampleRate, channel, sampleSize)
	if err != nil {
		return nil, err
	}

	// Return a new opusDecodingStream
	return &opusDecodingStream{
		id:               id,
		sampleRate:       sampleRate,
		sampleDurationMs: sampleDuration,
		channel:          channel,
		size:             sampleSize,

		decoder: dec.(*OpusDecoder),
	}, err
}

// NewEncodingOpusStream creates a new OpusStream for encoding
func NewEncodingOpusStream(id string, sampleRate, sampleDuration, channel int) (OpusStream, error) {
	// Similar to NewDecodingOpusStream, but for encoding
	// ... existing code ...
	sampleSize := sampleDuration * sampleRate / 1000
	enc, err := NewOpusEncoder(sampleRate, channel, sampleSize)
	if err != nil {
		return nil, err
	}
	return &opusEncodingStream{
		id:               id,
		sampleRate:       sampleRate,
		sampleDurationMs: sampleDuration,
		channel:          channel,
		size:             sampleSize,

		encoder: enc.(*OpusEncoder),
	}, err
}

func (ods *opusDecodingStream) SampleCount() int {
	return ods.size
}
func (ods *opusDecodingStream) ChannelCount() int {
	return ods.channel
}
func (ods *opusDecodingStream) SampleRate() int {
	return ods.sampleRate
}
func (ods *opusDecodingStream) SampleDurationMs() int {
	return ods.sampleDurationMs
}
func (ods *opusDecodingStream) ID() string {
	return ods.id
}

func (ods *opusDecodingStream) Decode(src []byte, dst []int16) (int, error) {
	return ods.decoder.Decode(src, dst)
}

// Write decodes Opus data and writes PCM to the sink
func (ods *opusDecodingStream) Write(data []byte) (int, error) {
	// Decode Opus data to PCM
	pcm := make([]int16, ods.size*ods.channel)
	n, err := ods.Decode(data, pcm)
	if err != nil {
		return 0, err
	}

	log.Printf("samples decoded: %v, os.size: %v, data size: %v\n", n, ods.size, len(data))

	// Write decoded PCM to sink if available
	if ods.sink != nil {
		_, err := ods.sink.Write(Int16ToByteSlice(pcm[:n*ods.channel]))
		if err != nil {
			return 0, err
		}
	}

	// Write PCM data to buffer
	return ods.decoder.buffer.Write(pcm[:n*ods.channel])
}

// ReadPCM reads raw PCM data from the decoding buffer
func (ods *opusDecodingStream) ReadPCM(dst []int16) (int, error) {
	if ods.decoder == nil {
		return 0, errors.New("stream is not decoding supported")
	}
	return ods.decoder.buffer.Read(dst)
}

// Read reads raw PCM data and converts it to a byte array
func (ods *opusDecodingStream) Read(dst []byte) (int, error) {
	if ods.decoder == nil {
		return 0, errors.New("stream is not decoding supported")
	}
	int16Buf := make([]int16, len(dst)/2)
	n, err := ods.decoder.buffer.Read(int16Buf)
	if err != nil {
		return n, err
	}
	bb := Int16ToByteSlice(int16Buf[:n])
	if len(dst) < len(bb) {
		return 0, io.ErrShortBuffer
	}
	n1 := copy(dst[:len(bb)], bb)
	return n1, nil
}

func (ods *opusDecodingStream) Encode([]int16, []byte) (int, error) {
	return 0, errors.New("decoding stream doesn't support encoding")
}

func (ods *opusDecodingStream) WritePCM([]int16) (int, error) {
	return 0, errors.New("decoding stream doesn't support writing pcm")
}

func (ods *opusDecodingStream) Connect(writer io.Writer) error {
	if ods.sink != nil {
		return errors.New("stream already connected to other sink")
	}
	ods.sink = writer
	return nil
}

func (oes *opusEncodingStream) SampleCount() int {
	return oes.size
}
func (oes *opusEncodingStream) ChannelCount() int {
	return oes.channel
}
func (oes *opusEncodingStream) SampleRate() int {
	return oes.sampleRate
}
func (oes *opusEncodingStream) SampleDurationMs() int {
	return oes.sampleDurationMs
}

func (oes *opusEncodingStream) ID() string {
	return oes.id
}

func (*opusEncodingStream) ReadPCM([]int16) (int, error) {
	return 0, errors.New("encoding stream doesn't support reading pcm")
}

func (oes *opusEncodingStream) Write(data []byte) (int, error) {
	pcm := ByteSliceToInt16(data)
	return oes.WritePCM(pcm)
}

func (oes *opusEncodingStream) WritePCM(data []int16) (int, error) {
	byteData := make([]byte, 1024)
	n, err := oes.Encode(data, byteData)
	if err != nil {
		return 0, err
	}
	if oes.sink != nil {
		_, err = oes.sink.Write(byteData[:n])
		if err != nil {
			return 0, err
		}
	}
	return oes.encoder.buffer.Write(byteData[:n])
}

// Read reads encoded Opus data from the encoder's buffer
func (oes *opusEncodingStream) Read(dst []byte) (int, error) {
	if oes.encoder == nil {
		return 0, errors.New("encoder is not initialized")
	}

	n, err := oes.encoder.buffer.Read(dst)
	if err != nil {
		if err == ErrEmptyBuffer {
			return 0, io.EOF
		}
		return n, err
	}

	return n, nil
}

func (oes *opusEncodingStream) Encode(src []int16, dst []byte) (int, error) {
	return oes.encoder.Encode(src, dst)
}

func (*opusEncodingStream) Decode([]byte, []int16) (int, error) {
	return 0, errors.New("encoding stream doesn't support decoding")
}

func (oes *opusEncodingStream) Connect(writer io.Writer) error {
	if oes.sink != nil {
		return errors.New("stream already connected to other reader")
	}
	oes.sink = writer
	return nil
}
