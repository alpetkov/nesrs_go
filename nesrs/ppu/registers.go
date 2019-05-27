package ppu

// PPU Registers indexes.
const (
	CtrlRegID             = iota // PPU Control Register (W)
	MaskRegID                    // PPU Mask Register (W)
	StatusRegID                  // PPU Status Register (R)
	SpriteRAMAddressRegID        // SPR-RAM Address Register (W)
	SpriteRAMIORegID             // SPR-RAM I/O Register (RW)
	ScrollRegID                  // Scroll Register (W2)
	VRAMAddressRegID             // VRAM Address Register (W2)
	VRAMIORegID                  // VRAM I/O Register (RW)
)

const (
	ctrlExecNMIOnVblank           = 0x80 // bit 7
	ctrlMasterSlaveSelection      = 0x40 // bit 6
	ctrlSpriteSize                = 0x20 // bit 5
	ctrlBackgrounPatternTableAddr = 0x10 // bit 4
	ctrlSpritePatternTableAddr    = 0x08 // bit 3
	ctrlAddrInc                   = 0x04 // bit 2
	ctrlNametableAddr             = 0x03 // bit 1 & 0
)

const (
	maskColorIntensity       = 0xE0 // bits 5,6 and 7
	maskSpriteVisibility     = 0x10 // bit 4
	maskBackgroundVisibility = 0x08 // bit 3
	maskSpriteClipping       = 0x04 // bit 2
	maskBackgroundClipping   = 0x02 // bit 1
	maskDisableColorburst    = 0x01 // bit 0
)

const (
	statusVblankOccurrence     = 0x80 // bit 7
	statusSpriteZeroOccurrence = 0x40 // bit 6
	statusScanlineSpriteCount  = 0x20 // bit 5
	statusVRAMWriteFlag        = 0x10 // bit 4
)

type ctrlRegister struct {
	value int
}
type maskRegister struct {
	value int
}
type spriteRAMAddressRegister struct {
	value int
}
type statusRegister struct {
	value int
}
type vramAddressScrollRegister struct {
	address     int  // VRAM address entered.
	lastValue   int  // Stores the last read VRAM value
	tempAddress int  // VRAM temp address. VRAM is entered in two steps. Also can be interpret as 0yyy NNYY YYYX XXXX (fineY, name table, tileY, tileX)
	toggle      bool // VRAM address step toggle
	bgFineX     int  // xxx (Background fineX)
}

func (ctrl *ctrlRegister) execNMIOnVblEnabled() bool {
	return (ctrl.value & ctrlExecNMIOnVblank) != 0
}

func (ctrl *ctrlRegister) setExecNMIOnVblEnabled(enabled bool) {
	if enabled {
		ctrl.value |= ctrlExecNMIOnVblank
	} else {
		ctrl.value &^= ctrlExecNMIOnVblank
	}
}

func (ctrl *ctrlRegister) getBackgroundPatternTableAddress() int {
	if (ctrl.value & ctrlBackgrounPatternTableAddr) != 0 {
		return 0x1000
	}

	return 0x0000
}

func (ctrl *ctrlRegister) is16PixelsSprite() bool {
	return (ctrl.value & ctrlSpriteSize) != 0
}

func (mask *maskRegister) isBackgroundVisibilityEnabled() bool {
	return (mask.value & maskBackgroundVisibility) != 0
}

func (mask *maskRegister) isSpriteVisibilityEnabled() bool {
	return (mask.value & maskSpriteVisibility) != 0
}

func (mask *maskRegister) isBackgroundClippingEnabled() bool {
	return (mask.value & maskBackgroundClipping) == 0
}

func (mask *maskRegister) isSpriteClippingEnabled() bool {
	return (mask.value & maskSpriteClipping) == 0
}

func (status *statusRegister) isInVblank() bool {
	return (status.value & statusVblankOccurrence) != 0
}

func (status *statusRegister) setInVblank(hasOccurred bool) {
	if hasOccurred {
		status.value |= statusVblankOccurrence
	} else {
		status.value &^= statusVblankOccurrence
	}
}

func (vram *vramAddressScrollRegister) backgroundFineY() int {
	// temp 0yyy NNYY YYYX XXXX
	return (vram.address >> 12) & 0x7
}

func (vram *vramAddressScrollRegister) setBackgroundFineY(fineY int) {
	// temp 0yyy NNYY YYYX XXXX
	vram.address &= 0x8FFF
	vram.address |= fineY << 12
}

func (vram *vramAddressScrollRegister) backgroundTileX() int {
	// temp 0yyy NNYY YYYX XXXX
	return vram.address & 0x1F
}

func (vram *vramAddressScrollRegister) setBackgroundTileX(tileX int) {
	// temp 0yyy NNYY YYYX XXXX
	vram.address &= 0xFFE0
	vram.address |= tileX
}

func (vram *vramAddressScrollRegister) backgroundTileY() int {
	// temp 0yyy NNYY YYYX XXXX
	return (vram.address >> 5) & 0x1F
}

func (vram *vramAddressScrollRegister) setBackgroundTileY(tileY int) {
	// temp 0yyy NNYY YYYX XXXX
	vram.address &= 0xFC1F
	vram.address |= tileY << 5
}

func (vram *vramAddressScrollRegister) nameTableIndex() int {
	return (vram.address >> 10) & 0x3
}

func (vram *vramAddressScrollRegister) setNameTableIndex(nameTableIndex int) {
	vram.address &= 0xF3FF
	vram.address |= nameTableIndex << 10
}
