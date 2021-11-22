package main

import (
	"bufio"
	"canvas"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"
)

type ColorID int
type State int
type Direction int

const (
	ForwardDir = iota
	RightDir
	BackwardDir
	LeftDir
)

const (
	NorthDir = ForwardDir
	SouthDir = BackwardDir
)

type Signal struct {
	state State
	color ColorID
}

type Action struct {
	state State
	color ColorID
	turn  Direction
}

type Turmite struct {
	rules      map[Signal]Action // rules store the rules stored in the mite file, once read, retain all the time in this object
	x, y       int               // position
	currentDir Direction
	state      State
}

type Field [][]ColorID

// NewField creates a new square field of the given edge size.
func NewField(size int) Field {
	f := make([][]ColorID, size)
	for i := range f {
		f[i] = make([]ColorID, size)
	}
	return f
}

// DrawField draws the field to a PNG in the given filename. Assumes that
// field[y][x] is at the cell (y,x), where the origin is in the top-right
// corner.
func (f Field) DrawField(filename string) {
	const scale = 5
	n := len(f)
	out := canvas.CreateNewCanvas(n*scale, n*scale)

	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			out.SetFillColor(f[y][x].ToColor())
			out.ClearRect(x*scale, y*scale, (x+1)*scale, (y+1)*scale)
		}
	}

	out.SaveToPNG(filename)
}

// ToRGB returns the red, green, blue values for a given color id.
func (c ColorID) ToColor() color.Color {
	colors := [][]uint8{
		{0, 0, 0},
		{125, 0, 0},
		{0, 125, 0},
		{0, 0, 125},
		{125, 0, 125},
		{255, 255, 255},
	}
	return canvas.MakeColor(colors[c][0], colors[c][1], colors[c][2])
}

// DirFromString returns a direction constant given an English string.
func DirFromString(s string) (Direction, error) {
	switch strings.ToLower(s) {
	case "forward", "f":
		return ForwardDir, nil
	case "backward", "back":
		return BackwardDir, nil
	case "left", "l":
		return LeftDir, nil
	case "right", "r":
		return RightDir, nil
	default:
		return 0, fmt.Errorf("unknown direction type: %s", s)
	}
}

// PositiveMod computes n % m, returning a number in [0,m-1].
func PositiveMod(n, m int) int {
	return ((n % m) + m) % m
}

// Left returns the direction turing 90 degrees left of d.
func (d Direction) Left() Direction {
	return Direction(PositiveMod(int(d)-1, 4))
}

// Right returns the direction turning 90 degrees right of d.
func (d Direction) Right() Direction {
	return Direction(PositiveMod(int(d)+1, 4))
}

// ReadTurmite reads a file that specifies the turmite rules. The file should
// have lines of the format:
//
//  state color -> state color direction
//
// where state is a lowercase letter a-z; color is an integer;  direction is a
// direction understood by DirFromString. The returned Turmite will be
// positioned at the center of the field and facing north (aka ForwardDir).
func ReadTurmite(filename string, size int) (*Turmite, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// the initial state, center, half x and half y, facing north
	tur := Turmite{
		x:          size / 2,
		y:          size / 2,
		currentDir: NorthDir,
		state:      0,
		rules:      make(map[Signal]Action),
	}

	scanner := bufio.NewScanner(file)
	for lineno := 1; scanner.Scan(); lineno++ {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		var color_in, color_out ColorID
		var dirString string
		var state_in_char, state_out_char rune

		// scan the argument string, storing successive space-separated values into successive arguments as determined by the format
		n, err := fmt.Sscanf(line, "%c %d -> %c %d %s",
			&state_in_char,
			&color_in,
			&state_out_char,
			&color_out,
			&dirString)
		if err != nil || n != 5 {
			return nil, fmt.Errorf("Badly formatted line: %d", lineno)
		}
		state_in := State(state_in_char - 'a')
		state_out := State(state_out_char - 'a')
		dir, err := DirFromString(dirString)
		if err != nil {
			return nil, err
		}

		// read the rules from mite file and attach it to the rules in this mite object
		tur.rules[Signal{state: state_in, color: color_in}] = Action{
			state: state_out,
			color: color_out,
			turn:  dir,
		}
	}
	fmt.Printf("Read turmite with %d rules\n", len(tur.rules))
	return &tur, nil
}

// Step moves the turmite one step using the given field. Return an error if the
// turmite gets stuck with no rule to apply.
func (t *Turmite) Step(field Field) error {
	// field is the input,which would store the board

	// sense the color and find the suitable rules ==> obtain the signal based on the current location
	currState := t.state
	currColor := field[t.x][t.y]

	currSignal := Signal{
		state: currState,
		color: currColor,
	}

	// find the rules based on the signal
	nextAction := t.rules[currSignal]

	// set its state to a new value in a...z.
	t.state = nextAction.state

	// change the color of the square that it is on to some color
	field[t.x][t.y] = nextAction.color

	// Turn degrees relative to the direction it is facing
	t.currentDir = nextAction.turn

	// Walk one step in the direction it is facing
	// 朝着t.currentDir 的方向，已经更新过的方向前进1step 问题是1step的方向，落实到格子上需要研究

	return nil
}

func main() {
	var program, pngfile string
	var fieldSize, iters int

	flag.StringVar(&program, "prog", "zip.mite", "File containing the turmite program")
	flag.IntVar(&fieldSize, "s", 100, "Size of the field")
	flag.IntVar(&iters, "steps", 100, "Number of steps")
	flag.StringVar(&pngfile, "o", "output.png", "Filename to draw output")
	flag.Parse()

	if program == "" {
		log.Fatal("Must supply a program file with -prog.")
	}

	mite, err := ReadTurmite(program, fieldSize)
	if err != nil {
		log.Fatal(err)
	}
	field := NewField(fieldSize)
	for i := 0; i < iters; i++ {
		err := mite.Step(field)
		if err != nil {
			log.Fatal(err)
		}
	}
	field.DrawField(pngfile)
}
