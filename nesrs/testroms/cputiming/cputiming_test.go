package cputiming

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
		{"cpu_timing_test.nes", "cpu_timing_test.ppm", 24},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../cputiming/"+tt.romPath, "../cputiming/"+tt.ppmPath, tt.seconds, t)
		})
	}
}
