# MariaKit

A Go package and CLI tool for generating type-safe Go code from MariaDB database schemas. It automatically creates Go structs, column constants, and enum values by inspecting your database structure through MariaDB's information schema.

MariaKit handles complex MariaDB features including JSON columns with custom type mappings, spatial data types (Point, LineString), and enum values while generating properly formatted Go code.

## Features

- 🔍 Automatically inspects MariaDB schema using information_schema
- 📝 Generates type-safe Go constants for all column names with `_Name` suffix
- 🎯 Generates Go type aliases for every table column 
- 🏗️ Creates Go structs with proper type mappings from MariaDB to Go
- 🔢 Generates constants for all enum values found in the database
- ✅ Proper TINYINT(1) to boolean mapping (MariaDB's actual boolean type)
- 🎯 Supports selective generation (constants, structs, types, or enums only)
- 📁 Clean, organized output with automatic Go formatting
- 🛠️ Standalone CLI tool with no external dependencies
- 🎨 Custom JSON column type mapping with YAML configuration
- 🔧 Automatic detection of JSON columns via `json_valid()` constraints
- 📦 Includes specialized MariaDB types (JSON, Point, LineString, StringArray)
- 🏷️ Preserves database column comments in generated structs and type aliases
- 🔗 Handles primary key relationships and nullable types

MariaKit let's you write your SQL queries in a type-safe way, with automatic type mapping and enum value handling, without imposing any ORM or dependencies to your Go application.

## Installation

### As a CLI Tool

```bash
# Clone the repository
git clone https://github.com/louis77/mariakit.git
cd mariakit

# Build the CLI tool
go build -o mariakit ./cmd/mariakit

# Move to your PATH (optional)
sudo mv mariakit /usr/local/bin/
```

### As a Go Package

```bash
go get github.com/louis77/mariakit/schema
```
### Types Package

The `types` package provides specialized database types for MariaDB/MySQL integration, designed to handle complex data types that require custom marshaling/unmarshaling:

```bash
go get github.com/louis77/mariakit/types
```

#### JSON[T] Type

A generic type for handling JSON columns with compile-time type safety:

```go
import "github.com/louis77/mariakit/types"

// Define a struct for JSON data
type UserPreferences struct {
    Theme     string            `json:"theme"`
    Language  string            `json:"language"`
    Settings  map[string]string `json:"settings"`
}

// Use in your table struct
type Users struct {
    ID         int32
    Name       string
    Email      string
    Preferences types.JSON[UserPreferences] `db:"preferences"`
    CreatedAt  time.Time
}

// Usage
user := Users{
    Preferences: types.JSON[UserPreferences]{
        Data: UserPreferences{
            Theme: "dark",
            Language: "en",
        },
        Valid: true,
    },
}

// Database operations work seamlessly
err := db.Get(&user, "SELECT * FROM users WHERE id = ?", userID)
```

**Features:**
- ✅ Generic type parameter for compile-time type safety
- ✅ Automatic JSON marshaling/unmarshaling
- ✅ NULL handling with `Valid` field
- ✅ Compatible with `database/sql/driver` interface

#### StringArray Type

Handles MariaDB JSON arrays stored as string arrays:

```go
import "github.com/louis77/mariakit/types"

type Articles struct {
    ID       int32
    Title    string
    Content  string
    Tags     types.StringArray `db:"tags"`        // ["golang", "database", "json"]
    Categories types.StringArray `db:"categories"` // ["tech", "programming"]
}

// Usage
article := Articles{
    Tags: types.StringArray{"golang", "database", "json"},
    Categories: types.StringArray{"tech", "programming"},
}

// Database operations
err := db.Get(&article, "SELECT * FROM articles WHERE id = ?", articleID)
```

**Features:**
- ✅ Automatic JSON marshaling of string slices
- ✅ Compatible with MariaDB JSON columns
- ✅ Simple slice-like interface
- ✅ NULL-safe operations

#### Point Type

Handles MariaDB POINT geometry data for latitude/longitude coordinates:

```go
import "github.com/louis77/mariakit/types"

type Locations struct {
    ID       int32
    Name     string
    Address  string
    Position types.Point `db:"position"` // Stores lat/lng as POINT
}

// Usage
location := Locations{
    Name: "Central Park",
    Position: types.Point{
        X: -73.9654, // Longitude
        Y: 40.7829,  // Latitude
    },
}

// Database operations
err := db.Get(&location, "SELECT * FROM locations WHERE id = ?", locationID)
```

**Features:**
- ✅ Automatic WKB (Well-Known Binary) encoding/decoding
- ✅ Handles both text and binary POINT formats
- ✅ Proper longitude/latitude coordinate handling
- ✅ Compatible with MariaDB spatial functions

#### LineString Type

Handles MariaDB LINESTRING geometry for paths and routes:

```go
import "github.com/louis77/mariakit/types"

type Routes struct {
    ID       int32
    Name     string
    Path     types.LineString `db:"path"` // Series of connected points
}

// Usage
route := Routes{
    Name: "Morning Jog",
    Path: types.LineString{
        Points: []types.Point{
            {X: -73.9654, Y: 40.7829}, // Start
            {X: -73.9632, Y: 40.7845}, // Point 2
            {X: -73.9610, Y: 40.7861}, // End
        },
    },
}
```

**Features:**
- ✅ Handles complex geometric paths
- ✅ Automatic WKB encoding/decoding
- ✅ Efficient storage of connected coordinates
- ✅ Compatible with MariaDB spatial operations

### Advanced Usage Examples

#### Combining Multiple Types

```go
import "github.com/louis77/mariakit/types"

type ComplexEntity struct {
    ID          int32
    Name        string
    Location    types.Point                           `db:"location"`
    Tags        types.StringArray                     `db:"tags"`
    Metadata    types.JSON[map[string]interface{}]    `db:"metadata"`
    Route       types.LineString                      `db:"route"`
    CreatedAt   time.Time
}
```

#### Custom JSON Types

```go
// Define custom types for your JSON data
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
    ZipCode string `json:"zip_code"`
}

type UserProfile struct {
    Bio       string                 `json:"bio"`
    Address   Address                `json:"address"`
    Interests []string               `json:"interests"`
    Settings  map[string]interface{} `json:"settings"`
}

type Users struct {
    ID      int32
    Name    string
    Email   string
    Profile types.JSON[UserProfile] `db:"profile"`
}
```

See [types/README.md](types/README.md) for detailed documentation on available types including `JSON[T]`, `StringArray`, `Point`, and `LineString`.

## Usage

### CLI Tool

#### Basic Usage

```bash
mariakit -conn="user:password@tcp(localhost:3306)/database"
```

#### Generate All Code Types

```bash
mariakit \
  -conn="user:password@tcp(localhost:3306)/database" \
  -output="./generated"
```

#### Generate Specific Code Types

##### Column Constants Only
```bash
mariakit \
  -conn="user:password@tcp(localhost:3306)/database" \
  -type=constants \
  -output="./generated"
```

##### Structs Only
```bash
mariakit \
  -conn="user:password@tcp(localhost:3306)/database" \
  -type=structs \
  -output="./generated"
```

##### Column Type Aliases Only
```bash
mariakit \
  -conn="user:password@tcp(localhost:3306)/database" \
  -type=types \
  -output="./generated"
```

##### Enum Constants Only
```bash
mariakit \
  -conn="user:password@tcp(localhost:3306)/database" \
  -type=enums \
  -output="./generated"
```

### Go Package

```go
package main

import (
    "context"
    "log"

    "github.com/louis77/mariakit/schema"
)

func main() {
    // Create schema generator
    generator, err := schema.NewSchemaGenerator("user:password@tcp(localhost:3306)/database")
    if err != nil {
        log.Fatal(err)
    }
    defer generator.Close()

    ctx := context.Background()

    // Generate all code types
    files, err := generator.GenerateAll(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Access generated code
    columnConstants := files["column_constants.go"]
    structs := files["structs.go"]
    columnTypes := files["column_types.go"]
    enumConstants := files["enum_constants.go"]

    // Or generate specific types
    constants, err := generator.GenerateColumnConstants(ctx)
    if err != nil {
        log.Fatal(err)
    }

    _ = constants // Use the generated code
}
```

## JSON Column Support

MariaKit automatically detects JSON columns in your MariaDB database by looking for `LONGTEXT` columns with `json_valid()` CHECK constraints. By default, these columns are mapped to `types.JSON[any]`, but you can customize this behavior using a configuration file.

### Configuration File

Create a `mariakit.yaml` file in your project directory to specify custom type mappings for JSON columns:

```yaml
json_mappings:
  users.preferences:
    type: mytypes.UserPreferences
    import: github.com/mycompany/mytypes
  orders.metadata:
    type: OrderMetadata
    import: github.com/mycompany/models
  products.specifications:
    type: types.JSON[ProductSpecs]
    import: github.com/mycompany/product
  widgets.config:
    type: map[string]interface{}
```

#### Configuration Structure

Each JSON mapping consists of:
- **Key**: `table.column` format (e.g., `users.preferences`)
- **type**: The Go type to use for this column
- **import**: (Optional) The package import path needed for the custom type

#### Examples

**Using Custom Structs:**
```yaml
json_mappings:
  users.profile:
    type: UserProfile
    import: github.com/myapp/models
```

**Using Built-in Types:**
```yaml
json_mappings:
  cache.data:
    type: map[string]interface{}
```

**Using Generic JSON Types:**
```yaml
json_mappings:
  products.metadata:
    type: types.JSON[ProductMetadata]
    import: github.com/myapp/product
```

### JSON Column Detection

MariaKit detects JSON columns using the following criteria:

1. Column type is `LONGTEXT`
2. A CHECK constraint exists containing `json_valid(column_name)`

This is the standard pattern used in MariaDB for JSON validation:

```sql
ALTER TABLE users ADD COLUMN preferences LONGTEXT 
CHECK (json_valid(preferences));
```

### Generated Output

**Without Configuration:**
```go
type Users struct {
    ID          int32             `db:"id"`
    Name        string            `db:"name"`
    Preferences types.JSON[any]   `db:"preferences"`
}
```

**With Custom Configuration:**
```go
import (
    "database/sql"
    "time"
    
    "github.com/myapp/models"
    "github.com/louis77/mariakit/types"
)

type Users struct {
    ID          int32              `db:"id"`
    Name        string             `db:"name"`
    Preferences models.UserProfile `db:"preferences"`
}
```

## Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-conn` | MariaDB connection string (required) | "" |
| `-output` | Output directory for generated files | "./generated" |
| `-type` | Type of code to generate: `all`, `constants`, `structs`, `types`, `enums` | "all" |
| `-config` | Path to configuration file | "mariakit.yaml" |
| `-help` | Show help message | false |

## Connection String Format

The connection string should follow the MariaDB connection format (using MySQL driver):

```
user:password@tcp(host:port)/database?parseTime=true&charset=utf8mb4
```

Examples:
- `root:password@tcp(localhost:3306)/myapp`
- `user:pass@tcp(192.168.1.100:3306)/production?parseTime=true`

## Generated Files

MariaKit generates clean, organized Go code split into separate files for better maintainability. When generating all code types, the following files are created:

### `column_constants.go`
Contains constants for all column names with `_Name` suffix:
```go
// Users table column constants
const (
    Users_Id_Name = "id"
    Users_Name_Name = "name"
    Users_Email_Name = "email"
    Users_CreatedAt_Name = "created_at"
)
```

### `structs.go`
Contains Go structs for all tables:
```go
// Users represents the users table
type Users struct {
    ID        int32     `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
}
```

### `column_types.go`
Contains Go type aliases for every table column:
```go
// Users table column type aliases
type Users_Id = int32
type Users_Name = string
type Users_Email = string
type Users_CreatedAt = time.Time
```

### `enum_constants.go`
Contains constants for all enum values:
```go
// Users table enum constants
const (
    Users_Status_Active = "active"
    Users_Status_Inactive = "inactive"
    Users_Role_Admin = "admin"
    Users_Role_User = "user"
)
```

## Type Mappings

The generator maps MariaDB types to appropriate Go types:

| MariaDB Type | Go Type | Nullable Go Type |
|------------|---------|------------------|
| TINYINT, INT | int32 | sql.NullInt32 |
| BIGINT | int64 | sql.NullInt64 |
| FLOAT | float32 | sql.NullFloat64 |
| DOUBLE, DECIMAL | float64 | sql.NullFloat64 |
| VARCHAR, TEXT | string | sql.NullString |
| DATE, DATETIME, TIMESTAMP | time.Time | sql.NullTime |
| BOOLEAN, BIT, TINYINT(1) | bool | sql.NullBool |
| BLOB, BINARY | []byte | []byte |
| ENUM | string | sql.NullString |
| LONGTEXT with json_valid() | types.JSON[any] | types.JSON[any] |

## Examples

### Example 1: Basic Generation

```bash
# Generate all code for a local database
mariakit \
  -conn="root:password@tcp(localhost:3306)/myapp"
