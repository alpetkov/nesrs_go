package cpu

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/alpetkov/nesrs_go/nesrs/cartridge"
)

func TestNestest(t *testing.T) {
	lines := readNestestLogLines("./nestest.log")

	cartridge := readNestestRom("./nestest.nes")

	cpuMemory := NESCPUMemory{cartridge: cartridge}
	cpu := New(&cpuMemory)

	cpu.A = 0x00
	cpu.X = 0x00
	cpu.Y = 0x00
	cpu.S = 0xFD
	cpu.P = 0x24
	cpu.PC = 0xC000
	cpu.OpCycles = 0

	cycles := 0
	for _, line := range lines {
		actualLogLine := ""
		actualLogLine += toHex(cpu.PC)
		actualLogLine += "    "
		actualLogLine += "A:" + toHex(cpu.A)
		actualLogLine += " "
		actualLogLine += "X:" + toHex(cpu.X)
		actualLogLine += " "
		actualLogLine += "Y:" + toHex(cpu.Y)
		actualLogLine += " "
		actualLogLine += "P:" + toHex(cpu.P)
		actualLogLine += " "
		actualLogLine += "SP:" + toHex(cpu.S)

		cycles += cpu.OpCycles * 3
		cycles %= 341
		actualLogLine += " "
		actualLogLine += "CYC:" + pad(cycles)

		if line != actualLogLine {
			t.Errorf("\nWrong %v\nRight %v", actualLogLine, line)
		}

		cpu.ExecuteOp()
	}

	fmt.Println("All good!")
}

func readNestestLogLines(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		newLine := line[0:4]
		newLine += "    "
		newLine += line[strings.Index(line, "A:"):strings.Index(line, " SL")]

		lines = append(lines, newLine)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}

func readNestestRom(filePath string) *cartridge.Cartridge {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	return cartridge.New(file)
}

func toHex(value int) string {
	res := fmt.Sprintf("%X", value)
	if len(res)%2 != 0 {
		res = "0" + res
	}

	return res
}

func pad(value int) string {
	valueStr := fmt.Sprintf("%v", value)
	if value < 10 {
		return "  " + valueStr
	} else if value < 100 {
		return " " + valueStr
	} else {
		return valueStr
	}
}
