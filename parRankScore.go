package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"./RankScore"
)

var wg sync.WaitGroup

func main() {

	expnMatrixFile := flag.String("filename", "provide filename with full path in double quotes", "input filename with full path")
	pathwayDB := flag.String("Pathway", "wiki", "input pathway DB")
	cpu := flag.Int("nCPU", 2, "provide --nCPU 4")
	help := flag.Bool("help", false, "")
	flag.Parse()
	if *expnMatrixFile == "" {
		fmt.Println("expn matrix filepath required")
		os.Exit(1)
	}
	if *help == true {
		fmt.Println("Usage::")
		fmt.Println(`RankScoreClI --filename "/User/dir/filename.csv" --Pathway "wiki" --nCPU 4`)
		return
	}
	var fname string
	if *pathwayDB == "wiki" {
		fname = "/Users/pk/learn_stuff/GOlang/Les1/wikipathway.v7.2.symbols.gmt"
	} else {
		fname = "/Users/pk/learn_stuff/GOlang/Les1/ReactomePathways.gmt"
	}

	f, err := os.Open(fname)
	if err != nil {
		log.Fatal("No file", err)
	}
	pid_glist := RankScore.ReadPathwayFile2(f)
	f.Close()

	csvFile, _ := os.Open(*expnMatrixFile)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	// each row[0] value is sample name and subsequent values are gene expn values
	var row [][]string

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		row = append(row, line)

	}

	sampleNames := make([]string, len(row)-1)
	for i := 0; i < len(sampleNames); i++ {
		sampleNames[i] = row[i+1][0]
	}

	fmt.Println("sampleNames =>", sampleNames)

	// primitives for goroutine
	pchan := make(chan RankScore.GeneListStruct)
	reschan := make(chan ResScoresStruct)
	var gr int
	gr = *cpu

	go passPathway(pchan, pid_glist)
	wg.Add(gr)
	for i := 1; i <= gr; i++ {
		go synPathway(reschan, pchan, row, i)
	}

	var samplePathwayMatx []ResScoresStruct
	for i := 0; i < len(pid_glist); i++ {
		v := <-reschan
		samplePathwayMatx = append(samplePathwayMatx, v)
	}

	fresult, err := os.Create("result.csv")
	defer fresult.Close()

	if err != nil {

		log.Fatalln("failed to open file", err)
	}

	wRes := csv.NewWriter(fresult)
	defer wRes.Flush()

	var header []string
	header = append([]string{"pathway"}, sampleNames...)
	wRes.Write(header)
	for _, val := range samplePathwayMatx {
		rec := make([]string, (len(sampleNames) + 1))
		//rec[0] = pid[i]
		rec[0] = val.Id
		for i, v := range val.score {
			rec[i+1] = fmt.Sprint(v)
		}
		if err := wRes.Write(rec); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}
	wg.Wait()
}

// end main here
// SynPathway take ourput channel, input channel row[][] matrix of data, worker number
//its main worker that will run as go routine in parrallel
func synPathway(out chan<- ResScoresStruct, in <-chan RankScore.GeneListStruct, row [][]string, worker int) {
	defer wg.Done()
	for {
		v, ok := <-in
		if !ok {
			fmt.Printf("worker : %d : shutting down\n", worker)
			return
		}
		out <- processRows(row, v)
		fmt.Println("Processing:", v.ID)

	}

}

//this function takes slice of gene_list struct and pass it to chan
// that will be used as input chan for SynPathway
func passPathway(out chan<- RankScore.GeneListStruct, pid_glist []RankScore.GeneListStruct) {
	for _, x := range pid_glist {
		out <- x
	}
	close(out)
}

//"ProcessRows return map of pathway as id and scores as val slice"
func processRows(row [][]string, glist RankScore.GeneListStruct) ResScoresStruct {
	nSample := len(row) - 1
	genes := row[0][1:len(row[0])]
	var pathwayScores ResScoresStruct
	allScores2 := make([]float32, nSample)
	// process each row in this loop
	for j := 0; j < nSample; j++ {
		var s2 []string
		s2 = append(s2, row[j+1]...) // row[0] is sample name, value are from 1:
		s3 := RankScore.RowToFloatVec(s2)
		allScores2[j] = float32(RankScore.RankScore(genes, s3, glist.GeneList))
	}
	pathwayScores.Id = glist.ID
	pathwayScores.score = allScores2
	return pathwayScores

}

// ResScoresStruct stores results of rankScores
type ResScoresStruct struct {
	Id    string
	score []float32
}
