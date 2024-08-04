# Audio Multiplexer

The `multiplexer` package provides functionality to manage multiple Opus audio streams, decode Opus data to PCM, interleave and multiplex PCM data, and re-encode the multiplexed PCM data back to Opus format. This is useful for real-time audio mixing, VoIP applications, and multi-participant audio conferencing systems.

## Features

- **Manage Multiple Audio Streams**: Add and handle multiple Opus audio streams, each with its own decoder and buffer.
- **Decode Opus Data**: Convert Opus-encoded data to PCM (Pulse Code Modulation) format.
- **Multiplex Audio Streams**: Interleave and combine PCM data from multiple streams into a single output.
- **Encode PCM to Opus**: Convert the combined PCM data back into Opus-encoded format for efficient transmission or storage.

# Usage

## Creating an OpusMultiplexer

```go
sampleDuration := 20  // Sample duration in milliseconds
sampleRate := 48000   // Sample rate in Hz
channels := 2         // Number of audio channels

multiplexer, err := NewOpusMultiplexer(sampleDuration, sampleRate, channels)
if err != nil {
    log.Fatalf("Failed to create multiplexer: %v", err)
}
```

## Adding a Stream

```go
streamID := "stream1"
clockRate := 48000  // Clock rate in Hz

err = multiplexer.AddStream(streamID, clockRate, sampleDuration, channels)
if err != nil {
    log.Fatalf("Failed to add stream: %v", err)
}
```

## Processing Opus Data

```go
opusData := ... // Your Opus-encoded data

err = multiplexer.Process(opusData, streamID)
if err != nil {
    log.Fatalf("Failed to process Opus data: %v", err)
}
```

## Reading PCM Data

```go
pcmData := multiplexer.ReadPCM16()
if len(pcmData) == 0 {
    log.Println("No PCM data available")
}
```

## Reading Opus Data

```go
encodedData, err := multiplexer.ReadOpusBytes()
if err != nil {
    log.Fatalf("Failed to read Opus bytes: %v", err)
}
if len(encodedData) == 0 {
    log.Println("No Opus data available")
}
```

# Testing

A comprehensive test suite is provided to ensure the end-to-end functionality of the multiplexer, including decoding, multiplexing, and re-encoding audio data. To run the tests, use the following command:

```sh
go test ./...
```

# License

This project is licensed under the MIT License. See the LICENSE file for details.

# Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.
