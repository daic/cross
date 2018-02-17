package main

import (
	"fmt"
	"math/rand"
	"time"
)

const dimension = 3
const n = dimension * dimension
const nNei = 50
const nPop = 1000
const nGeneration = 1000
const krestik = 1
const nolik = 2

// доли мутации и типы их

var mutProb []float32
var mutType []int
var mutBoard []int

var s [n]byte
var pos [n * 4]float32

type nei struct {
	src1 byte
	src2 byte
	dst  byte
	k1   float32
	k2   float32
}
type net []nei

var pop []net
var nLife, nWin, nDraw, nLose, sorted []int
var iGeneration int

func main() {
	tt := time.Now().UnixNano()
	rand.Seed(tt)
	pop = make([]net, nPop)
	nLife = make([]int, nPop)
	nWin = make([]int, nPop)
	nLose = make([]int, nPop)
	nDraw = make([]int, nPop)
	sorted = make([]int, nPop)
	mutProb = make([]float32, 4)
	mutType = make([]int, 4)
	mutBoard = make([]int, 5)

	mutRule()
	for p := 0; p < nPop; p++ {
		pop[p] = make([]nei, nNei, nNei)
		sorted[p] = p
	}
	rand.Seed(42)
	for p := 0; p < nPop; p++ {
		pop[p].makeRandom()
		nLife[p] = 0
	}
	iGeneration = 0
	for iGen := 0; iGen < nGeneration; iGen++ {
		iGeneration++
		oneLife()
		zeroStat()
		for p1 := 0; p1 < nPop; p1++ {
			for p2 := 0; p2 < nPop; p2++ {
				oneGame(p1, p2)
			}
		}
		sortPop()
		printResult()
		mutation()
	}
}
func mutRule() {
	mutProb = []float32{20, 35, 35, 10}
	mutType = []int{0, 1, 2, 3}
	for i := 0; i < len(mutProb); i++ {
		mutBoard[i] = 0
	}
	for i := 1; i < len(mutProb); i++ {
		mutBoard[i] = int(mutProb[i-1]*nPop/100.0) + mutBoard[i-1]
	}
	mutBoard[4] = nPop
	fmt.Println("mutBoard=", mutBoard)
}
func mutOne(sk int, mType int) {
	switch mType {
	case 0:
	case 1:
		shiftMutation(sorted[sk], 3, 0.2)
	case 2:
		sexMutation(sorted[sk], randRange(mutBoard[0], mutBoard[1]), randRange(mutBoard[1], mutBoard[2]))
	case 3:
		pop[sorted[sk]].makeRandom()
		nLife[sorted[sk]] = 0
	}
}
func randRange(b, e int) int {
	return sorted[rand.Intn(e)+b]
}
func mutation() {
	numType := len(mutProb)
	for sk := 0; sk < nPop; sk++ {
		for mi := 0; mi < numType; mi++ {
			if sk >= mutBoard[mi] && sk < mutBoard[mi+1] {
				mutOne(sk, mutType[mi])
			}
		}
	}
}
func sexMutation(p, p1, p2 int) {
	nLife[p] = 0
	sh := 1
	for i := 0; i < nNei; i++ {
		if sh == 1 {
			neiCopy(pop[p][i], pop[p1][i])
			sh = 2
		} else if sh == 2 {
			neiCopy(pop[p][i], pop[p2][i])
			sh = 1
		}

	}

}
func neiCopy(ne, ne1 nei) {
	ne.k1 = ne1.k1
	ne.k2 = ne1.k2
	ne.src1 = ne1.src1
	ne.src2 = ne1.src2
	ne.dst = ne1.dst
}
func shiftMutation(p int, period int, shift float32) {
	for i := 0; i < nNei; i++ {
		if rand.Intn(period) == 0 {
			pop[p][i].k1 = pop[p][i].k1 + (rand.Float32()*2-1.0)*shift
			pop[p][i].k2 = pop[p][i].k1 + (rand.Float32()*2-1.0)*shift
		}
	}
}
func printResult() {
	win := 0
	lose := 0
	draw := 0
	del := 0
	for p := 0; p < nPop; p++ {
		win = win + nWin[p]
		lose = lose + nLose[p]
		draw = draw + nDraw[p]
	}
	bestLife := 0
	iBestLife := -1
	for p := 0; p < nPop; p++ {
		if nLife[p] > bestLife {
			bestLife = nLife[p]
			iBestLife = p
		}
	}
	rangBest := 0
	for sk := 0; sk < nPop; sk++ {
		if sorted[sk] == iBestLife {
			rangBest = sk
			break
		}
	}
	for sk := mutBoard[2]; sk < nPop; sk++ {
		if nLife[sorted[sk]] > 1 {
			del++
		}
	}
	fmt.Println("G=", iGeneration, ",(draw=", draw/2, ") best(", sorted[0], "-", nLife[sorted[0]], ")(",
		nWin[sorted[0]], nLose[sorted[0]], nDraw[sorted[0]], ") worst(", sorted[nPop-1], "-", nLife[sorted[nPop-1]], ")(",
		nWin[sorted[nPop-1]], nLose[sorted[nPop-1]], nDraw[sorted[nPop-1]], "), bLife(", iBestLife, "-", nLife[iBestLife], "-", rangBest, ")(",
		nWin[iBestLife], nLose[iBestLife], nDraw[iBestLife], ") del(", del, ")")

}
func score(p int) int {
	return nWin[p]*2 + nDraw[p]
}
func sortPop() {
	for {
		shift := 0
		for sk := 0; sk < nPop-1; sk++ {
			if score(sorted[sk]) < score(sorted[sk+1]) {
				shift++
				sorted[sk], sorted[sk+1] = sorted[sk+1], sorted[sk]
			}
		}
		if shift == 0 {
			break
		}
	}
}
func oneGame(p1, p2 int) {
	setNull()
	win := 0
	turn := 1
	for t := 0; t < n; t++ {
		setPos()
		if turn == 1 {
			play(p1)
		} else if turn == 2 {
			invertPos()
			play(p2)
		}
		tec := -1
		for i := 0; i < n; i++ {
			if s[i] != 0 {
				continue
			} else if tec == -1 {
				tec = i
			} else if pos[3*n+i] > pos[3*n+tec] {
				tec = i
			}
		}
		s[tec] = byte(turn)
		win = result()
		if turn == 1 {
			turn = 2
		} else {
			turn = 1
		}
		/*if win != 0 && p1 == 501 && p2 == 502 {
			fmt.Println("(", s[0], s[1], s[2], ")(", p1, p2, ")")
			fmt.Println("(", s[3], s[4], s[5], ")")
			fmt.Println("(", s[6], s[7], s[8], ")")
			fmt.Println(pos)
		}*/
		if win != 0 {
			break
		}
	}
	if win == 1 {
		nWin[p1]++
		nLose[p2]++
	} else if win == 2 {
		nWin[p2]++
		nLose[p1]++
	} else {
		nDraw[p1]++
		nDraw[p2]++
	}
}
func play(p int) {
	for i := 0; i < nNei; i++ {
		pos[pop[p][i].dst] = pos[pop[p][i].src1]*pop[p][i].k1 + pos[pop[p][i].src2]*pop[p][i].k2
	}
}
func result() int {
	for p := 1; p <= 2; p++ {
		if (s[0] == byte(p)) && (s[1] == byte(p)) && (s[2] == byte(p)) {
			return p
		} else if (s[3] == byte(p)) && (s[4] == byte(p)) && (s[5] == byte(p)) {
			return p
		} else if (s[6] == byte(p)) && (s[7] == byte(p)) && (s[8] == byte(p)) {
			return p
		} else if (s[0] == byte(p)) && (s[3] == byte(p)) && (s[6] == byte(p)) {
			return p
		} else if (s[1] == byte(p)) && (s[4] == byte(p)) && (s[7] == byte(p)) {
			return p
		} else if (s[2] == byte(p)) && (s[5] == byte(p)) && (s[8] == byte(p)) {
			return p
		} else if (s[0] == byte(p)) && (s[4] == byte(p)) && (s[8] == byte(p)) {
			return p
		} else if (s[2] == byte(p)) && (s[4] == byte(p)) && (s[6] == byte(p)) {
			return p
		}
	}
	return 0
}
func setNull() {
	for i := 0; i < n; i++ {
		s[i] = 0
	}
}
func setPos() {
	for i := 0; i < n*2; i++ {
		pos[i] = 0.0
	}
	for i := 0; i < n; i++ {
		if s[i] == krestik {
			pos[i] = 1.0
		} else if s[i] == nolik {
			pos[n+i] = 1.0
		}
	}
	for i := n * 2; i < n*4; i++ {
		pos[i] = 1.0
	}
}
func invertPos() {
	for i := 0; i < n; i++ {
		pos[i], pos[n+i] = pos[n+i], pos[i]
	}
}
func zeroStat() {
	for k := 0; k < nPop; k++ {
		nWin[k] = 0
		nLose[k] = 0
		nDraw[k] = 0
	}
}
func oneLife() {
	for p := 0; p < nPop; p++ {
		nLife[p]++
	}
}
func (p net) makeRandom() {
	for i := 0; i < nNei; i++ {
		p[i].src1 = byte(rIntn(4 * n))
		p[i].src2 = byte(rIntn(n * 4))
		p[i].dst = byte(rIntn(n*2)) + n*2
		p[i].k1 = -1.0 + rand.Float32()*2
		p[i].k2 = -1.0 + rand.Float32()*2 - 1.0
	}
}

func rIntn(max int) int {
	return rand.Intn(max)
}
func rFloat32() float32 {
	return rand.Float32()
}
