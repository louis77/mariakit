package types

import (
	"testing"
)

func TestVector_Float32(t *testing.T) {
	// Test creating and converting a Vector[float32]
	data := []float32{1.0, 2.5, 3.14, -4.2}
	v := NewVector(data)

	if !v.Valid {
		t.Error("Vector should be valid")
	}

	if v.Dimension != 4 {
		t.Errorf("Expected dimension 4, got %d", v.Dimension)
	}

	if len(v.Data) != 4 {
		t.Errorf("Expected data length 4, got %d", len(v.Data))
	}

	// Test Value() method
	value, err := v.Value()
	if err != nil {
		t.Errorf("Value() error: %v", err)
	}

	if value == nil {
		t.Error("Value() should not return nil for valid vector")
	}

	// Test Scan() method with binary data
	var v2 Vector[float32]
	err = v2.Scan(value)
	if err != nil {
		t.Errorf("Scan() error: %v", err)
	}

	if !v2.Valid {
		t.Error("Scanned vector should be valid")
	}

	if v2.Dimension != v.Dimension {
		t.Errorf("Expected dimension %d, got %d", v.Dimension, v2.Dimension)
	}

	// Compare data
	for i := range v.Data {
		if v.Data[i] != v2.Data[i] {
			t.Errorf("Data mismatch at index %d: expected %f, got %f", i, v.Data[i], v2.Data[i])
		}
	}
}

func TestVector_Float64(t *testing.T) {
	// Test creating and converting a Vector[float64]
	data := []float64{1.0, 2.5, 3.141592653589793, -4.2}
	v := NewVector(data)

	// Test Value() and Scan() roundtrip
	value, err := v.Value()
	if err != nil {
		t.Errorf("Value() error: %v", err)
	}

	var v2 Vector[float64]
	err = v2.Scan(value)
	if err != nil {
		t.Errorf("Scan() error: %v", err)
	}

	// Compare data
	for i := range v.Data {
		if v.Data[i] != v2.Data[i] {
			t.Errorf("Data mismatch at index %d: expected %f, got %f", i, v.Data[i], v2.Data[i])
		}
	}
}

func TestVector_Int32(t *testing.T) {
	// Test creating and converting a Vector[int32]
	data := []int32{1, -2, 3, 4}
	v := NewVector(data)

	// Test Value() and Scan() roundtrip
	value, err := v.Value()
	if err != nil {
		t.Errorf("Value() error: %v", err)
	}

	var v2 Vector[int32]
	err = v2.Scan(value)
	if err != nil {
		t.Errorf("Scan() error: %v", err)
	}

	// Compare data
	for i := range v.Data {
		if v.Data[i] != v2.Data[i] {
			t.Errorf("Data mismatch at index %d: expected %d, got %d", i, v.Data[i], v2.Data[i])
		}
	}
}

func TestVector_StringScan(t *testing.T) {
	// Test parsing from string representation
	var v Vector[float64]
	err := v.Scan("[1.0, 2.5, 3.14, -4.2]")
	if err != nil {
		t.Errorf("Scan string error: %v", err)
	}

	if !v.Valid {
		t.Error("Vector should be valid after string scan")
	}

	if v.Dimension != 4 {
		t.Errorf("Expected dimension 4, got %d", v.Dimension)
	}

	expected := []float64{1.0, 2.5, 3.14, -4.2}
	for i, exp := range expected {
		if v.Data[i] != exp {
			t.Errorf("Data mismatch at index %d: expected %f, got %f", i, exp, v.Data[i])
		}
	}
}

func TestVector_EmptyVector(t *testing.T) {
	// Test empty vector
	var v Vector[float32]
	err := v.Scan("[]")
	if err != nil {
		t.Errorf("Scan empty vector error: %v", err)
	}

	if !v.Valid {
		t.Error("Empty vector should be valid")
	}

	if v.Dimension != 0 {
		t.Errorf("Expected dimension 0, got %d", v.Dimension)
	}

	if len(v.Data) != 0 {
		t.Errorf("Expected empty data, got length %d", len(v.Data))
	}
}

func TestVector_NullValue(t *testing.T) {
	// Test null value
	var v Vector[float64]
	err := v.Scan(nil)
	if err != nil {
		t.Errorf("Scan null error: %v", err)
	}

	if v.Valid {
		t.Error("Vector should not be valid after scanning null")
	}
}

func TestVector_String(t *testing.T) {
	// Test String() method
	data := []float32{1.0, 2.5, 3.14}
	v := NewVector(data)
	
	str := v.String()
	expected := "[1, 2.5, 3.14]"
	if str != expected {
		t.Errorf("Expected string %s, got %s", expected, str)
	}

	// Test invalid vector
	var v2 Vector[float32]
	str2 := v2.String()
	if str2 != "NULL" {
		t.Errorf("Expected 'NULL' for invalid vector, got %s", str2)
	}
}

func TestVector_Len(t *testing.T) {
	data := []int64{1, 2, 3, 4, 5}
	v := NewVector(data)
	
	if v.Len() != 5 {
		t.Errorf("Expected length 5, got %d", v.Len())
	}
}
