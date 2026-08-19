// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

type ExpectedOutput struct {
	Entities      []ExpectedEntity       `json:"entities"`
	Relationships []ExpectedRelationship `json:"relationships"`
}

type ExpectedEntity struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ExpectedRelationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

type TestCaseResult struct {
	Name            string
	EntityPrecision float64
	EntityRecall    float64
	RelPrecision    float64
	RelRecall       float64
	Passed          bool
	Errors          []string
}

func main() {
	corpusDir := "garuda-bench/corpus/cases"
	if len(os.Args) > 1 {
		corpusDir = os.Args[1]
	}

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read corpus directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⚖️  GARUDA BENCHMARK RUNNER (V5 Truth-First Suite)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var allPassed = true
	var totalEntP, totalEntR, totalRelP, totalRelR float64
	var caseCount float64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		casePath := filepath.Join(corpusDir, entry.Name())
		res := evaluateCase(casePath)
		caseCount++

		totalEntP += res.EntityPrecision
		totalEntR += res.EntityRecall
		totalRelP += res.RelPrecision
		totalRelR += res.RelRecall

		status := "✅ PASS"
		if !res.Passed {
			status = "❌ FAIL"
			allPassed = false
		}

		fmt.Printf("\n[%s] %s\n", status, res.Name)
		fmt.Printf("   Entities:      Precision: %5.1f%% | Recall: %5.1f%%\n", res.EntityPrecision*100, res.EntityRecall*100)
		fmt.Printf("   Relationships: Precision: %5.1f%% | Recall: %5.1f%%\n", res.RelPrecision*100, res.RelRecall*100)

		for _, errStr := range res.Errors {
			fmt.Printf("   ⚠️  %s\n", errStr)
		}
	}

	if caseCount > 0 {
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📊 AGGREGATE BENCHMARK SCORECARD (%d fixtures)\n", int(caseCount))
		fmt.Printf("   Overall Entity Precision:       %5.1f%%\n", (totalEntP/caseCount)*100)
		fmt.Printf("   Overall Entity Recall:          %5.1f%%\n", (totalEntR/caseCount)*100)
		fmt.Printf("   Overall Relationship Precision: %5.1f%%\n", (totalRelP/caseCount)*100)
		fmt.Printf("   Overall Relationship Recall:    %5.1f%%\n", (totalRelR/caseCount)*100)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	if !allPassed {
		os.Exit(1)
	}
}

func evaluateCase(casePath string) TestCaseResult {
	caseName := filepath.Base(casePath)
	res := TestCaseResult{Name: caseName, Passed: true}

	expectedFile := filepath.Join(casePath, "expected.json")
	expData, err := os.ReadFile(expectedFile)
	if err != nil {
		res.Passed = false
		res.Errors = append(res.Errors, fmt.Sprintf("Missing expected.json: %v", err))
		return res
	}

	var expected ExpectedOutput
	if err := json.Unmarshal(expData, &expected); err != nil {
		res.Passed = false
		res.Errors = append(res.Errors, fmt.Sprintf("Malformed expected.json: %v", err))
		return res
	}

	actual, err := analyzer.Extract(casePath)
	if err != nil {
		res.Passed = false
		res.Errors = append(res.Errors, fmt.Sprintf("Analysis crashed: %v", err))
		return res
	}

	var symbolEntities []analyzer.Entity
	for _, e := range actual.Entities {
		if e.Kind != analyzer.KindFile && e.Kind != analyzer.KindPackage && e.Kind != analyzer.KindDirectory && e.Name != "main" {
			symbolEntities = append(symbolEntities, e)
		}
	}

	matchedEntities := 0
	for _, exp := range expected.Entities {
		found := false
		for _, act := range symbolEntities {
			if act.Name == exp.Name && strings.EqualFold(string(act.Kind), exp.Kind) {
				found = true
				break
			}
		}
		if found {
			matchedEntities++
		} else {
			res.Errors = append(res.Errors, fmt.Sprintf("Missing expected entity: '%s' (%s)", exp.Name, exp.Kind))
		}
	}

	if len(symbolEntities) > 0 {
		res.EntityPrecision = float64(matchedEntities) / float64(len(symbolEntities))
	}
	if len(expected.Entities) > 0 {
		res.EntityRecall = float64(matchedEntities) / float64(len(expected.Entities))
	}

	matchedRels := 0
	for _, exp := range expected.Relationships {
		found := false
		for _, act := range actual.Relationships {
			if strings.EqualFold(act.From, exp.From) && strings.EqualFold(act.To, exp.To) && strings.EqualFold(act.Type, exp.Type) {
				found = true
				matchedRels++
				break
			}
		}
		if !found {
			res.Errors = append(res.Errors, fmt.Sprintf("Missing relationship: %s -[%s]-> %s", exp.From, exp.Type, exp.To))
		}
	}

	if len(actual.Relationships) > 0 {
		res.RelPrecision = float64(matchedRels) / float64(len(actual.Relationships))
	} else if len(expected.Relationships) == 0 {
		res.RelPrecision = 1.0
	}

	if len(expected.Relationships) > 0 {
		res.RelRecall = float64(matchedRels) / float64(len(expected.Relationships))
	} else {
		res.RelRecall = 1.0
	}

	if res.EntityRecall < 0.99 || res.RelRecall < 0.99 || len(res.Errors) > 0 {
		res.Passed = false
	}

	return res
}
