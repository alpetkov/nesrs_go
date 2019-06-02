package ppu

import "math/bits"

type backgroundTileLatch struct {
	tileIndex            int // 8 bits
	tileDataLow          int // 8 bits
	tileDataHigh         int // 8 bits
	attributePaletteData int // 2 bits
}

type backgroundRenderPipeline struct {
	tileDataLow              int // 16 bits (16 pixels - 2 tiles' row pixels)
	tileDataHigh             int // 16 bits (16 pixels - 2 tiles' row pixels)
	attributePalleteDataLow  int
	attributePalleteDataHigh int
}

type backgroundRenderer struct {
	ctrlReg    *ctrlRegister
	maskReg    *maskRegister
	vramReg    *vramAddressScrollRegister
	vramMemory *vramMemory
	tileLatch  backgroundTileLatch
	pipeline   backgroundRenderPipeline
}

// 0x0 - 0x2000, 0x1 - 0x2400, 0x2 - 0x2800, 0x3 - 0x2C00
var nametableIndexToAddress = [...]int{0x2000, 0x2400, 0x2800, 0x2C00}

// 0x0 | 0x1
// 0x2 | 0x3
var nametableIndexToHorizontalIndex = [...]int{0x1, 0x0, 0x3, 0x2}
var nametableIndexToVerticalndex = [...]int{0x2, 0x3, 0x0, 0x1}

func newBackgroundRenderer(ctrlReg *ctrlRegister, maskReg *maskRegister, vramReg *vramAddressScrollRegister, vramMemory *vramMemory) *backgroundRenderer {
	renderer := backgroundRenderer{ctrlReg: ctrlReg, maskReg: maskReg, vramReg: vramReg, vramMemory: vramMemory}
	return &renderer
}

func (renderer *backgroundRenderer) executeScanlineBackgroundCycle(currentCycle int,
	scanlineOffscreenBuffer []int, shouldRender bool) {

	if 0 <= currentCycle && currentCycle <= 255 {
		// Memory fetch phase 1 through 128
		fetchTilePhaseCycle := currentCycle & 0x07

		if shouldRender {
			// Handle render pipeline feed at the beginning of the phase
			if fetchTilePhaseCycle == 0 {
				renderer.loadPipeline()
			}
			renderer.renderTileData(
				currentCycle,
				scanlineOffscreenBuffer,
				true /*drawPixel*/)
		}

		// Fetch tile (8 cc - do it at once as opposed to 4 x 2cc)
		if fetchTilePhaseCycle == 7 {
			renderer.fetchTileData()
		}
	} else if 256 <= currentCycle && currentCycle <= 319 {
		if currentCycle == 256 {
			renderer.incrementBackgroundFineY()
		}

		// Memory fetch phase 129 through 160
		// Do nothing

	} else if 320 <= currentCycle && currentCycle <= 335 {
		// Memory fetch phase 161 through 168
		fetchTilePhaseCycle := currentCycle & 0x07

		// Handle render pipeline feed at the beginning of the phase
		if fetchTilePhaseCycle == 0 {
			renderer.loadPipeline()
		}
		renderer.renderTileData(
			currentCycle,
			scanlineOffscreenBuffer,
			false /*drawPixel*/)

		// Fetch tile (8 cc - do it at once as opposed to 4 x 2cc)
		if fetchTilePhaseCycle == 7 {
			renderer.fetchTileData()
		}

	} else {
		// Memory fetch phase 169 through 170
		// Do nothing
	}
}

func (renderer *backgroundRenderer) loadPipeline() {
	// Load pipeline from latch.

	// Load in MSB
	renderer.pipeline.tileDataLow =
		(reverseByte(renderer.tileLatch.tileDataLow) << 8) |
			(renderer.pipeline.tileDataLow & 0x00FF)

	// Load in MSB
	renderer.pipeline.tileDataHigh =
		(reverseByte(renderer.tileLatch.tileDataHigh) << 8) |
			(renderer.pipeline.tileDataHigh & 0x00FF)

	// Load in MSB
	attributePalleteDataLow := 0x00
	if (renderer.tileLatch.attributePaletteData & 0x01) != 0 {
		attributePalleteDataLow = 0xFF // replicate 8 times
	}
	renderer.pipeline.attributePalleteDataLow =
		attributePalleteDataLow<<8 | (renderer.pipeline.attributePalleteDataLow & 0x00FF)

	// Load in MSB
	attributePalleteDataHigh := 0x00
	if (renderer.tileLatch.attributePaletteData & 0x02) != 0 {
		attributePalleteDataHigh = 0xFF // replicate 8 times
	}
	renderer.pipeline.attributePalleteDataHigh =
		attributePalleteDataHigh<<8 | (renderer.pipeline.attributePalleteDataHigh & 0xFF)
}

func (renderer *backgroundRenderer) renderTileData(currentCycle int, scanlineOffscreenBuffer []int,
	drawPixel bool) {

	// Handle background pixel rendering. Pixel is rendered every cycle for total of 256 pixels
	if drawPixel {
		if currentCycle > 7 || !renderer.maskReg.isBackgroundClippingEnabled() {
			renderer.renderBackgroundPixel(currentCycle, scanlineOffscreenBuffer)
		}
	}

	// Shift pipeline
	renderer.pipeline.tileDataLow >>= 1
	renderer.pipeline.tileDataHigh >>= 1
	renderer.pipeline.attributePalleteDataLow >>= 1
	renderer.pipeline.attributePalleteDataHigh >>= 1
}

