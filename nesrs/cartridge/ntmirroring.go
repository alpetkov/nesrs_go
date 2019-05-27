package cartridge

const (
	ntIndexA = iota
	ntIndexB
	ntIndexC
	ntIndexD
)

const (
	ntMirroringHorizontal = iota
	ntMirroringVertical
	ntMirroringOneScreenA
	ntMirroringOneScreenB
	ntMirroringFourScreen
)

func getNameTableOffset(address int) int {
	// [0x2000, 0x23FF] -> address - 0x2000, [0x2400, 0x27FF] -> address - 0x2400
	// [0x2800, 0x2BFF] -> address - 0x2800, [0x2C00, 0x2FFF] -> address - 0x2C00
	return (address & 0x03FF)
}

func getNameTableIndex(ppuAddress int, ntMirroringType int) int {

	switch ntMirroringType {
	case ntMirroringHorizontal:

		// Horizontal Mirroring:
		// [0x2000, 0x23FF] -> NTA, [0x2400, 0x27FF] -> NTA
		// [0x2800, 0x2BFF] -> NTB, [0x2C00, 0x2FFF] -> NTB
		if (ppuAddress & 0x0800) == 0 {
			return ntIndexA
		}

		return ntIndexB

	case ntMirroringVertical:
		// Vertical Mirroring:
		// [0x2000, 0x23FF] -> NTA, [0x2400, 0x27FF] -> NTB
		// [0x2800, 0x2BFF] -> NTA, [0x2C00, 0x2FFF] -> NTB
		if (ppuAddress & 0x0400) == 0 {
			return ntIndexA
		}

		return ntIndexB

	case ntMirroringOneScreenA:
		// One Screen Mirroring: All address points to the same data.
		// [0x2000, 0x23FF] -> NTA, [0x2400, 0x27FF] -> NTA
		// [0x2800, 0x2BFF] -> NTA, [0x2C00, 0x2FFF] -> NTA
		// Enabled by a mapper usually.
		return ntIndexA

	case ntMirroringOneScreenB:
		// One Screen Mirroring: All address points to the same data.
		// [0x2000, 0x23FF] -> NTB, [0x2400, 0x27FF] -> NTB
		// [0x2800, 0x2BFF] -> NTB, [0x2C00, 0x2FFF] -> NTB
		// Enabled by a mapper usually.
		return ntIndexB

	case ntMirroringFourScreen:
		// 4 screen Mirroring: Each addresses have their own memory space.
		// Enable by a mapper usually.
		// [0x2000, 0x23FF] -> NTA, [0x2400, 0x27FF] -> NTB
		// [0x2800, 0x2BFF] -> NTC, [0x2C00, 0x2FFF] -> NTD
		bucket := (ppuAddress & 0x0C00)

		switch bucket {
		case 0x0000:
			return ntIndexA

		case 0x0300:
			return ntIndexB

		case 0x0800:
			return ntIndexC

		default:
			return ntIndexD
		}
	}

	// Keep compiler silent
	return ntIndexA
}
