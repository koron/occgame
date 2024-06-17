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

func isAdjacent(a, b Pos) bool {
	return (a.Y == b.Y && a.X+1 == b.X) || (a.X == b.X && a.Y+1 == b.Y)
}

func (g *Occgame) CanConnect(a, b Pos) bool {
	// If they point to the same location, they will not be able to connect.
	if a == b {
		return false
	}
	// If they point different tokens, they will not be able to connect.
	ta, tb := g.get(a), g.get(b)
	if ta == 0 || ta != tb {
		return false
	}

	// Swap points if needed: A should lefter than B, or upper than B.
	if a.X > b.X || (a.X == b.X && a.Y > b.Y) {
		a, b = b, a
	}

	// If A and B are adjacent, they can obviously be connected.
	if isAdjacent(a, b) {
		return true
	}

	// TODO:

	return false
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
