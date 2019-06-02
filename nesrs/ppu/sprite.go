package ppu

type spriteRenderPipeline struct {
	tileDataLow          int // 8 bits (8 pixels - 1 tile row pixels)
	tileDataHigh         int // 8 bits (8 pixels - 1 tile row pixels)
	attributePaletteData int // 2 bits (for 8 pixels)
	isHighPriority       bool
	xPosition            int // 8 bits // X position on the screen
	isSpriteZero         bool
}

type spriteRenderer struct {
	ctrlReg             *ctrlRegister
	maskReg             *maskRegister
	statusReg           *statusRegister
	vramMemory          *vramMemory
	sprMemory           *spriteMemory
	pipelineMemory      [8]spriteRenderPipeline
	isSpriteZeroInRange bool
}

const (
	sprAttrRevertVertically   = 0x80 // bit 7
	sprAttrRevertHorizontally = 0x40 // bit 6
	sprAttrPriority           = 0x20 // bit 5
	sprAttrPalette            = 0x3  // bits 0 & 1
)

func newSpriteRenderer(ctrlReg *ctrlRegister, maskReg *maskRegister, statusReg *statusRegister, vramMemory *vramMemory, spriteMemory *spriteMemory) *spriteRenderer {
	renderer := spriteRenderer{
		ctrlReg:    ctrlReg,
		maskReg:    maskReg,
		statusReg:  statusReg,
		vramMemory: vramMemory,
		sprMemory:  spriteMemory}

	return &renderer
}

func (renderer *spriteRenderer) reset() {
	noSpriteRenderData := spriteRenderPipeline{}
	for i := range renderer.pipelineMemory {
		renderer.pipelineMemory[i] = noSpriteRenderData
	}
}

func (renderer *spriteRenderer) clearSpriteZeroInRangeFlag() {
	renderer.isSpriteZeroInRange = false
}

func (renderer *spriteRenderer) executeScanlineSpriteCycle(currentCycle int, currentScanline int, scanlineOffscreenBuffer []int) {

	// Render sprite pixel for current scanline
	if 0 <= currentCycle && currentCycle <= 255 {
		if currentScanline > FirstRenderScanline { // No sprites on first scanline
			if currentCycle > 7 || !renderer.maskReg.isSpriteClippingEnabled() {
				renderer.renderSpritePixel(currentCycle, scanlineOffscreenBuffer)
			}
		}
	}

	// Evaluate/Fetch sprites for next scanline

	if 0 <= currentCycle && currentCycle <= 63 {
		// Init
		if currentCycle == 0 {
			for i := 0; i < 32; i++ {
				renderer.sprMemory.writeTemp(i, 0xFF)
			}
		}

	} else if 64 <= currentCycle && currentCycle <= 255 {
		// Sprite evaluation for next scanline
		if currentCycle == 64 {
			renderer.evaluateSprites(currentScanline)
		}

	} else if 256 <= currentCycle && currentCycle <= 319 {
		if currentCycle == 260 {
			renderer.fetchSpriteTileData(currentScanline)
		}
	}
}

func (renderer *spriteRenderer) renderSpritePixel(currentCycle int, scanlineOffscreenBuffer []int) {
	for _, pipeline := range renderer.pipelineMemory {
		// Go through all sprites evaluated for the scanline

		fineX := currentCycle - pipeline.xPosition
		if 0 <= fineX && fineX <= 7 {
			bitPosition := 1 << uint(7-fineX)

			tilePaletteDataLowBit := 0
			if (pipeline.tileDataLow & bitPosition) != 0 {
				tilePaletteDataLowBit = 1
			}
			tilePaletteDataHighBit := 0
			if (pipeline.tileDataHigh & bitPosition) != 0 {
				tilePaletteDataHighBit = 1
			}

			if tilePaletteDataLowBit != 0 || tilePaletteDataHighBit != 0 {
				// First non transparent sprite. We stop here!

				bgPixel := scanlineOffscreenBuffer[currentCycle]

				// Determine pixels to draw
				if (bgPixel&0x3) == 0 ||
					pipeline.isHighPriority ||
					!renderer.maskReg.isBackgroundVisibilityEnabled() {

					// BG transparent or SPRITE is high priority -> draw sprite
					paletteIndex := ((pipeline.attributePaletteData << 2) | (tilePaletteDataHighBit << 1) | tilePaletteDataLowBit) & 0xF

					scanlineOffscreenBuffer[currentCycle] = 0x10 | paletteIndex
				}

				// Sprite zero hit test
				if renderer.maskReg.isBackgroundVisibilityEnabled() &&
					renderer.maskReg.isSpriteVisibilityEnabled() &&
					currentCycle <= 254 &&
					(bgPixel&0x3) != 0 &&
					pipeline.isSpriteZero {

					// Sprite zero hit
					renderer.statusReg.value |= statusSpriteZeroOccurrence
				}

				break
			}
		}
	}
}

