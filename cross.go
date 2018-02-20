package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const dimension = 3
const n = dimension * dimension
const mem = n * 4
const nGeneration = 1000
const krestik = 1
const nolik = 2

var nNei = 50
var nPop = 500

// доли мутации и типы их
var mutProb []float32
var mutType []int
var mutBoard []int

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
var sMut, cMut []int
var iGeneration int
var nameG *string
var dir string
var wg sync.WaitGroup
var start, end time.Time
var elapsed time.Duration

func main() {
	nameG = flag.String("n", "default", "Каталог набора")
	loadData()
	tt := time.Now().UnixNano()
	rand.Seed(tt)

	for iGen := 0; iGen < nGeneration; iGen++ {
		iGeneration++
		start = time.Now()
		oneLife()
		zeroStat()
		for p1 := 0; p1 < nPop; p1++ {
			for p2 := 0; p2 < nPop; p2++ {
				wg.Add(1)
				go oneGame(p1, p2)
			}
			wg.Wait()
		}
		sortPop()
		end = time.Now()
		elapsed = end.Sub(start)
		if iGeneration%10 == 0 {
			printResult()
		}
		start = time.Now()
		mutation()
		end = time.Now()
		elapsed = end.Sub(start)
		if iGeneration%10 == 0 {
			fmt.Println("Mutation duration", elapsed)
		}
		if iGeneration%50 == 0 {
			saveCurrentData(dir)
			groupAnalyse()
		}
	}
}
func loadData() {
	var data struct {
		NNei    int
		NPop    int
		MutProb [4]float32
	}
	data.MutProb = [4]float32{49, 49, 1, 1}
	mutProb = make([]float32, 4)
	mutProb = []float32{49, 49, 1, 1}
	dir, _ = os.Getwd()
	nameF := dir + "\\" + *nameG + "\\rule.json"
	conffile, err := os.Open(nameF)
	newData := false
	if err == nil {
		dec := json.NewDecoder(conffile)
		_ = dec.Decode(&data)
		nNei = data.NNei
		nPop = data.NPop
		for i := 0; i < 4; i++ {
			mutProb[i] = data.MutProb[i]
		}
		fmt.Println("Read", data)
		conffile.Close()
	} else {
		conffile.Close()
		os.Mkdir(*nameG, 0700)
		file, _ := os.Create(nameF)
		data.NNei = nNei
		data.NPop = nPop
		for i := 0; i < 4; i++ {
			data.MutProb[i] = mutProb[i]
		}
		enc := json.NewEncoder(file)
		_ = enc.Encode(data)
		fmt.Println("Write", data)
		file.Close()
		newData = true
	}
	pop = make([]net, nPop)
	nLife = make([]int, nPop)
	nWin = make([]int, nPop)
	nLose = make([]int, nPop)
	nDraw = make([]int, nPop)
	sMut = make([]int, nPop)
	cMut = make([]int, nPop)
	sorted = make([]int, nPop)
	mutType = make([]int, 4)
	mutBoard = make([]int, 5)
	mutType = []int{0, 1, 2, 3}
	mutRule()
	for p := 0; p < nPop; p++ {
		pop[p] = make([]nei, nNei)
		sorted[p] = p
	}
	for p := 0; p < nPop; p++ {
		pop[p].makeRandom()
		nLife[p] = 0
	}
	iGeneration = 0
	if newData == false {
		loadCurrentData(dir)
	}
}
func loadCurrentData(dir string) {
	var data struct {
		IGen  int
		NLife []int
		SMut  []int
		CMut  []int
	}
	data.NLife = nLife
	data.SMut = sMut
	data.CMut = cMut
	nameF := dir + "\\" + *nameG + "\\stat.json"
	conffile, err := os.Open(nameF)
	if err == nil {
		dec := json.NewDecoder(conffile)
		_ = dec.Decode(&data)
		iGeneration = data.IGen
		fmt.Println("loadData", data)
	}
	var popData struct {
		Src1 [][]byte
		Src2 [][]byte
		Dst  [][]byte
		K1   [][]float32
		K2   [][]float32
	}
	popData.Src1 = make([][]byte, nPop)
	popData.Src2 = make([][]byte, nPop)
	popData.Dst = make([][]byte, nPop)
	popData.K1 = make([][]float32, nPop)
	popData.K2 = make([][]float32, nPop)
	for k := 0; k < nPop; k++ {
		popData.Src1[k] = make([]byte, nNei)
		popData.Src2[k] = make([]byte, nNei)
		popData.Dst[k] = make([]byte, nNei)
		popData.K1[k] = make([]float32, nNei)
		popData.K2[k] = make([]float32, nNei)
	}
	nameF = dir + "\\" + *nameG + "\\pop.json"
	conffile, err = os.Open(nameF)
	if err == nil {
		dec := json.NewDecoder(conffile)
		_ = dec.Decode(&popData)
		for k := 0; k < nPop; k++ {
			for i := 0; i < nNei; i++ {
				pop[k][i].src1 = popData.Src1[k][i]
				pop[k][i].src2 = popData.Src2[k][i]
				pop[k][i].dst = popData.Dst[k][i]
				pop[k][i].k1 = popData.K1[k][i]
				pop[k][i].k2 = popData.K2[k][i]
			}
		}
	}
}
func saveCurrentData(dir string) {
	var data struct {
		IGen  int
		NLife []int
		SMut  []int
		CMut  []int
	}
	data.NLife = nLife
	data.SMut = sMut
	data.CMut = cMut
	data.IGen = iGeneration
	nameF := dir + "\\" + *nameG + "\\stat.json"
	conffile, err := os.Create(nameF)
	if err == nil {
		enc := json.NewEncoder(conffile)
		_ = enc.Encode(data)
		iGeneration = data.IGen
		//fmt.Println("saveData", data)
		conffile.Close()
	}
	var popData struct {
		Src1 [][]byte
		Src2 [][]byte
		Dst  [][]byte
		K1   [][]float32
		K2   [][]float32
	}
	popData.Src1 = make([][]byte, nPop)
	popData.Src2 = make([][]byte, nPop)
	popData.Dst = make([][]byte, nPop)
	popData.K1 = make([][]float32, nPop)
	popData.K2 = make([][]float32, nPop)
	for k := 0; k < nPop; k++ {
		popData.Src1[k] = make([]byte, nNei)
		popData.Src2[k] = make([]byte, nNei)
		popData.Dst[k] = make([]byte, nNei)
		popData.K1[k] = make([]float32, nNei)
		popData.K2[k] = make([]float32, nNei)
		for i := 0; i < nNei; i++ {
			popData.Src1[k][i] = pop[k][i].src1
			popData.Src2[k][i] = pop[k][i].src2
			popData.Dst[k][i] = pop[k][i].dst
			popData.K1[k][i] = pop[k][i].k1
			popData.K2[k][i] = pop[k][i].k2
		}
	}
	nameF = dir + "\\" + *nameG + "\\pop.json"
	conffile, err = os.Create(nameF)
	if err == nil {
		enc := json.NewEncoder(conffile)
		_ = enc.Encode(popData)
		conffile.Close()
	}
}
func mutRule() {
	for i := 0; i < len(mutProb); i++ {
		mutBoard[i] = 0
	}
	for i := 1; i < len(mutProb); i++ {
		mutBoard[i] = int(mutProb[i-1]*float32(nPop)/100.0) + mutBoard[i-1]
	}
	mutBoard[4] = nPop
	fmt.Println("mutBoard=", mutBoard)
}
func mutOne(sk int, mType int) {
	defer wg.Done()
	//commonMutation(sorted[sk])
	switch mType {
	case 0:
	case 1:
		copyMutation2(sorted[sk], randRange(mutBoard[0], mutBoard[1]))
		//shiftMutation(sorted[sk], 3, 0.02)
	case 2:
		//copyMutation2(sorted[sk], randRange(mutBoard[0], mutBoard[1]))
		sexMutation(sorted[sk], randRange(mutBoard[0], mutBoard[1]), randRange(mutBoard[1], mutBoard[2]))
		/*if rand.Intn(2) == 0 {
			copyMutation(sorted[sk], randRange(mutBoard[0], mutBoard[1]), 10, 0.1, 0.005)
		} else {
			sexMutation(sorted[sk], randRange(mutBoard[0], mutBoard[1]), randRange(mutBoard[1], mutBoard[2]))
		}*/
	case 3:
		randomMutation(sorted[sk])
	}
}
func randomMutation(p int) {
	pop[p].makeRandom()
	nLife[p] = 0
	sMut[p] = 0
	cMut[p] = 0
}
func randRange(b, e int) int {
	return sorted[rand.Intn(e-b)+b]
}
func mutation() {
	numType := len(mutProb)
	for mi := 0; mi < numType; mi++ {
		for sk := 0; sk < nPop; sk++ {
			if sk >= mutBoard[mi] && sk < mutBoard[mi+1] {
				wg.Add(1)
				go mutOne(sk, mutType[mi])
			}
		}
		wg.Wait()
	}
}
func commonMutation(p int) {
	if nLose[p] == 0 {
		return
	}
	cMut[p]++
	prob := float32(nLose[p]) / float32(nWin[p]+nLose[p]+nDraw[p])
	for i := 0; i < nNei; i++ {
		if rand.Float32() < prob {
			pop[p][i].k1 = pop[p][i].k1 + (rand.Float32()*2-1.0)*0.05
			pop[p][i].k2 = pop[p][i].k2 + (rand.Float32()*2-1.0)*0.05
		}
		if rand.Float32() < prob*0.01 {
			pop[p][i].src1 = byte(rIntn(4 * n))
		}
		if rand.Float32() < prob*0.01 {
			pop[p][i].src2 = byte(rIntn(4 * n))
		}
		if rand.Float32() < prob*0.01 {
			pop[p][i].dst = byte(rIntn(4 * n))
		}
	}
}
func sexMutation(p, p1, p2 int) {
	nLife[p] = 0
	sMut[p]++
	for i := 0; i < nNei; i++ {
		if i < int(nNei/2) {
			neiCopy(pop[p][i], pop[p1][i])
		} else {
			neiCopy(pop[p][i], pop[p2][i])
		}
	}
	commonMutation(p)
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
func copyMutation2(p, p1 int) {
	for i := 0; i < nNei; i++ {
		neiCopy(pop[p][i], pop[p1][i])
	}
	nLife[p] = nLife[p1]
	sMut[p] = sMut[p1]
	cMut[p] = cMut[p1]
	commonMutation(p)
}
func copyMutation(p, p1, period int, shift float32, prob float32) {
	nLife[p] = 0
	for i := 0; i < nNei; i++ {
		neiCopy(pop[p][i], pop[p1][i])
		if rand.Intn(period) == 0 {
			pop[p][i].k1 = pop[p][i].k1 + (rand.Float32()*2-1.0)*shift
			pop[p][i].k2 = pop[p][i].k1 + (rand.Float32()*2-1.0)*shift

		}
		if rand.Float32() < prob {
			pop[p][i].src1 = byte(rIntn(4 * n))
		}
		if rand.Float32() < prob {
			pop[p][i].src2 = byte(rIntn(4 * n))
		}
		if rand.Float32() < prob {
			pop[p][i].dst = byte(rIntn(4 * n))
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
	fmt.Println("G=", iGeneration, ",(draw=", draw/2, ") best(", sorted[0], "-", nLife[sorted[0]], "sMut=", sMut[sorted[0]], "cMut=", cMut[sorted[0]], ")(",
		nWin[sorted[0]], nLose[sorted[0]], nDraw[sorted[0]], ") \n worst(", sorted[nPop-1], "-", nLife[sorted[nPop-1]], "sMut=", sMut[sorted[nPop-1]], "cMut=", cMut[sorted[nPop-1]], ")(",
		nWin[sorted[nPop-1]], nLose[sorted[nPop-1]], nDraw[sorted[nPop-1]], ") del(", del, ") \n bLife(", iBestLife, "-", nLife[iBestLife], "rang=", rangBest, "sMut=", sMut[iBestLife], "cMut=", cMut[iBestLife], ")(",
		nWin[iBestLife], nLose[iBestLife], nDraw[iBestLife], ")", elapsed)

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
	defer wg.Done()
	var s [n]byte
	for i := 0; i < n; i++ {
		s[i] = 0
	}
	var pos [mem]float32
	win := 0
	turn := 1
	for t := 0; t < n; t++ {
		setPos(s, &pos)
		if turn == 1 {
			play(p1, &pos)
		} else if turn == 2 {
			invertPos(&pos)
			play(p2, &pos)
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
		win = result(s)
		if turn == 1 {
			turn = 2
		} else {
			turn = 1
		}
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
func play(p int, pos *[mem]float32) {
	for i := 0; i < nNei; i++ {
		pos[pop[p][i].dst] = pos[pop[p][i].src1]*pop[p][i].k1 + pos[pop[p][i].src2]*pop[p][i].k2
	}
}
func result(s [n]byte) int {
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
func setPos(s [n]byte, pos *[mem]float32) {
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
func invertPos(pos *[mem]float32) {
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
func groupAnalyse() {
	var groups, charRep []int
	groups = make([]int, nPop)
	charRep = make([]int, nPop)
	numGroup := 0
	var maxCoin, midLife float32
	midLife = 0.0
	coin := func(i, k int) float32 {
		nCoin := 0
		nAll := 0
		for m := 0; m < nNei; m++ {
			nAll = nAll + 3
			if pop[charRep[i]][m].src1 == pop[k][m].src1 {
				nCoin++
			}
			if pop[charRep[i]][m].src2 == pop[k][m].src2 {
				nCoin++
			}
			if pop[charRep[i]][m].dst == pop[k][m].dst {
				nCoin++
			}
		}
		return float32(nCoin) / float32(nAll)
	}
	for k := 0; k < nPop; k++ {
		b := true
		midLife = midLife + float32(nLife[k])
		for i := 0; i < numGroup; i++ {
			co := coin(i, k)
			if co > maxCoin {
				maxCoin = co
			}
			if co > 0.8 {
				groups[i]++
				b = false
				break
			}
		}
		if b == true {
			charRep[numGroup] = k
			numGroup++
		}
	}
	for i := 0; i < 10; i++ {
		if groups[i] > 5 {
			fmt.Println(i, ")", groups[i])
		}
	}
	midLife = midLife / float32(nPop)
	fmt.Println("nGroup=", numGroup, "Pop=", nPop, "MaxCoin=", maxCoin, "midLife=", midLife)
}
