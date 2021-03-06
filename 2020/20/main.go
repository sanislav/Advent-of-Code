package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	// "os"
	"regexp"
)

type Pixels []string

func (p Pixels) Rotations() []Pixels {
	rot := make([]Pixels, 8)
	for r := 0; r < 8; r += 2 {
		for _, s := range p {
			rot[r] = append(rot[r], flip(s))
		}
		for i := range p {
			rot[r+1] = append(rot[r+1], flip(p.col(i)))
		}
		p = rot[r+1]
	}
	return rot
}

func (p Pixels) col(i int) (col string) {
	for _, s := range p {
		col += string(s[i])
	}
	return
}

func flip(s string) string {
	r := ""
	for _, c := range s {
		r = string(c) + r
	}

	return r
}

func neighboursMap(sensors map[string][]string) map[string][]string {
	m := map[string][]string{}
	for i1, e1 := range sensors {
		for i2, e2 := range sensors {
			if i1 != i2 {
				for _, v1 := range e1 {
					for _, v2 := range e2 {
						if v1 == v2 {
							alreadyExists := false
							for _, v := range(m[i1]) {
								if v == i2 {
									alreadyExists = true
									break
								}
							}
							if ! alreadyExists {
								m[i1] = append(m[i1], i2)
							}
						}
					}
				}
			}
		}
	}

	return m
}

func intersection(a, b []string) (c []string) {
	m := make(map[string]bool)
	unique := make(map[string]bool)

	if (len(a) == 0) {
		return b
	}
	if (len(b) == 0) {
		return a
	}

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			if _, isUnique := unique[item]; ! isUnique {
				c = append(c, item)
				unique[item] = true
			}
		}
	}

	return
}

func rebuildImage(m map[string][]string) [12][12]string {
	start := ""
	for k, n := range m {
		if len(n) == 2 {
			start = k
			break
		}
	}

	grid := [12][12]string{}
	used := map[string]bool{}
	l1 := m[start][0]
	l2 := m[start][1]
	used[start] = true
	used[l1] = true
	used[l2] = true

	grid[0][0] = start
	grid[0][1] = l1
	grid[1][0] = l2

	for len(used) != 12 * 12 {
		for i := 0; i < 12; i++ {
			for j := 0; j < 12; j ++ {
				if grid[i][j] != "" {
					continue
				}

				possibilitiesLeft := []string{}
				possibilitiesTop := []string{}

				if j > 0 {
					left := grid[i][j-1]

					for _, v := range m[left] {
						if _, alreadyUsed := used[v]; !alreadyUsed {
							possibilitiesLeft = append(possibilitiesLeft, v)
						}
					}
				}
				if i > 0 {
					top := grid[i-1][j]
					for _, v := range m[top] {
						if _, alreadyUsed := used[v]; !alreadyUsed {
							for _, v := range m[top] {
								if _, alreadyUsed := used[v]; !alreadyUsed {
									possibilitiesTop = append(possibilitiesTop, v)
								}
							}
						}
					}
				}
				options := intersection(possibilitiesLeft, possibilitiesTop)

				if len(options) == 1 {
					grid[i][j] = options[0]
					used[options[0]] = true
				}

				if i == 11 && j == 11 {
					i = 0;
					j = 0;
				}
			}
		}
	}

	return grid
}