```

### Example 2: Production Database

```bash
# Generate structs only for production database
mariakit \
  -conn="app_user:secure_pass@tcp(db.example.com:3306)/production?parseTime=true" \
  -type=structs \
  -output="./internal/models"
```

### Example 3: CI/CD Integration

```bash
# Generate code in CI pipeline
mariakit \
  -conn="$DATABASE_URL" \
  -output="./generated"
```

## Error Handling

The generator will exit with an error if:

- Connection string is not provided
- Database connection fails
- Schema inspection fails
- Output directory cannot be created
- File writing fails

All errors are logged with descriptive messages to help with troubleshooting.

## Troubleshooting

### Common Issues

**Connection Failed:**
- Verify your MariaDB server is running and accessible
- Check that the connection string format is correct
- Ensure the database user has SELECT privileges on `information_schema`
- Test the connection string with a MariaDB client first

**No JSON Columns Detected:**
- Verify your JSON columns are `LONGTEXT` type with `json_valid()` CHECK constraints. 
  You can add CHECK constraints like this:
  ```sql
    alter table your_table
        add constraint colname
            check (json_valid(`colname`));
  ```
- Use `SHOW CREATE TABLE your_table` to verify constraint syntax
- JSON columns without `json_valid()` constraints will be treated as regular `LONGTEXT` (Go strings)

**Configuration Not Applied:**
- Check that `mariakit.yaml` exists in the current directory or specify path with `-config`
- Verify YAML syntax is correct (use a YAML validator)
- Ensure table.column names match exactly (case-sensitive)

**Import Errors in Generated Code:**
- Verify custom package import paths are accessible
- Check that custom types are exported (start with capital letter)
- Ensure go.mod includes all required dependencies

## Notes

- The generated code includes a header comment indicating it's auto-generated
- Generated code should not be manually edited as it will be overwritten
- The generator uses the `information_schema` to inspect the database schema
- Enum values are extracted from the MariaDB `COLUMN_TYPE` field
- Table and column names are converted to CamelCase for Go naming conventions
- All generated Go files are automatically formatted using `go/format`
- Configuration files are optional - the tool works with sensible defaults
- Supports both nullable and non-nullable columns with appropriate Go types

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
