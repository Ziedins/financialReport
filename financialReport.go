package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

var EvaluationFields = [...]string{
	"Symbol",
	"StockPrice",
	"NumberOfShares",
	"NetIncome",
	"Eps",
	"P/E ratio",
}

var (
	incomeSatementMethod   = "income-statement"
	enterpriseValuesMethod = "enterprise-values"
	financialsUrlBase      = "https://financialmodelingprep.com/stable/"
	resultFileName         = "finances.xlsx"
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

	ReportedCurrency                        string  `json:"reportedCurrency"`
	Cik                                     string  `json:"cik"`
	FilingDate                              string  `json:"filingDate"`
	AcceptedDate                            string  `json:"acceptedDate"`
	FiscalYear                              string  `json:"fiscalYear"`
	Period                                  string  `json:"period"`
	Revenue                                 int64   `json:"revenue"`
	CostOfRevenue                           int64   `json:"costOfRevenue"`
	GrossProfit                             int64   `json:"grossProfit"`
	ResearchAndDevelopmentExpenses          int64   `json:"researchAndDevelopmentExpenses"`
	GeneralAndAdministrativeExpenses        int     `json:"generalAndAdministrativeExpenses"`
	SellingAndMarketingExpenses             int     `json:"sellingAndMarketingExpenses"`
	SellingGeneralAndAdministrativeExpenses int64   `json:"sellingGeneralAndAdministrativeExpenses"`
	OtherExpenses                           int     `json:"otherExpenses"`
	OperatingExpenses                       int64   `json:"operatingExpenses"`
	CostAndExpenses                         int64   `json:"costAndExpenses"`
	NetInterestIncome                       int     `json:"netInterestIncome"`
	InterestIncome                          int     `json:"interestIncome"`
	InterestExpense                         int     `json:"interestExpense"`
	DepreciationAndAmortization             int64   `json:"depreciationAndAmortization"`
	Ebitda                                  int64   `json:"ebitda"`
	Ebit                                    int64   `json:"ebit"`
	NonOperatingIncomeExcludingInterest     int     `json:"nonOperatingIncomeExcludingInterest"`
	OperatingIncome                         int64   `json:"operatingIncome"`
	TotalOtherIncomeExpensesNet             int     `json:"totalOtherIncomeExpensesNet"`
	IncomeBeforeTax                         int64   `json:"incomeBeforeTax"`
	IncomeTaxExpense                        int64   `json:"incomeTaxExpense"`
	NetIncomeFromContinuingOperations       int64   `json:"netIncomeFromContinuingOperations"`
	NetIncomeFromDiscontinuedOperations     int     `json:"netIncomeFromDiscontinuedOperations"`
	OtherAdjustmentsToNetIncome             int     `json:"otherAdjustmentsToNetIncome"`
	NetIncome                               int64   `json:"netIncome"`
	NetIncomeDeductions                     int     `json:"netIncomeDeductions"`
	BottomLineNetIncome                     int64   `json:"bottomLineNetIncome"`
	Eps                                     float64 `json:"eps"`
	EpsDiluted                              float64 `json:"epsDiluted"`
	WeightedAverageShsOut                   int64   `json:"weightedAverageShsOut"`
	WeightedAverageShsOutDil                int64   `json:"weightedAverageShsOutDil"`
}

func main() {
	apiKey := flag.String("apiKey", "enterApiKey", "Api key for financialmodelingprep.com")
	flag.Parse()
	args := flag.Args()
	var symbolEvaluations []Evaluations

	for _, symbol := range args {
		var evaluation Evaluations

		fetchEvaluationData(&evaluation, symbol, enterpriseValuesMethod, *apiKey)
		fetchEvaluationData(&evaluation, symbol, incomeSatementMethod, *apiKey)

		if evaluation.Symbol == "" {
			fmt.Printf("client: financials not found for : %s\n", symbol)
		} else {
			fmt.Printf("Financials gathered for : %s\n", symbol)
			symbolEvaluations = append(symbolEvaluations, evaluation)
		}

	}
	createExcelIfNotExists(symbolEvaluations)
	fmt.Printf("Data stored in excel : %s\n", resultFileName)
}

func fetchEvaluationData(evaluation *Evaluations, symbol string, method string, apiKey string) {
	url := fmt.Sprintf(financialsUrlBase+method+"?symbol=%s&apikey=%s&limit=1", symbol, apiKey)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
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
		addDataToEvaluation(evaluation, resBody)
	}
}

func addDataToEvaluation(evaluation *Evaluations, data []uint8) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.Token()
	if err := dec.Decode(&evaluation); err != nil {
		fmt.Printf("client: could not parse respone: %s\n", err)
		os.Exit(1)
	}
}

func createExcelIfNotExists(symbolEvaluations []Evaluations) {
	now := time.Now()
	sheetName := now.Format("January2006")
	_, err := os.Stat(resultFileName)
	if err == nil {
		fmt.Printf("client: file already exists : %s\n", resultFileName)
	}
	f := excelize.NewFile()

	sheetIndex, _ := f.NewSheet(sheetName)
	for i := range EvaluationFields {
		cell := string('A'+i) + "1"
		f.SetCellValue(sheetName, cell, EvaluationFields[i])
	}

	i := 2
	for _, evaluation := range symbolEvaluations {
		cell := fmt.Sprintf("A%v", i)
		f.SetCellValue(sheetName, cell, evaluation.Symbol)
		cell = fmt.Sprintf("B%v", i)
		f.SetCellValue(sheetName, cell, evaluation.StockPrice)
		cell = fmt.Sprintf("C%v", i)
		f.SetCellValue(sheetName, cell, evaluation.NumberOfShares)
		cell = fmt.Sprintf("D%v", i)
		f.SetCellValue(sheetName, cell, evaluation.NetIncome)
		cell = fmt.Sprintf("E%v", i)
		f.SetCellValue(sheetName, cell, evaluation.Eps)
		cell = fmt.Sprintf("F%v", i)
		profitEarningsRatio := evaluation.StockPrice / evaluation.Eps
		f.SetCellValue(sheetName, cell, profitEarningsRatio)
		i++
	}
	f.SetActiveSheet(sheetIndex)
	if err := f.SaveAs(resultFileName); err != nil {
		fmt.Println(err)
	}
}
