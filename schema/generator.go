package schema

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// SchemaGenerator generates Go code from MariaDB schema
type SchemaGenerator struct {
	db     *sql.DB
	config *Config
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(connectionString string) (*SchemaGenerator, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("cannot create connector: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	return &SchemaGenerator{db: db}, nil
}

// NewSchemaGeneratorWithConfig creates a new schema generator with custom configuration
func NewSchemaGeneratorWithConfig(connectionString string, config *Config) (*SchemaGenerator, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("cannot create connector: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	return &SchemaGenerator{db: db, config: config}, nil
}

// Close closes the database connection
func (sg *SchemaGenerator) Close() error {
	if sg.db != nil {
		return sg.db.Close()
	}
	return nil
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name        string
	Columns     []ColumnInfo
	PrimaryKeys []string
}

// ColumnInfo represents information about a database column
type ColumnInfo struct {
	Name                 string
	Type                 string
	Nullable             bool
	DefaultValue         sql.NullString
	Comment              sql.NullString
	IsEnum               bool
	EnumValues           []string
	IsJSON               bool
	IsGenerated          bool
	GenerationType       sql.NullString // VIRTUAL or STORED
	GenerationExpression sql.NullString
}

// EnumInfo represents information about an enum type
type EnumInfo struct {
	TableName  string
	ColumnName string
	Values     []string
}

// GetTables retrieves all table names from the database
func (sg *SchemaGenerator) GetTables(ctx context.Context) ([]string, error) {
	query := `
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := sg.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// GetTableInfo retrieves detailed information about a table
func (sg *SchemaGenerator) GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error) {
	// Get column information
	columnsQuery := `
		SELECT
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_COMMENT,
			COALESCE(IS_GENERATED, 'NO') as IS_GENERATED,
			GENERATION_EXPRESSION,
			EXTRA
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := sg.db.QueryContext(ctx, columnsQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable, isGenerated, extra string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.DefaultValue, &col.Comment, &isGenerated, &col.GenerationExpression, &extra); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		col.Nullable = nullable == "YES"
		col.IsGenerated = isGenerated == "YES"
		
		// Extract generation type from EXTRA field
		if col.IsGenerated {
			if strings.Contains(strings.ToLower(extra), "virtual") {
				col.GenerationType.String = "VIRTUAL"
				col.GenerationType.Valid = true
			} else if strings.Contains(strings.ToLower(extra), "stored") {
				col.GenerationType.String = "STORED"
				col.GenerationType.Valid = true
			}
		}

		// Check if this is an enum column
		if strings.HasPrefix(col.Type, "enum(") {
			col.IsEnum = true
			col.EnumValues = sg.parseEnumValues(col.Type)
		}

		// Check if this is a JSON column (LONGTEXT with json_valid() constraint)
		if strings.ToLower(col.Type) == "longtext" {
			isJSON, err := sg.checkJSONConstraint(ctx, tableName, col.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to check JSON constraint for column %s: %w", col.Name, err)
			}
			col.IsJSON = isJSON
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	// Get primary keys
	pkQuery := `
		SELECT COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION
	`

	pkRows, err := sg.db.QueryContext(ctx, pkQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary keys for table %s: %w", tableName, err)
	}
	defer pkRows.Close()

	var primaryKeys []string
	for pkRows.Next() {
		var pk string
		if err := pkRows.Scan(&pk); err != nil {
			return nil, fmt.Errorf("failed to scan primary key: %w", err)
		}
		primaryKeys = append(primaryKeys, pk)
	}

	return &TableInfo{
		Name:        tableName,
		Columns:     columns,
		PrimaryKeys: primaryKeys,
	}, nil
}

// GetAllEnums retrieves all enum columns from all tables
func (sg *SchemaGenerator) GetAllEnums(ctx context.Context) ([]EnumInfo, error) {
	query := `
		SELECT
			TABLE_NAME,
			COLUMN_NAME,
			COLUMN_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND COLUMN_TYPE LIKE 'enum%'
		ORDER BY TABLE_NAME, COLUMN_NAME
	`

	rows, err := sg.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enums: %w", err)
	}
	defer rows.Close()

	var enums []EnumInfo
	for rows.Next() {
		var enum EnumInfo
		var columnType string
		if err := rows.Scan(&enum.TableName, &enum.ColumnName, &columnType); err != nil {
			return nil, fmt.Errorf("failed to scan enum info: %w", err)
		}
		enum.Values = sg.parseEnumValues(columnType)
		enums = append(enums, enum)
	}

	return enums, rows.Err()
}

// parseEnumValues extracts enum values from MariaDB enum type string
func (sg *SchemaGenerator) parseEnumValues(enumType string) []string {
	// enumType looks like: enum('value1','value2','value3')
	if !strings.HasPrefix(enumType, "enum(") || !strings.HasSuffix(enumType, ")") {
		return nil
	}

	// Extract the values part
	valuesStr := enumType[5 : len(enumType)-1] // Remove "enum(" and ")"

	// Split by comma and clean up quotes
	parts := strings.Split(valuesStr, ",")
	values := make([]string, len(parts))

	for i, part := range parts {
		// Remove surrounding quotes
		values[i] = strings.Trim(part, "'")
	}

	return values
}

// checkJSONConstraint checks if a LONGTEXT column has a json_valid() CHECK constraint
func (sg *SchemaGenerator) checkJSONConstraint(ctx context.Context, tableName, columnName string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM information_schema.CHECK_CONSTRAINTS cc
		JOIN information_schema.TABLE_CONSTRAINTS tc 
			ON cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME 
			AND cc.CONSTRAINT_SCHEMA = tc.TABLE_SCHEMA
		WHERE tc.TABLE_SCHEMA = DATABASE()
		AND tc.TABLE_NAME = ?
		AND tc.CONSTRAINT_TYPE = 'CHECK'
		AND cc.CHECK_CLAUSE LIKE CONCAT('%json_valid(%', ?, '%)%')
	`

	var count int
	err := sg.db.QueryRowContext(ctx, query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query JSON constraints: %w", err)
	}

	return count > 0, nil
}

// GenerateColumnConstants generates Go constants for all column names
func (sg *SchemaGenerator) GenerateColumnConstants(ctx context.Context, packageName string) (string, error) {
	tables, err := sg.GetTables(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tables: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("// Code generated by MariaDB Schema Generator. DO NOT EDIT.\n")
	builder.WriteString("// Generated on: " + time.Now().Format(time.RFC3339) + "\n\n")
	builder.WriteString("package " + packageName + "\n\n")

	for _, tableName := range tables {
		tableInfo, err := sg.GetTableInfo(ctx, tableName)
		if err != nil {
			return "", fmt.Errorf("failed to get table info for %s: %w", tableName, err)
		}

		// Generate constants for this table
		builder.WriteString(fmt.Sprintf("// %s table column constants\n", sg.toCamelCase(tableName)))
		builder.WriteString("const (\n")

		for _, col := range tableInfo.Columns {
			constName := sg.toConstantName(tableName, col.Name)
			builder.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constName, col.Name))
		}

		builder.WriteString(")\n\n")
	}

	return builder.String(), nil
}

// GenerateStructs generates Go structs for all tables
func (sg *SchemaGenerator) GenerateStructs(ctx context.Context, packageName string) (string, error) {
	tables, err := sg.GetTables(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tables: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("// Code generated by MariaDB Schema Generator. DO NOT EDIT.\n")
	builder.WriteString("// Generated on: " + time.Now().Format(time.RFC3339) + "\n\n")
	builder.WriteString("package " + packageName + "\n\n")
	builder.WriteString("import (\n")
	builder.WriteString("\t\"database/sql\"\n")
	builder.WriteString("\t\"time\"\n")

	// Add custom imports from config
	customImports := sg.getCustomImports()
	if len(customImports) > 0 {
		builder.WriteString("\n")
		for _, imp := range customImports {
			builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
	}

	builder.WriteString("\n")
	builder.WriteString("\t\"github.com/louis77/mariakit/types\"\n")
	builder.WriteString(")\n\n")

	for _, tableName := range tables {
		tableInfo, err := sg.GetTableInfo(ctx, tableName)
		if err != nil {
			return "", fmt.Errorf("failed to get table info for %s: %w", tableName, err)
		}

		// Generate struct for this table
		structName := sg.toStructName(tableName)
		builder.WriteString(fmt.Sprintf("// %s represents the %s table\n", structName, tableName))
		builder.WriteString(fmt.Sprintf("type %s struct {\n", structName))

		for _, col := range tableInfo.Columns {
			fieldName := sg.toFieldName(col.Name)
			goType := sg.mysqlTypeToGoType(col.Type, col.Nullable, col.IsJSON, tableName, col.Name)

			// Add db tag with comments
			tag := fmt.Sprintf("`db:\"%s\"`", col.Name)
			var comments []string
			
			if col.Comment.Valid && col.Comment.String != "" {
				comments = append(comments, col.Comment.String)
			}
			
			if col.IsGenerated {
				genType := "VIRTUAL"
				if col.GenerationType.Valid && col.GenerationType.String != "" {
					genType = col.GenerationType.String
				}
				genComment := fmt.Sprintf("Generated (%s): %s", genType, col.GenerationExpression.String)
				comments = append(comments, genComment)
			}
			
			if len(comments) > 0 {
				tag = fmt.Sprintf("`db:\"%s\"` // %s", col.Name, strings.Join(comments, "; "))
			}

			builder.WriteString(fmt.Sprintf("\t%s %s %s\n", fieldName, goType, tag))
		}

		builder.WriteString("}\n\n")
	}

	return builder.String(), nil
}

