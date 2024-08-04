package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

func main1() {
	// Open input PCM files
	inputFile1, err := os.Open("1.pcm")
	if err != nil {
		log.Fatalf("Error opening input file 1: %v", err)
	}
	defer inputFile1.Close()

	inputFile2, err := os.Open("2.pcm")
	if err != nil {
		log.Fatalf("Error opening input file 2: %v", err)
	}
	defer inputFile2.Close()

	// Create output PCM file
	outputFile, err := os.Create("output.pcm")
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outputFile.Close()

	// Buffer to hold PCM samples
	var sample1, sample2 [2]int16
	var mixedSample [2]int16

	// Mix the PCM data
	for {
		// Read samples from both input files
		err1 := binary.Read(inputFile1, binary.LittleEndian, &sample1)
		err2 := binary.Read(inputFile2, binary.LittleEndian, &sample2)

		if err1 == io.EOF || err2 == io.EOF {
			break
		}
		if err1 != nil {
			log.Fatalf("Error reading from input file 1: %v", err1)
		}
		if err2 != nil {
			log.Fatalf("Error reading from input file 2: %v", err2)
		}

		// Mix the samples by averaging them
		mixedSample[0] = (sample1[0]/2 + sample2[0]/2)
		mixedSample[1] = (sample1[1]/2 + sample2[1]/2)

		// Write the mixed samples to the output file
		err = binary.Write(outputFile, binary.LittleEndian, &mixedSample)
		if err != nil {
			log.Fatalf("Error writing to output file: %v", err)
		}
	}

	fmt.Println("Audio multiplexing complete!")
}
