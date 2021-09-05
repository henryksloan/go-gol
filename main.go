package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/pborman/getopt/v2"
)

const BrailleBase = 0x2800

func coordToBrailleIndex(row, col int) int {
	if row == 0 {
		return []int{0, 1, 2, 6}[col]
	} else {
		return []int{3, 4, 5, 7}[col]
	}
}

func renderPixelsToBraille(pixels *[][]bool, textBuf *[][]rune) {
	for i := range *pixels {
		for j, pixel := range (*pixels)[i] {
			(*textBuf)[i/4][j/2] &= ^(1 << coordToBrailleIndex(j%2, i%4))
			if pixel {
				(*textBuf)[i/4][j/2] |= (1 << coordToBrailleIndex(j%2, i%4))
			}
		}
	}
}

func printTextBuffer(textBuf *[][]rune) {
	for i := range *textBuf {
		for j := range (*textBuf)[i] {
			fmt.Print(string((*textBuf)[i][j]))
		}
		fmt.Println()
	}
}

func redrawTextBuffer(pixels *[][]bool, textBuf *[][]rune, cursorX, cursorY int) {
	renderPixelsToBraille(pixels, textBuf)
	moveCursor(-cursorY, -cursorX)
	printTextBuffer(textBuf)
	moveCursorUp(len(*textBuf))
	moveCursor(cursorY, cursorX)
}

func setCursorVisible(visible bool) {
	var command rune
	if visible {
		command = 'h'
	} else {
		command = 'l'
	}
	fmt.Printf("\x1b[?25%c", command)
}

func moveCursorUp(lines int) {
	fmt.Printf("\x1b[%dA", lines)
}

func moveCursor(rows, cols int) {
	if rows < 0 {
		fmt.Printf("\x1b[%dA", -rows)
	} else if rows > 0 {
		fmt.Printf("\x1b[%dB", rows)
	}

	if cols < 0 {
		fmt.Printf("\x1b[%dD", -cols)
	} else if cols > 0 {
		fmt.Printf("\x1b[%dC", cols)
	}
}

func randomizePixels(pixels *[][]bool) {
	for row := range *pixels {
		for col := range (*pixels)[row] {
			(*pixels)[row][col] = rand.Intn(2) == 1
		}
	}
}

func togglePixelUnderCursor(pixels *[][]bool, textBuf *[][]rune, cursorX, cursorY int, pixelRow, pixelCol int) {
	pixel := &((*pixels)[cursorY*4+pixelRow][cursorX*2+pixelCol])
	*pixel = !(*pixel)
	redrawTextBuffer(pixels, textBuf, cursorX, cursorY)
}

func clearUnderCursor(pixels *[][]bool, textBuf *[][]rune, cursorX, cursorY int) {
	for pixelRow := 0; pixelRow < 4; pixelRow++ {
		for pixelCol := 0; pixelCol < 2; pixelCol++ {
			(*pixels)[cursorY*4+pixelRow][cursorX*2+pixelCol] = false
		}
	}
	redrawTextBuffer(pixels, textBuf, cursorX, cursorY)
}

func nLiveNeighbors(pixels *[][]bool, row, col int) int {
	var n int

	yOffs := []int{0}
	if row != 0 {
		yOffs = append(yOffs, -1)
	}
	if row != len(*pixels)-1 {
		yOffs = append(yOffs, 1)
	}

	xOffs := []int{0}
	if col != 0 {
		xOffs = append(xOffs, -1)
	}
	if col != len((*pixels)[0])-1 {
		xOffs = append(xOffs, 1)
	}

	for _, yOff := range yOffs {
		for _, xOff := range xOffs {
			if (yOff != 0 || xOff != 0) && (*pixels)[row+yOff][col+xOff] {
				n++
			}
		}
	}

	return n
}

func permuteGOL(pixels *[][]bool) {
	oldPixels := make([][]bool, len(*pixels))
	for i := range *pixels {
		oldPixels[i] = make([]bool, len((*pixels)[i]))
		copy(oldPixels[i], (*pixels)[i])
	}

	for row := range *pixels {
		for col := range (*pixels)[row] {
			alive := oldPixels[row][col]
			neighbors := nLiveNeighbors(&oldPixels, row, col)

			switch {
			case alive && (neighbors < 2 || neighbors > 3):
				(*pixels)[row][col] = false
			case !alive && neighbors == 3:
				(*pixels)[row][col] = true
			}
		}
	}
}

