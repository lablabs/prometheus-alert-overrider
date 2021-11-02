package main

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestOverride(t *testing.T) {
	cases, err := os.ReadDir("./testCases")
	require.NoError(t, err)

	for _, inputFile := range cases {
		if strings.Contains(inputFile.Name(), ".input.") {
			expectedPath := "./testCases/" + strings.Replace(inputFile.Name(), ".input", ".expected", 1)

			t.Run(strings.Replace(inputFile.Name(), ".input.yml", "", 1), func(t *testing.T) {
				file, err := loadAlertFile("./testCases/" + inputFile.Name())
				require.NoError(t, err)
				for _, group := range file.Groups {
					for _, rule := range group.Rules {
						if len(rule.OverrideRules) > 0 {
							file.Override(*rule)
						}
					}
				}

				output := file.Exporter()

				expectedFile, err := os.Open(expectedPath)
				require.NoError(t, err)

				expected, err := ioutil.ReadAll(expectedFile)
				require.NoError(t, err)

				require.Equal(t, string(expected), output)
			})
		}
	}
}