func generateImgPixels(grid [12][12]string, tileToPixels map[string]Pixels) Pixels {
	image := [120]string{}
	// rotationsUsed := map[string]int{}
	order := map[string]Pixels{}

	rotations := tileToPixels[grid[0][0]].Rotations()
	rotationsR := tileToPixels[grid[0][1]].Rotations()
	rotationsB := tileToPixels[grid[1][0]].Rotations()

	// InitialRotation:
	for index, r := range rotations {
		edgeRight := r.col(len(r) - 1)
		edgeBottom := r[len(r) - 1]
		matches := []int{}
		matchR := -1
		matchB := -1
		for ir, rr := range rotationsR {
			edgeLR := rr.col(0)

			if (edgeRight == edgeLR) {
				matches = append(matches, index)
				matchR = ir
			}
		}

		if len(matches) == 0 {
			continue
		}

		for ib, rb := range rotationsB {
			edgeUR := rb[0]

			if (edgeBottom == edgeUR) {
				if (len(matches) == 0) {
					matches = append(matches, index)
				} else {
					for im, m := range matches {
						if m != index {
							matches = append(matches[:im], matches[im+1:]...)
						}
					}
				}
				matchB = ib
			}
		}

		if len(matches) == 1 {
			order[grid[0][0]] = rotations[matches[0]]
			order[grid[0][1]] = rotationsR[matchR]
			order[grid[1][0]] = rotationsB[matchB]
			break
		}
	}

	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[0]); j++ {
			sensorId := grid[i][j]
			if len(order[sensorId]) > 0 {
				continue
			}

			// flip / rotate to match neighbours
			rotations := tileToPixels[sensorId].Rotations()

			upN, leftN := Pixels{}, Pixels{}
			if i > 0 && len(order[grid[i-1][j]]) > 0 {
				upN = order[grid[i-1][j]]
			}
			if j > 0 && len(order[grid[i][j-1]]) > 0 {
				leftN = order[grid[i][j-1]]
			}

			for _, r := range rotations {
				edgeL := r.col(0)
				edgeT := r[0]

				if len(leftN) > 0 {
					edgeR := leftN.col(len(leftN) - 1)

					if (edgeL != edgeR) {
						continue
					}
				}

				if len(upN) > 0 {
					edgeB := upN[len(upN) - 1]

					if (edgeT != edgeB) {
						continue
					}
				}

				order[grid[i][j]] = r

				break
			}
		}
	}

	for i := 0; i < len(grid); i ++ {
		for j := 0; j < len(grid[0]); j ++ {
			position := order[grid[i][j]]
			for ir := 0; ir < len(position); ir++ {
				image[(i*10) + ir] += position[ir][1:len(position[ir]) - 1]
			}
		}
	}

	cutImg := Pixels{}
	for i := 0; i < len(image); i++ {
		if (i % 10 == 0) || (i + 1) % 10 == 0 {
			continue
		}

		cutImg = append(cutImg, image[i])
	}

	return cutImg
}

func main() {
	input, _ := ioutil.ReadFile("input.txt")
	sections := strings.Split(strings.TrimSpace(string(input)), "\n\n")
	tilesToPixels := map[string]Pixels{}
	sensors := map[string][]string{}
	for _, s := range sections {
		parts := strings.Split(s, "\n")
		id := strings.Split(strings.TrimRight(parts[0], ":"), "Tile ")

		top := string(parts[1])
		bottom := string(parts[10])
		left := ""
		right := ""
		pixels := Pixels{}

		for i := 1; i <= 10; i++ {
			left += string(parts[i][0])
			right += string(parts[i][9])
			pixels = append(pixels, parts[i])
		}

		sensors[id[1]] = []string{top, flip(top), bottom, flip(bottom), left, flip(left), right, flip (right)}
		tilesToPixels[id[1]] = pixels
	}

	m := neighboursMap(sensors)

	ans := 1
	for id, l := range m {
		if len(l) == 2 { // each corner piece only has 2 neighbours
			intID, _ := strconv.Atoi(id)
			ans *= intID
		}
	}

	fmt.Println(ans)

	sensorPositions := rebuildImage(m)

	img := generateImgPixels(sensorPositions, tilesToPixels)
	monster := []string{"..................#.", "#....##....##....###", ".#..#..#..#..#..#..."}
	nmonster := 0
	for _, r := range img.Rotations() {
		for y := 0; y < len(r)-len(monster); y++ {
		findMonster:
			for x := 0; x < len(r[0])-len(monster[0]); x++ {
				for i, s := range monster {
					if match, _ := regexp.MatchString(s, r[y+i][x:x+len(s)]); !match {
						continue findMonster
					}
				}
				nmonster++
			}
		}
	}
	fmt.Println(strings.Count(strings.Join(img, ""), "#") - nmonster*strings.Count(strings.Join(monster, ""), "#"))
}