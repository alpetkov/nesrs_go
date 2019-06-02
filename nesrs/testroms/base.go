package testroms

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/alpetkov/nesrs_go/nesrs"
	"github.com/alpetkov/nesrs_go/nesrs/ppu"
)

func TestRom(romPath string, screenshotPath string, numberOfSeconds int, t *testing.T) {
	romBytes, _ := readFile(romPath)

	ppmVideoReceiver := new(ppu.PPMVideoReceiver)

	nes := nesrs.New(romBytes, ppmVideoReceiver)
	nes.Start()

	start := time.Now().Unix()
	end := start

	go nes.Run()

	for end-start < int64(numberOfSeconds) {
		time.Sleep(1 * time.Second)
		end = time.Now().Unix()
	}
	nes.Stop()

	var b bytes.Buffer
	ppmVideoReceiver.Write(&b)
	actualContent := b.Bytes()

	expectedContent, _ := readFile(screenshotPath)

	if !bytes.Equal(actualContent, expectedContent) {
		// Dump actual content so it can be reviewed.
		fmt.Printf("Test failed for file %s\n", romPath)
		file, _ := ioutil.TempFile("", "test*.ppm")
		fmt.Printf("Printing to file %s\n", file.Name())
		ppmVideoReceiver.Write(file)
		file.Close()

		t.Errorf("\nWrong %v\nRight %v", actualContent, expectedContent)
	} else {
		fmt.Printf("Test pass for file %s\n", romPath)
	}
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}
