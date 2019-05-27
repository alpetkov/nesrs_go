package cpu

import "fmt"

// Status register's flags
const (
	flagC = 0x01 // Carry flag. 1 -> Carry occurred
	flagZ = 0x02 // Zero flag. 1 -> Result is zero
	flagI = 0x04 // Interrupt flag. 1 -> IRQ disabled
	flagD = 0x08 // Decimal mode flag. 1 -> Decimal arithmetic
	flagB = 0x10 // Break flag. 1 -> BRK instruction occurred
	flagR = 0x20 // Not used. Always is set to 1
	flagV = 0x40 // Overflow flag. 1 -> Overflow occurred
	flagN = 0x80 // Negative flag. 1 -> Result is negative
)

// Interrupt types
const (
	RESET = iota + 1
	NMI
	IRQ
)

// Addressing modes
const (
	ACC = iota + 1
	IMPL
	IMM
	ABS
	ZP
	ZPX
	ZPY
	ABSX
	ABSX2
	ABSY
	ABSY2
	REL
	INDX
	INDY
	INDY2
	IND
	IND2
)

var opCyclesLength = [256]int{
	/*       0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F*/
	/*0x00*/ 7, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
	/*0x10*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	/*0x20*/ 6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
	/*0x30*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	/*0x40*/ 6, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
	/*0x50*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	/*0x60*/ 6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
	/*0x70*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	/*0x80*/ 2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	/*0x90*/ 2, 6, 2, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
	/*0xA0*/ 2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	/*0xB0*/ 2, 5, 2, 5, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4,
	/*0xC0*/ 2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	/*0xD0*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	/*0xE0*/ 2, 6, 3, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	/*0xF0*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7}

// CPU - 6502 CPU for NES.
type CPU struct {
	// CPU's registers
	A  int // Accumulator (8 bits)
	X  int // X index register (8 bits)
	Y  int // Y index register (8 bits)
	S  int // Stack pointer (8 bits)
	P  int // Status register (set of flags) (8 bits)
	PC int // Program counter (16 bits)

	// Number of cycles of the last executed op
	OpCycles int

	// Pending interrupt
	pendingInterrupt int

	// 64Kb of CPU's addressable memory
	memory CPUMemory

	decimalModeSupported bool
}

// New CPU.
func New(memory CPUMemory) *CPU {
	cpu := CPU{memory: memory, decimalModeSupported: false}

	return &cpu
}

// Init the CPU.
func (cpu *CPU) Init() int {
	cpu.A = 0x00
	cpu.X = 0x00
	cpu.Y = 0x00
	cpu.S = 0xFF
	cpu.P = flagB | flagR | flagI
	cpu.OpCycles = 7

	cpu.PC = (cpu.readMemory(0xFFFD) << 8) | cpu.readMemory(0xFFFC)

	return cpu.OpCycles
}

// Reset the CPU.
func (cpu *CPU) Reset() {
	cpu.requestInterrupt(RESET)
}

// NMI - sends NMI to the CPU
func (cpu *CPU) NMI() {
	fmt.Println("CPU NMI")
	cpu.requestInterrupt(NMI)
}

// IRQ - sends IRQ to the CPU
func (cpu *CPU) IRQ() {
	cpu.requestInterrupt(IRQ)
}

// ExecuteOp - Execute CPU OP
func (cpu *CPU) ExecuteOp() int {
	if cpu.pendingInterrupt != 0 {
		cpu.executePendingInterruptOp()
	} else {
		opCode := cpu.readMemory(cpu.PC)
		cpu.PC++
		cpu.OpCycles = opCyclesLength[opCode]
		cpu.executeOp(opCode)
	}

	return cpu.OpCycles
}

//
// Interrupt handling
//

func (cpu *CPU) requestInterrupt(interruptType int) {
	if cpu.pendingInterrupt == 0 {
		cpu.pendingInterrupt = interruptType
	} else if cpu.pendingInterrupt == IRQ {
		cpu.pendingInterrupt = interruptType
	} else if cpu.pendingInterrupt == NMI {
		if interruptType == RESET {
			cpu.pendingInterrupt = interruptType
		}
	} else if cpu.pendingInterrupt == RESET {
		// Already requested
	}
}

func (cpu *CPU) executePendingInterruptOp() {
	cpu.OpCycles = 0

	if cpu.pendingInterrupt == RESET {
		cpu.OpCycles = 7
		cpu.A = 0x00
		cpu.X = 0x00
		cpu.Y = 0x00
		cpu.S = 0xFF
		cpu.P = flagZ | flagR
		cpu.PC = (cpu.readMemory(0xFFFD) << 8) | cpu.readMemory(0xFFFC)

	} else if cpu.pendingInterrupt == NMI {
		cpu.OpCycles = 7
		cpu.push((cpu.PC >> 8) & 0xFF)
		cpu.push(cpu.PC &^ 0x00FF)
		cpu.push(cpu.P &^ flagB)
		cpu.P = cpu.P &^ flagD
		cpu.PC = (cpu.readMemory(0xFFFB) << 8) | cpu.readMemory(0xFFFA)

	} else if cpu.pendingInterrupt == IRQ && ((cpu.P & flagI) == 0) {
		cpu.OpCycles = 7
		cpu.push((cpu.PC >> 8) & 0xFF)
		cpu.push(cpu.PC & 0x00FF)
		cpu.push(cpu.P &^ flagB)
		cpu.P = cpu.P &^ flagD
		cpu.P = cpu.P &^ flagI
		cpu.PC = (cpu.readMemory(0xFFFF) << 8) | cpu.readMemory(0xFFFE)
	}

	cpu.pendingInterrupt = 0
}

//
// Memory Management
//

func (cpu *CPU) readMemory(address int) int {
	return cpu.memory.Read(address)
}

func (cpu *CPU) writeMemory(address int, value int) {
	additionalWriteCycles := cpu.memory.Write(address, value)
	if additionalWriteCycles > 0 {
		cpu.OpCycles += additionalWriteCycles
	}
}

func isPageBoundaryCrossed(address1 int, address2 int) bool {
	return (address1 >> 8) != (address2 >> 8)
}

//
// Addressing modes
//

// 1. Accumulator addressing - ACC
// 2. Implied addressing - IMPL

// 3. Immediate addressing - IMM
func (cpu *CPU) calculateMemoryAddressIMM() int {
	res := cpu.PC
	cpu.PC++
	return res
}

// 4. Absolute addressing - ABS
func (cpu *CPU) calculateMemoryAddressABS() int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	high := cpu.readMemory(cpu.PC)
	cpu.PC++
	return 0xFFFF & ((high << 8) | low)
}

// 5. Zero page addressing - ZP
func (cpu *CPU) calculateMemoryAddressZP() int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	return low
}

// 6. Indexed zero page addressing with register X - ZP,X
func (cpu *CPU) calculateMemoryAddressZPX() int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	return 0x00FF & (cpu.X + low)
}

// 7. Indexed zero page addressing with register Y - ZP,Y
func (cpu *CPU) calculateMemoryAddressZPY() int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	return 0x00FF & (cpu.Y + low)
}

// 8. Indexed absolute addressing with register X - ABS,X
func (cpu *CPU) calculateMemoryAddressABSX(countAdditionalCycleOnPageBoundaryCrossed bool) int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	high := cpu.readMemory(cpu.PC)
	cpu.PC++
	address := 0xFFFF & ((high << 8) | low)
	resultAddress := 0xFFFF & (address + cpu.X)
	if countAdditionalCycleOnPageBoundaryCrossed && isPageBoundaryCrossed(address, resultAddress) {
		cpu.OpCycles++
	}

	return resultAddress
}

// 9. Indexed absolute addressing with register Y - ABS,Y
func (cpu *CPU) calculateMemoryAddressABSY(countAdditionalCycleOnPageBoundaryCrossed bool) int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	high := cpu.readMemory(cpu.PC)
	cpu.PC++
	address := 0xFFFF & ((high << 8) | low)
	resultAddress := 0xFFFF & (address + cpu.Y)
	// TODO
	if countAdditionalCycleOnPageBoundaryCrossed && isPageBoundaryCrossed(address, resultAddress) {
		cpu.OpCycles++
	}

	return resultAddress
}

// 10. Relative addressing - REL
func (cpu *CPU) calculateMemoryAddressREL() int {
	inc := cpu.readMemory(cpu.PC)
	cpu.PC++
	var offset = 0
	var isPositive = true
	if (inc & 0x80) == 0 {
		// Positive or Zero
		offset = inc & 0x7F
	} else {
		// Negative
		offset = 0x7F + 1 - (inc & 0x7F)
		isPositive = false
	}

	var address int
	if isPositive {
		address = cpu.PC + offset
	} else {
		address = cpu.PC - offset
	}

	return 0xFFFF & address
}

// 11. Indexed indirect (pre-indexed) addressing with register X - (IND,X)
func (cpu *CPU) calculateMemoryAddressINDX() int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	address := 0x00FF & (low + cpu.X)
	nextAddress := 0x00FF & (address + 1)
	return 0xFFFF & ((cpu.readMemory(nextAddress) << 8) | cpu.readMemory(address))
}

// 12. Indirect indexed (post-indexed) addressing with register Y - (IND),Y
func (cpu *CPU) calculateMemoryAddressINDY(countAdditionalCycleOnPageBoundaryCrossed bool) int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++

	lowAddress := cpu.readMemory(low)
	highAddress := cpu.readMemory(0x00FF & (low + 1))
	address := 0xFFFF & ((highAddress << 8) | lowAddress)
	resultAddress := 0xFFFF & (address + cpu.Y)
	if countAdditionalCycleOnPageBoundaryCrossed && isPageBoundaryCrossed(address, resultAddress) {
		cpu.OpCycles++
	}

	return resultAddress
}

// 13. Absolute indirect addressing - IND
func (cpu *CPU) calculateMemoryAddressIND(isPageWrappingNotAllowed bool) int {
	low := cpu.readMemory(cpu.PC)
	cpu.PC++
	high := cpu.readMemory(cpu.PC)
	cpu.PC++
	address := 0xFFFF & ((high << 8) | low)
	nextAddress := address + 1
	if (address & 0xFF) == 0xFF {
		if isPageWrappingNotAllowed {
			nextAddress = address & 0xFF00
		}
	}
	return 0xFFFF & ((cpu.readMemory(nextAddress) << 8) | cpu.readMemory(address))
}

func (cpu *CPU) calculateMemoryAddress(mode int) int {
	address := 0

	switch mode {
	case ACC:
	case IMPL:
	case IMM:
		address = cpu.calculateMemoryAddressIMM()
	case ABS:
		address = cpu.calculateMemoryAddressABS()
	case ZP:
		address = cpu.calculateMemoryAddressZP()
	case ZPX:
		address = cpu.calculateMemoryAddressZPX()
	case ZPY:
		address = cpu.calculateMemoryAddressZPY()
	case ABSX:
		address = cpu.calculateMemoryAddressABSX(false)
	case ABSX2:
		address = cpu.calculateMemoryAddressABSX(true)
	case ABSY:
		address = cpu.calculateMemoryAddressABSY(false)
	case ABSY2:
		address = cpu.calculateMemoryAddressABSY(true)
	case REL:
		address = cpu.calculateMemoryAddressREL()
	case INDX:
		address = cpu.calculateMemoryAddressINDX()
	case INDY:
		address = cpu.calculateMemoryAddressINDY(false)
	case INDY2:
		address = cpu.calculateMemoryAddressINDY(true)
	case IND:
		address = cpu.calculateMemoryAddressIND(false)
	case IND2:
		address = cpu.calculateMemoryAddressIND(true)
	}

	return address
}

//
// Stack manipulation
//

func (cpu *CPU) push(value int) {
	cpu.writeMemory(0x0100|cpu.S, value)
	if cpu.S == 0x00 {
		cpu.S = 0xFF
	} else {
		cpu.S--
	}
}

func (cpu *CPU) pop() int {
	if cpu.S == 0xFF {
		cpu.S = 0x00
	} else {
		cpu.S++
	}

	return cpu.readMemory(0x0100 | cpu.S)
}

//
// CPU running cycle
//

func (cpu *CPU) executeOp(opCode int) {
	switch opCode {
	/*1.ADC*/
	case 0x69 /*IMM*/ :
		cpu.opADC(IMM)
	case 0x65 /*ZP*/ :
		cpu.opADC(ZP)
	case 0x75 /*ZP,X*/ :
		cpu.opADC(ZPX)
	case 0x6D /*ABS*/ :
		cpu.opADC(ABS)
	case 0x7D /*ABS,X+*/ :
		cpu.opADC(ABSX2)
	case 0x79 /*ABS,Y+*/ :
		cpu.opADC(ABSY2)
	case 0x61 /*(IND,X)*/ :
		cpu.opADC(INDX)
	case 0x71 /*(IND),Y+*/ :
		cpu.opADC(INDY2)

	/*2.AND*/
	case 0x29 /*IMM*/ :
		cpu.opAND(IMM)
	case 0x25 /*ZP*/ :
		cpu.opAND(ZP)
	case 0x35 /*ZP,X*/ :
		cpu.opAND(ZPX)
	case 0x2D /*ABS*/ :
		cpu.opAND(ABS)
	case 0x3D /*ABS,X+*/ :
		cpu.opAND(ABSX2)
	case 0x39 /*ABS,Y+*/ :
		cpu.opAND(ABSY2)
	case 0x21 /*(IND,X)*/ :
		cpu.opAND(INDX)
	case 0x31 /*(IND),Y+*/ :
		cpu.opAND(INDY2)

	/*3.ASL*/
	case 0x0A /*ACC*/ :
		cpu.opASL(ACC)
	case 0x06 /*ZP*/ :
		cpu.opASL(ZP)
	case 0x16 /*ZP,X*/ :
		cpu.opASL(ZPX)
	case 0x0E /*ABS*/ :
		cpu.opASL(ABS)
	case 0x1E /*ABS,X*/ :
		cpu.opASL(ABSX)

	/*4.BCC*/
	case 0x90 /*REL*/ :
		cpu.opBCC(REL)

	/*5.BCS*/
	case 0xB0 /*REL*/ :
		cpu.opBCS(REL)

	/*6.BEQ*/
	case 0xF0 /*REL*/ :
		cpu.opBEQ(REL)

	/*7.BIT*/
	case 0x24 /*ZP*/ :
		cpu.opBIT(ZP)
	case 0x2C /*ABS*/ :
		cpu.opBIT(ABS)

	/*8.BMI*/
	case 0x30 /*REL*/ :
		cpu.opBMI(REL)

	/*9.BNE*/
	case 0xD0 /*REL*/ :
		cpu.opBNE(REL)

	/*10.BPL*/
	case 0x10 /*REL*/ :
		cpu.opBPL(REL)

	/*11.BRK*/
	case 0x00 /*IMPL*/ :
		cpu.opBRK()

	/*12.BVC*/
	case 0x50 /*REL*/ :
		cpu.opBVC(REL)

	/*13.BVS*/
	case 0x70 /*REL*/ :
		cpu.opBVS(REL)

	/*14.CLC*/
	case 0x18 /*IMPL*/ :
		cpu.P = cpu.P &^ flagC

	/*15.CLD*/
	case 0xD8 /*IMPL*/ :
		cpu.P = cpu.P &^ flagD

	/*16.CLI*/
	case 0x58 /*IMPL*/ :
		cpu.P = cpu.P &^ flagI

	/*17.CLV*/
	case 0xB8 /*IMPL*/ :
		cpu.P = cpu.P &^ flagV

	/*18.CMP*/
	case 0xC9 /*IMM*/ :
		cpu.opCMP(IMM)
	case 0xC5 /*ZP*/ :
		cpu.opCMP(ZP)
	case 0xD5 /*ZP,X*/ :
		cpu.opCMP(ZPX)
	case 0xCD /*ABS*/ :
		cpu.opCMP(ABS)
	case 0xDD /*ABS,X+*/ :
		cpu.opCMP(ABSX2)
	case 0xD9 /*ABS,Y+*/ :
		cpu.opCMP(ABSY2)
	case 0xC1 /*(IND,X)*/ :
		cpu.opCMP(INDX)
	case 0xD1 /*(IND),Y+*/ :
		cpu.opCMP(INDY2)

	/*19.CPX*/
	case 0xE0 /*IMM*/ :
		cpu.opCPX(IMM)
	case 0xE4 /*ZP*/ :
		cpu.opCPX(ZP)
	case 0xEC /*ABS*/ :
		cpu.opCPX(ABS)

	/*20.CPY*/
	case 0xC0 /*IMM*/ :
		cpu.opCPY(IMM)
	case 0xC4 /*ZP*/ :
		cpu.opCPY(ZP)
	case 0xCC /*ABS*/ :
		cpu.opCPY(ABS)

	/*21.DEC*/
	case 0xC6 /*ZP*/ :
		cpu.opDEC(ZP)
	case 0xD6 /*ZP,X*/ :
		cpu.opDEC(ZPX)
	case 0xCE /*ABS*/ :
		cpu.opDEC(ABS)
	case 0xDE /*ABS,X*/ :
		cpu.opDEC(ABSX)

	/*22.DEX*/
	case 0xCA /*IMPL*/ :
		cpu.X = cpu.opDecrease(cpu.X)

	/*23.DEY*/
	case 0x88 /*IMPL*/ :
		cpu.Y = cpu.opDecrease(cpu.Y)

	/*24.EOR*/
	case 0x49 /*IMM*/ :
		cpu.opEOR(IMM)
	case 0x45 /*ZP*/ :
		cpu.opEOR(ZP)
	case 0x55 /*ZP,X*/ :
		cpu.opEOR(ZPX)
	case 0x4D /*ABS*/ :
		cpu.opEOR(ABS)
	case 0x5D /*ABS,X+*/ :
		cpu.opEOR(ABSX2)
	case 0x59 /*ABS,Y+*/ :
		cpu.opEOR(ABSY2)
	case 0x41 /*(IND,X)*/ :
		cpu.opEOR(INDX)
	case 0x51 /*(IND),Y+*/ :
		cpu.opEOR(INDY2)

	/*25.INC*/
	case 0xE6 /*ZP*/ :
		cpu.opINC(ZP)
	case 0xF6 /*ZP,X*/ :
		cpu.opINC(ZPX)
	case 0xEE /*ABS*/ :
		cpu.opINC(ABS)
	case 0xFE /*ABS,X*/ :
		cpu.opINC(ABSX)

	/*26.INX*/
	case 0xE8 /*IMPL*/ :
		cpu.X = cpu.opIncrease(cpu.X)

	/*27.INY*/
	case 0xC8 /*IMPL*/ :
		cpu.Y = cpu.opIncrease(cpu.Y)

	/*28.JMP*/
	case 0x4C /*ABS*/ :
		cpu.opJMP(ABS)
	case 0x6C /*IND*/ :
		cpu.opJMP(IND2)

	/*29.JSR*/
	case 0x20 /*ABS*/ :
		cpu.opJSR(ABS)

	/*30.LDA*/
	case 0xA9 /*IMM*/ :
		cpu.opLDA(IMM)
	case 0xA5 /*ZP*/ :
		cpu.opLDA(ZP)
	case 0xB5 /*ZP,X*/ :
		cpu.opLDA(ZPX)
	case 0xAD /*ABS*/ :
		cpu.opLDA(ABS)
	case 0xBD /*ABS,X+*/ :
		cpu.opLDA(ABSX2)
	case 0xB9 /*ABS,Y+*/ :
		cpu.opLDA(ABSY2)
	case 0xA1 /*(IND,X)*/ :
		cpu.opLDA(INDX)
	case 0xB1 /*(IND),Y+*/ :
		cpu.opLDA(INDY2)

	/*31.LDX*/
	case 0xA2 /*IMM*/ :
		cpu.opLDX(IMM)
	case 0xA6 /*ZP*/ :
		cpu.opLDX(ZP)
	case 0xB6 /*ZP,Y*/ :
		cpu.opLDX(ZPY)
	case 0xAE /*ABS*/ :
		cpu.opLDX(ABS)
	case 0xBE /*ABS,Y+*/ :
		cpu.opLDX(ABSY2)

	/*32.LDY*/
	case 0xA0 /*IMM*/ :
		cpu.opLDY(IMM)
	case 0xA4 /*ZP*/ :
		cpu.opLDY(ZP)
	case 0xB4 /*ZP,X*/ :
		cpu.opLDY(ZPX)
	case 0xAC /*ABS*/ :
		cpu.opLDY(ABS)
	case 0xBC /*ABS,X+*/ :
		cpu.opLDY(ABSX2)

	/*33.LSR*/
	case 0x4A /*ACC*/ :
		cpu.opLSR(ACC)
	case 0x46 /*ZP*/ :
		cpu.opLSR(ZP)
	case 0x56 /*ZP,X*/ :
		cpu.opLSR(ZPX)
	case 0x4E /*ABS*/ :
		cpu.opLSR(ABS)
	case 0x5E /*ABS,X*/ :
		cpu.opLSR(ABSX)

	/*34.NOP*/
	case 0xEA /*IMPL*/ :
		break

	/*35.ORA*/
	case 0x09 /*IMM*/ :
		cpu.opORA(IMM)
	case 0x05 /*ZP*/ :
		cpu.opORA(ZP)
	case 0x15 /*ZP,X*/ :
		cpu.opORA(ZPX)
	case 0x0D /*ABS*/ :
		cpu.opORA(ABS)
	case 0x1D /*ABS,X+*/ :
		cpu.opORA(ABSX2)
	case 0x19 /*ABS,Y+*/ :
		cpu.opORA(ABSY2)
	case 0x01 /*(IND,X)*/ :
		cpu.opORA(INDX)
	case 0x11 /*(IND),Y+*/ :
		cpu.opORA(INDY2)

	/*36.PHA*/
	case 0x48 /*IMPL*/ :
		cpu.opPHA()

	/*37.PHP*/
	case 0x08 /*IMPL*/ :
		cpu.opPHP()

	/*38.PLA*/
	case 0x68 /*IMPL*/ :
		cpu.opPLA()

	/*39.PLP*/
	case 0x28 /*IMPL*/ :
		cpu.opPLP()

	/*40.ROL*/
	case 0x2A /*ACC*/ :
		cpu.opROL(ACC)
	case 0x26 /*ZP*/ :
		cpu.opROL(ZP)
	case 0x36 /*ZP,X*/ :
		cpu.opROL(ZPX)
	case 0x2E /*ABS*/ :
		cpu.opROL(ABS)
	case 0x3E /*ABS,X*/ :
		cpu.opROL(ABSX)

	/*41.ROR*/
	case 0x6A /*ACC*/ :
		cpu.opROR(ACC)
	case 0x66 /*ZP*/ :
		cpu.opROR(ZP)
	case 0x76 /*ZP,X*/ :
		cpu.opROR(ZPX)
	case 0x6E /*ABS*/ :
		cpu.opROR(ABS)
	case 0x7E /*ABS,X*/ :
		cpu.opROR(ABSX)

	/*42.RTI*/
	case 0x40 /*IMPL*/ :
		cpu.opRTI()

	/*43.RTS*/
	case 0x60 /*IMPL*/ :
		cpu.opRTS()

	/*44.SBC*/
	case 0xE9 /*IMM*/ :
		cpu.opSBC(IMM)
	case 0xE5 /*ZP*/ :
		cpu.opSBC(ZP)
	case 0xF5 /*ZP,X*/ :
		cpu.opSBC(ZPX)
	case 0xED /*ABS*/ :
		cpu.opSBC(ABS)
	case 0xFD /*ABS,X+*/ :
		cpu.opSBC(ABSX2)
	case 0xF9 /*ABS,Y+*/ :
		cpu.opSBC(ABSY2)
	case 0xE1 /*(IND,X)*/ :
		cpu.opSBC(INDX)
	case 0xF1 /*(IND),Y+*/ :
		cpu.opSBC(INDY2)

	/*45.SEC*/
	case 0x38 /*IMPL*/ :
		cpu.P = cpu.P | flagC

	/*46.SED*/
	case 0xF8 /*IMPL*/ :
		cpu.P = cpu.P | flagD

	/*47.SEI*/
	case 0x78 /*IMPL*/ :
		cpu.P = cpu.P | flagI

	/*48.STA*/
	case 0x85 /*ZP*/ :
		cpu.opSTA(ZP)
	case 0x95 /*ZP,X*/ :
		cpu.opSTA(ZPX)
	case 0x8D /*ABS*/ :
		cpu.opSTA(ABS)
	case 0x9D /*ABS,X*/ :
		cpu.opSTA(ABSX)
	case 0x99 /*ABS,Y*/ :
		cpu.opSTA(ABSY)
	case 0x81 /*(IND,X)*/ :
		cpu.opSTA(INDX)
	case 0x91 /*(IND),Y*/ :
		cpu.opSTA(INDY)

	/*49.STX*/
	case 0x86 /*ZP*/ :
		cpu.opSTX(ZP)
	case 0x96 /*ZP,Y*/ :
		cpu.opSTX(ZPY)
	case 0x8E /*ABS*/ :
		cpu.opSTX(ABS)

	/*50.STY*/
	case 0x84 /*ZP*/ :
		cpu.opSTY(ZP)
	case 0x94 /*ZP,X*/ :
		cpu.opSTY(ZPX)
	case 0x8C /*ABS*/ :
		cpu.opSTY(ABS)

	/*51.TAX*/
	case 0xAA /*IMPL*/ :
		cpu.opTAX()

	/*52.TAY*/
	case 0xA8 /*IMPL*/ :
		cpu.opTAY()

	/*53.TSX*/
	case 0xBA /*IMPL*/ :
		cpu.opTSX()

	/*54.TXA*/
	case 0x8A /*IMPL*/ :
		cpu.opTXA()

	/*55.TXS*/
	case 0x9A /*IMPL*/ :
		cpu.opTXS()

	/*56.TYA*/
	case 0x98 /*IMPL*/ :
		cpu.opTYA()

		// Unofficial opcodes

	/*DOP*/
	case 0x04 /*ZP*/ :
		cpu.opDOP(ZP)
	case 0x14 /*ZP,X*/ :
		cpu.opDOP(ZPX)
	case 0x34 /*ZP,X*/ :
		cpu.opDOP(ZPX)
	case 0x44 /*ZP*/ :
		cpu.opDOP(ZP)
	case 0x54 /*ZP,X*/ :
		cpu.opDOP(ZPX)
	case 0x64 /*ZP*/ :
		cpu.opDOP(ZP)
	case 0x74 /*ZP,X*/ :
		cpu.opDOP(ZPX)
	case 0x80 /*IMM*/ :
		cpu.opDOP(IMM)
	case 0x82 /*IMM*/ :
		cpu.opDOP(IMM)
	case 0x89 /*IMM*/ :
		cpu.opDOP(IMM)
	case 0xC2 /*IMM*/ :
		cpu.opDOP(IMM)
	case 0xD4 /*ZP,X*/ :
		cpu.opDOP(ZPX)
	case 0xE2 /*IMM*/ :
		cpu.opDOP(IMM)
	case 0xF4 /*ZP,X*/ :
		cpu.opDOP(ZPX)

	/*TOP*/
	case 0x0C /*ABS*/ :
		cpu.opTOP(ABS)
	case 0x1C /*ABS,X*/ :
		cpu.opTOP(ABSX2)
	case 0x3C /*ABS,X*/ :
		cpu.opTOP(ABSX2)
	case 0x5C /*ABS,X*/ :
		cpu.opTOP(ABSX2)
	case 0x7C /*ABS,X*/ :
		cpu.opTOP(ABSX2)
	case 0xDC /*ABS,X*/ :
		cpu.opTOP(ABSX2)
	case 0xFC /*ABS,X*/ :
		cpu.opTOP(ABSX2)

	/*LAX*/
	case 0xA7 /*ZP*/ :
		cpu.opLAX(ZP)
	case 0xB7 /*ZP,Y*/ :
		cpu.opLAX(ZPY)
	case 0xAF /*ABS*/ :
		cpu.opLAX(ABS)
	case 0xBF /*ABS,Y*/ :
		cpu.opLAX(ABSY)
	case 0xA3 /*(IND,X)*/ :
		cpu.opLAX(INDX)
	case 0xB3 /*(IND),Y+*/ :
		cpu.opLAX(INDY2)

	/*AAX*/
	case 0x87 /*ZP*/ :
		cpu.opAAX(ZP)
	case 0x97 /*ZP,Y*/ :
		cpu.opAAX(ZPY)
	case 0x83 /*(IND,X)*/ :
		cpu.opAAX(INDX)
	case 0x8F /*ABS*/ :
		cpu.opAAX(ABS)

	/*SBC*/
	case 0xEB /*IMM*/ :
		cpu.opSBC(IMM)

	/*DCP*/
	case 0xC7 /*ZP*/ :
		cpu.opDCP(ZP)
	case 0xD7 /*ZP,X*/ :
		cpu.opDCP(ZPX)
	case 0xCF /*ABS*/ :
		cpu.opDCP(ABS)
	case 0xDF /*ABS,X*/ :
		cpu.opDCP(ABSX)
	case 0xDB /*ABS,Y*/ :
		cpu.opDCP(ABSY)
	case 0xC3 /*(IND,X)*/ :
		cpu.opDCP(INDX)
	case 0xD3 /*(IND),Y*/ :
		cpu.opDCP(INDY)

	/*ISC*/
	case 0xE7 /*ZP*/ :
		cpu.opISC(ZP)
	case 0xF7 /*ZPX*/ :
		cpu.opISC(ZPX)
	case 0xEF /*ABS*/ :
		cpu.opISC(ABS)
	case 0xFF /*ABS,X*/ :
		cpu.opISC(ABSX)
	case 0xFB /*ABS,Y*/ :
		cpu.opISC(ABSY)
	case 0xE3 /*(IND,X)*/ :
		cpu.opISC(INDX)
	case 0xF3 /*(IND),Y*/ :
		cpu.opISC(INDY)

	/*SLO*/
	case 0x07 /*ZP*/ :
		cpu.opSLO(ZP)
	case 0x17 /*ZP,X*/ :
		cpu.opSLO(ZPX)
	case 0x0F /*ABS*/ :
		cpu.opSLO(ABS)
	case 0x1F /*ABS,X*/ :
		cpu.opSLO(ABSX)
	case 0x1B /*ABS,Y*/ :
		cpu.opSLO(ABSY)
	case 0x03 /*(IND,X)*/ :
		cpu.opSLO(INDX)
	case 0x13 /*(IND),Y*/ :
		cpu.opSLO(INDY)

	/*RLA*/
	case 0x27 /*ZP*/ :
		cpu.opRLA(ZP)
	case 0x37 /*ZP,X*/ :
		cpu.opRLA(ZPX)
	case 0x2F /*ABS*/ :
		cpu.opRLA(ABS)
	case 0x3F /*ABS,X*/ :
		cpu.opRLA(ABSX)
	case 0x3B /*ABS,Y*/ :
		cpu.opRLA(ABSY)
	case 0x23 /*(IND,X)*/ :
		cpu.opRLA(INDX)
	case 0x33 /*(IND),Y*/ :
		cpu.opRLA(INDY)

	/*SRE*/
	case 0x47 /*ZP*/ :
		cpu.opSRE(ZP)
	case 0x57 /*ZP,X*/ :
		cpu.opSRE(ZPX)
	case 0x4F /*ABS*/ :
		cpu.opSRE(ABS)
	case 0x5F /*ABS,X*/ :
		cpu.opSRE(ABSX)
	case 0x5B /*ABS,Y*/ :
		cpu.opSRE(ABSY)
	case 0x43 /*(IND,X)*/ :
		cpu.opSRE(INDX)
	case 0x53 /*(IND),Y*/ :
		cpu.opSRE(INDY)

	/*RRA*/
	case 0x67 /*ZP*/ :
		cpu.opRRA(ZP)
	case 0x77 /*ZP,X*/ :
		cpu.opRRA(ZPX)
	case 0x6F /*ABS*/ :
		cpu.opRRA(ABS)
	case 0x7F /*ABS,X*/ :
		cpu.opRRA(ABSX)
	case 0x7B /*ABS,Y*/ :
		cpu.opRRA(ABSY)
	case 0x63 /*(IND,X)*/ :
		cpu.opRRA(INDX)
	case 0x73 /*(IND),Y*/ :
		cpu.opRRA(INDY)
	}
}

