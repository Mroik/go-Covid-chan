package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const covidSource string = "https://www.worldometers.info/coronavirus/"

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
	if len(results) > 5 {
		return results[1], results[3], results[5], nil
	} else {
		err = errors.New("Country not found")
		return "", "", "", err
	}
}

func getTop() ([5]string, error) {
	var results [5]string
	regex, err := regexp.Compile("(?sm)(<tr style.*?>[[:digit:]]</td>.*?</a>)")
	if err != nil {
		return results, err
	}
	body, err := getStats()
	if err != nil {
		return results, err
	}
	temp := regex.FindAllStringSubmatch(body, 5)
	regex, err = regexp.Compile(">[[:word:]]+</a>")
	if err != nil {
		return results, err
	}
	for x := 0; x < 5; x++ {
		results[x] = strings.Split(temp[x][1], "<td")[2]
		results[x] = regex.FindString(results[x])
	}
	regex, err = regexp.Compile("[[:word:]]+")
	if err != nil {
		return results, err
	}
	for x := 0; x < 5; x++ {
		results[x] = regex.FindString(results[x])
	}
	return results, nil
}
