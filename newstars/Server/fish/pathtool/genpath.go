package main

import (
	"bufio"
	"fmt"
	"io"
	"newstars/framework/util/decimal"
	"os"
	"strings"
)

func main() {

	fi, err := os.Open("paths.json")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	count := 0
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		lineData := strings.TrimSpace(string(a))

		arr := strings.Split(lineData, ",")
		if strings.LastIndex(lineData, ",") < 0 {
			continue
		}
		var index int
		if len(lineData) == strings.LastIndex(lineData, ",")+1 {
			index = len(arr) - 2
		} else {
			index = len(arr) - 1
		}
		if index < 0 {
			continue
		}

		str := strings.Replace(arr[index], "]", "", -1)
		dec, _ := decimal.NewFromString(str)
		dec = dec.Mul(decimal.New(1000, 0)).Round(0)
		ivalue := dec.IntPart() + 1000
		// if dec.GreaterThan(decimal.New(ivalue, 0)) {
		// 	ivalue = ivalue + 1
		// }

		fmt.Printf("insert into fish_path_t(id,deadtime,rate) value(%v,%v,100);\n", count, ivalue)
		// fmt.Printf("paths[%v]=%v\n", count, ivalue)
		count++
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