func (cpu *CPU) opADC(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	if (cpu.P&flagD) == 0 || !cpu.decimalModeSupported {
		res := cpu.A + value
		if (cpu.P & flagC) != 0 {
			res++
		}

		cpu.P = cpu.P &^ (flagN | flagV | flagZ | flagC) // Clear flags
		cpu.P = cpu.P | (res & flagN)                    // N
		if ((cpu.A ^ (res & 0xFF)) &^ (cpu.A ^ value) & 0x80) != 0 {
			cpu.P = cpu.P | flagV // V
		}
		if (res & 0xFF) == 0 {
			cpu.P = cpu.P | flagZ // Z
		}
		if res > 0xFF {
			cpu.P = cpu.P | flagC // C
		}

		cpu.A = res & 0xFF
	} else {
		carry := 0
		if (cpu.P & flagC) != 0 {
			carry = 1
		}
		AL := (cpu.A & 15) + (value & 15) + carry // Calculate the lower nybble.
		AH := (cpu.A >> 4) + (value >> 4)
		if AL > 15 {
			AH++
		}

		if AL > 9 {
			AL += 6 // BCD fix up for lower nybble.
		}

		/* Negative and Overflow flags are set with the same logic than in
		Binary mode, but after fixing the lower nybble. */
		if (AH & 8) != 0 {
			cpu.P = cpu.P | flagN // N
		}
		if ((((AH << 4) ^ cpu.A) & 128) != 0) && (((cpu.A ^ value) & 128) == 0) {
			cpu.P = cpu.P | flagV // V
		}
		// Z flag is set just like in Binary mode.
		if cpu.A+value+carry != 0 {
			cpu.P = cpu.P | flagZ // Z
		}

		if AH > 9 {
			AH += 6 // BCD fix up for upper nybble.
		}
		/* Carry is the only flag set after fixing the result. */
		if AH > 15 {
			cpu.P = cpu.P | flagC // C
		}

		cpu.A = ((AH << 4) | (AL & 15)) & 255
	}
}

