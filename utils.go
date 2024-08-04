package multiplexer

// Helper function to convert int16 PCM data to byte slice
func int16ToByteSlice(samples []int16) []byte {
	byteSlice := make([]byte, len(samples)*2)
	for i, sample := range samples {
		byteSlice[i*2] = byte(sample)
		byteSlice[i*2+1] = byte(sample >> 8)
	}
	return byteSlice
}

// Helper function to convert byte slice to int16 PCM data
func byteSliceToInt16(samples []byte) []int16 {
	pcm := make([]int16, len(samples)/2)
	for i := 0; i < len(samples); i += 2 {
		pcm[i/2] = int16(samples[i]) | int16(samples[i+1])<<8
	}
	return pcm
}
