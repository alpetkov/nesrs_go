package ppu

import (
	"github.com/alpetkov/nesrs_go/nesrs/cartridge"
)

// NES dimensions.
const (
	NESHeight = 240
	NESWidth  = 256
)

// VideoReceiver - handles PPU frames.
type VideoReceiver interface {
	ReceiveFrame(frame []int)
}

// VBLReceiver - handles PPU VBlank signals.
type VBLReceiver interface {
	ReceiveVBL()
}

// PPU - NES Picture Processing Unit.
type PPU struct {
	ctrlReg                    *ctrlRegister              // Registers
	maskReg                    *maskRegister              //
	statusReg                  *statusRegister            //
	sprRAMAddressReg           *spriteRAMAddressRegister  //
	vramAddressScrollReg       *vramAddressScrollRegister //
	vramMemory                 *vramMemory                // Memory
	sprMemory                  *spriteMemory              //
	backgroundRenderer         *backgroundRenderer        // Renderers
	spriteRenderer             *spriteRenderer            //
	scanlineOffscreenBuffer    [NESWidth]int              // Pixel buffers
	frameBuffer                [NESHeight * NESWidth]int  //
	currentCycle               int                        // Counters
	currentScanline            int                        //
	currentScanlineCyclesCount int                        //
	isOddFrame                 bool                       //
	canSetVblForFrame          bool                       //
	vblReceiver                VBLReceiver                // Receivers
	videoReceiver              VideoReceiver              //
}

// New instance of PPU.
func New(cartridge *cartridge.Cartridge, vblReceiver VBLReceiver, videoReceiver VideoReceiver) *PPU {
	ctrlReg := &ctrlRegister{}
	maskReg := &maskRegister{}
	statusReg := &statusRegister{}
	spriteRAMAddressReg := &spriteRAMAddressRegister{}
	vramAddressScrollReg := &vramAddressScrollRegister{}

	vramMemory := newVRAMMemory(cartridge)
	spriteMemory := &spriteMemory{}

	backgroundRenderer := newBackgroundRenderer(ctrlReg, maskReg, vramAddressScrollReg, vramMemory)
	spriteRenderer := newSpriteRenderer(ctrlReg, maskReg, statusReg, vramMemory, spriteMemory)

	ppu := PPU{
		ctrlReg:              ctrlReg,
		maskReg:              maskReg,
		statusReg:            statusReg,
		sprRAMAddressReg:     spriteRAMAddressReg,
		vramAddressScrollReg: vramAddressScrollReg,
		vramMemory:           vramMemory,
		sprMemory:            spriteMemory,
		backgroundRenderer:   backgroundRenderer,
		spriteRenderer:       spriteRenderer,
		vblReceiver:          vblReceiver,
		videoReceiver:        videoReceiver}

	return &ppu
}

// Init PPU.
func (ppu *PPU) Init() {
	ppu.currentCycle = -1
	ppu.currentScanline = 0
	ppu.currentScanlineCyclesCount = CyclesCountInScanline
	ppu.isOddFrame = true
	ppu.canSetVblForFrame = true

	// Registers
	ppu.ctrlReg.value = 0x00
	ppu.maskReg.value = 0x06
	ppu.statusReg.value = 0x00
	ppu.sprRAMAddressReg.value = 0x00
	ppu.vramAddressScrollReg.tempAddress = 0x0
	ppu.vramAddressScrollReg.bgFineX = 0
	ppu.vramAddressScrollReg.toggle = false
	ppu.vramAddressScrollReg.address = 0x0
	ppu.vramAddressScrollReg.lastValue = 0x0
}

// Reset PPU.
func (ppu *PPU) Reset() {
	ppu.currentCycle = -1
	ppu.currentScanline = 0
	ppu.currentScanlineCyclesCount = CyclesCountInScanline
	ppu.isOddFrame = true
	ppu.canSetVblForFrame = true

	// Registers
	ppu.ctrlReg.value = 0x00
	ppu.maskReg.value = 0x06
	ppu.statusReg.value &= 0x80
	ppu.vramAddressScrollReg.tempAddress = 0x0
	ppu.vramAddressScrollReg.bgFineX = 0
	ppu.vramAddressScrollReg.toggle = false
	ppu.vramAddressScrollReg.lastValue = 0x0
}

