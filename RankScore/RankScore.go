package RankScore

import (
	"strconv"

	"gonum.org/v1/gonum/floats"
)

// rankSumScore return single score for a sample and pathway

func RankScore(genes []string, sample []float64, pathwayGeneList []string) float64 {
	indx := make([]int, len(genes))

	for i := range genes {
		indx[i] = i
	}
	floats.Argsort(sample, indx)

	sortedGenes := make([]string, len(genes))
	for i, v := range indx {
		sortedGenes[i] = genes[v]
	}
	mi := MatchingGeneIndex(sortedGenes, pathwayGeneList)

	var RankSum int = 0
	for _, v := range mi {
		RankSum = RankSum + (v + 1)
	}

	var RankSumScore float64
	RankSumScore = float64(RankSum) / (float64(len(genes)) * float64(len(mi)))
	return RankSumScore
}

// function converts row of string to float 64
func RowToFloatVec(row []string) []float64 {
	n := len(row) - 1
	sample := make([]float64, n)
	// first value of row is sample name as read from cpm csv file
	// for i starts from 1 and skips sample name
	for i := 0; i < n; i++ {
		if s, err := strconv.ParseFloat(row[i+1], 64); err == nil {
			sample[i] = s
		}
	}
	return sample
}

// Rank Score function
func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func Include(vs []string, t string) bool {
	return Index(vs, t) >= 0
}

func MatchingGeneIndex(BL []string, SL []string) []int {
	var matchIndex []int

	for i, v := range BL {
		if Include(SL, v) == true {
			matchIndex = append(matchIndex, i)
		}

	}

	return matchIndex
}
