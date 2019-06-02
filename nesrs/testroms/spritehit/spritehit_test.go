package spritehit

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
		{"01.basics.nes", "01.basics.ppm", 6},
		{"02.alignment.nes", "02.alignment.ppm", 2},
		{"03.corners.nes", "03.corners.ppm", 2},
		{"04.flip.nes", "04.flip.ppm", 2},
		{"05.left_clip.nes", "05.left_clip.ppm", 2},
		{"06.right_edge.nes", "06.right_edge.ppm", 2},
		{"07.screen_bottom.nes", "07.screen_bottom.ppm", 2},
		{"08.double_height.nes", "08.double_height.ppm", 2},
		{"09.timing_basics.nes", "09.timing_basics.ppm", 4},
		{"10.timing_order.nes", "10.timing_order.ppm", 2},
		{"11.edge_timing.nes", "11.edge_timing.ppm", 2},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../spritehit/"+tt.romPath, "../spritehit/"+tt.ppmPath, tt.seconds, t)
		})
	}
}