// ExecuteCycles runs PPU cycles.
func (ppu *PPU) ExecuteCycles(ppuCycles int) {
	for i := 0; i < ppuCycles; i++ {
		ppu.currentCycle++

		//
		// Determine scanline & frame
		//

		if ppu.currentCycle == ppu.currentScanlineCyclesCount {
			// New scanline
			ppu.currentCycle = 0
			ppu.currentScanline++
			ppu.currentScanlineCyclesCount = CyclesCountInScanline

			if ppu.currentScanline == ScanlineCountInFrame {
				// New frame
				ppu.currentScanline = 0
				ppu.isOddFrame = !ppu.isOddFrame
				for i := range ppu.frameBuffer {
					ppu.frameBuffer[i] = 0x00
				}
			}
		}

		//
		// Scanline rendering
		//

		if VblankStartScanline == ppu.currentScanline {
			if ppu.currentCycle == 0 {
				// Set VBlank flag
				if ppu.canSetVblForFrame {
					ppu.statusReg.setInVblank(true)
				}

				// Clear spr ram address
				ppu.sprRAMAddressReg.value = 0

				// Clear sprite renderer pipeline
				ppu.spriteRenderer.reset()

			} else if ppu.currentCycle == 2 {
				if ppu.ctrlReg.execNMIOnVblEnabled() {
					ppu.vblReceiver.ReceiveVBL()
				}

				// Clear VBL lock
				ppu.canSetVblForFrame = true
			}

		} else if DummyRenderScanline == ppu.currentScanline {
			if ppu.currentCycle == 0 {
				// Clear VBlank flag
				ppu.statusReg.setInVblank(false)

				// Fix flags
				ppu.statusReg.value &^= statusScanlineSpriteCount
				ppu.statusReg.value &^= statusSpriteZeroOccurrence

				// ???
				//ppu.statusReg._value = 0

				// Clear sprite zero flag
				ppu.spriteRenderer.clearSpriteZeroInRangeFlag()

				if ppu.isOddFrame && ppu.maskReg.isBackgroundVisibilityEnabled() {
					ppu.currentScanlineCyclesCount = CyclesCountInScanline - 1
				}

			} else if ppu.currentCycle == 256 /*or 257*/ {
				if ppu.isRenderingEnabled() {
					// v:0000010000011111=t:0000010000011111
					ppu.vramAddressScrollReg.address &= 0xFBE0
					ppu.vramAddressScrollReg.address |= (ppu.vramAddressScrollReg.tempAddress & 0x041F)
				}

			} else if ppu.currentCycle == 303 /*or 304*/ {
				// Frame start
				if ppu.isRenderingEnabled() {
					ppu.vramAddressScrollReg.address = ppu.vramAddressScrollReg.tempAddress
				}
			}

			// Render
			if ppu.isRenderingEnabled() {
				ppu.backgroundRenderer.executeScanlineBackgroundCycle(
					ppu.currentCycle,
					ppu.scanlineOffscreenBuffer[:],
					false)

				ppu.spriteRenderer.executeScanlineSpriteCycle(
					ppu.currentCycle,
					ppu.currentScanline,
					ppu.scanlineOffscreenBuffer[:])
			}

		} else if FirstRenderScanline <= ppu.currentScanline &&
			ppu.currentScanline <= LastRenderScanline {

			if ppu.currentCycle == 0 {
				// Clear offscreen buffers
				for i := range ppu.scanlineOffscreenBuffer {
					ppu.scanlineOffscreenBuffer[i] = 0x00
				}
			} else if ppu.currentCycle == 256 /*or 257*/ {
				if ppu.isRenderingEnabled() {
					// v:0000010000011111=t:0000010000011111
					ppu.vramAddressScrollReg.address &= 0xFBE0
					ppu.vramAddressScrollReg.address |= (ppu.vramAddressScrollReg.tempAddress & 0x041F)
				}
			}

			// Render
			if ppu.isRenderingEnabled() {
				ppu.backgroundRenderer.executeScanlineBackgroundCycle(
					ppu.currentCycle,
					ppu.scanlineOffscreenBuffer[:],
					true)

				ppu.spriteRenderer.executeScanlineSpriteCycle(
					ppu.currentCycle,
					ppu.currentScanline,
					ppu.scanlineOffscreenBuffer[:])
			}

			// Send to video
			if ppu.currentCycle == ppu.currentScanlineCyclesCount-1 {
				scanlineBuffer := ppu.getScanlineVideo()
				offset := (ppu.currentScanline - FirstRenderScanline) * NESWidth
				for i := range scanlineBuffer {
					ppu.frameBuffer[offset+i] = scanlineBuffer[i]
				}
				if ppu.currentScanline == LastRenderScanline {
					ppu.videoReceiver.ReceiveFrame(ppu.frameBuffer[:])
				}
			}
		}
	}
}

