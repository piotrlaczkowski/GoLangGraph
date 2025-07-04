package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/piotrlaczkowski/GoLangGraph/examples"
)

func main() {
	fmt.Println("🚀 GoLangGraph Examples")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Choose an example to run:")
	fmt.Println("1. Quick Start Demo - Minimal code examples")
	fmt.Println("2. Simple Agent Demo - Basic agent with tools")
	fmt.Println("3. Database Persistence Demo - Database integration")
	fmt.Println("4. Ultimate Minimal Demo - One-line agent creation")
	fmt.Println("5. Run All Examples")
	fmt.Println()
	fmt.Print("Enter your choice (1-5): ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		fmt.Println("\n🚀 Running Quick Start Demo...")
		examples.QuickStartDemo()
	case "2":
		fmt.Println("\n🚀 Running Simple Agent Demo...")
		examples.SimpleAgentDemo()
	case "3":
		fmt.Println("\n🚀 Running Database Persistence Demo...")
		examples.RunDatabasePersistenceDemo()
	case "4":
		fmt.Println("\n🚀 Running Ultimate Minimal Demo...")
		examples.RunUltimateMinimalDemo()
	case "5":
		fmt.Println("\n🚀 Running All Examples...")
		runAllExamples()
	default:
		fmt.Println("Invalid choice. Please run again and select 1-5.")
		os.Exit(1)
	}
}

func runAllExamples() {
	separator := strings.Repeat("=", 60)

	fmt.Println(separator)
	fmt.Println("1. Quick Start Demo")
	fmt.Println(separator)
	examples.QuickStartDemo()

	fmt.Println("\n" + separator)
	fmt.Println("2. Simple Agent Demo")
	fmt.Println(separator)
	examples.SimpleAgentDemo()

	fmt.Println("\n" + separator)
	fmt.Println("3. Database Persistence Demo")
	fmt.Println(separator)
	examples.RunDatabasePersistenceDemo()

	fmt.Println("\n" + separator)
	fmt.Println("4. Ultimate Minimal Demo")
	fmt.Println(separator)
	examples.RunUltimateMinimalDemo()

	fmt.Println("\n🎉 All examples completed!")
}
