package multiplexer

import (
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

func TestMultiplexer(t *testing.T) {
	sampleDuration := 20
	sampleRate := 48000
	channel := 2
	mr, err := NewOpusMultiplexer(sampleDuration, sampleRate, channel)

	log.Println("multiplexer created")
	assert.NoError(t, err)
	assert.NotNil(t, mr)

	ticker := time.NewTicker(time.Duration(mr.sampleDurationMs) * time.Millisecond)
	dataCh := make(chan sampleData)
	signalClose := make(chan struct{}, 1)
	var inputFile1, inputFile2 *os.File
	defer func() {
		if inputFile1 != nil {
			inputFile1.Close()
		}
		if inputFile2 != nil {
			inputFile2.Close()
		}
	}()

	t.Run("add new stream", func(t *testing.T) {
		log.Println("adding new streams")
		err := mr.AddStream("1", sampleRate, sampleDuration, channel)
		assert.NoError(t, err)
		err = mr.AddStream("2", sampleRate, sampleDuration, channel)
		assert.NoError(t, err)
	})

	t.Run("push raw opus data", func(t *testing.T) {
		reader1, file1, err := newOggReader("testdata/1.ogg")
		assert.NoError(t, err)
		assert.NotNil(t, file1)
		assert.NotNil(t, reader1)
		inputFile1 = file1

		reader2, file2, err := newOggReader("testdata/2.ogg")
		assert.NoError(t, err)
		assert.NotNil(t, file2)
		assert.NotNil(t, reader2)
		inputFile1 = file2

		closeCh := make(chan struct{}, 2)
		go readOpusPayloadFromOgg("1", reader1, dataCh, closeCh)
		go readOpusPayloadFromOgg("2", reader2, dataCh, closeCh)
		go func() {
			log.Println("start processing raw opus encoded data")
			counter := 2
			for {
				select {
				case data := <-dataCh:
					err := mr.Process(data.data, data.id)
					assert.NoError(t, err)
				case <-closeCh:
					counter -= 1
					if counter == 0 {
						time.Sleep(10 * time.Millisecond)
						ticker.Stop()
						close(signalClose)
						return
					}
				}
			}
		}()
	})

	t.Run("setup output buffer", func(t *testing.T) {
		log.Println("starting ticker for reading decoded samples")
		// Create the output OGG file
		outputFile, err := os.Create("test_output.pcm")
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer outputFile.Close()
		for {
			select {
			case <-ticker.C:
				data, err := mr.ReadOpusBytes()
				assert.NoError(t, err)
				n, err := outputFile.Write(data)
				assert.NoError(t, err, "failed to write opus data to file: %w", err)
				log.Printf("saved: %v bytes", n)
			case <-signalClose:
				return
			}
		}
	})

	t.Run("verify the output", func(t *testing.T) {
		actual, err := os.Open("test_output.pcm")
		assert.NoError(t, err)
		defer actual.Close()
		expected, err := os.Open("testdata/expected.pcm")
		assert.NoError(t, err)
		defer expected.Close()

		for j := 0; j < 20; j++ {
			buf1 := make([]byte, 1024)
			n, err := actual.Read(buf1)
			assert.NoError(t, err)
			buf1 = buf1[:n]

			buf2 := make([]byte, 1024)
			n1, err := expected.Read(buf2)
			assert.NoError(t, err)
			buf2 = buf2[:n1]

			assert.Equal(t, n1, n)

			for i := 0; i < n1; i++ {
				if buf1[i] != buf2[i] {
					t.Logf("actual: %v, expected: %v", buf1[i], buf2[i])
					t.FailNow()
				}
			}
		}
	})
}
