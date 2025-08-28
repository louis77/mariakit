# MariaDB Types Package

This package provides common database types used with MariaDB/MySQL databases.

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
```

## Database Integration

All types implement the necessary interfaces for seamless integration with database/sql:

- `driver.Valuer` for converting Go types to database values
- `sql.Scanner` for converting database values to Go types

The types handle both binary (WKB) and text formats for geometry data, and JSON serialization for complex types.