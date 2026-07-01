package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
)

// ValidatorConfig holds configuration for the validator
type ValidatorConfig struct {
	PythonPath    string
	GoToken       string
	PythonToken   string
	OutputDir     string
	Verbose       bool
	MethodsToTest []string
}

// ValidationResult represents the result of a validation test
type ValidationResult struct {
	Method       string        `json:"method"`
	Passed       bool          `json:"passed"`
	GoResult     interface{}   `json:"go_result,omitempty"`
	PythonResult interface{}   `json:"python_result,omitempty"`
	Error        string        `json:"error,omitempty"`
	Duration     time.Duration `json:"duration"`
}

// ValidationReport represents the full validation report
type ValidationReport struct {
	Timestamp   time.Time          `json:"timestamp"`
	TotalTests  int                `json:"total_tests"`
	Passed      int                `json:"passed"`
	Failed      int                `json:"failed"`
	SuccessRate float64            `json:"success_rate"`
	Results     []ValidationResult `json:"results"`
}

func main() {
	config := parseFlags()

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Run validation
	validator := NewValidator(config)
	report, err := validator.Run()
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Save report
	reportPath := filepath.Join(config.OutputDir, fmt.Sprintf("validation_report_%d.json", time.Now().Unix()))
	if err := saveReport(report, reportPath); err != nil {
		log.Fatalf("Failed to save report: %v", err)
	}

	// Print summary
	printSummary(report)

	// Exit with non-zero if any tests failed
	if report.Failed > 0 {
		os.Exit(1)
	}
}

func parseFlags() *ValidatorConfig {
	config := &ValidatorConfig{}

	flag.StringVar(&config.PythonPath, "python", "/Users/eshaffer321/code/monarchmoney", "Path to Python client")
	flag.StringVar(&config.GoToken, "go-token", os.Getenv("MONARCH_TOKEN"), "Token for Go client")
	flag.StringVar(&config.PythonToken, "python-token", os.Getenv("MONARCH_TOKEN"), "Token for Python client")
	flag.StringVar(&config.OutputDir, "output", "./validation_results", "Output directory for results")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")

	// Parse method list
	methodList := flag.String("methods", "", "Comma-separated list of methods to test (empty for all)")

	flag.Parse()

	if *methodList != "" {
		config.MethodsToTest = strings.Split(*methodList, ",")
	} else {
		// Default methods to test
		config.MethodsToTest = []string{
			"get_accounts",
			"get_account_type_options",
			"get_transactions",
			"get_transaction_categories",
			"get_transaction_tags",
			"get_budgets",
			"get_institutions",
		}
	}

	return config
}

// Validator handles the validation process
type Validator struct {
	config       *ValidatorConfig
	goClient     *monarch.Client
	pythonScript string
}

// NewValidator creates a new validator
func NewValidator(config *ValidatorConfig) *Validator {
	// Create Go client
	goClient, err := monarch.NewClientWithToken(config.GoToken)
	if err != nil {
		log.Fatalf("Failed to create Go client: %v", err)
	}

	return &Validator{
		config:       config,
		goClient:     goClient,
		pythonScript: filepath.Join(config.PythonPath, "validation_script.py"),
	}
}

// Run executes the validation tests
func (v *Validator) Run() (*ValidationReport, error) {
	// Create Python validation script
	if err := v.createPythonScript(); err != nil {
		return nil, fmt.Errorf("failed to create Python script: %w", err)
	}

	report := &ValidationReport{
		Timestamp: time.Now(),
		Results:   make([]ValidationResult, 0),
	}

	ctx := context.Background()

	// Test each method
	for _, method := range v.config.MethodsToTest {
		if v.config.Verbose {
			fmt.Printf("Testing %s...\n", method)
		}

		result := v.testMethod(ctx, method)
		report.Results = append(report.Results, result)

		if result.Passed {
			report.Passed++
		} else {
			report.Failed++
		}
	}

	report.TotalTests = len(report.Results)
	if report.TotalTests > 0 {
		report.SuccessRate = float64(report.Passed) / float64(report.TotalTests) * 100
	}

	return report, nil
}

// testMethod tests a single method
func (v *Validator) testMethod(ctx context.Context, method string) ValidationResult {
	start := time.Now()
	result := ValidationResult{
		Method: method,
	}

	// Get Go result
	goResult, goErr := v.executeGoMethod(ctx, method)
	if goErr != nil {
		result.Error = fmt.Sprintf("Go error: %v", goErr)
		result.Duration = time.Since(start)
		return result
	}

	// Get Python result
	pythonResult, pyErr := v.executePythonMethod(method)
	if pyErr != nil {
		result.Error = fmt.Sprintf("Python error: %v", pyErr)
		result.Duration = time.Since(start)
		return result
	}

	// Compare results
	result.GoResult = goResult
	result.PythonResult = pythonResult
	result.Passed = v.compareResults(goResult, pythonResult)
	result.Duration = time.Since(start)

	if !result.Passed && v.config.Verbose {
		fmt.Printf("  Mismatch in %s:\n", method)
		fmt.Printf("    Go: %v\n", goResult)
		fmt.Printf("    Python: %v\n", pythonResult)
	}

	return result
}

