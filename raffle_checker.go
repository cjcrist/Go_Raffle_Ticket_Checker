package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// This is required since the server returns json wrapped in a list.
type LottoAPIResponse struct {
	Collection []DataHead
}


type DataHead struct {
	Data `json:"data"`
}


type Data struct {
	GameByCode `json:"gameByCode"`
}


type GameByCode struct {
	LogicalGameIdentifier string `json:"logicalGameIdentifier"`
	DrawResultsBetweenDates []DrawResultsBetweenDate `json:"drawResultsBetweenDates"`
	Typename string `json:"__typename"`
}


type DrawResultsBetweenDate struct {
	DrawDate string `json:"drawDate"`
	DrawSequence int `json:"drawSequence"`
	HasPayoutData bool `json:"hasPayoutData"`
	WinningNumber `json:"winningNumbers"`
	Typename string `json:"__typename"`
}


type WinningNumber struct {
	DrawNumbers []int `json:"drawNumbers, omitempty"`
	Powerball string `json:"powerball, omitempty"`
	Powerplay string `json:"powerplay, omitempty"`
	Megaball string `json:"megaball, omitempty"`
	Megaplier string `json:"megaplier, omitempty"`
	Luckyball string `json:"luckyball, omitempty"`
	Typename string `json:"__typename, omitempty"`
}


func makeHTTPRequest(url string, jsonStr []byte) []byte {
	client := &http.Client{}
	req, reqErr := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	if reqErr != nil {
		panic(reqErr)
	}

	resp, respErr := client.Do(req)
	if respErr != nil {
		panic(respErr)
	}

	defer resp.Body.Close()
	jsonResponseData, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		panic(readErr)
	}
	return jsonResponseData
}


func printAllNumbers(winners map[string][]int){
	// An array of keys in winners
	keys := make([]string, 0, len(winners))
	for k := range winners {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println("The winning numbers are:")
	//Sorted print
	for _, k := range keys {
		date, err := time.Parse(time.RFC3339, k)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s %02d, %d  --|--> %v%v%v\n", date.Month().String()[:3], date.Day(), date.Year(),+
			winners[k][0],	winners[k][1], winners[k][2])
	}
}


func checkWinningNumber(user []string, winners map[string][]int) {
	winDict := make(map[string]string)
	// Compare winning numbers with user input numbers. Store in winDict
	for k, v := range winners {
		// Each winNumber is stored as an int array of len(3). User input is stored as a string.
		// There may be a more efficient way of comparing values
		num1 := strconv.Itoa(v[0])
		num2 := strconv.Itoa(v[1])
		num3 := strconv.Itoa(v[2])
		winNumber := num1 + num2 + num3
		for _, userInputNum := range user {
			if userInputNum == winNumber {
				winDict[k] = winNumber
			}
		}
	}
	if len(winDict) != 0 {
		if len(winDict) > 1 {
			fmt.Println("\t~~~~~Wowsers, you won multiple times!!~~~~~")
		}
		fmt.Println("\t\t~~~~~Congratulations!~~~~~")
		for k, v := range winDict {
			date, err := time.Parse(time.RFC3339, k)
			if err != nil {
				panic(err)
			}
			fmt.Printf("\t[!] Your ticket #%s won on %s %02d, %d!\n", v, date.Month().String()[:3],+
				date.Day(),	date.Year())
		}
	} else {
		fmt.Print("\n\t\t    ¯\\_(ツ)_/¯   \n\t~~~You have no winning numbers!~~~\n")
	}
}


