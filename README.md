# Audio Multiplexer And Transcoder

The `multiplexer` package provides functionality to manage multiple Opus and G711 audio streams, decode Opus/G711 data to PCM, interleave and multiplex PCM data, and re-encode the multiplexed PCM data back to Opus format. This is useful for real-time audio mixing, VoIP applications, and multi-participant audio conferencing systems.

## Features

- **Manage Multiple Audio Streams**: Add and handle multiple Opus and G711 audio streams, each with its own decoder and buffer.
- **Decode Opus/G711 Data**: Convert Opus-encoded data to PCM (Pulse Code Modulation) format.
- **Multiplex Audio Streams**: Interleave and combine PCM data from multiple streams into a single output.
- **Encode PCM to Opus/G711**: Convert the combined PCM data back into Opus/G711-encoded format for efficient transmission or storage.

# Usage

## Creating an Multiplexer

```go

multiplexer, err := NewMultiplexer()
if err != nil {
    log.Fatalf("Failed to create multiplexer: %v", err)
}
```

## Reading PCM data

```go
pcmData := multiplexer.ReadPCM16()
if len(pcmData) == 0 {
    log.Println("No PCM data available")
}
```

## Reading Opus Data

```go
buf:=make([]byte, 2048)
n, err := multiplexer.Read(buf)
if err != nil {
    log.Fatalf("Failed to read Opus bytes: %v", err)
}
if n == 0 {
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