// executeGoMethod executes a method using the Go client
func (v *Validator) executeGoMethod(ctx context.Context, method string) (interface{}, error) {
	switch method {
	case "get_accounts":
		accounts, err := v.goClient.Accounts.List(ctx)
		if err != nil {
			return nil, err
		}
		// Convert to comparable format
		return v.normalizeAccounts(accounts), nil

	case "get_account_type_options":
		types, err := v.goClient.Accounts.GetTypes(ctx)
		if err != nil {
			return nil, err
		}
		return v.normalizeAccountTypes(types), nil

	case "get_transactions":
		// Get last 10 transactions
		result, err := v.goClient.Transactions.Query().
			Limit(10).
			Execute(ctx)
		if err != nil {
			return nil, err
		}
		return v.normalizeTransactions(result.Transactions), nil

	case "get_transaction_categories":
		categories, err := v.goClient.Transactions.Categories().List(ctx)
		if err != nil {
			return nil, err
		}
		return v.normalizeCategories(categories), nil

	case "get_transaction_tags":
		tags, err := v.goClient.Tags.List(ctx)
		if err != nil {
			return nil, err
		}
		return v.normalizeTags(tags), nil

	case "get_budgets":
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, -1)

		budgets, err := v.goClient.Budgets.List(ctx, startOfMonth, endOfMonth)
		if err != nil {
			return nil, err
		}
		return v.normalizeBudgets(budgets), nil

	case "get_institutions":
		institutions, err := v.goClient.Institutions.List(ctx)
		if err != nil {
			return nil, err
		}
		return v.normalizeInstitutions(institutions), nil

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// executePythonMethod executes a method using the Python client
func (v *Validator) executePythonMethod(method string) (interface{}, error) {
	// Execute Python script
	cmd := exec.Command("python3", v.pythonScript, method, v.config.PythonToken)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("Python script failed: %s", string(exitErr.Stderr))
		}
		return nil, err
	}

	// Parse JSON output
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Python output: %w", err)
	}

	return result, nil
}

// compareResults compares Go and Python results
func (v *Validator) compareResults(goResult, pythonResult interface{}) bool {
	// Normalize both to JSON and compare
	goJSON, err := json.Marshal(goResult)
	if err != nil {
		return false
	}

	pyJSON, err := json.Marshal(pythonResult)
	if err != nil {
		return false
	}

	// Parse back to maps for comparison
	var goMap, pyMap interface{}
	_ = json.Unmarshal(goJSON, &goMap)
	_ = json.Unmarshal(pyJSON, &pyMap)

	// Use reflect.DeepEqual for comparison
	return reflect.DeepEqual(goMap, pyMap)
}

// Normalization functions to convert Go structs to comparable format

func (v *Validator) normalizeAccounts(accounts []*monarch.Account) []map[string]interface{} {
	result := make([]map[string]interface{}, len(accounts))
	for i, acc := range accounts {
		result[i] = map[string]interface{}{
			"id":             acc.ID,
			"displayName":    acc.DisplayName,
			"currentBalance": acc.CurrentBalance,
			"isAsset":        acc.IsAsset,
			"isManual":       acc.IsManual,
		}
	}
	return result
}

func (v *Validator) normalizeAccountTypes(types []*monarch.AccountType) []map[string]interface{} {
	result := make([]map[string]interface{}, len(types))
	for i, t := range types {
		result[i] = map[string]interface{}{
			"type":    t.Type.Name,
			"display": t.Type.Display,
		}
	}
	return result
}

func (v *Validator) normalizeTransactions(txns []*monarch.Transaction) []map[string]interface{} {
	result := make([]map[string]interface{}, len(txns))
	for i, txn := range txns {
		result[i] = map[string]interface{}{
			"id":       txn.ID,
			"amount":   txn.Amount,
			"date":     txn.Date.Format("2006-01-02"),
			"merchant": txn.Merchant,
		}
	}
	return result
}

func (v *Validator) normalizeCategories(categories []*monarch.TransactionCategory) []map[string]interface{} {
	result := make([]map[string]interface{}, len(categories))
	for i, cat := range categories {
		result[i] = map[string]interface{}{
			"id":   cat.ID,
			"name": cat.Name,
			"icon": cat.Icon,
		}
	}
	return result
}