func (renderer *backgroundRenderer) renderBackgroundPixel(currentCycle int, scanlineOffscreenBuffer []int) {
	// Determine palette data
	fineX := renderer.vramReg.bgFineX
	bitPosition := 1 << uint8(fineX)

	tilePaletteDataLowBit := 0
	if (renderer.pipeline.tileDataLow & bitPosition) != 0 {
		tilePaletteDataLowBit = 1
	}
	tilePaletteDataHighBit := 0
	if (renderer.pipeline.tileDataHigh & bitPosition) != 0 {
		tilePaletteDataHighBit = 2
	}
	attributePaletteDataLowBit := 0
	if (renderer.pipeline.attributePalleteDataLow & bitPosition) != 0 {
		attributePaletteDataLowBit = 4
	}
	attributePaletteDataHighBit := 0
	if (renderer.pipeline.attributePalleteDataHigh & bitPosition) != 0 {
		attributePaletteDataHighBit = 8
	}

	paletteIndex := attributePaletteDataHighBit | attributePaletteDataLowBit |
		tilePaletteDataHighBit | tilePaletteDataLowBit

	// Palette mirroring
	if paletteIndex == 0x04 || paletteIndex == 0x08 || paletteIndex == 0x0C {
		paletteIndex = 0x00
	}

	scanlineOffscreenBuffer[currentCycle] = paletteIndex
}

func (renderer *backgroundRenderer) fetchTileData() {
	//
	// Name table read
	//
	nameTableIndex := renderer.vramReg.nameTableIndex()
	nameTableAddress := nametableIndexToAddress[nameTableIndex]
	tileX := renderer.vramReg.backgroundTileX()
	tileY := renderer.vramReg.backgroundTileY()

	tileAddress := nameTableAddress + 32*tileY + tileX
	renderer.tileLatch.tileIndex = renderer.vramMemory.read(tileAddress)

	//
	// Attribute table read
	//
	attributeTableX := tileX / 4
	attributeTableY := tileY / 4
	// 32x30 (960) tiles in nametable. The last 64 (actually 60) bytes are for attribute data.
	// Each attribute byte is for 32x32 pixels (4x4 tiles).
	attributeAddress := nameTableAddress + 960 + 8*attributeTableY + attributeTableX
	attributeByte := renderer.vramMemory.read(attributeAddress)

	attributeFineX := tileX % 4
	attributeFineY := tileY % 4
	if attributeFineY < 2 {
		if attributeFineX < 2 {
			// Square 0 (top left)
			renderer.tileLatch.attributePaletteData = attributeByte & 0x03
		} else {
			// Square 1 (top right)
			renderer.tileLatch.attributePaletteData = (attributeByte & 0x0C) >> 2
		}
	} else {
		if attributeFineX < 2 {
			// Square 2 (bottom left)
			renderer.tileLatch.attributePaletteData = (attributeByte & 0x30) >> 4
		} else {
			// Square 3 (bottom right)
			renderer.tileLatch.attributePaletteData = (attributeByte & 0xC0) >> 6
		}
	}

	// tileX is calculated and stored. Increment for next fetch.
	renderer.incrementBackgroundTileX()

	//
	// Pattern table bitmap #0 read
	//
	fineY := renderer.vramReg.backgroundFineY()
	backgroundPatternTableAddress := renderer.ctrlReg.getBackgroundPatternTableAddress()
	tileDataLowAddress := backgroundPatternTableAddress + renderer.tileLatch.tileIndex*16 + fineY
	renderer.tileLatch.tileDataLow = renderer.vramMemory.read(tileDataLowAddress)

	//
	// Pattern table bitmap #1 read
	//
	tileDataHighAddress := tileDataLowAddress + 8
	renderer.tileLatch.tileDataHigh = renderer.vramMemory.read(tileDataHighAddress)
}

func (renderer *backgroundRenderer) incrementBackgroundTileX() {
	tileX := renderer.vramReg.backgroundTileX()
	nameTableIndex := renderer.vramReg.nameTableIndex()

	tileX++
	if tileX > 31 {
		tileX = 0
		nameTableIndex = nametableIndexToHorizontalIndex[nameTableIndex]
	}

	renderer.vramReg.setBackgroundTileX(tileX)
	renderer.vramReg.setNameTableIndex(nameTableIndex)
}

func (renderer *backgroundRenderer) incrementBackgroundFineY() {
	fineY := renderer.vramReg.backgroundFineY()
	nameTableIndex := renderer.vramReg.nameTableIndex()
	tileY := renderer.vramReg.backgroundTileY()

	fineY++
	if fineY > 7 {
		fineY = 0
		tileY++
		if tileY == 31 {
			tileY = 0
		} else if tileY == 30 {
			tileY = 0
			nameTableIndex = nametableIndexToVerticalndex[nameTableIndex]
		}
	}

	renderer.vramReg.setBackgroundFineY(fineY)
	renderer.vramReg.setBackgroundTileY(tileY)
	renderer.vramReg.setNameTableIndex(nameTableIndex)
}

func reverseByte(b int) int {
	rev := bits.Reverse32(uint32(b))
	rev = (rev >> 24) & 0xFF
	return int(rev)
}
