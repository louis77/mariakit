# MariaDB Types Package

This package provides common database types used with MariaDB databases.

## Types

### JSON[T]

A generic JSON type that implements the `sql.Scanner` and `driver.Valuer` interfaces for storing JSON data in database columns.

```go
type JSON[T any] struct {
    Data  T
    Valid bool
}
```

### StringArray

A type for storing arrays of strings as JSON in database columns.

```go
type StringArray []string
```

### Point

A geometric point type for storing latitude/longitude coordinates.

```go
type Point struct {
    X float64 `json:"x"` // Longitude
    Y float64 `json:"y"` // Latitude
}
```

### LineString

A geometric line string type for storing sequences of points.

```go
type LineString struct {
    Points []Point
}
```

### Vector[T]

A generic vector type for storing embeddings and multi-dimensional numerical arrays, corresponding to MariaDB's VECTOR datatype.

```go
type Vector[T VectorElement] struct {
    Data      []T
    Dimension int
    Valid     bool
}
```

Supported element types:
- `float32` (for MariaDB VECTOR with FLOAT elements)
- `float64` (for MariaDB VECTOR with DOUBLE elements)  
- `int32` (for MariaDB VECTOR with INT elements)
- `int64` (for MariaDB VECTOR with BIGINT elements)

## Usage

```go
import "github.com/louis77/mariakit/types"

// JSON type
type User struct {
    ID       int
    Metadata types.JSON[map[string]interface{}] `db:"metadata"`
}

// String array
type Product struct {
    ID     int
    Tags   types.StringArray `db:"tags"`
}

// Geometry types
type Location struct {
    ID     int
    Point  types.Point      `db:"location"`
    Route  types.LineString `db:"route"`
}

// Vector types for embeddings
type Document struct {
    ID        int
    Content   string
    Embedding types.Vector[float32] `db:"embedding"` // For VECTOR(128,FLOAT)
}

type AIModel struct {
    ID         int
    Name       string
    Weights    types.Vector[float64] `db:"weights"`    // For VECTOR(256,DOUBLE)
    Parameters types.Vector[int32]   `db:"parameters"` // For VECTOR(64,INT)
}
```

## Database Integration

All types implement the necessary interfaces for seamless integration with database/sql:

- `driver.Valuer` for converting Go types to database values
- `sql.Scanner` for converting database values to Go types

The types handle both binary (WKB) and text formats for geometry data, and JSON serialization for complex types.