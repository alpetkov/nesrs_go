package branchtiming

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
		{"1.Branch_Basics.nes", "1.Branch_Basics.ppm", 2},
		{"2.Backward_Branch.nes", "2.Backward_Branch.ppm", 2},
		{"3.Forward_Branch.nes", "3.Forward_Branch.ppm", 2},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../branchtiming/"+tt.romPath, "../branchtiming/"+tt.ppmPath, tt.seconds, t)
		})
	}
}
