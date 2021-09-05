# go-gol
A terminal-based implementation of Conway's Game of Life using Go and Unicode braille characters 

## Usage
`go run .`

or use `go run . --help` to see the different options

### Controls
Press `q` to pause/unpause the simulation  
Press `c` to clear the board  
Press `space` to pause/unpause the simulation

When paused:  
Use `hjkl` (vi keys) to move the cursor  
Press `0` to clear the selected 2x4 cell block  
Toggle cells within the current 2x4 cell block with the numpad (NumLock on):
```
/*
89
56
23
```

If you don't have a numpad, use the `--no-numpad` option, and use the keys:
```
12
34
56
78
```

![Screenshot of a large game of life board running in a terminal](/screenshots/gol.png)
