package main

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
)

type Occgame struct {
	W       int
	H       int
	Quartet int
	Board   []uint8
	LastTok uint8
}

type Pos struct {
	X int
	Y int
}

func NewGame() (*Occgame, error) {
	w, h := 12, 6
	quartet := 16
	return &Occgame{
		W:       w,
		H:       h,
		Quartet: quartet,
		Board:   make([]uint8, w*h),
	}, nil
}

func (g *Occgame) clearAll() {
	for i := range g.Board {
		g.Board[i] = 0
	}
}

func (g *Occgame) isCleard() bool {
	for _, v := range g.Board {
		if v != 0 {
			return false
		}
	}
	return true
}

// boardIndex returns an index for p Pos on the board.
func (g *Occgame) boardIndex(p Pos) int {
	return p.X + p.Y*g.W
}

// set modifies a token at p Pos in the board.
func (g *Occgame) set(p Pos, tok uint8) {
	g.Board[g.boardIndex(p)] = tok
}

// get gets a token at p Pos in the board.
func (g *Occgame) get(p Pos) uint8 {
	return g.Board[g.boardIndex(p)]
}

func (g *Occgame) Init() {
	n := g.W * g.H

	// shuffled position lists.
	pp := make([]Pos, 0, n)
	for i := 0; i < n; i++ {
		pp = append(pp, Pos{X: i % g.W, Y: i / g.W})
	}
	rand.Shuffle(n, func(i, j int) {
		pp[i], pp[j] = pp[j], pp[i]
	})

	// fill the board with tokens in shuffled.
	var tok uint8 = 1
	for i := 0; i < g.Quartet && len(pp) >= 4; i++ {
		for j := 0; j < 4; j++ {
			var p Pos
			p, pp = pp[0], pp[1:]
			g.set(p, tok)
		}
		tok++
	}
	for len(pp) >= 2 {
		for j := 0; j < 2; j++ {
			var p Pos
			p, pp = pp[0], pp[1:]
			g.set(p, tok)
		}
		tok++
	}
	g.LastTok = tok
}

func (g *Occgame) DumpBoard(w io.Writer, indent string) {
	x := 0
	for i := 0; i < g.H; i++ {
		io.WriteString(w, indent)
		for j := 0; j < g.W; j++ {
			tok := g.Board[x]
			x++
			if tok == 0 {
				io.WriteString(w, " ")
				continue
			}
			io.WriteString(w, string([]rune{'A' + rune(tok) - 1}))
		}
		io.WriteString(w, "\n")
	}
}

func (g *Occgame) Classify() map[uint8][]Pos {
	m := make(map[uint8][]Pos)
	for i, tok := range g.Board {
		if tok == 0 {
			continue
		}
		m[tok] = append(m[tok], Pos{X: i % g.W, Y: i / g.W})
	}
	return m
}

// isAdjacent checks two Pos are adjacent (隣接してたらtrue)
func isAdjacent(a, b Pos) bool {
	return (a.Y == b.Y && a.X+1 == b.X) || (a.X == b.X && a.Y+1 == b.Y)
}

// emptyHorz checks for horizontal empties
func (g *Occgame) emptyHorz(y, x0, x1 int) bool {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		if g.Board[y*g.W+x] != 0 {
			return false
		}
	}
	return true
}

// emptyVert checks for vertical empties
func (g *Occgame) emptyVert(x, y0, y1 int) bool {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		if g.Board[x+y*g.W] != 0 {
			return false
		}
	}
	return true
}

func (g *Occgame) lookFarVert(p Pos) (up, down int) {
	up, down = p.Y, p.Y
	for up > 0 && g.get(Pos{X: p.X, Y: up - 1}) == 0 {
		up--
	}
	for down < g.H-1 && g.get(Pos{X: p.X, Y: down + 1}) == 0 {
		down++
	}
	return up, down
}

func (g *Occgame) lookFarHorz(p Pos) (left, right int) {
	left, right = p.X, p.X
	for left > 0 && g.get(Pos{X: left - 1, Y: p.Y}) == 0 {
		left--
	}
	for right < g.W-1 && g.get(Pos{X: right - 1, Y: p.Y}) == 0 {
		right++
	}
	return left, right
}

func (g *Occgame) CanConnect(a, b Pos) []Pos {
	// If they point to the same location, they will not be able to connect.
	if a == b {
		return nil
	}
	// If they point different tokens, they will not be able to connect.
	ta, tb := g.get(a), g.get(b)
	if ta == 0 || ta != tb {
		return nil
	}

	// Swap points if needed: A should lefter than B, or upper than B.
	if a.X > b.X || (a.X == b.X && a.Y > b.Y) {
		a, b = b, a
	}

	// If A and B are adjacent, they can obviously be connected.
	if isAdjacent(a, b) {
		return []Pos{a, b}
	}

	// Check two points can be connected with a straight line.
	if a.Y == b.Y {
		// Check the horizontal.
		if g.emptyHorz(a.Y, a.X, b.X) {
			return []Pos{a, b}
		}
	}
	if a.X == b.X {
		// Check the vertical.
		if g.emptyVert(a.X, a.Y, b.Y) {
			return []Pos{a, b}
		}
	}

	// Check two points can be connected with two straight lines. (single corner)
	if a.X != b.X && a.Y != b.Y {
		if g.emptyHorz(a.Y, a.X+1, b.X) && ((a.Y < b.Y && g.emptyVert(b.X, a.Y+1, b.Y)) || (a.Y > b.Y && g.emptyVert(b.X, b.Y, a.Y-1))) {
			return []Pos{a, {X: b.X, Y: a.Y}, b}
		}
		if g.emptyHorz(b.Y, a.X+1, b.X) && ((a.Y < b.Y && g.emptyVert(a.X, a.Y+1, b.Y)) || (a.Y > b.Y && g.emptyVert(a.X, b.Y, a.Y-1))) {
			return []Pos{a, {X: a.X, Y: b.Y}, b}
		}
	}

	// Check two points can be connected with two straight lines. (double corner)
	if a.X != b.X {
		// 上下方向にチェック
		au, ad := g.lookFarVert(a)
		bu, bd := g.lookFarVert(b)
		up := max(au, bu)
		down := min(ad, bd)
		if up < down {
			// TODO:
			// Check links outside of the box.
			if up == 0 {
				return []Pos{a, {X: a.X, Y: -1}, {X: b.X, Y: -1}, b}
			}
			if down == g.H-1 {
				return []Pos{a, {X: a.X, Y: g.W}, {X: b.X, Y: g.W}, b}
			}
		}
	}
	if a.Y != b.Y {
		// 左右方向にチェック
		al, ar := g.lookFarHorz(a)
		bl, br := g.lookFarHorz(b)
		left := max(al, bl)
		right := min(ar, br)
		if left < right {
			// TODO:
			// Check links outside of the box.
			if left == 0 {
				return []Pos{a, {X: -1, Y: a.Y}, {X: -1, Y: b.Y}, b}
			}
			if right == g.W-1 {
				return []Pos{a, {X: g.H, Y: a.Y}, {X: g.H, Y: b.Y}, b}
			}
		}
	}

	return nil
}

func main() {
	g, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}
	g.Init()
	fmt.Println("#0")
	g.DumpBoard(os.Stdout, "  ")
	m := g.Classify()
	fmt.Printf("m: %+v\n", m)
}
