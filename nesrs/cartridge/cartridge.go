package cartridge

import (
	"io"
)

// Memory for storing cartridge PRG and CHR ROM/RAM.
type memory struct {
	prgROM                [][]int
	prgRAM                []int
	isPrgRAMBatteryBacked bool
	chrMem                [][]int
	isChrMemRAM           bool
	ntMirroringType       int
}

// Cartridge for NES.
type Cartridge struct {
	memory       *memory
	mapperNumber int
	prgROMMap    [32]int
	chrMemMap    [8]int
}

func byteToInt(b []byte) []int {
	result := make([]int, len(b))

	for i := range b {
		result[i] = int(b[i])
	}

	return result
}

// New Cartridge.
func New(reader io.Reader) *Cartridge {
	// Headers
	headers := make([]byte, 16)
	reader.Read(headers)

	// Read trainer
	if headers[6]&0x04 != 0 {
		trainer := make([]byte, 512)
		reader.Read(trainer)
	}

	// Read PRG ROM
	prgROM := make([][]int, int(headers[4])*16)
	for i := range prgROM {
		bank := make([]byte, 1024)
		reader.Read(bank)
		prgROM[i] = byteToInt(bank)
	}

	// Read CHR ROM
	var chrMem [][]int
	isChrMemRAM := false
	if headers[5] > 0 {
		chrROM := make([][]int, int(headers[5])*8)
		for i := range chrROM {
			bank := make([]byte, 1024)
			reader.Read(bank)
			chrROM[i] = byteToInt(bank)
		}
		chrMem = chrROM
	} else {
		// CHR RAM
		isChrMemRAM = true
		chrMem = make([][]int, 8)
		for i := range chrMem {
			chrMem[i] = make([]int, 1024)
		}
	}

	// PRG RAM
	prgRAMSize := int(headers[8])
	if prgRAMSize == 0 {
		prgRAMSize = 1
	}
	prgRAM := make([]int, prgRAMSize*8*1024)

	// Mirror type
	var ntMirroringType int
	if (headers[6] & 0x8) != 0 {
		ntMirroringType = ntMirroringFourScreen
	} else {
		if (headers[6] & 0x1) == 0 {
			ntMirroringType = ntMirroringHorizontal
		} else {
			ntMirroringType = ntMirroringVertical
		}
	}

	// Mapper number
	isMapperNumberUpperNibbleSupported := true
	for i := 11; i < len(headers); i++ {
		if headers[i] != 0 {
			isMapperNumberUpperNibbleSupported = false
			break
		}
	}
	mapperNumber := int(((headers[6] & 0xF0) >> 4) & 0x0F)
	if isMapperNumberUpperNibbleSupported {
		mapperNumber += int(headers[7] & 0xF0)
	}

	memory := &memory{prgROM, prgRAM, false, chrMem, isChrMemRAM, ntMirroringType}
	cartridge := createCartridge(memory, mapperNumber)

	return cartridge
}

func createCartridge(memory *memory, mapperNumber int) *Cartridge {
	cartridge := Cartridge{memory: memory, mapperNumber: mapperNumber}
	for i := 0; i < 32; i++ {
		cartridge.prgROMMap[i] = i & (cap(memory.prgROM) - 1)
	}

	for i := 0; i < 8; i++ {
		cartridge.chrMemMap[i] = i & (cap(memory.chrMem) - 1)
	}

	return &cartridge
}

// ReadPrgMemory from Cartridge.
func (cartridge *Cartridge) ReadPrgMemory(cpuAddress int) int {
	page := (cpuAddress & 0xF000)

	if page == 0x4000 || page == 0x5000 {
		// Expansion ROM
		return cartridge.readExpansionRom(cpuAddress)

	} else if page == 0x6000 || page == 0x7000 {
		// RAM
		return cartridge.memory.prgRAM[cpuAddress&0x1FFF] // 8KB

	} else {
		// ROM
		return cartridge.memory.prgROM[cartridge.prgROMMap[(cpuAddress&0x7FFF)>>10]][(cpuAddress & 0x03FF)]
	}
}

// WritePrgMemory to Cartridge.
func (cartridge *Cartridge) WritePrgMemory(cpuAddress int, value int) {
	page := (cpuAddress & 0xF000)

	if page == 0x6000 || page == 0x7000 {
		// RAM
		cartridge.memory.prgRAM[cpuAddress&0x1FFF] = value
	}
}

// ReadChrMemory from Cartridge.
func (cartridge *Cartridge) ReadChrMemory(ppuAddress int) int {
	if 0x0000 <= ppuAddress && ppuAddress <= 0x1FFF {
		return cartridge.memory.chrMem[cartridge.chrMemMap[(ppuAddress&0x1FFF)>>10]][ppuAddress&0x03FF]
	}

	return 0
}

// WriteChrMemory to Cartridge.
func (cartridge *Cartridge) WriteChrMemory(ppuAddress int, value int) {
	if 0x0000 <= ppuAddress && ppuAddress <= 0x1FFF {
		if cartridge.memory.isChrMemRAM {
			cartridge.memory.chrMem[cartridge.chrMemMap[(ppuAddress&0x1FFF)>>10]][ppuAddress&0x03FF] = value
		}
	}
}

// ReadNameTable from Cartridge.
func (cartridge *Cartridge) ReadNameTable(ppuAddress int, ppuNTRAM [][]int) int {
	ntIndex := getNameTableIndex(ppuAddress, cartridge.memory.ntMirroringType)
	nameTableOffset := getNameTableOffset(ppuAddress)

	switch ntIndex {
	case ntIndexA:
		return ppuNTRAM[0][nameTableOffset]
	case ntIndexB:
		return ppuNTRAM[1][nameTableOffset]
	case ntIndexC:
		return 0
	case ntIndexD:
		return 0
	}

	return 0
}

// WriteNameTable to Cartridge.
func (cartridge *Cartridge) WriteNameTable(ppuAddress int, value int, ppuNTRAM [][]int) {
	ntIndex := getNameTableIndex(ppuAddress, cartridge.memory.ntMirroringType)
	nameTableOffset := getNameTableOffset(ppuAddress)

	switch ntIndex {
	case ntIndexA:
		ppuNTRAM[0][nameTableOffset] = value
	case ntIndexB:
		ppuNTRAM[1][nameTableOffset] = value
	}
}

func (cartridge *Cartridge) readExpansionRom(cpuAddress int) int {
	return 0
}