func main() {
	url := "https://www.michiganlottery.com/api"
	jan1 := time.Date(time.Now().Year(), time.January, 01, 10, 0, 0, 0, time.Local)
	startDate:= jan1.Format(time.RFC3339)
	today := time.Now()
	endDate := today.Format(time.RFC3339)
	var jsonStr = []byte(`[
		{
		"query": "query Game($gameCode: String!, $startDateString: String!, $endDateString: String!) ` +
		`{\n  gameByCode(code: $gameCode) {\n    logicalGameIdentifier\n    drawResultsBetweenDates(startDateString: ` +
		`$startDateString, endDateString: $endDateString) {\n      drawDate\n      drawSequence\n      hasPayoutData\n`+
		`winningNumbers {\n        drawNumbers\n        powerball\n        powerplay\n        megaball\n        ` +
		`megaplier\n        luckyball\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n}\n",
			"variables": {
			"gameCode": "3",
				"startDateString": "` + startDate + `",
				"endDateString": "` + endDate + `"
			},
		"operationName": "Game"
		}
	]`)
	// make HTTP POST request and retrieve data
	var jsonResponseData = makeHTTPRequest(url, jsonStr)
	// points to the head of the data being returned, which is json wrapped in a list.
	lotto := make([]DataHead, 0)
	lottoErrs := json.Unmarshal(jsonResponseData, &lotto)
	if lottoErrs != nil {
		panic(lottoErrs)
	}
	// Server returns each winning number as an array of int len(3)
	winDatesAndNumbers := make(map[string][]int)
	// Iterates through all of the numbers and maps the winning numbers to winDatesAndNumbers dictionary
	for _, lottoData := range lotto {
		for i, _ := range lottoData.DrawResultsBetweenDates{
			iDate := lottoData.DrawResultsBetweenDates[i].DrawDate
			t, _ := time.Parse(time.RFC3339, iDate)
			if t.Weekday() == time.Saturday {
				winDatesAndNumbers[iDate] = lottoData.DrawResultsBetweenDates[i].DrawNumbers
			}
		}
	}
	year := today.Year()
	fmt.Printf("[*][*][*]\n" +
		"This tool was created to check your raffle ticket number(s) against the winning numbers in the \n" +
		"%d Weapon-a-Week Raffle.\n" +
		"[*][*][*]\n\n", year)
	usage := "[*] Usage: \n" +
		"[*] A: Shows all winning numbers and draw dates for the lottery as of today.\n" +
		"[*] Q: Exits the program immediately.\n" +
		"[*] H: Shows this help menu.\n" +
		"[*]\n" +
		"[*]\n" +
		"[*] Checking your numbers: \n[*] \tSearching the database to see if you have a winning number is easy.\n" +
		"[*] \tEnter each of your numbers 1 at a time, followed by pressing ENTER.\n" +
		"[*] \tWhen you are finished entering your numbers, press the ENTER to proceed.\n\n"
	fmt.Print(usage)
	fmt.Println("\n[*] Enter a number, or enter 'A' to view all numbers.\n[*] Press ENTER when finished.")

	userNums := make([]string, 0)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		if input.Text() == "" {
			fmt.Print("\n")
			break
		}
		switch input.Text() {
		case "A", "a":
			printAllNumbers(winDatesAndNumbers)
			os.Exit(0)
		case "Q", "q":
			fmt.Println("Goodbye!")
			os.Exit(0)
		case "H", "h":
			fmt.Print(usage)
			fmt.Println("\n[*] Enter a number, or enter 'A' to view all numbers.\n[*] Press ENTER when finished.")
			continue
		default:
			if m, _ := regexp.MatchString("^[0-9]+$", input.Text()); !m {
				fmt.Println("[*] Oops, you typed a letter or a space.  Your ticket number should be a 3 digit " +
					"number. Enter each 3 digit number once and without spaces. \n" +
					"[*] Type 'H' for help, or enter a valid ticket number.")
				continue
			} else if len(input.Text()) != 3 {
				fmt.Println("[*] Um, that's not quite right. Your ticket number should be a 3 digit number. \n" +
					"[*] Type 'H' for help, or enter a valid ticket number.")
				continue
			} else {
				userNums = append(userNums, input.Text())
				fmt.Println("[*] Enter your next number, or press ENTER to search the database.")
				continue
			}
		}
	}
	// Check the user input numbers against the winning numbers.
	checkWinningNumber(userNums, winDatesAndNumbers)
	fmt.Print("\n\n")
	os.Exit(0)
}