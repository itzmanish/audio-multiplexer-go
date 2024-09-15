package avmuxer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDecodingOpusStream(t *testing.T) {
	stream, err := NewDecodingOpusStream("stream1", 48000, 20, 2)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	assert.Equal(t, 48000, stream.SampleRate())
	assert.Equal(t, 2, stream.ChannelCount())
	assert.Equal(t, 20, stream.SampleDurationMs())
	assert.Equal(t, 960, stream.SampleCount()) // sampleDuration * sampleRate / 1000
}

func TestNewEncodingOpusStream(t *testing.T) {
	stream, err := NewEncodingOpusStream("stream1", 48000, 20, 2)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	assert.Equal(t, 48000, stream.SampleRate())
	assert.Equal(t, 2, stream.ChannelCount())
	assert.Equal(t, 20, stream.SampleDurationMs())
	assert.Equal(t, 960, stream.SampleCount()) // sampleDuration * sampleRate / 1000
}

func TestDecodingStream_Write(t *testing.T) {
	stream, err := NewDecodingOpusStream("stream1", 48000, 20, 2)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	encodedData := make([]byte, 960)
	n, err := stream.Write(encodedData)
	assert.NoError(t, err)
	assert.Equal(t, len(encodedData), n)
}

func TestEncodingStream_WritePCM(t *testing.T) {
	stream, err := NewEncodingOpusStream("stream1", 48000, 20, 2)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	pcmData := make([]int16, 960) // Example PCM data
	n, err := stream.WritePCM(pcmData)
	assert.NoError(t, err)
	assert.Equal(t, 960, n)
}

func TestDecodingStream_ReadPCM(t *testing.T) {
	stream, err := NewDecodingOpusStream("stream1", 48000, 20, 2)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	dst := make([]int16, 960)
	n, err := stream.ReadPCM(dst)
	assert.NoError(t, err)
	assert.Equal(t, 960, n)
}

func TestEncodingStream_Read(t *testing.T) {
	stream, err := NewEncodingOpusStream("stream1", 48000, 20, 2)
	assert.NoError(t, err)
	assert.NotNil(t, stream)

	dst := make([]byte, 4096)
	n, err := stream.Read(dst)
	assert.NoError(t, err)
	assert.True(t, n > 0)
}
