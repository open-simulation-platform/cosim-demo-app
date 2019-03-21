package cse

import (
	"log"
	"strconv"
	"strings"
)

func parseFloat(argument string) float64 {
	f, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		log.Fatal(err)
		return 0.0
	}
	return f
}


func strCat(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}