// ReadRegister of PPU.
func (ppu *PPU) ReadRegister(register int) int {
	switch register {

	case StatusRegID:
		{
			// $2002 R toggle = 0
			status := ppu.statusReg.value

			ppu.statusReg.setInVblank(false)
			ppu.vramAddressScrollReg.toggle = false

			// Reading one PPU cycle before VBL unsets it
			if WasteScanline == ppu.currentScanline &&
				ppu.currentCycle == ppu.currentScanlineCyclesCount-1 {
				ppu.canSetVblForFrame = false
				ppu.ctrlReg.setExecNMIOnVblEnabled(false)

			} else if VblankStartScanline == ppu.currentScanline &&
				ppu.currentCycle <= 1 {
				ppu.ctrlReg.setExecNMIOnVblEnabled(false)
			}

			return status
		}

	case SpriteRAMIORegID:
		{
			return ppu.sprMemory.read(ppu.sprRAMAddressReg.value)
		}

	case VRAMIORegID:
		{
			var result int

			vramAddress := ppu.vramAddressScrollReg.address & 0x3FFF
			if 0 <= vramAddress && vramAddress < 0x3F00 {
				result = ppu.vramAddressScrollReg.lastValue
				ppu.vramAddressScrollReg.lastValue = ppu.vramMemory.read(vramAddress)
			} else { // Background palette
				result = ppu.vramMemory.read(vramAddress)
				// $2000-$2fff are mirrored at $3000-$3fff
				ppu.vramAddressScrollReg.lastValue = ppu.vramMemory.read(vramAddress - 0x1000)
			}

			ppu.vramAddressScrollReg.address += ppu.getVramAddressInc()
			ppu.vramAddressScrollReg.address &= 0x7FFF

			return result
		}
	}

	return 0
}

// WriteRegister of PPU.
func (ppu *PPU) WriteRegister(register int, value int) {
	switch register {

	case CtrlRegID:
		{
			// $2000 W %---- --NN
			// temp    %---- NN--   ---- ----
			// t:0000|1100|0000|0000=d:0000|0011
			ppu.vramAddressScrollReg.tempAddress &= 0x73FF
			ppu.vramAddressScrollReg.tempAddress |= ((value << 10) & 0x0C00)

			if ppu.statusReg.isInVblank() && !ppu.ctrlReg.execNMIOnVblEnabled() &&
				((value & ctrlExecNMIOnVblank) != 0) {
				ppu.vblReceiver.ReceiveVBL()
			}

			ppu.ctrlReg.value = value
		}

	case MaskRegID:
		{
			ppu.maskReg.value = value
		}

	case SpriteRAMAddressRegID:
		{
			ppu.sprRAMAddressReg.value = value & 0xFF
		}

	case SpriteRAMIORegID:
		{
			ppu.sprMemory.write(ppu.sprRAMAddressReg.value, value)

			ppu.sprRAMAddressReg.value++
			ppu.sprRAMAddressReg.value &= 0xFF
		}

	case ScrollRegID:
		{
			if !ppu.vramAddressScrollReg.toggle {
				// First write

				ppu.vramAddressScrollReg.tempAddress &= 0x7FE0
				ppu.vramAddressScrollReg.tempAddress |= ((value >> 3) & 0x1F)
				ppu.vramAddressScrollReg.bgFineX = (value & 0x07)
			} else {
				// Second write

				ppu.vramAddressScrollReg.tempAddress &= 0x0C1F
				ppu.vramAddressScrollReg.tempAddress |= ((value << 2) & 0x03E0)
				ppu.vramAddressScrollReg.tempAddress |= ((value << 12) & 0x7000)
			}

			ppu.vramAddressScrollReg.toggle = !ppu.vramAddressScrollReg.toggle
		}

	case VRAMAddressRegID:
		{
			if !ppu.vramAddressScrollReg.toggle {
				// Upper address byte (first write)

				ppu.vramAddressScrollReg.tempAddress &= 0x00FF
				ppu.vramAddressScrollReg.tempAddress |= ((value << 8) & 0x3F00)
			} else {
				// Lower address byte (second write)

				ppu.vramAddressScrollReg.tempAddress &= 0x7F00
				ppu.vramAddressScrollReg.tempAddress |= (value & 0x00FF)
				ppu.vramAddressScrollReg.address = ppu.vramAddressScrollReg.tempAddress
			}

			ppu.vramAddressScrollReg.toggle = !ppu.vramAddressScrollReg.toggle
		}

	case VRAMIORegID:
		{
			if (ppu.statusReg.value & statusVRAMWriteFlag) == 0 {
				ppu.vramMemory.write(ppu.vramAddressScrollReg.address&0x3FFF, value)
			}

			ppu.vramAddressScrollReg.address += ppu.getVramAddressInc()
			ppu.vramAddressScrollReg.address &= 0x7FFF
		}
	}
}

func (ppu *PPU) getVramAddressInc() int {
	if (ppu.ctrlReg.value & ctrlAddrInc) != 0 {
		return 32
	}

	return 1
}

func (ppu *PPU) getScanlineVideo() []int {
	scanlineVideoBuffer := make([]int, len(ppu.scanlineOffscreenBuffer))
	for i := range ppu.scanlineOffscreenBuffer {
		paletteAddress := 0x3F00 | (ppu.scanlineOffscreenBuffer[i] & 0x1F)
		colorIndex := ppu.vramMemory.read(paletteAddress)
		rgb := paletteRGB[colorIndex&0x3F]
		scanlineVideoBuffer[i] = rgb
	}

	return scanlineVideoBuffer
}

func (ppu *PPU) isRenderingEnabled() bool {
	return ppu.maskReg.isBackgroundVisibilityEnabled() || ppu.maskReg.isSpriteVisibilityEnabled()
}