// GenerateEnumConstants generates Go constants for all enum values
func (sg *SchemaGenerator) GenerateEnumConstants(ctx context.Context, packageName string) (string, error) {
	enums, err := sg.GetAllEnums(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get enums: %w", err)
	}

	if len(enums) == 0 {
		return "// No enum types found in the database\n", nil
	}

	var builder strings.Builder
	builder.WriteString("// Code generated by MariaDB Schema Generator. DO NOT EDIT.\n")
	builder.WriteString("// Generated on: " + time.Now().Format(time.RFC3339) + "\n\n")
	builder.WriteString("package " + packageName + "\n\n")

	// Group enums by table for better organization
	tableEnums := make(map[string][]EnumInfo)
	for _, enum := range enums {
		tableEnums[enum.TableName] = append(tableEnums[enum.TableName], enum)
	}

	// Sort table names for consistent output
	var tableNames []string
	for tableName := range tableEnums {
		tableNames = append(tableNames, tableName)
	}
	sort.Strings(tableNames)

	for _, tableName := range tableNames {
		enums := tableEnums[tableName]
		builder.WriteString(fmt.Sprintf("// %s table enum constants\n", sg.toCamelCase(tableName)))

		for _, enum := range enums {
			builder.WriteString("const (\n")

			for _, value := range enum.Values {
				constName := sg.toEnumConstantName(tableName, enum.ColumnName, value)
				builder.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constName, value))
			}

			builder.WriteString(")\n\n")
		}
	}

	return builder.String(), nil
}

