package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"./RankScore"
)

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/upload", uploadFile)

	fmt.Println("Server started at localhost:3000")
	http.ListenAndServe(":3000", nil)
}

// we handlers
func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	tmpl, _ := template.ParseFiles("rankScore.html")
	err := tmpl.Execute(w, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if err := r.ParseMultipartForm(1024); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f, handler, err := r.FormFile("pathwayFile")
	fmt.Println(handler.Filename)
	// read pathway file
	//f1, err := os.Open(handler.Filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dup2: %v\n", err)
	}

	//var pid, glist = RankScore.ReadPathwayFile(f)
	pid_glist := RankScore.ReadPathwayFile2(f)
	f.Close()

	//fmt.Println("pid::: ", pid[0:5])

	csvFile, _, err := r.FormFile("matxFile")
	//csvFile, _ := os.Open(matx)
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
	// first line is header nSample is len(row)-1
	sampleNames := make([]string, len(row)-1)
	for i := 0; i < len(sampleNames); i++ {
		sampleNames[i] = row[i+1][0]
	}

	fmt.Println("sampleNames =>", sampleNames)
	// sample 1 genelist and cpm val
	//genes := row[0][1:len(row[0])]

	//plenth := len(pid)
	plenth := len(pid_glist)
	//samplePathwayMatx := make([][]float32, plenth)

	samplePathwayMatx := make([]ResScoresStruct, plenth)
	//var samplePathwayMatx [][]float64
	//fmt.Println(len(samplePathwayMatx))

	//fmt.Println(len(glist))

	for i := 0; i < plenth; i++ {
		//nSample := len(sampleNames)
		samplePathwayMatx[i] = ProcessRows(row, pid_glist[i])
		fmt.Println("iteration : ", i, "::pathway::", pid_glist[i].ID)

	}

	//fmt.Println(samplePathwayMatx[0:5])

	alias := r.FormValue("result_file")

	dir, err := os.Getwd()
	fmt.Println("file save here:", dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileLocation := filepath.Join(dir, "files", alias)
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	defer targetFile.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//fresult, err := os.Create(os.Args[2])
	//defer fresult.Close()

	wRes := csv.NewWriter(targetFile)
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

	w.Write([]byte("\n done"))

}

type ResScoresStruct struct {
	Id    string
	score []float32
}

// "ProcessRows return map of pathway as id and scores as val slice"
func ProcessRows(row [][]string, glist RankScore.GeneListStruct) ResScoresStruct {
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
