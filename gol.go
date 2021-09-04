package main

import (
    "fmt"
    "math"
    "math/rand"
    "time"
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
            (*textBuf)[i / 4][j / 2] &= ^(1 << coordToBrailleIndex(j % 2, i % 4))
            if pixel {
                (*textBuf)[i / 4][j / 2] |= (1 << coordToBrailleIndex(j % 2, i % 4))
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
    if row != len(*pixels) - 1 {
        yOffs = append(yOffs, 1)
    }

    xOffs := []int{0}
    if col != 0 {
        xOffs = append(xOffs, -1)
    }
    if col != len((*pixels)[0]) - 1 {
        xOffs = append(xOffs, 1)
    }

    for _, yOff := range yOffs {
        for _, xOff := range xOffs {
            if (yOff != 0 || xOff != 0) && (*pixels)[row + yOff][col + xOff]{
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
            case alive && (neighbors < 2 || neighbors > 3): (*pixels)[row][col] = false
            case !alive && neighbors == 3: (*pixels)[row][col] = true
            }
        }
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())

    var NRows = 80
    var NCols = 100
    var NTextRows = int(math.Ceil(float64(NRows) / 4.0))
    var NTextCols = int(math.Ceil(float64(NCols) / 2.0))

    pixels := make([][]bool, NRows)
    for i := range pixels {
        pixels[i] = make([]bool, NCols)
    }
    randomizePixels(&pixels)

    textBuf := make([][]rune, NTextRows)
    for row := range textBuf {
        textBuf[row] = make([]rune, NTextCols)
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
