package main

import (
    "fmt"
    "math/rand"
    "time"
)

const BrailleBase = 0x2800

const NRows = 80
const NCols = 100
const NTextRows = NRows / 4
const NTextCols = NCols / 2

func coordToBrailleIndex(row, col int) int {
    if row == 0 {
        return []int{0, 1, 2, 6}[col]
    } else {
        return []int{3, 4, 5, 7}[col]
    }
}

func renderPixelsToBraille(pixels *[NRows][NCols]bool, textBuf *[NTextRows][NTextCols]rune) {
    for i := range pixels {
        for j, pixel := range pixels[i] {
            textBuf[i / 4][j / 2] &= ^(1 << coordToBrailleIndex(j % 2, i % 4))
            if pixel {
                textBuf[i / 4][j / 2] |= (1 << coordToBrailleIndex(j % 2, i % 4))
            }
        }
    }
}

func printTextBuffer(textBuf *[NTextRows][NTextCols]rune) {
    for i := range textBuf {
        for j := range textBuf[i] {
            fmt.Print(string(textBuf[i][j]))
        }
        fmt.Println()
    }
}

func moveCursorUp(lines int) {
    fmt.Printf("\x1b[%dA", lines)
}

func nLiveNeighbors(pixels *[NRows][NCols]bool, row, col int) int {
    var n int

    yOffs := []int{0}
    if row != 0 {
        yOffs = append(yOffs, -1)
    }
    if row != len(pixels) - 1 {
        yOffs = append(yOffs, 1)
    }

    xOffs := []int{0}
    if col != 0 {
        xOffs = append(xOffs, -1)
    }
    if col != len(pixels[0]) - 1 {
        xOffs = append(xOffs, 1)
    }

    for _, yOff := range yOffs {
        for _, xOff := range xOffs {
            if (yOff != 0 || xOff != 0) && pixels[row + yOff][col + xOff]{
                n++
            }
        }
    }

    return n
}

func permuteGOL(pixels *[NRows][NCols]bool) {
    oldPixels := *pixels
    for row := range pixels {
        for col := range pixels[row] {
            alive := oldPixels[row][col]
            neighbors := nLiveNeighbors(&oldPixels, row, col)

            switch {
            case alive && (neighbors < 2 || neighbors > 3): pixels[row][col] = false
            case !alive && neighbors == 3: pixels[row][col] = true
            }
        }
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())

    var pixels [NRows][NCols]bool
    for row := range pixels {
        for col := range pixels[row] {
            pixels[row][col] = rand.Intn(2) == 1
        }
    }

    var textBuf [NTextRows][NTextCols]rune
    for row := range textBuf {
        for col := range textBuf[row] {
            textBuf[row][col] = BrailleBase
        }
    }

    for {
        renderPixelsToBraille(&pixels, &textBuf)
        printTextBuffer(&textBuf)

        permuteGOL(&pixels)

        time.Sleep(250 * time.Millisecond)
        moveCursorUp(len(textBuf))
    }
}
