package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

type Evaluations struct {
	Symbol                      string  `json:"symbol"`
	Date                        string  `json:"date"`
	StockPrice                  float64 `json:"stockPrice"`
	NumberOfShares              int64   `json:"numberOfShares"`
	MarketCapitalization        int64   `json:"marketCapitalization"`
	MinusCashAndCashEquivalents int64   `json:"minusCashAndCashEquivalents"`
	AddTotalDebt                int64   `json:"addTotalDebt"`
	EnterpriseValue             int64   `json:"enterpriseValue"`
}

func main() {
	args := os.Args[1:]
	var symbolEvaluations []Evaluations
	for _, arg := range args {

		url := fmt.Sprintf("https://financialmodelingprep.com/stable/enterprise-values?symbol=%s&apikey=%s&limit=1", arg, "apiKey")
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fmt.Printf("client: could not create request: %s\n", err)
			os.Exit(1)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("client: got response!\n")
		fmt.Printf("client: status code: %d\n", res.StatusCode)
		if res.StatusCode == 200 {
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("client: could not read response body: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("client: response body: %s\n", resBody)
			var enterpriseResponses []Evaluations
			if err := json.Unmarshal(resBody, &enterpriseResponses); err != nil {
				fmt.Printf("client: could parse response: %s\n", err)
				os.Exit(1)
			}

			fmt.Println(enterpriseResponses[0])
			fmt.Println(err)
			symbolEvaluations = append(symbolEvaluations, enterpriseResponses[0])
		}
		createExcelIfNotExists(symbolEvaluations)
	}
}

func createExcelIfNotExists(symbolEvaluations []Evaluations) {
	filename := "finances.xlsx"
	now := time.Now()
	sheetName := now.Format("January2006")
	_, err := os.Stat(filename)
	if err == nil {
		fmt.Printf("file already exists : %s\n, error: %s\n", filename, err)
	}
	f := excelize.NewFile()
	sheetIndex, _ := f.NewSheet(sheetName)
	i := 1
	for _, evaluation := range symbolEvaluations {

		cell := fmt.Sprintf("A%v", i)
		fmt.Println(cell, evaluation.Symbol)
		f.SetCellValue(sheetName, cell, evaluation.Symbol)
		cell = fmt.Sprintf("B%v", i)
		f.SetCellValue(sheetName, cell, evaluation.StockPrice)
		i++
	}
	f.SetActiveSheet(sheetIndex)
	if err := f.SaveAs(filename); err != nil {
		fmt.Println(err)
	}
}
