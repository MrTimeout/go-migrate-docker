package main

import (
	"bufio"
	"os"
	"regexp"
)

func groupRegexFirstSubMatchFromFile(filename string, re *regexp.Regexp) ([]string, error) {
	var result []string

	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if matches := re.FindStringSubmatch(scanner.Text()); len(matches) == 2 {
			result = append(result, matches[1])
		}
	}

	return result, nil
}

func matchesOfStringArr(re *regexp.Regexp, input []string) []string {
	var result []string

	for _, i := range input {
		if re.MatchString(i) {
			result = append(result, i)
		}
	}

	return result
}