func (cpu *CPU) opAND(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	cpu.A = cpu.A & value
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (cpu.A & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if cpu.A == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
}

func (cpu *CPU) opASL(mode int) {
	if mode == ACC {
		cpu.A = cpu.opASLInt(cpu.A)
	} else {
		address := cpu.calculateMemoryAddress(mode)
		value := cpu.readMemory(address)

		newValue := cpu.opASLInt(value)
		cpu.writeMemory(address, newValue)
	}
}

func (cpu *CPU) opASLInt(value int) int {
	res := (value << 1) & 0xFF
	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags
	if (res & 0x80) > 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x80) > 0 {
		cpu.P = cpu.P | flagC // C
	}

	return res
}

func (cpu *CPU) opBCC(mode int) {
	cpu.opBranch((cpu.P&flagC) == 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBCS(mode int) {
	cpu.opBranch((cpu.P&flagC) != 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBEQ(mode int) {
	cpu.opBranch((cpu.P&flagZ) != 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBMI(mode int) {
	cpu.opBranch((cpu.P&flagN) != 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBNE(mode int) {
	cpu.opBranch((cpu.P&flagZ) == 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBPL(mode int) {
	cpu.opBranch((cpu.P&flagN) == 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBVC(mode int) {
	cpu.opBranch((cpu.P&flagV) == 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBVS(mode int) {
	cpu.opBranch((cpu.P&flagV) != 0, cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opBranch(condition bool, jumpAddress int) {
	if condition {
		if isPageBoundaryCrossed(cpu.PC, jumpAddress) {
			cpu.OpCycles += 2
		} else {
			cpu.OpCycles++
		}

		cpu.PC = jumpAddress
	}
}

func (cpu *CPU) opBIT(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	cpu.P = cpu.P &^ (flagN | flagV | flagZ) // Clear flags
	if (value & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if (value & 0x40) != 0 {
		cpu.P = cpu.P | flagV // V
	}
	if (cpu.A & value) == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
}

func (cpu *CPU) opBRK() {
	cpu.PC++                // skip next bite (usually it is a NOP or number that is analyzed by the interrupt handler)
	cpu.push(cpu.PC >> 8)   // push high bits
	cpu.push(cpu.PC & 0xFF) // push low bits
	cpu.P = cpu.P | flagB   // B
	cpu.push(cpu.P)
	cpu.PC = (cpu.readMemory(0xFFFF) << 8) | cpu.readMemory(0xFFFE)
}

func (cpu *CPU) opCMP(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.opCompare(cpu.A, value)
}

func (cpu *CPU) opCPX(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.opCompare(cpu.X, value)
}

func (cpu *CPU) opCPY(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.opCompare(cpu.Y, value)
}

func (cpu *CPU) opCompare(register int, value int) {
	res := (register - value) & 0xFF
	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags
	if (res & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if register >= value {
		cpu.P = cpu.P | flagC // C
	}
}

func (cpu *CPU) opDEC(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.writeMemory(address, cpu.opDecrease(value))
}

func (cpu *CPU) opDecrease(value int) int {
	res := (value - 1) & 0xFF
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (res & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}

	return res
}

func (cpu *CPU) opEOR(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.opEORInt(value)
}

func (cpu *CPU) opEORInt(value int) {
	cpu.A = cpu.A ^ value
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (cpu.A & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if cpu.A == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
}

func (cpu *CPU) opINC(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.writeMemory(address, cpu.opIncrease(value))
}

func (cpu *CPU) opIncrease(value int) int {
	res := (value + 1) & 0xFF
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (res & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}

	return res
}

func (cpu *CPU) opJMP(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	cpu.PC = address
}

func (cpu *CPU) opJSR(mode int) {
	address := cpu.calculateMemoryAddress(mode)

	cpu.PC--
	cpu.push(cpu.PC >> 8)
	cpu.push(cpu.PC & 0xFF)
	cpu.PC = address
}

func (cpu *CPU) opLDA(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	cpu.A = cpu.opLoad(value)
}

func (cpu *CPU) opLDX(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	cpu.X = cpu.opLoad(value)
}

func (cpu *CPU) opLDY(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	cpu.Y = cpu.opLoad(value)
}

func (cpu *CPU) opLoad(value int) int {
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (value & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if value == 0 {
		cpu.P = cpu.P | flagZ // Z
	}

	return value
}

func (cpu *CPU) opLSR(mode int) {
	if mode == ACC {
		cpu.A = cpu.opLSRInt(cpu.A)
	} else {
		address := cpu.calculateMemoryAddress(mode)
		value := cpu.readMemory(address)

		newValue := cpu.opLSRInt(value)
		cpu.writeMemory(address, newValue)
	}
}

func (cpu *CPU) opLSRInt(value int) int {
	res := 0x7F & (value >> 1)
	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // clear flags
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x01) != 0 {
		cpu.P = cpu.P | flagC // C
	}

	return res
}

func (cpu *CPU) opORA(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)
	cpu.opORAInt(value)
}

func (cpu *CPU) opORAInt(value int) {
	cpu.A = cpu.A | value
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (cpu.A & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if cpu.A == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
}

func (cpu *CPU) opPHA() {
	cpu.push(cpu.A)
}

func (cpu *CPU) opPHP() {
	cpu.push(cpu.P | flagB)
}

func (cpu *CPU) opPLA() {
	cpu.A = cpu.pop()

	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	cpu.P = cpu.P | (cpu.A & flagN)  // N
	if cpu.A == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
}

func (cpu *CPU) opPLP() {
	cpu.P = (cpu.pop() &^ flagB) | flagR
}

func (cpu *CPU) opROL(mode int) {
	if mode == ACC {
		cpu.A = cpu.opROLInt(cpu.A)
	} else {
		address := cpu.calculateMemoryAddress(mode)
		value := cpu.readMemory(address)

		newValue := cpu.opROLInt(value)
		cpu.writeMemory(address, newValue)
	}
}

func (cpu *CPU) opROLInt(value int) int {
	res := (value << 1) & 0xFF
	if (cpu.P & flagC) != 0 {
		res = res | 1
	}
	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags

	if (res & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x80) != 0 {
		cpu.P = cpu.P | flagC // C
	}

	return res
}

func (cpu *CPU) opROR(mode int) {
	if mode == ACC {
		cpu.A = cpu.opRORInt(cpu.A)
	} else {
		address := cpu.calculateMemoryAddress(mode)
		value := cpu.readMemory(address)

		newValue := cpu.opRORInt(value)
		cpu.writeMemory(address, newValue)
	}
}

func (cpu *CPU) opRORInt(value int) int {
	res := (value >> 1) & 0xFF
	if (cpu.P & flagC) != 0 {
		res = res | 0x80
	}
	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags
	if (res & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x01) != 0 {
		cpu.P = cpu.P | flagC // C
	}

	return res
}

func (cpu *CPU) opRTI() {
	cpu.P = cpu.pop() &^ flagB
	cpu.P = cpu.P | flagR
	PCL := cpu.pop()
	PCH := cpu.pop()
	cpu.PC = (PCH << 8) | PCL
}

func (cpu *CPU) opRTS() {
	PCL := cpu.pop()
	PCH := cpu.pop()
	cpu.PC = (PCH << 8) | PCL
	cpu.PC++
}

func (cpu *CPU) opSBC(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	value := cpu.readMemory(address)

	if (cpu.P&flagD) == 0 || !cpu.decimalModeSupported {
		res := cpu.A - value
		if (cpu.P & flagC) == 0 {
			res--
		}

		cpu.P = cpu.P &^ (flagN | flagV | flagZ | flagC) // Clear flags
		cpu.P = cpu.P | (res & flagN)                    // N
		if ((cpu.A ^ value) & (cpu.A ^ (0xFF & res)) & 0x80) != 0 {
			cpu.P = cpu.P | flagV // V
		}
		if (res & 0xFF) == 0 {
			cpu.P = cpu.P | flagZ // Z
		}
		if (res & 0x100) == 0 {
			cpu.P = cpu.P | flagC // C
		}
		// if (res > 0xFF) {
		// 	cpu.P = cpu.P | flagC // C
		// }

		cpu.A = 0xFF & res
	} else {
		borrow := 0
		if (cpu.P & flagC) == 0 {
			borrow = 1
		}
		AL := (cpu.A & 15) - (value & 15) - borrow // Calculate the lower nybble.
		if (AL & 16) != 0 {
			AL -= 6 // BCD fix up for lower nybble.
		}

		AH := (cpu.A >> 4) - (value >> 4) - (AL & 16) // Calculate the upper nybble.
		if (AH & 16) != 0 {
			AH -= 6 // BCD fix up for upper nybble.
		}

		if ((cpu.A - value - borrow) & 128) != 0 {
			cpu.P = cpu.P | flagN // N
		}
		if (((cpu.A-value-borrow)^value)&128) != 0 && ((cpu.A^value)&128) != 0 {
			cpu.P = cpu.P | flagV // V
		}
		if ((cpu.A - value - borrow) & 255) != 0 {
			cpu.P = cpu.P | flagZ // Z
		}
		if ((cpu.A - value - borrow) & 256) != 0 {
			cpu.P = cpu.P | flagC // C
		}

		cpu.A = ((AH << 4) | (AL & 15)) & 255
	}
}

func (cpu *CPU) opSTA(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	cpu.writeMemory(address, cpu.A)
}

func (cpu *CPU) opSTX(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	cpu.writeMemory(address, cpu.X)
}

func (cpu *CPU) opSTY(mode int) {
	address := cpu.calculateMemoryAddress(mode)
	cpu.writeMemory(address, cpu.Y)
}

func (cpu *CPU) opTAX() {
	cpu.X = cpu.A
	cpu.transfer(cpu.X)
}

func (cpu *CPU) opTAY() {
	cpu.Y = cpu.A
	cpu.transfer(cpu.Y)
}

func (cpu *CPU) opTSX() {
	cpu.X = cpu.S
	cpu.transfer(cpu.X)
}

func (cpu *CPU) opTXA() {
	cpu.A = cpu.X
	cpu.transfer(cpu.A)
}

func (cpu *CPU) opTXS() {
	cpu.S = cpu.X
}

func (cpu *CPU) opTYA() {
	cpu.A = cpu.Y
	cpu.transfer(cpu.A)
}

func (cpu *CPU) transfer(toRegister int) {
	cpu.P = cpu.P &^ (flagN | flagZ) // Clear flags
	if (toRegister & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if toRegister == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
}

func (cpu *CPU) opDOP(mode int) {
	/*DOP double NOP*/
	cpu.readMemory(cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opTOP(mode int) {
	/*TOP triple NOP*/
	cpu.readMemory(cpu.calculateMemoryAddress(mode))
}

func (cpu *CPU) opLAX(mode int) {
	/*LAX Load accumulator and X register with memory
	  Status flags: N,Z
	*/

	address := cpu.calculateMemoryAddress(mode)
	cpu.A = cpu.opLoad(cpu.readMemory(address))
	cpu.X = cpu.A
}

func (cpu *CPU) opAAX(mode int) {
	/*AAX (SAX) [AXS] AND X register with accumulator and store result in memory. */
	address := cpu.calculateMemoryAddress(mode)

	result := cpu.A & cpu.X
	cpu.writeMemory(address, result)
}

func (cpu *CPU) opDCP(mode int) {
	/*DCP (DCP) [DCM]*/

	address := cpu.calculateMemoryAddress(mode)

	value := cpu.readMemory(address)
	value = 0xFF & (value - 1)
	//      if (value != 0) {
	//         value--
	//      } else {
	//         value := 0xFF
	//      }

	valueToTest := cpu.A - value

	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags
	if (valueToTest & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if valueToTest == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (valueToTest & 0x100) == 0 {
		cpu.P = cpu.P | flagC // C
	}

	cpu.writeMemory(address, value)
}

func (cpu *CPU) opISC(mode int) {
	/*ISC (ISB) [INS] Increase memory by one, then subtract memory from accumulator (with
	  borrow). Status flags: N,V,Z,C*/

	address := cpu.calculateMemoryAddress(mode)

	value := cpu.readMemory(address)
	value = 0xFF & (value + 1)

	//result = cpu.A - value + ((cpu.P & flagC) != 0 ? 1 : 0)
	result := cpu.A - value
	if (cpu.P & flagC) == 0 {
		result--
	}

	cpu.P = cpu.P &^ (flagN | flagV | flagZ | flagC) // Clear flags
	if ((cpu.A ^ value) & (cpu.A ^ (result & 0xFF)) & 0x80) != 0 {
		cpu.P = cpu.P | flagV // V
	}
	if (result & 0x100) == 0 {
		cpu.P = cpu.P | flagC // C
	}

	cpu.A = result & 0xFF
	if (cpu.A & 0xFF) == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	cpu.P = cpu.P | (cpu.A & flagN) // N

	//      res = cpu.A - value - ((cpu.P & flagC) != 0 ? 0 : 1)

	//      cpu.P = cpu.P &^ (flagN | flagV | flagZ | flagC) // Clear flags
	//      cpu.P = cpu.P | (res & flagN) // N
	//      cpu.P = cpu.P | (((cpu.A ^ value) & (cpu.A ^ (0xFF & res)) & 0x80) != 0 ? flagV : 0) // V
	//      cpu.P = cpu.P | ((res & 0xFF) == 0 ? flagZ // Z
	//      cpu.P = cpu.P | ((res & 0x100) != 0 ? 0: flagC) // C
	//
	//      cpu.A = 0xFF & res

	cpu.writeMemory(address, value)
}

func (cpu *CPU) opSLO(mode int) {
	/*SLO (SLO) [ASO]
	Shift left one bit in memory, then OR accumulator with memory. =
	Status flags: N,Z,C*/
	address := cpu.calculateMemoryAddress(mode)

	value := cpu.readMemory(address)
	result := (value << 1) & 0xFF

	cpu.A = cpu.A | result

	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags
	cpu.P = cpu.P | (cpu.A & flagN)          // N
	if cpu.A == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x80) != 0 {
		cpu.P = cpu.P | flagC // C
	}

	cpu.writeMemory(address, result)
}

func (cpu *CPU) opRLA(mode int) {
	/*RLA (RLA) [RLA]
	Rotate one bit left in memory, then AND accumulator with memory. Status
	flags: N,Z,C */
	address := cpu.calculateMemoryAddress(mode)

	value := cpu.readMemory(address)

	res := (value << 1) & 0xFF
	if (cpu.P & flagC) != 0 {
		res = res | 1
	}

	cpu.A = cpu.A & res

	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // Clear flags
	if (cpu.A & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if cpu.A == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x80) != 0 {
		cpu.P = cpu.P | flagC // C
	}

	cpu.writeMemory(address, res)
}

func (cpu *CPU) opSRE(mode int) {
	/*SRE (SRE) [LSE]
	Shift right one bit in memory, then EOR accumulator with memory. Status
	flags: N,Z,C*/
	address := cpu.calculateMemoryAddress(mode)

	value := cpu.readMemory(address)

	res := 0x7F & (value >> 1)
	cpu.A = cpu.A ^ res

	cpu.P = cpu.P &^ (flagN | flagZ | flagC) // clear flags
	if (cpu.A & 0x80) != 0 {
		cpu.P = cpu.P | flagN // N
	}
	if res == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if (value & 0x01) != 0 {
		cpu.P = cpu.P | flagC // C
	}

	cpu.writeMemory(address, res)
}

func (cpu *CPU) opRRA(mode int) {
	/*RRA (RRA) [RRA]
	Rotate one bit right in memory, then add memory to accumulator (with carry).
	Status flags: N,V,Z,C*/
	address := cpu.calculateMemoryAddress(mode)

	value1 := cpu.readMemory(address)
	value := (value1 >> 1) & 0xFF
	if (cpu.P & flagC) != 0 {
		value = value | 0x80
	}

	res := cpu.A + value
	if (value1 & 0x01) != 0 {
		res++
	}

	cpu.P = cpu.P &^ (flagN | flagV | flagZ | flagC) // Clear flags
	cpu.P = cpu.P | (res & flagN)                    // N
	if (((cpu.A ^ (res & 0xFF)) &^ (cpu.A ^ value)) & 0x80) != 0 {
		cpu.P = cpu.P | flagV // V
	}

	if (res & 0xFF) == 0 {
		cpu.P = cpu.P | flagZ // Z
	}
	if res > 0xFF {
		cpu.P = cpu.P | flagC // C
	}

	cpu.A = res & 0xFF

	cpu.writeMemory(address, value)
}
