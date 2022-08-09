package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	utils "github.com/eskriett/confusables"
)

var errDownload = errors.New("unable to download confusables")

const (
	url = "https://www.unicode.org/Public/security/latest/confusables.txt"
)

const sourceFile = `package confusables

// THIS FILE WAS AUTOGENERATED - DO NOT EDIT

var confusables = map[rune]string{
{{- range $key, $value := .Confusables}}
	{{ $key }}: {{ $value }},
{{- end}}
}

var descriptions = map[string]string{
{{- range $key, $value := .Descriptions}}
	{{ $key }}: {{ $value }},
{{- end}}
}
`

func main() {
	if err := buildTable(); err != nil {
		log.Fatal("unable to build tables: ", err)
	}
}

func buildTable() error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errDownload
	}

	confusables := map[string]string{}
	descriptions := map[string]string{}

	// Extract confusables from downloaded file
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if err := parseLine(line, confusables, descriptions); err != nil && !errors.Is(err, utils.ErrIgnoreLine) {
			return err
		}
	}

	amendments, err := os.Open("scripts/amendments.txt")
	if err != nil {
		return err
	}

	defer amendments.Close()

	scanner = bufio.NewScanner(amendments)
	for scanner.Scan() {
		line := scanner.Text()

		if err := parseLine(line, confusables, descriptions); err != nil && !errors.Is(err, utils.ErrIgnoreLine) {
			return err
		}
	}

	// Output a mapping file
	tmpl := template.New("tables.go")

	tmpl, err = tmpl.Parse(sourceFile)
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	f, err := os.Create("tables.go")
	if err != nil {
		return fmt.Errorf("unable to create tables.go: %w", err)
	}

	defer f.Close()

	if err := tmpl.Execute(f, struct {
		Confusables  map[string]string
		Descriptions map[string]string
	}{
		Confusables:  confusables,
		Descriptions: descriptions,
	}); err != nil {
		return fmt.Errorf("unable to execute template: %w", err)
	}

	return nil
}

func parseLine(line string, confusables map[string]string, descriptions map[string]string) error {
	entry, err := utils.ParseLine(line)
	if err != nil {
		return err
	}

	sourceStr := string(entry.Source)
	if _, ok := descriptions[sourceStr]; !ok {
		descriptions[strconv.Quote(sourceStr)] = strconv.Quote(entry.Description.From)
	}

	if _, ok := descriptions[entry.Target]; !ok {
		descriptions[strconv.Quote(entry.Target)] = strconv.Quote(entry.Description.To)
	}

	confusables[fmt.Sprintf("0x%.8X", entry.Source)] = fmt.Sprintf("%+q", entry.Target)

	return nil
}
