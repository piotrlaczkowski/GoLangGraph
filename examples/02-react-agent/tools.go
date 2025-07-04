// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - ReAct Agent Tools

package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// executeCalculator performs mathematical calculations
func (a *ReActAgent) executeCalculator(input string) string {
	input = strings.ToLower(input)

	// Handle square root
	if strings.Contains(input, "sqrt") {
		re := regexp.MustCompile(`sqrt\((\d+(?:\.\d+)?)\)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if num, err := strconv.ParseFloat(matches[1], 64); err == nil {
				result := math.Sqrt(num)
				return fmt.Sprintf("√%.2f = %.2f", num, result)
			}
		}

		// Try to extract number after "sqrt of" or similar
		re = regexp.MustCompile(`sqrt(?:\s+of)?\s+(\d+(?:\.\d+)?)`)
		matches = re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if num, err := strconv.ParseFloat(matches[1], 64); err == nil {
				result := math.Sqrt(num)
				return fmt.Sprintf("√%.2f = %.2f", num, result)
			}
		}
	}

	// Handle powers
	if strings.Contains(input, "^") || strings.Contains(input, "power") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\^\s*(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 2 {
			if base, err := strconv.ParseFloat(matches[1], 64); err == nil {
				if exp, err := strconv.ParseFloat(matches[2], 64); err == nil {
					result := math.Pow(base, exp)
					return fmt.Sprintf("%.2f^%.2f = %.2f", base, exp, result)
				}
			}
		}
	}

	// Handle basic arithmetic
	if strings.Contains(input, "+") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\+\s*(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 2 {
			if a, err := strconv.ParseFloat(matches[1], 64); err == nil {
				if b, err := strconv.ParseFloat(matches[2], 64); err == nil {
					result := a + b
					return fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result)
				}
			}
		}
	}

	if strings.Contains(input, "-") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*-\s*(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 2 {
			if a, err := strconv.ParseFloat(matches[1], 64); err == nil {
				if b, err := strconv.ParseFloat(matches[2], 64); err == nil {
					result := a - b
					return fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result)
				}
			}
		}
	}

	if strings.Contains(input, "*") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\*\s*(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 2 {
			if a, err := strconv.ParseFloat(matches[1], 64); err == nil {
				if b, err := strconv.ParseFloat(matches[2], 64); err == nil {
					result := a * b
					return fmt.Sprintf("%.2f × %.2f = %.2f", a, b, result)
				}
			}
		}
	}

	if strings.Contains(input, "/") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*/\s*(\d+(?:\.\d+)?)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 2 {
			if a, err := strconv.ParseFloat(matches[1], 64); err == nil {
				if b, err := strconv.ParseFloat(matches[2], 64); err == nil {
					if b != 0 {
						result := a / b
						return fmt.Sprintf("%.2f ÷ %.2f = %.2f", a, b, result)
					} else {
						return "Error: Division by zero"
					}
				}
			}
		}
	}

	return "I couldn't parse that mathematical expression. Try formats like 'sqrt(144)', '2^3', or '10 + 5'"
}

// executeUnitConverter converts between different units
func (a *ReActAgent) executeUnitConverter(input string) string {
	input = strings.ToLower(input)

	// Temperature conversions
	if strings.Contains(input, "fahrenheit") && strings.Contains(input, "celsius") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*fahrenheit\s*to\s*celsius`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
				c := (f - 32) * 5 / 9
				return fmt.Sprintf("%.2f°F = %.2f°C", f, c)
			}
		}
	}

	if strings.Contains(input, "celsius") && strings.Contains(input, "fahrenheit") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*celsius\s*to\s*fahrenheit`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if c, err := strconv.ParseFloat(matches[1], 64); err == nil {
				f := c*9/5 + 32
				return fmt.Sprintf("%.2f°C = %.2f°F", c, f)
			}
		}
	}

	// Length conversions
	if strings.Contains(input, "meters") && strings.Contains(input, "feet") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*meters?\s*to\s*feet`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if m, err := strconv.ParseFloat(matches[1], 64); err == nil {
				ft := m * 3.28084
				return fmt.Sprintf("%.2f meters = %.2f feet", m, ft)
			}
		}
	}

	if strings.Contains(input, "feet") && strings.Contains(input, "meters") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*feet\s*to\s*meters?`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if ft, err := strconv.ParseFloat(matches[1], 64); err == nil {
				m := ft / 3.28084
				return fmt.Sprintf("%.2f feet = %.2f meters", ft, m)
			}
		}
	}

	// Weight conversions
	if strings.Contains(input, "pounds") && strings.Contains(input, "kilograms") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*pounds?\s*to\s*kilograms?`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if lb, err := strconv.ParseFloat(matches[1], 64); err == nil {
				kg := lb * 0.453592
				return fmt.Sprintf("%.2f pounds = %.2f kilograms", lb, kg)
			}
		}
	}

	if strings.Contains(input, "kilograms") && strings.Contains(input, "pounds") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*kilograms?\s*to\s*pounds?`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			if kg, err := strconv.ParseFloat(matches[1], 64); err == nil {
				lb := kg / 0.453592
				return fmt.Sprintf("%.2f kilograms = %.2f pounds", kg, lb)
			}
		}
	}

	return "I couldn't parse that conversion. Try formats like '100 fahrenheit to celsius' or '10 meters to feet'"
}

// executeDataAnalyzer analyzes numerical data
func (a *ReActAgent) executeDataAnalyzer(input string) string {
	input = strings.ToLower(input)

	// Extract numbers from the input
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)`)
	matches := re.FindAllString(input, -1)

	if len(matches) == 0 {
		return "No numbers found in the input. Please provide numbers to analyze."
	}

	var numbers []float64
	for _, match := range matches {
		if num, err := strconv.ParseFloat(match, 64); err == nil {
			numbers = append(numbers, num)
		}
	}

	if len(numbers) == 0 {
		return "Could not parse any valid numbers from the input."
	}

	// Calculate statistics based on what was requested
	if strings.Contains(input, "mean") || strings.Contains(input, "average") {
		mean := calculateMean(numbers)
		return fmt.Sprintf("Mean of %v = %.2f", numbers, mean)
	}

	if strings.Contains(input, "median") {
		median := calculateMedian(numbers)
		return fmt.Sprintf("Median of %v = %.2f", numbers, median)
	}

	if strings.Contains(input, "mode") {
		mode := calculateMode(numbers)
		if len(mode) == 1 {
			return fmt.Sprintf("Mode of %v = %.2f", numbers, mode[0])
		} else if len(mode) > 1 {
			return fmt.Sprintf("Modes of %v = %v (multimodal)", numbers, mode)
		} else {
			return fmt.Sprintf("No mode found for %v (all values are unique)", numbers)
		}
	}

	if strings.Contains(input, "sum") || strings.Contains(input, "total") {
		sum := calculateSum(numbers)
		return fmt.Sprintf("Sum of %v = %.2f", numbers, sum)
	}

	if strings.Contains(input, "max") || strings.Contains(input, "maximum") {
		max := calculateMax(numbers)
		return fmt.Sprintf("Maximum of %v = %.2f", numbers, max)
	}

	if strings.Contains(input, "min") || strings.Contains(input, "minimum") {
		min := calculateMin(numbers)
		return fmt.Sprintf("Minimum of %v = %.2f", numbers, min)
	}

	// Default: provide comprehensive analysis
	mean := calculateMean(numbers)
	median := calculateMedian(numbers)
	min := calculateMin(numbers)
	max := calculateMax(numbers)
	sum := calculateSum(numbers)

	return fmt.Sprintf("Analysis of %v:\n• Count: %d\n• Sum: %.2f\n• Mean: %.2f\n• Median: %.2f\n• Min: %.2f\n• Max: %.2f",
		numbers, len(numbers), sum, mean, median, min, max)
}

// Helper functions for statistical calculations

func calculateMean(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum / float64(len(numbers))
}

func calculateMedian(numbers []float64) float64 {
	sorted := make([]float64, len(numbers))
	copy(sorted, numbers)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func calculateMode(numbers []float64) []float64 {
	frequency := make(map[float64]int)
	for _, num := range numbers {
		frequency[num]++
	}

	maxFreq := 0
	for _, freq := range frequency {
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	if maxFreq == 1 {
		return []float64{} // No mode (all values unique)
	}

	var modes []float64
	for num, freq := range frequency {
		if freq == maxFreq {
			modes = append(modes, num)
		}
	}

	sort.Float64s(modes)
	return modes
}

func calculateSum(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func calculateMax(numbers []float64) float64 {
	max := numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
	}
	return max
}

func calculateMin(numbers []float64) float64 {
	min := numbers[0]
	for _, num := range numbers {
		if num < min {
			min = num
		}
	}
	return min
}
