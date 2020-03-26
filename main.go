package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type AlertFile struct {
	Groups []Group `yaml:"groups,omitempty"`
}

type Group struct {
	Name  string  `yaml:"name"`
	Rules []*Rule `yaml:"rules"`
}

type Rule struct {
	AlertName   string            `yaml:"alert,omitempty"`
	Record      string            `yaml:"record,omitempty"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`

	Enabled       *bool    `yaml:"enabled,omitempty"`
	OverrideRules []string `yaml:"override,omitempty"`
}

func (rules *AlertFile) Exporter() string {

	// Remove override parameters for exporting purposes
	for i, group := range rules.Groups {
		for j, rule := range group.Rules {
			if rule.Enabled != nil && !*rule.Enabled {
				// Remove from slice
				rules.Groups[i].Rules = append(group.Rules[:j], group.Rules[j+1:]...)
			} else {
				group.Rules[j].Enabled = nil
				group.Rules[j].OverrideRules = nil
			}
		}
	}

	d, err := yaml.Marshal(&rules)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return string(d)
}

func (alertFile AlertFile) Override(overrideRule Rule) {
	for i, group := range alertFile.Groups {
		for j, rule := range group.Rules {

			// Skip if the rule is the same and override rules itself
			if rule.AlertName == overrideRule.AlertName || rule.AlertName == "" || len(rule.OverrideRules) > 0 {
				continue
			}

			for _, override := range overrideRule.OverrideRules {
				matched, _ := regexp.MatchString("\\b"+override+"\\b", rule.AlertName)
				if matched {
					negatedFilter := NegateFilterExpression(overrideRule.Expr)
					alertFile.Groups[i].Rules[j].Expr = AppendFilters(negatedFilter, rule.Expr)
					break
				}
			}
		}
	}
}

func LoadRules(input []byte) (*AlertFile, error) {
	var alertRules AlertFile
	err := yaml.Unmarshal(input, &alertRules)
	if err != nil {
		return nil, err
	}

	return &alertRules, nil
}

func containsOperator(needle string, haystack string) bool {
	result, _ := regexp.MatchString(needle, haystack)
	return result
}

func extractFilterExpressions(input string) string {
	re := regexp.MustCompile("{.*}")
	return re.FindAllString(input, 1)[0]
}

func AppendFilters(input string, targetExpr string) string {

	// Handle no brackets case
	if !containsOperator("{", targetExpr) && !containsOperator("}", targetExpr) {
		words := strings.Fields(targetExpr)
		// Append brackets so the expr gets picked up by later proccessing
		words[0] += "{}"
		targetExpr = strings.Join(words, " ")
	}

	var separator = ","
	// Handle empty brackets case
	if containsOperator("{}", targetExpr) {
		separator = ""
	}

	bracketIndex := strings.Index(targetExpr, "}")
	targetExpr = targetExpr[:bracketIndex] + separator + input + targetExpr[bracketIndex:]
	return targetExpr
}

func NegateFilterExpression(input string) string {
	exprFilter := extractFilterExpressions(input)

	exprFilterBody := strings.Replace(exprFilter, "{", "", -1)
	exprFilterBody = strings.Replace(exprFilterBody, "}", "", -1)

	exprFilterBodyElements := strings.Split(exprFilterBody, ",")

	for i, elem := range exprFilterBodyElements {
		if containsOperator("!=", elem) {
			exprFilterBodyElements[i] = strings.Replace(elem, "!=", "=", 1)
		} else if containsOperator("!~", elem) {
			exprFilterBodyElements[i] = strings.Replace(elem, "!~", "~", 1)
		} else if containsOperator("=", elem) {
			exprFilterBodyElements[i] = strings.Replace(elem, "=", "!=", 1)
		} else if containsOperator("=~", elem) {
			exprFilterBodyElements[i] = strings.Replace(elem, "=~", "!~", 1)
		}
	}

	return strings.Join(exprFilterBodyElements[:], ",")
}

func getFilePaths(rootPath string) []string {
	var filePaths []string

	files, err := ioutil.ReadDir(rootPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		filePaths = append(filePaths, rootPath+"/"+f.Name())
	}

	return filePaths
}

func loadAlertFile(filePath string) *AlertFile {
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Errorf("Failed to read file " + filePath)
		return nil
	}
	alertFile, err := LoadRules(input)
	if err != nil || len(alertFile.Groups) == 0 {
		return nil
	}
	return alertFile
}

func main() {

	if len(os.Args) != 2 {
		panic("No arguments provided")
	}

	filePaths := os.Args[1]
	files := getFilePaths(filePaths)

	var alertFiles []AlertFile

	for _, file := range files {
		alertFile := loadAlertFile(file)
		if alertFile != nil {
			alertFiles = append(alertFiles, *alertFile)
		}
	}

	var alertFile AlertFile = AlertFile{}
	for _, elem := range alertFiles {
		alertFile.Groups = append(alertFile.Groups, elem.Groups...)
	}

	for _, group := range alertFile.Groups {
		for _, rule := range group.Rules {
			if len(rule.OverrideRules) > 0 {
				alertFile.Override(*rule)
			}
		}
	}

	fmt.Print(alertFile.Exporter())
}
