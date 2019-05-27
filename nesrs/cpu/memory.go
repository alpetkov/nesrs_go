package cpu

import (
	"github.com/alpetkov/nesrs_go/nesrs/cartridge"
	"github.com/alpetkov/nesrs_go/nesrs/ppu"
)

// CPUMemory is the CPU addressable memory.
type CPUMemory interface {
	Read(adress int) int
	Write(adress int, value int) int
}

// TestCPUMemory for testing purposes.
type TestCPUMemory struct {
	mem [64 * 1024]int // 64KB addressable memory
}

// Read from TestCPUMemory.
func (memory *TestCPUMemory) Read(address int) int {
	return memory.mem[address]
}

// Write for TestCPUMemory.
func (memory *TestCPUMemory) Write(address int, value int) int {
	memory.mem[address] = value
	return 0
}

// NESCPUMemory for NES.
type NESCPUMemory struct {
	ram       [0x800]int
	cartridge *cartridge.Cartridge
	ppu       *ppu.PPU
}

// SetCartridge .
func (memory *NESCPUMemory) SetCartridge(cartridge *cartridge.Cartridge) {
	memory.cartridge = cartridge
}

// SetPPU .
func (memory *NESCPUMemory) SetPPU(ppu *ppu.PPU) {
	memory.ppu = ppu
}

// Read from NES.
func (memory *NESCPUMemory) Read(address int) int {
	page := address & 0xF000

	if page == 0x0000 || page == 0x1000 {
		// RAM
		return memory.ram[address&0x07FF]

	} else if page == 0x2000 || page == 0x3000 {
		// PPU
		if memory.ppu != nil {
			return memory.ppu.ReadRegister(address & 0x0007)
		}

	} else if page == 0x4000 {
		// I/O Registers or Expansion ROM

		if address == 0x4015 {
			// APU
			// if memory.apu != nil {
			// 	return memory.apu.ReadRegister(address)
			// }

			return 0
		} else if address == 0x4016 {
			// Controller 1
			// if memory.controller1 != nils {
			// 	return memory.controller1.Read()
			// }
			return 0

		} else if address == 0x4017 {
			// Controller 2
			// if memory.controller2 != nil {
			// 	return memory.controller2.Read()
			// }

		} else if address >= 0x4020 {
			// Expansion ROM/Cartridge
			if memory.cartridge != nil {
				return memory.cartridge.ReadPrgMemory(address)
			}
		}

	} else {
		// Cartridge

		if memory.cartridge != nil {
			return memory.cartridge.ReadPrgMemory(address)
		}
	}

	return 0
}

// Write to NES.
func (memory *NESCPUMemory) Write(address int, value int) int {
	page := (address & 0xF000)

	if page == 0x0000 || page == 0x1000 {
		// RAM
		memory.ram[address&0x07FF] = value

	} else if page == 0x2000 || page == 0x3000 {
		// PPU

		if memory.ppu != nil {
			memory.ppu.WriteRegister(address&0x0007, value)
		}

	} else if page == 0x4000 {
		// I/O Registers or Expansion ROM

		if address <= 0x4013 || address == 0x4015 || address == 0x4017 {
			// APU
			// if memory.apu != nil {
			// 	memory.apu.WriteRegister(address, value)
			// }

		} else if address == 0x4014 {
			// DMA
			if memory.ppu != nil {
				memAddress := value << 8
				for i := 0; i <= 0xFF; i++ {
					memValue := memory.Read(memAddress)
					// Writes to 0x2004 which is mapped to ppu's spr ram register
					memory.ppu.WriteRegister(ppu.SpriteRAMIORegID, memValue)
					memAddress++
				}
			}

			//513 cycles 1 for read
			//1 write, final is read
			return 513

		} else if address == 0x4016 {
			// Controller 1
			// if ppu.controller1 != nil {
			// 	ppu.controller1.Write(value & 0x01)
			// }

			// Controller 2
			// if ppu.controller2 != nil {
			// 	ppu.controller2.Write(value & 0x01)
			// }

		} else if address >= 0x4020 {
			// Expansion ROM
			if memory.cartridge != nil {
				memory.cartridge.WritePrgMemory(address, value)
			}
		}

	} else {
		// Cartridge
		if memory.cartridge != nil {
			memory.cartridge.WritePrgMemory(address, value)
		}
	}

	return 0
}
