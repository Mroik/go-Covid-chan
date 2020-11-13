package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var covidSource string = "https://www.worldometers.info/coronavirus/"

func getStats() (string, error) {
	resp, err := http.Get(covidSource)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return string(body), nil
}

//This function filters covid stats by country
func getCountry(country string) (string, string, string, error) {
	regex, err := regexp.Compile("(?sim)>" + country + "<.*?data-continent")
	if err != nil {
		return "", "", "", err
	}
	body, err := getStats()
	if err != nil {
		return "", "", "", err
	}
	results := strings.Split(regex.FindString(body), "<td")
	for x := 0; x < len(results); x++ {
		temp := strings.Split(results[x], ">")
		if len(temp) > 1 {
			results[x] = temp[1]
		}
		temp = strings.Split(results[x], "<")
		if len(temp) > 1 {
			results[x] = temp[0]
		}
	}
	return results[1], results[3], results[5], nil
}
