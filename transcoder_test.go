package avmuxer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStream is a mock implementation of the Stream interface
type MockStream struct {
	mock.Mock
}

func (ms *MockStream) ReadPCM(pcm []int16) (int, error) {
	args := ms.Called(pcm)
	return args.Int(0), args.Error(1)
}

func (ms *MockStream) Write([]byte) (int, error) {
	return 0, errors.New("Write not supported in this mock")
}

func (ms *MockStream) Read([]byte) (int, error) {
	return 0, errors.New("Read not supported in this mock")
}

func (ms *MockStream) WritePCM([]int16) (int, error) {
	return 0, errors.New("WritePCM not supported in this mock")
}

// MockEncoder is a mock implementation of the Encoder interface
type MockEncoder struct {
	mock.Mock
}

func (me *MockEncoder) Encode(src []int16, dst []byte) (int, error) {
	args := me.Called(src, dst)
	return args.Int(0), args.Error(1)
}

func (me *MockEncoder) SampleSize() int {
	args := me.Called()
	return args.Int(0)
}

func (me *MockEncoder) ChannelCount() int {
	args := me.Called()
	return args.Int(0)
}

func TestTranscoder_AddSource(t *testing.T) {
	transcoder := NewTranscoder()

	mockStream := new(MockStream)
	err := transcoder.AddSource(mockStream)
	assert.NoError(t, err, "Expected no error when adding the first source")

	err = transcoder.AddSource(mockStream)
	assert.EqualError(t, err, "source is already present", "Expected an error when adding a second source")
}

func TestTranscoder_AddEncoder(t *testing.T) {
	transcoder := NewTranscoder()

	mockEncoder := new(MockEncoder)
	err := transcoder.AddEncoder(mockEncoder)
	assert.NoError(t, err, "Expected no error when adding the first encoder")

	err = transcoder.AddEncoder(mockEncoder)
	assert.EqualError(t, err, "encoder is already present", "Expected an error when adding a second encoder")
}

func TestTranscoder_Read(t *testing.T) {
	transcoder := NewTranscoder()

	mockStream := new(MockStream)
	mockEncoder := new(MockEncoder)

	// Set up the mock responses
	mockStream.On("ReadPCM", mock.Anything).Return(160, nil)
	mockEncoder.On("SampleSize").Return(160)
	mockEncoder.On("Encode", mock.Anything, mock.Anything).Return(20, nil)

	// Test reading without binding input stream
	dst := make([]byte, 100)
	_, err := transcoder.Read(dst)
	assert.EqualError(t, err, "input stream is not binded", "Expected an error when reading without an input stream")

	// Bind the input stream and encoder
	_ = transcoder.AddSource(mockStream)
	_ = transcoder.AddEncoder(mockEncoder)

	// Test successful read
	n, err := transcoder.Read(dst)
	assert.NoError(t, err, "Expected no error when reading with valid input stream and encoder")
	assert.Equal(t, 20, n, "Expected 20 bytes to be read")

	// Validate that the mock methods were called
	mockStream.AssertCalled(t, "ReadPCM", mock.Anything)
	mockEncoder.AssertCalled(t, "Encode", mock.Anything, mock.Anything)
}

func TestTranscoder_ReadPCM(t *testing.T) {
	transcoder := NewTranscoder()

	mockStream := new(MockStream)
	_ = transcoder.AddSource(mockStream)

	// Set up the mock response
	mockStream.On("ReadPCM", mock.Anything).Return(160, nil)

	// Test ReadPCM
	pcm := make([]int16, 160)
	n, err := transcoder.ReadPCM(pcm)
	assert.NoError(t, err, "Expected no error when reading PCM data")
	assert.Equal(t, 160, n, "Expected 160 samples to be read")

	// Validate that the mock method was called
	mockStream.AssertCalled(t, "ReadPCM", mock.Anything)
}

func TestTranscoder_WritePCM(t *testing.T) {
	transcoder := NewTranscoder()

	_, err := transcoder.WritePCM(nil)
	assert.EqualError(t, err, "transcoder stream doesn't support write method", "Expected an error when trying to write PCM data to transcoder")
}
