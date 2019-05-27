package ppu

import "github.com/alpetkov/nesrs_go/nesrs/cartridge"

type vramMemory struct {
	ntVRAM               [][]int // Name table VRAM (A + B)(2Kb) (aka CIRAM)
	backgroundPaletteRAM [16]int // Background Palette RAM (16b)
	spritePaletteRAM     [16]int // Sprite Palette RAM (16b)
	cartridge            *cartridge.Cartridge
}

type spriteMemory struct {
	ram        [0x100]int // Sprite RAM (256b) (64 sprites)
	tempMemory [0x20]int  // Sprite temporary Memory (32b) (8 sprites)
}

func newVRAMMemory(cartridge *cartridge.Cartridge) *vramMemory {
	ntVRAM := make([][]int, 2)
	for i := range ntVRAM {
		ntVRAM[i] = make([]int, 1024)
	}
	vramMemory := vramMemory{ntVRAM: ntVRAM, cartridge: cartridge}

	return &vramMemory
}

func (vramMemory *vramMemory) read(address int) int {
	decodedAddress := decodePPUAddress(address)

	if 0 <= decodedAddress && decodedAddress <= 0x1FFF {
		// CHR ROM/RAM
		return vramMemory.cartridge.ReadChrMemory(decodedAddress)

	} else if 0x2000 <= decodedAddress && decodedAddress < 0x3000 {
		// Name table
		return vramMemory.cartridge.ReadNameTable(decodedAddress, vramMemory.ntVRAM)

	} else if 0x3F00 <= decodedAddress && decodedAddress <= 0x3F0F {
		// Background Palette
		return vramMemory.backgroundPaletteRAM[decodedAddress&0x000F]

	} else if 0x3F10 <= decodedAddress && decodedAddress <= 0x3F1F {
		// Sprite Palette
		return vramMemory.spritePaletteRAM[decodedAddress&0x000F]
	}

	return 0

}

func (vramMemory *vramMemory) write(address int, value int) {
	decodedAddress := decodePPUAddress(address)

	if 0 <= decodedAddress && decodedAddress <= 0x1FFF {
		// CHR ROM/RAM
		vramMemory.cartridge.WriteChrMemory(decodedAddress, value)

	} else if 0x2000 <= decodedAddress && decodedAddress <= 0x2FFF {
		// Name table
		vramMemory.cartridge.WriteNameTable(decodedAddress, value, vramMemory.ntVRAM)

	} else if 0x3F00 <= decodedAddress && decodedAddress <= 0x3F0F {
		// Background Palette
		vramMemory.backgroundPaletteRAM[decodedAddress&0x000F] = value

	} else if 0x3F10 <= decodedAddress && decodedAddress <= 0x3F1F {
		// Sprite Palette
		vramMemory.spritePaletteRAM[decodedAddress&0x000F] = value
	}
}

func decodePPUAddress(address int) int {
	// Size Mirroring
	address = address & 0x3FFF

	// Name tables & palette size mirroring
	if 0x3000 <= address && address <= 0x3EFF {
		// Mirror of Name Tables (0x2000 - 0x2EFF)
		address = 0x2000 | (address & 0x0FFF)

	} else if 0x3F20 <= address && address <= 0x3FFF {
		// Mirror of Background + Sprite Palettes ($3F00-$3F1F)
		address = 0x3F00 | (address % 0x20)
	}

	if 0x3F00 <= address && address <= 0x3F1F {
		// Background + Sprite Palettes Mirroring
		// Addresses $3F10/$3F14/$3F18/$3F1C are mirrors of $3F00/$3F04/$3F08/$3F0C.
		switch address {
		case 0x3F10:
			address = 0x3F00
		case 0x3F14:
			address = 0x3F04
		case 0x3F18:
			address = 0x3F08
		case 0x3F1C:
			address = 0x3F0C
		}
	}

	return address
}

func (spriteMemory *spriteMemory) read(offset int) int {
	return spriteMemory.ram[offset]
}

func (spriteMemory *spriteMemory) write(offset int, value int) {
	spriteMemory.ram[offset] = value
}

func (spriteMemory *spriteMemory) readTemp(address int) int {
	return spriteMemory.tempMemory[address]
}

func (spriteMemory *spriteMemory) writeTemp(address int, value int) {
	spriteMemory.tempMemory[address] = value
}