// GenerateAll generates all types of code (constants, structs, and enums)
func (sg *SchemaGenerator) GenerateAll(ctx context.Context, packageName string) (map[string]string, error) {
	columnConstants, err := sg.GenerateColumnConstants(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate column constants: %w", err)
	}

	structs, err := sg.GenerateStructs(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate structs: %w", err)
	}

	enumConstants, err := sg.GenerateEnumConstants(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate enum constants: %w", err)
	}

	return map[string]string{
		"column_constants.go": columnConstants,
		"structs.go":          structs,
		"enum_constants.go":   enumConstants,
	}, nil
}

// Helper functions for name conversion

func (sg *SchemaGenerator) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func (sg *SchemaGenerator) toConstantName(tableName, columnName string) string {
	table := sg.toCamelCase(tableName)
	column := sg.toCamelCase(columnName)
	return fmt.Sprintf("%s_%s", table, column)
}

func (sg *SchemaGenerator) toStructName(tableName string) string {
	return sg.toCamelCase(tableName)
}

func (sg *SchemaGenerator) toFieldName(columnName string) string {
	return sg.toCamelCase(columnName)
}

func (sg *SchemaGenerator) toEnumConstantName(tableName, columnName, value string) string {
	table := sg.toCamelCase(tableName)
	column := sg.toCamelCase(columnName)
	val := sg.toCamelCase(value)
	return fmt.Sprintf("%s_%s_%s", table, column, val)
}

func (sg *SchemaGenerator) mysqlTypeToGoType(mysqlType string, nullable bool, isJSON bool, tableName, columnName string) string {
	// Handle JSON types (detected LONGTEXT with json_valid() constraint)
	if isJSON {
		// Check for custom JSON mapping
		if sg.config != nil {
			if mapping, exists := sg.config.GetJSONMapping(tableName, columnName); exists {
				return mapping.Type
			}
		}
		return "types.JSON[any]"
	}

	// Handle enum types
	if strings.HasPrefix(mysqlType, "enum(") {
		if nullable {
			return "sql.NullString"
		}
		return "string"
	}

	// Extract base type (remove size specifications)
	baseType := mysqlType
	if idx := strings.Index(baseType, "("); idx > 0 {
		baseType = baseType[:idx]
	}

	var goType string
	switch strings.ToLower(baseType) {
	case "tinyint", "smallint", "mediumint", "int", "integer":
		if nullable {
			goType = "sql.NullInt32"
		} else {
			goType = "int32"
		}
	case "bigint":
		if nullable {
			goType = "sql.NullInt64"
		} else {
			goType = "int64"
		}
	case "float", "real":
		if nullable {
			goType = "sql.NullFloat64"
		} else {
			goType = "float32"
		}
	case "double", "decimal", "numeric":
		if nullable {
			goType = "sql.NullFloat64"
		} else {
			goType = "float64"
		}
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext":
		if nullable {
			goType = "sql.NullString"
		} else {
			goType = "string"
		}
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		goType = "[]byte"
	case "date", "datetime", "timestamp":
		if nullable {
			goType = "sql.NullTime"
		} else {
			goType = "time.Time"
		}
	case "time":
		if nullable {
			goType = "sql.NullString"
		} else {
			goType = "string"
		}
	case "year":
		if nullable {
			goType = "sql.NullInt32"
		} else {
			goType = "int32"
		}
	case "bit", "bool", "boolean":
		if nullable {
			goType = "sql.NullBool"
		} else {
			goType = "bool"
		}
	case "json":
		goType = "[]byte" // Simplified for standalone package
	case "point":
		goType = "[]byte" // Simplified for standalone package
	case "geometry":
		goType = "[]byte"
	default:
		// Unknown type, default to string
		if nullable {
			goType = "sql.NullString"
		} else {
			goType = "string"
		}
	}

	return goType
}

// getCustomImports returns all unique import paths needed for custom JSON mappings
func (sg *SchemaGenerator) getCustomImports() []string {
	if sg.config == nil {
		return nil
	}
	return sg.config.GetRequiredImports()
}
