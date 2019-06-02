package spritehittiming

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
		{"sprite_hit_timing.nes", "sprite_hit_timing.ppm", 4},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../spritehittiming/"+tt.romPath, "../spritehittiming/"+tt.ppmPath, tt.seconds, t)
		})
	}
}
