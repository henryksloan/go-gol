package main

/*
#ifdef _WIN32
#include<conio.h>
#endif

#ifdef linux
#include <stdio.h>
#include <unistd.h>
#include <termios.h>
char getch(){
    char ch = 0;
    struct termios old = {0};
    fflush(stdout);
    if( tcgetattr(0, &old) < 0 ) perror("tcsetattr()");
    old.c_lflag &= ~ICANON;
    old.c_lflag &= ~ECHO;
    old.c_cc[VMIN] = 1;
    old.c_cc[VTIME] = 0;
    if( tcsetattr(0, TCSANOW, &old) < 0 ) perror("tcsetattr ICANON");
    if( read(0, &ch,1) < 0 ) perror("read()");
    old.c_lflag |= ICANON;
    old.c_lflag |= ECHO;
    if(tcsetattr(0, TCSADRAIN, &old) < 0) perror("tcsetattr ~ICANON");
    return ch;
}
#endif
*/
import "C"

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

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

func randomizePixels(pixels *[][]bool) {
	for row := range *pixels {
		for col := range (*pixels)[row] {
			(*pixels)[row][col] = rand.Intn(2) == 1
		}
	}
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
	getopt.Parse()
	tick := time.Duration(*tickInt)

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

	getCharChan := make(chan rune, 1)
	go func() {
		for {
			getCharChan <- rune(C.getch())
		}
	}()

	pause := false
	setCursorVisible(false)
	for {
		select {
		case char := <-getCharChan:
			switch char {
			case 27: // Escape key
				exit()
			case ' ':
				pause = !pause
			}
		default:
		}

		if !pause {
			renderPixelsToBraille(&pixels, &textBuf)
			printTextBuffer(&textBuf)

			permuteGOL(&pixels)

			time.Sleep(tick * time.Millisecond)
			moveCursorUp(len(textBuf))
		}
	}
}
