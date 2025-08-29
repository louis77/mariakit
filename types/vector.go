package types

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Vector represents a MariaDB VECTOR datatype for storing embeddings
// It supports vectors with float32, float64, int32, and int64 element types
type Vector[T VectorElement] struct {
	Data      []T
	Dimension int
	Valid     bool
}

// VectorElement defines the supported element types for vectors
type VectorElement interface {
	~float32 | ~float64 | ~int32 | ~int64
}

// NewVector creates a new Vector with the given data
func NewVector[T VectorElement](data []T) Vector[T] {
	return Vector[T]{
		Data:      data,
		Dimension: len(data),
		Valid:     true,
	}
}

// Value implements the driver.Valuer interface
func (v Vector[T]) Value() (driver.Value, error) {
	if !v.Valid || len(v.Data) == 0 {
		return nil, nil
	}

	// MariaDB expects vectors in a specific binary format
	// We'll store as a binary representation with type information
	var elementSize int
	var elementType byte

	// Determine element type and size
	switch any(v.Data[0]).(type) {
	case float32:
		elementSize = 4
		elementType = 1 // FLOAT
	case float64:
		elementSize = 8
		elementType = 2 // DOUBLE
	case int32:
		elementSize = 4
		elementType = 3 // INT
	case int64:
		elementSize = 8
		elementType = 4 // BIGINT
	default:
		return nil, fmt.Errorf("unsupported vector element type")
	}

	// Create binary data: [type:1][dimension:4][data:dimension*elementSize]
	data := make([]byte, 1+4+len(v.Data)*elementSize)
	
	// Write element type
	data[0] = elementType
	
	// Write dimension
	binary.LittleEndian.PutUint32(data[1:5], uint32(len(v.Data)))
	
	// Write vector elements
	offset := 5
	for _, elem := range v.Data {
		switch elementType {
		case 1: // float32
			binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(float32(any(elem).(float32))))
		case 2: // float64
			binary.LittleEndian.PutUint64(data[offset:offset+8], math.Float64bits(float64(any(elem).(float64))))
		case 3: // int32
			binary.LittleEndian.PutUint32(data[offset:offset+4], uint32(any(elem).(int32)))
		case 4: // int64
			binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(any(elem).(int64)))
		}
		offset += elementSize
	}

	return data, nil
}

// Scan implements the sql.Scanner interface
func (v *Vector[T]) Scan(value interface{}) error {
	if value == nil {
		v.Valid = false
		return nil
	}

	var data []byte
	switch val := value.(type) {
	case string:
		// Handle text representation like "[1.0, 2.0, 3.0]"
		return v.scanFromString(val)
	case []byte:
		data = val
	default:
		return fmt.Errorf("unsupported type for Vector: %T", value)
	}

	// Parse binary data
	if len(data) < 5 {
		return fmt.Errorf("vector data too short: %d bytes", len(data))
	}

	// Read element type
	elementType := data[0]
	
	// Read dimension
	dimension := int(binary.LittleEndian.Uint32(data[1:5]))
	
	// Determine element size
	var elementSize int
	switch elementType {
	case 1, 3: // float32, int32
		elementSize = 4
	case 2, 4: // float64, int64
		elementSize = 8
	default:
		return fmt.Errorf("unknown vector element type: %d", elementType)
	}

	// Check data length
	expectedLen := 5 + dimension*elementSize
	if len(data) < expectedLen {
		return fmt.Errorf("vector data too short for dimension %d: got %d bytes, expected %d", 
			dimension, len(data), expectedLen)
	}

	// Parse elements
	elements := make([]T, dimension)
	offset := 5
	
	for i := 0; i < dimension; i++ {
		var elem interface{}
		
		switch elementType {
		case 1: // float32
			bits := binary.LittleEndian.Uint32(data[offset : offset+4])
			elem = math.Float32frombits(bits)
		case 2: // float64
			bits := binary.LittleEndian.Uint64(data[offset : offset+8])
			elem = math.Float64frombits(bits)
		case 3: // int32
			elem = int32(binary.LittleEndian.Uint32(data[offset : offset+4]))
		case 4: // int64
			elem = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		}
		
		elements[i] = T(elem.(T))
		offset += elementSize
	}

	v.Data = elements
	v.Dimension = dimension
	v.Valid = true

	return nil
}

// scanFromString parses vector from string representation like "[1.0, 2.0, 3.0]"
func (v *Vector[T]) scanFromString(s string) error {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return fmt.Errorf("invalid vector string format: %s", s)
	}

	// Remove brackets
	s = s[1 : len(s)-1]
	s = strings.TrimSpace(s)
	
	if s == "" {
		v.Data = []T{}
		v.Dimension = 0
		v.Valid = true
		return nil
	}

	// Split by comma
	parts := strings.Split(s, ",")
	elements := make([]T, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		
		// Parse based on target type
		var elem interface{}
		var err error
		
		switch any(elements[0]).(type) {
		case float32:
			var f float64
			f, err = strconv.ParseFloat(part, 32)
			elem = float32(f)
		case float64:
			elem, err = strconv.ParseFloat(part, 64)
		case int32:
			var i int64
			i, err = strconv.ParseInt(part, 10, 32)
			elem = int32(i)
		case int64:
			elem, err = strconv.ParseInt(part, 10, 64)
		default:
			return fmt.Errorf("unsupported vector element type")
		}
		
		if err != nil {
			return fmt.Errorf("failed to parse vector element '%s': %v", part, err)
		}
		
		elements[i] = T(elem.(T))
	}

	v.Data = elements
	v.Dimension = len(elements)
	v.Valid = true

	return nil
}

// String returns string representation of the vector
func (v Vector[T]) String() string {
	if !v.Valid {
		return "NULL"
	}
	
	if len(v.Data) == 0 {
		return "[]"
	}

	parts := make([]string, len(v.Data))
	for i, elem := range v.Data {
		parts[i] = fmt.Sprintf("%v", elem)
	}
	
	return "[" + strings.Join(parts, ", ") + "]"
}

// Len returns the dimension of the vector
func (v Vector[T]) Len() int {
	return v.Dimension
}

// IsValid returns true if the vector contains valid data
func (v Vector[T]) IsValid() bool {
	return v.Valid
}
