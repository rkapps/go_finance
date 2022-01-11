package utils

import (
	"bufio"
	"log"
	"os"
	"strings"
)

//LoadLinesFromFile loads lines from a file
func LoadLinesFromFile(fileName string) []string {
	var lines []string
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Cannot open file: %s", fileName)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

//LoadFromFile loads lines from a csv file
func LoadFromFile(fileName string, sep string) [][]string {

	var entries [][]string
	lines := LoadLinesFromFile(fileName)

	for _, line := range lines {
		// log.Println(scanner.Text())
		entry := strings.Split(line, sep)
		entries = append(entries, entry)
	}
	return entries
}

//WriteLinesToFile writes lines to a file
func WriteLinesToFile(fileName string, lines []string) error {

	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Cannot open file: %s", fileName)
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		writer.WriteString(line)
		writer.WriteString("\n")
	}
	writer.Flush()

	return nil
}

//WriteToFile writes lines to a file
func WriteToFile(fileName string, entries [][]string, sep string) error {

	var lines []string
	for _, entry := range entries {
		line := strings.Join(entry, sep)
		lines = append(lines, line)
	}
	return WriteLinesToFile(fileName, lines)
}
