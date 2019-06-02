package spriteoverflow

import (
	"testing"

	"github.com/alpetkov/nesrs_go/nesrs/testroms"
)

func TestPPU(t *testing.T) {

	data := []struct {
		romPath string
		ppmPath string
		seconds int
	}{
		{"1.Basics.nes", "1.Basics.ppm", 2},
		{"2.Details.nes", "2.Details.ppm", 2},
		{"3.Timing.nes", "3.Timing.ppm", 2},
		{"4.Obscure.nes", "4.Obscure.ppm", 2},
		{"5.Emulator.nes", "5.Emulator.ppm", 2},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../spriteoverflow/"+tt.romPath, "../spriteoverflow/"+tt.ppmPath, tt.seconds, t)
		})
	}
}
