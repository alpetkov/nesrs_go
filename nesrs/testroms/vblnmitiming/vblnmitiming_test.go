package vblnmitiming

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
		{"1.frame_basics.nes", "1.frame_basics.ppm", 10},
		{"2.vbl_timing.nes", "2.vbl_timing.ppm", 6},
		{"3.even_odd_frames.nes", "3.even_odd_frames.ppm", 5},
		{"4.vbl_clear_timing.nes", "4.vbl_clear_timing.ppm", 5},
		{"5.nmi_suppression.nes", "5.nmi_suppression.ppm", 5},
		{"6.nmi_disable.nes", "6.nmi_disable.ppm", 3},
		{"7.nmi_timing.nes", "7.nmi_timing.ppm", 3},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../vblnmitiming/"+tt.romPath, "../vblnmitiming/"+tt.ppmPath, tt.seconds, t)
			t.Log("")
		})
	}
}