func exit() {
	setCursorVisible(true)
	os.Exit(0)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	size := getopt.StringLong("size", 's', "100x80", "the size of the game board, <width>x<height>")
	tickInt := getopt.IntLong("tick", 't', 50, "the time between ticks (in milliseconds)")
	noNumpad := getopt.BoolLong("no-numpad", 'n', "turn on key bindings that don't require a numpad")
	help := getopt.BoolLong("help", 'h', "show this message")
	getopt.Parse()

	if *help {
		getopt.Usage()
		return
	}

	tick := time.Duration(*tickInt)

	var cellToggleKeys [][]rune
	if *noNumpad {
		cellToggleKeys = [][]rune{
			{'1', '2'},
			{'3', '4'},
			{'5', '6'},
			{'7', '8'},
		}
	} else {
		cellToggleKeys = [][]rune{
			{'/', '*'},
			{'8', '9'},
			{'5', '6'},
			{'2', '3'},
		}
	}

	sizeSplit := strings.Split(*size, "x")
	const sizeParseError = "size must be in the form <width>x<height>"
	if len(sizeSplit) != 2 {
		fmt.Println(sizeParseError)
		os.Exit(2)
	}
	nRows, err := strconv.Atoi(sizeSplit[1])
	if err != nil {
		fmt.Println(sizeParseError)
		os.Exit(2)
	}
	nCols, err := strconv.Atoi(sizeSplit[0])
	if err != nil {
		fmt.Println(sizeParseError)
		os.Exit(2)
	}

	var NTextRows = int(math.Ceil(float64(nRows) / 4.0))
	var NTextCols = int(math.Ceil(float64(nCols) / 2.0))

	pixels := make([][]bool, nRows)
	for i := range pixels {
		pixels[i] = make([]bool, nCols)
	}
	randomizePixels(&pixels)

	textBuf := make([][]rune, NTextRows)
	for row := range textBuf {
		textBuf[row] = make([]rune, NTextCols)
		for col := range textBuf[row] {
			textBuf[row][col] = BrailleBase
		}
	}

	// When the user terminates the process, show the cursor
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			exit()
		}
	}()

	keyChan := make(chan struct {
		rune
		keyboard.Key
	}, 1)
	go func() {
		for {
			char, key, err := keyboard.GetSingleKey()
			if err == nil {
				keyChan <- struct {
					rune
					keyboard.Key
				}{char, key}
			}
		}
	}()

	pause := false
	pauseCursorX, pauseCursorY := 0, 0
	setCursorVisible(false)
	for {
	charLoop:
		for {
			select {
			case press := <-keyChan:
				char := press.rune
				key := press.Key
				switch char {
				case 'q':
					exit()
				case 'c':
					for i := range pixels {
						pixels[i] = make([]bool, nCols)
					}
					if pause {
						redrawTextBuffer(&pixels, &textBuf, pauseCursorX, pauseCursorY)
					}
				}

				if char == 0 {
					switch key {
					case keyboard.KeyCtrlC:
						exit()
					case keyboard.KeySpace:
						if pause {
							pause = false
							moveCursor(-pauseCursorY, -pauseCursorX)
						} else {
							pause = true
							moveCursor(pauseCursorY, pauseCursorX)
						}
						setCursorVisible(pause)
					}
				}

				if pause {
					for row := range cellToggleKeys {
						for col, key := range cellToggleKeys[row] {
							if char == key {
								togglePixelUnderCursor(&pixels, &textBuf, pauseCursorX, pauseCursorY, row, col)
							}
						}
					}

					switch char {
					case 'k':
						if pauseCursorY > 0 {
							pauseCursorY -= 1
							moveCursor(-1, 0)
						}
					case 'j':
						if pauseCursorY < NTextRows-1 {
							pauseCursorY += 1
							moveCursor(1, 0)
						}
					case 'l':
						if pauseCursorX < NTextCols-1 {
							pauseCursorX += 1
							moveCursor(0, 1)
						}
					case 'h':
						if pauseCursorX > 0 {
							pauseCursorX -= 1
							moveCursor(0, -1)
						}

					case '0':
						clearUnderCursor(&pixels, &textBuf, pauseCursorX, pauseCursorY)
					}
				}
			default:
				break charLoop
			}
		}

		if !pause {
			permuteGOL(&pixels)

			renderPixelsToBraille(&pixels, &textBuf)
			printTextBuffer(&textBuf)

			time.Sleep(tick * time.Millisecond)
			moveCursorUp(len(textBuf))
		}
	}
}
