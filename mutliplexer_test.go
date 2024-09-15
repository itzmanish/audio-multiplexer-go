package avmuxer

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pion/webrtc/v4/pkg/media/oggreader"
	"github.com/stretchr/testify/assert"
)

type sampleData struct {
	data []byte
	id   string
}

func newOggReader(inputFile string) (*oggreader.OggReader, *os.File, error) {
	// Open the input OGG file
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Error opening input file: %v", err)
	}
	// Decode the Opus data from the OGG file
	oggReader, _, err := oggreader.NewWith(file)
	return oggReader, file, err
}

func readOpusPayloadFromOgg(id string, oggReader *oggreader.OggReader, ch chan<- sampleData, closeCh chan struct{}) error {
	maxReadUptoIndex := 100
	var lastGranule uint64 = 0
	for {
		opusData, header, err := oggReader.ParseNextPage()
		if err != nil {
			if err == io.EOF {
				closeCh <- struct{}{}
				break
			} else {
				return err
			}
		}
		sampleCount := float64(header.GranulePosition - lastGranule)
		lastGranule = header.GranulePosition
		if sampleCount == 0 {
			continue
		}
		log.Printf("ogg header: %+v, lastGranule: %v", header, lastGranule)
		if (header.GranulePosition / uint64(sampleCount)) == uint64(maxReadUptoIndex) {
			closeCh <- struct{}{}
			break
		}
		ch <- sampleData{data: opusData, id: id}
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

func TestOpusStream_Decode(t *testing.T) {
	// Load an Ogg file with Opus data from the testdata folder
	reader, file, err := newOggReader("testdata/1.ogg")
	assert.NoError(t, err)
	defer file.Close()

	// Initialize the Opus decoder stream
	stream, err := NewDecodingOpusStream("testStream", 48000, 20, 2)
	assert.NoError(t, err)

	ch := make(chan sampleData)
	closeCh := make(chan struct{})
	go readOpusPayloadFromOgg("1", reader, ch, closeCh)
	writeFile, err := os.Create("testdata/1.pcm")
	assert.NoError(t, err)
	defer writeFile.Close()
	for {
		select {
		case data := <-ch:
			// Create a buffer for the decoded PCM data
			pcm := make([]int16, stream.SampleCount()*stream.ChannelCount())

			// Decode the Opus data
			_, err = stream.Write(data.data)
			assert.NoError(t, err)

			_, err = stream.ReadPCM(pcm)
			assert.NoError(t, err)
			// Validate that the decoded PCM data is non-empty
			assert.NotEmpty(t, pcm)
			for _, sample := range pcm {
				err = binary.Write(writeFile, binary.LittleEndian, sample)
				assert.NoError(t, err)
			}
		case <-closeCh:
			return
		}
	}

}

func TestOpusStream_Encode(t *testing.T) {
	// Initialize the Opus encoder stream
	stream, err := NewEncodingOpusStream("testStream", 48000, 20, 2)
	assert.NoError(t, err)

	// Prepare sample PCM data to encode
	pcmByte := make([]byte, stream.SampleCount()*2)
	file, err := os.Open("testdata/1.pcm")
	assert.NoError(t, err)
	defer file.Close()

	i, err := file.Read(pcmByte)
	assert.NoError(t, err)

	pcm := byteSliceToInt16(pcmByte[:i])

	// Create a buffer to hold the encoded Opus data
	encoded := make([]byte, 1024)

	// Encode the PCM data
	n, err := stream.Encode(pcm, encoded)
	assert.NoError(t, err)

	// Validate that the encoded data is non-empty
	assert.NotZero(t, n)

	// Initialize the Opus encoder stream
	stream1, err := NewDecodingOpusStream("testStream1", 48000, 20, 2)
	assert.NoError(t, err)

	pcm1 := make([]int16, stream1.SampleCount())
	n1, err := stream1.Decode(encoded[:n], pcm1)
	assert.NoError(t, err)

	assert.NotEmpty(t, pcm1[:n1])
}

func TestMultiplexer_AddOpusSourceStream(t *testing.T) {
	mux := NewMultiplexer()

	// Initialize an Opus decoding stream
	stream, err := NewDecodingOpusStream("testStream", 48000, 20, 1)
	assert.NoError(t, err)

	// Add the stream to the multiplexer
	err = mux.AddSourceStream("testStream", stream)
	assert.NoError(t, err)

	// Verify that the stream was added
	mux.RLock()
	defer mux.RUnlock()
	assert.Contains(t, mux.sources, "testStream")
}

func TestMultiplexer_ReadPCM16(t *testing.T) {
	mux := NewMultiplexer()

	// Initialize and add an Opus decoding stream to the multiplexer
	stream1, err := NewDecodingOpusStream("1", 48000, 20, 2)
	assert.NoError(t, err)

	stream2, err := NewDecodingOpusStream("2", 48000, 20, 2)
	assert.NoError(t, err)

	err = mux.AddSourceStream("1", stream1)
	assert.NoError(t, err)
	err = mux.AddSourceStream("2", stream2)
	assert.NoError(t, err)

	// Load an Ogg file with Opus data from the testdata folder
	reader1, file1, err := newOggReader("testdata/1.ogg")
	assert.NoError(t, err)
	defer file1.Close()

	// Load an Ogg file with Opus data from the testdata folder
	reader2, file2, err := newOggReader("testdata/2.ogg")
	assert.NoError(t, err)
	defer file2.Close()

	ch := make(chan sampleData)
	closeCh := make(chan struct{}, 2)
	closedCount := 0
	go readOpusPayloadFromOgg("1", reader1, ch, closeCh)
	go readOpusPayloadFromOgg("2", reader2, ch, closeCh)
	writeFile, err := os.Create("testdata/muxed.pcm")
	assert.NoError(t, err)
	defer writeFile.Close()

	go func() {
		for {
			select {
			case data := <-ch:
				if data.id == "1" {
					// Simulate writing the Opus data to the stream
					_, err = stream1.Write(data.data)
				} else {
					_, err = stream2.Write(data.data)
				}
				assert.NoError(t, err)

			case <-closeCh:
				closedCount += 1
				if closedCount == 2 {
					return
				}
			}
		}
	}()
	startTime := time.Now()
	for {
		if startTime.Add(2*time.Second).Compare(time.Now()) < 0 {
			break
		}
		// Decode the Opus data
		pcm := mux.ReadPCM(960 * 2)

		// Validate that the decoded PCM data is non-empty
		assert.NotEmpty(t, pcm)
		for _, sample := range pcm {
			err = binary.Write(writeFile, binary.LittleEndian, sample)
			assert.NoError(t, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func generateTestPCM(sampleCount, repeatCount int) []int16 {
	pcm := make([]int16, sampleCount*repeatCount)
	for i := 0; i < repeatCount; i++ {
		for j := 0; j < sampleCount; j++ {
			pcm[i*sampleCount+j] = int16((j % 100) * 100)
		}
	}
	return pcm
}
