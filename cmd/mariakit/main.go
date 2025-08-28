package main

import (
	"context"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/louis77/mariakit/schema"
)

func main() {
	var (
		connectionString = flag.String("conn", "", "MariaDB connection string (required)")
		outputDir        = flag.String("output", "./generated", "Output directory for generated files")
		generateType     = flag.String("type", "all", "Type of code to generate: all, constants, structs, enums")
		configPath       = flag.String("config", "mariakit.yaml", "Path to configuration file")
		help             = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *connectionString == "" {
		log.Fatal("Connection string is required. Use -conn flag.")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Extract package name from output directory
	packageName := filepath.Base(*outputDir)

	// Load configuration
	config, err := schema.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if config file exists and report
	if _, err := os.Stat(*configPath); err == nil {
		fmt.Printf("📄 Using configuration file: %s\n", *configPath)
	} else {
		fmt.Printf("📄 No configuration file found at %s, using defaults\n", *configPath)
	}

	// Create schema generator with config
	generator, err := schema.NewSchemaGeneratorWithConfig(*connectionString, config)
	if err != nil {
		log.Fatalf("Failed to create schema generator: %v", err)
	}
	defer generator.Close()

	ctx := context.Background()

	fmt.Println("🔍 Inspecting MariaDB schema...")

	// Generate code based on type
	switch strings.ToLower(*generateType) {
	case "all":
		fmt.Println("📝 Generating all code types...")
		files, err := generator.GenerateAll(ctx, packageName)
		if err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}

		for filename, content := range files {
			outputPath := filepath.Join(*outputDir, filename)
			if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
				log.Fatalf("Failed to write file %s: %v", outputPath, err)
			}
			fmt.Printf("✅ Generated %s\n", outputPath)
		}

	case "constants":
		fmt.Println("📝 Generating column constants...")
		content, err := generator.GenerateColumnConstants(ctx, packageName)
		if err != nil {
			log.Fatalf("Failed to generate column constants: %v", err)
		}

		outputPath := filepath.Join(*outputDir, "column_constants.go")
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write file %s: %v", outputPath, err)
		}
		fmt.Printf("✅ Generated %s\n", outputPath)

	case "structs":
		fmt.Println("📝 Generating table structs...")
		content, err := generator.GenerateStructs(ctx, packageName)
		if err != nil {
			log.Fatalf("Failed to generate structs: %v", err)
		}

		outputPath := filepath.Join(*outputDir, "structs.go")
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write file %s: %v", outputPath, err)
		}
		fmt.Printf("✅ Generated %s\n", outputPath)

	case "enums":
		fmt.Println("📝 Generating enum constants...")
		content, err := generator.GenerateEnumConstants(ctx, packageName)
		if err != nil {
			log.Fatalf("Failed to generate enum constants: %v", err)
		}

		outputPath := filepath.Join(*outputDir, "enum_constants.go")
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write file %s: %v", outputPath, err)
		}
		fmt.Printf("✅ Generated %s\n", outputPath)

	default:
		log.Fatalf("Invalid generate type: %s. Use 'all', 'constants', 'structs', or 'enums'", *generateType)
	}

	// Format generated Go files
	fmt.Println("🔧 Formatting generated Go files...")
	if err := formatGeneratedFiles(*outputDir); err != nil {
		log.Printf("Warning: Failed to format generated files: %v", err)
	}

	fmt.Println("🎉 Schema code generation completed successfully!")
}

// formatGeneratedFiles formats all .go files in the specified directory using go/format
func formatGeneratedFiles(outputDir string) error {
	// Find all .go files in the output directory
	goFiles, err := filepath.Glob(filepath.Join(outputDir, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(goFiles) == 0 {
		return nil // No Go files to format
	}

	// Format each file using go/format
	for _, file := range goFiles {
		if err := formatFile(file); err != nil {
			return fmt.Errorf("failed to format %s: %w", file, err)
		}
	}

	fmt.Printf("✅ Formatted %d Go files\n", len(goFiles))
	return nil
}

// formatFile formats a single Go file using go/format
func formatFile(filename string) error {
	// Read the file
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Format the source code
	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("failed to format source: %w", err)
	}

	// Write the formatted code back to the file
	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write formatted file: %w", err)
	}

	return nil
}

func showHelp() {
	fmt.Println("MariaDB Schema Code Generator")
	fmt.Println()
	fmt.Println("This tool generates Go code from MariaDB database schema including:")
	fmt.Println("  - Column name constants for all tables")
	fmt.Println("  - Go structs for all tables with proper types")
	fmt.Println("  - Enum value constants for all enum columns")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [flags]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate all code types")
	fmt.Printf("  %s -conn='user:password@tcp(localhost:3306)/database' -output='./generated'\n", os.Args[0])
	fmt.Println()
	fmt.Println("  # Generate only column constants")
	fmt.Printf("  %s -conn='user:password@tcp(localhost:3306)/database' -type=constants\n", os.Args[0])
	fmt.Println()
	fmt.Println("  # Generate only structs")
	fmt.Printf("  %s -conn='user:password@tcp(localhost:3306)/database' -type=structs\n", os.Args[0])
	fmt.Println()
	fmt.Println("  # Generate only enum constants")
	fmt.Printf("  %s -conn='user:password@tcp(localhost:3306)/database' -type=enums\n", os.Args[0])
}
