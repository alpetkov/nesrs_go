package ppu

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
		{"palette_ram.nes", "palette_ram.ppm", 2},
		{"power_up_palette.nes", "power_up_palette.ppm", 2},
		{"sprite_ram.nes", "sprite_ram.ppm", 2},
		{"vbl_clear_time.nes", "vbl_clear_time.ppm", 2},
		{"vram_access.nes", "vram_access.ppm", 2},
	}

	for _, tt := range data {
		t.Run(tt.romPath, func(t *testing.T) {
			testroms.TestRom("../ppu/"+tt.romPath, "../ppu/"+tt.ppmPath, tt.seconds, t)
		})
	}
}