func (v *Validator) normalizeTags(tags []*monarch.Tag) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tags))
	for i, tag := range tags {
		result[i] = map[string]interface{}{
			"id":    tag.ID,
			"name":  tag.Name,
			"color": tag.Color,
		}
	}
	return result
}

func (v *Validator) normalizeBudgets(budgets []*monarch.Budget) []map[string]interface{} {
	result := make([]map[string]interface{}, len(budgets))
	for i, budget := range budgets {
		result[i] = map[string]interface{}{
			"id":       budget.ID,
			"amount":   budget.Amount,
			"spent":    budget.Spent,
			"rollover": budget.Rollover,
		}
	}
	return result
}

func (v *Validator) normalizeInstitutions(institutions []*monarch.Institution) []map[string]interface{} {
	result := make([]map[string]interface{}, len(institutions))
	for i, inst := range institutions {
		result[i] = map[string]interface{}{
			"id":     inst.ID,
			"name":   inst.Name,
			"status": inst.Status,
		}
	}
	return result
}

// createPythonScript creates the Python validation script
func (v *Validator) createPythonScript() error {
	script := `#!/usr/bin/env python3
import sys
import json
import asyncio
sys.path.insert(0, '` + v.config.PythonPath + `')

from monarchmoney import MonarchMoney

async def main():
    method = sys.argv[1]
    token = sys.argv[2]
    
    mm = MonarchMoney(token=token)
    
    if method == "get_accounts":
        result = await mm.get_accounts()
        accounts = result.get('accounts', [])
        normalized = [{
            'id': acc['id'],
            'displayName': acc['displayName'],
            'currentBalance': acc['currentBalance'],
            'isAsset': acc['isAsset'],
            'isManual': acc['isManual']
        } for acc in accounts]
        print(json.dumps(normalized))
        
    elif method == "get_account_type_options":
        result = await mm.get_account_type_options()
        types = result.get('accountTypeOptions', [])
        normalized = [{
            'type': t['type']['name'],
            'display': t['type']['display']
        } for t in types]
        print(json.dumps(normalized))
        
    elif method == "get_transactions":
        result = await mm.get_transactions(limit=10)
        txns = result.get('allTransactions', {}).get('results', [])
        normalized = [{
            'id': txn['id'],
            'amount': txn['amount'],
            'date': txn['date'][:10],
            'merchant': txn.get('merchant', {}).get('name', '')
        } for txn in txns]
        print(json.dumps(normalized))
        
    elif method == "get_transaction_categories":
        result = await mm.get_transaction_categories()
        categories = result.get('categories', [])
        normalized = [{
            'id': cat['id'],
            'name': cat['name'],
            'icon': cat.get('icon', '')
        } for cat in categories]
        print(json.dumps(normalized))
        
    elif method == "get_transaction_tags":
        result = await mm.get_transaction_tags()
        tags = result.get('tags', [])
        normalized = [{
            'id': tag['id'],
            'name': tag['name'],
            'color': tag['color']
        } for tag in tags]
        print(json.dumps(normalized))
        
    elif method == "get_budgets":
        from datetime import datetime
        now = datetime.now()
        start = now.replace(day=1).strftime('%Y-%m-%d')
        end = now.strftime('%Y-%m-%d')
        
        result = await mm.get_budgets(start_date=start, end_date=end)
        budgets = result.get('budgets', [])
        normalized = [{
            'id': b['id'],
            'amount': b['amount'],
            'spent': b.get('spent', 0),
            'rollover': b.get('rollover', False)
        } for b in budgets]
        print(json.dumps(normalized))
        
    elif method == "get_institutions":
        result = await mm.get_institutions()
        institutions = result.get('institutions', [])
        normalized = [{
            'id': inst['id'],
            'name': inst['name'],
            'status': inst.get('status', '')
        } for inst in institutions]
        print(json.dumps(normalized))

if __name__ == "__main__":
    asyncio.run(main())
`

	return os.WriteFile(v.pythonScript, []byte(script), 0755)
}

func saveReport(report *ValidationReport, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func printSummary(report *ValidationReport) {
	fmt.Println("\n=== Validation Report ===")
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("Passed: %d\n", report.Passed)
	fmt.Printf("Failed: %d\n", report.Failed)
	fmt.Printf("Success Rate: %.1f%%\n", report.SuccessRate)

	if report.Failed > 0 {
		fmt.Println("\nFailed Tests:")
		for _, result := range report.Results {
			if !result.Passed {
				fmt.Printf("  - %s: %s\n", result.Method, result.Error)
			}
		}
	}

	fmt.Printf("\nReport saved to: validation_results/\n")
}
