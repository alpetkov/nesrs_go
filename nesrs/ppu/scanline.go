package ppu

// Scanline helper.
const (
	ScanlineCountInFrame  = 262
	CyclesCountInScanline = 341

	VblankScanlinesInFrame = 20

	// Variant 1
	// 0-19 Vblank
	// 20 dummy scanline
	// 21-260 render scanlines
	// 261 waste scanline

	// Variant 2
	// 0-239 render scanline
	// 240 waste scanline
	// 241-260 Vblank
	// 261 dummy scanline

	// Platoon works only with vblank start set to 0 right now.
	VblankStartScanline = 0 //241
	VblankEndScanline   = VblankStartScanline + VblankScanlinesInFrame - 1

	DummyRenderScanline = VblankEndScanline + 1
	FirstRenderScanline = (DummyRenderScanline + 1) % ScanlineCountInFrame
	LastRenderScanline  = FirstRenderScanline + 239

	WasteScanline = LastRenderScanline + 1
)

type scanlineCounter struct {
	currentCycle               int
	currentScanline            int
	currentScanlineCyclesCount int
	isOddFrame                 bool
	canSetVblForFrame          bool
}
