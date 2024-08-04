package multiplexer

import (
	"errors"
	"log"
	"math"
	"sync"

	"gopkg.in/hraban/opus.v2"
)

type OpusMultiplexer struct {
	encoder          *opus.Encoder
	sampleRate       int
	channel          int
	sampleDurationMs int

	closeCh chan struct{}
	inputs  map[string]*Stream
}

type Stream struct {
	id             string
	sampleRate     int
	channel        int
	sampleDuration int
	size           int

	decoder *opus.Decoder
	buffer  FlushableBuffer
}

var int16BufferPool sync.Pool

func NewStream(id string, sampleRate, sampleDuration, channel int) (*Stream, error) {
	sampleSize := channel * sampleDuration * sampleRate / 1000
	decoder, err := opus.NewDecoder(sampleRate, channel)
	if err != nil {
		return nil, err
	}
	return &Stream{
		id:             id,
		sampleRate:     sampleRate,
		sampleDuration: sampleDuration,
		channel:        channel,
		size:           sampleSize,

		decoder: decoder,
		buffer:  NewFlushableBuffer(sampleSize),
	}, err
}

func NewOpusMultiplexer(sampleDuration, sampleRate int, channel int) (*OpusMultiplexer, error) {
	enc, err := opus.NewEncoder(sampleRate, channel, opus.AppAudio)
	if err != nil {
		return nil, err
	}
	err = enc.SetBitrate(int(opus.Fullband))
	if err != nil {
		return nil, err
	}
	// sampleSize := channel * sampleDuration * sampleRate / 1000
	return &OpusMultiplexer{
		sampleDurationMs: sampleDuration,
		sampleRate:       sampleRate,
		channel:          channel,
		closeCh:          make(chan struct{}),
		encoder:          enc,
		inputs:           make(map[string]*Stream),
	}, nil
}

func (mr *OpusMultiplexer) Stop() {
	mr.closeCh <- struct{}{}
	close(mr.closeCh)
}

// size should be calculated as clock_rate*sample_duration_in_ms/1000
func (mr *OpusMultiplexer) AddStream(id string, clockRate, sampleDurationMs, channel int) error {
	if _, ok := mr.inputs[id]; ok {
		return errors.New("stream already exists")
	}
	stream, err := NewStream(id, clockRate, sampleDurationMs, channel)
	if err != nil {
		return err
	}

	mr.inputs[id] = stream
	return nil
}

// data is opus data
func (mr *OpusMultiplexer) Process(data []byte, id string) error {
	stream, ok := mr.inputs[id]
	if !ok {
		return errors.New("stream is not initialized")
	}
	pcm := make([]int16, stream.size)
	n, err := stream.decoder.Decode(data, pcm)
	if err != nil {
		return err
	}
	validData := pcm[:n*stream.channel]
	isEmpty := true
	i := 0
	for _, b := range validData {
		if b != 0 {
			isEmpty = false
			break
		}
		i += 1
	}
	if !isEmpty {
		stream.buffer.Push(validData[i:])
	}
	return nil
}

func (mr *OpusMultiplexer) interleavedMultiplex() []int16 {
	buffs := [][]int16{}
	maxBufSize := 0
	for _, s := range mr.inputs {
		buf := s.buffer.Flush()
		if len(buf) > maxBufSize {
			maxBufSize = len(buf)
		}
		log.Println("size of buf after flush: ", len(buf), cap(buf))
		buffs = append(buffs, buf)
	}
	size := len(buffs)
	out := make([]int16, maxBufSize)
	for i := maxBufSize - 1; i >= 0; i-- {
		var sum int16
		for _, buf := range buffs {
			b := int16(0)
			if len(buf) < maxBufSize {
				diff := maxBufSize - len(buf)
				pos := i - diff
				if pos < 0 {
					continue
				}
				b = buf[i-diff]
			} else {
				b = buf[i]
			}
			sum += b / int16(size)
		}
		out[i] = sum
	}
	for _, buf := range buffs {
		if cap(buf) < math.MaxInt16 {
			buf = buf[:0]
			int16BufferPool.Put(&buf)
		}
	}
	return out
}

func (mr *OpusMultiplexer) ReadPCM16() []int16 {
	return mr.interleavedMultiplex()
}

func (mr *OpusMultiplexer) ReadOpusBytes() ([]byte, error) {
	data := mr.ReadPCM16()
	byteSlice := make([]byte, mr.sampleDurationMs*mr.sampleRate/1000)
	if len(data) == 0 {
		return byteSlice, nil
	}
	n, err := mr.encoder.Encode(data, byteSlice)
	if err != nil {
		return nil, err
	}
	log.Printf("encoded data: size(pcm): %v, n: %v, slicecap: %v", len(data), n, len(byteSlice))
	return byteSlice[:n], nil
}
