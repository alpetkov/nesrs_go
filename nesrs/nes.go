package nesrs

import (
	"bytes"

	"github.com/alpetkov/nesrs_go/nesrs/cartridge"
	"github.com/alpetkov/nesrs_go/nesrs/cpu"
	"github.com/alpetkov/nesrs_go/nesrs/ppu"
)

const (
	started = iota + 1
	paused
	stopped
)

// NES The.
type NES struct {
	cpu   *cpu.CPU
	ppu   *ppu.PPU
	state int
}

// CPUVBLReceiver .
type CPUVBLReceiver struct {
	cpu *cpu.CPU
}

// ReceiveVBL .
func (vblReceiver *CPUVBLReceiver) ReceiveVBL() {
	vblReceiver.cpu.NMI()
}

// New NES.
func New(rom []byte, videoReceiver ppu.VideoReceiver) *NES {
	// Assemble cpu
	cpuMemory := cpu.NESCPUMemory{}
	cpu := cpu.New(&cpuMemory)

	// Assemble cartridge
	cartridge := cartridge.New(bytes.NewReader(rom))

	// Assemble ppu
	vblReceiver := CPUVBLReceiver{cpu}
	ppu := ppu.New(cartridge, &vblReceiver, videoReceiver)

	// Memory-mapped devices
	cpuMemory.SetCartridge(cartridge)
	cpuMemory.SetPPU(ppu)

	nes := NES{cpu, ppu, stopped}

	return &nes
}

// Start NES.
func (nes *NES) Start() {
	nes.cpu.Init()
	nes.ppu.Init()
	nes.state = started
}

// Reset NES.
func (nes *NES) Reset() {
	nes.cpu.Reset()
	nes.ppu.Reset()
}

// Stop NES.
func (nes *NES) Stop() {
	nes.state = stopped
}

// Run NES.
func (nes *NES) Run() {
	for nes.state == started {
		cpuCycles := nes.cpu.ExecuteOp()

		ppuCycles := cpuCycles * 3
		nes.ppu.ExecuteCycles(ppuCycles)
	}
}
