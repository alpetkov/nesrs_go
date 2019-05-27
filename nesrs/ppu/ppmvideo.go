package ppu

import (
	"fmt"
	"io"
	"log"
)

// PPMVideoReceiver .
type PPMVideoReceiver struct {
	frame [NESWidth * NESHeight]int
}

// ReceiveFrame .
func (ppmVideoReceiver *PPMVideoReceiver) ReceiveFrame(frame []int) {
	copy(ppmVideoReceiver.frame[:], frame)
}

// Write .
func (ppmVideoReceiver *PPMVideoReceiver) Write(out io.Writer) {
	writePpm(out, NESWidth, NESHeight, ppmVideoReceiver.frame[:])
}

func writePpm(out io.Writer, width int, height int, pixels []int) {
	fmt.Fprintf(out, "P6 %d %d 255\n", width, height)

	for _, rgb := range pixels {
		bytes := []byte{
			byte((rgb >> 16) & 0xFF),
			byte((rgb >> 8) & 0xFF),
			byte(rgb & 0xFF)}
		i, err := out.Write(bytes)
		if err != nil {
			log.Fatal("Error writing to file")
		}
		if i != 3 {
			log.Fatal("Not 3 bytes written")
		}
	}
}