func (renderer *spriteRenderer) evaluateSprites(currentScanline int) {
	// Iterate over all 64 sprites and find the first 8 that are suitable for the next scanline.
	spriteIndexForNextScanline := 0
	for i := 0; i < 64; i++ {
		spriteMemoryIndex := i * 4

		yPosition := renderer.sprMemory.read(spriteMemoryIndex)

		if renderer.isSpriteInRangeForNextScanline(yPosition, currentScanline) {

			if spriteIndexForNextScanline < 8 {
				// 8 sprites are only visible for scanline.
				tileIndex := renderer.sprMemory.read(spriteMemoryIndex + 1)
				attributes := renderer.sprMemory.read(spriteMemoryIndex + 2)
				xPosition := renderer.sprMemory.read(spriteMemoryIndex + 3)

				spriteTempMemoryIndex := spriteIndexForNextScanline * 4

				renderer.sprMemory.writeTemp(spriteTempMemoryIndex, yPosition)
				renderer.sprMemory.writeTemp(spriteTempMemoryIndex+1, tileIndex)
				renderer.sprMemory.writeTemp(spriteTempMemoryIndex+2, attributes)
				renderer.sprMemory.writeTemp(spriteTempMemoryIndex+3, xPosition)

				spriteIndexForNextScanline++

				if i == 0 {
					// Sprite #0 is in range
					renderer.isSpriteZeroInRange = true
				}
			} else {
				// TODO FIXME Make this cycle perfect
				// More than 8 sprites suitable for next scanline.
				// Stop the evaluation and set the overflow flag.
				renderer.statusReg.value |= statusScanlineSpriteCount
				break
			}
		}
	}
}

func (renderer *spriteRenderer) isSpriteInRangeForNextScanline(sprYPosition int, currentScanline int) bool {
	isSpriteInRange := false

	spriteFineY :=
		(currentScanline + 1) -
			FirstRenderScanline -
			(sprYPosition + 1)

	if 0 <= spriteFineY && spriteFineY <= 7 {
		isSpriteInRange = true
	} else {
		if renderer.ctrlReg.is16PixelsSprite() {
			// 8x16 sprite
			if 0 <= spriteFineY && spriteFineY <= 15 {
				isSpriteInRange = true
			}
		}
	}

	return isSpriteInRange
}

func (renderer *spriteRenderer) fetchSpriteTileData(currentScanline int) {
	for spriteIndex := 0; spriteIndex < 8; spriteIndex++ {

		spriteAddress := spriteIndex * 4
		yPosition := renderer.sprMemory.readTemp(spriteAddress)
		tileIndex := renderer.sprMemory.readTemp(spriteAddress + 1)
		attributes := renderer.sprMemory.readTemp(spriteAddress + 2)
		xPosition := renderer.sprMemory.readTemp(spriteAddress + 3)

		fineY :=
			currentScanline -
				FirstRenderScanline -
				yPosition
		var spritePatternTableAddress int

		if !renderer.ctrlReg.is16PixelsSprite() {
			if (attributes & sprAttrRevertVertically) != 0 {
				fineY = 7 - fineY
			}
			spritePatternTableAddress = 0x0000
			if (renderer.ctrlReg.value & ctrlSpritePatternTableAddr) != 0 {
				spritePatternTableAddress = 0x1000
			}

		} else {
			if (attributes & sprAttrRevertVertically) != 0 {
				fineY = 15 - fineY
			}
			spritePatternTableAddress = 0x0000
			if (tileIndex & 0x1) != 0 {
				spritePatternTableAddress = 0x1000
			}
			tileIndex &= 0xFE // clear bit 0
			if fineY > 7 {
				// Pick second tile
				tileIndex++
				fineY -= 8
			}
		}

		spriteRenderData := spriteRenderPipeline{}

		if yPosition == 0xFF &&
			(tileIndex == 0xFE || tileIndex == 0xFF) &&
			attributes == 0xFF &&
			xPosition == 0xFF {

			// Although there is no sprite, we need to do dummy fetch so that the address line
			// is available (for Mapper04 for example).
			tileDataLowAddress := spritePatternTableAddress + tileIndex*16 + 0
			renderer.vramMemory.read(tileDataLowAddress)
			//_memory.read(tileDataLowAddress + 8)

			// No sprite
			spriteRenderData.tileDataLow = 0x0
			spriteRenderData.tileDataHigh = 0x0         // Transparent
			spriteRenderData.attributePaletteData = 0x0 // Irrelevant palette select index
			spriteRenderData.isHighPriority = false     // < background
			spriteRenderData.xPosition = 0x0            // Irrelevant
			spriteRenderData.isSpriteZero = false

		} else {
			tileDataLowAddress := spritePatternTableAddress + tileIndex*16 + fineY
			spriteRenderData.tileDataLow = renderer.vramMemory.read(tileDataLowAddress)
			spriteRenderData.tileDataHigh = renderer.vramMemory.read(tileDataLowAddress + 8)
			if (attributes & sprAttrRevertHorizontally) != 0 {
				spriteRenderData.tileDataLow = reverseByte(spriteRenderData.tileDataLow)
				spriteRenderData.tileDataHigh = reverseByte(spriteRenderData.tileDataHigh)
			}
			spriteRenderData.attributePaletteData = attributes & sprAttrPalette
			spriteRenderData.isHighPriority = (attributes & sprAttrPriority) == 0
			spriteRenderData.xPosition = xPosition
			spriteRenderData.isSpriteZero = renderer.isSpriteZeroInRange && (spriteIndex == 0)
		}

		renderer.pipelineMemory[spriteIndex] = spriteRenderData
	}
}
