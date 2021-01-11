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
	"strings"

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

	input := bufio.NewScanner(f)
	var list []string
	for input.Scan() {
		list = append(list, input.Text())
	}
	var pathwayId = make([]string, len(list))
	var geneList = make([][]string, len(list))

	for i, v := range list {
		res := strings.Split(v, "\t")

		//pathwayId[i] = res[1]
		pathwayId[i] = res[0]
		geneList[i] = res[2:len(res)]
	}
	//return pathwayId, geneList

	//var pid, glist = readFile(f1)
	pid := pathwayId
	glist := geneList
	f.Close()

	fmt.Println("pid::: ", pid[0:5])

	csvFile, _, err := r.FormFile("matxFile")
	//csvFile, _ := os.Open(matx)
	reader := csv.NewReader(bufio.NewReader(csvFile))

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
	// sample 1 genelist and cpm val
	genes := row[0][1:len(row[0])]

	plenth := len(pid)
	samplePathwayMatx := make([][]float32, plenth)
	//var samplePathwayMatx [][]float64
	fmt.Println(len(samplePathwayMatx))

	fmt.Println(len(glist))

	for i := 0; i < plenth; i++ {
		nSample := len(sampleNames)
		allScores2 := make([]float32, nSample)

		for j := 0; j < nSample; j++ {
			var s2 []string
			s2 = append(s2, row[j+1]...)
			s3 := RankScore.RowToFloatVec(s2)
			allScores2[j] = float32(RankScore.RankScore(genes, s3, glist[i]))
		}
		//samplePathwayMatx = append(samplePathwayMatx,allScores2)
		samplePathwayMatx[i] = allScores2
		fmt.Println("iteration : ", i, "::pathway::", pid[i])

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
	for i, record := range samplePathwayMatx {
		rec := make([]string, (len(sampleNames) + 1))
		rec[0] = pid[i]
		for i, v := range record {
			rec[i+1] = fmt.Sprint(v)
		}
		if err := wRes.Write(rec); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	w.Write([]byte("\n done"))

}

// func readFile(f *os.File) ([]string, [][]string) {
// 	input := bufio.NewScanner(f)
// 	var list []string
// 	for input.Scan() {
// 		list = append(list, input.Text())
// 	}
// 	var pathwayId = make([]string, len(list))
// 	var geneList = make([][]string, len(list))

// 	for i, v := range list {
// 		res := strings.Split(v, "\t")

// 		//pathwayId[i] = res[1]
// 		pathwayId[i] = res[0]
// 		geneList[i] = res[2:len(res)]
// 	}
// 	return pathwayId, geneList
// }
