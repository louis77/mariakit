package types

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"math"
)

const (
	WKBTypePoint      = 1
	WKBTypeLineString = 2
	//WKBTypePolygon            = 3
	//WKBTypeMultiPoint         = 4
	//WKBTypeMultiLineString    = 5
	//WKBTypeMultiPolygon       = 6
	//WKBTypeGeometryCollection = 7
)

func decodePoint(byteOrder binary.ByteOrder, data []byte) Point {
	var p Point
	p.X = math.Float64frombits(byteOrder.Uint64(data[0:8]))
	p.Y = math.Float64frombits(byteOrder.Uint64(data[8:16]))
	return p
}

type Point struct {
	// X is Longitude
	X float64 `json:"x"`
	// Y is Latitude
	Y float64 `json:"y"`
}

func (p Point) Value() (driver.Value, error) {
	data := make([]byte, 25)
	// SRID
	data[0] = 0
	data[1] = 0
	data[2] = 0
	data[3] = 0

	// Byte order indicator (endianness)
	data[4] = 1 // Little endian

	byteOrder := binary.LittleEndian
	byteOrder.PutUint32(data[5:9], WKBTypePoint)

	// Points
	byteOrder.PutUint64(data[9:17], math.Float64bits(p.X))
	byteOrder.PutUint64(data[17:25], math.Float64bits(p.Y))

	return data, nil
}

func (p *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type for Point: %T", value)
	}

	// Check if it's a text representation
	if len(data) > 5 && (data[0] == 'P' || data[0] == 'p') {
		// Handle text format like "POINT(x y)"
		var str = string(data)
		var x, y float64
		_, err := fmt.Sscanf(str, "POINT(%f %f)", &x, &y)
		if err != nil {
			// Try a more lenient parsing
			var pointPrefix string
			_, err = fmt.Sscanf(str, "%s(%f %f)", &pointPrefix, &x, &y)
			if err != nil {
				return fmt.Errorf("failed to parse Point from '%s': %v", str, err)
			}
		}
		p.X = x
		p.Y = y
		return nil
	}

	// Handle WKB (Well-Known Binary) format
	if len(data) < 25 {
		return fmt.Errorf("WKB data too short: %d bytes", len(data))
	}

	// The first 4 bytes are SRID, which we don't care about
	// and most likely is 0 anyway

	// Check the byte order (endianness)
	var byteOrder binary.ByteOrder
	if data[4] == 0 {
		byteOrder = binary.BigEndian
	} else if data[4] == 1 {
		byteOrder = binary.LittleEndian
	} else {
		return fmt.Errorf("invalid byte order indicator: %d", data[0])
	}

	// Check geometry type (should be 1 for Point)
	geometryType := byteOrder.Uint32(data[5:9])
	if geometryType != WKBTypePoint {
		return fmt.Errorf("expected geometry type 1 (Point), got %d", geometryType)
	}

	// Extract X and Y coordinates (double precision floating point, 8 bytes each)
	*p = decodePoint(byteOrder, data[9:25])

	return nil
}

type LineString struct {
	Points []Point
}

func (p *LineString) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type for LineString: %T", value)
	}

	// Handle WKB (Well-Known Binary) format
	if len(data) < 29 {
		return fmt.Errorf("WKB data too short: %d bytes", len(data))
	}

	// The first 4 bytes are SRID, which we don't care about
	// and most likely is 0 anyway

	// Check the byte order (endianness)
	var byteOrder binary.ByteOrder
	if data[4] == 0 {
		byteOrder = binary.BigEndian
	} else if data[4] == 1 {
		byteOrder = binary.LittleEndian
	} else {
		return fmt.Errorf("invalid byte order indicator: %d", data[0])
	}

	// Check geometry type (should be 1 for Point)
	geometryType := byteOrder.Uint32(data[5:9])
	if geometryType != WKBTypeLineString {
		return fmt.Errorf("expected geometry type 2 (LineString), got %d", geometryType)
	}

	numPoints := byteOrder.Uint32(data[9:13])

	points := make([]Point, numPoints)
	for i := range numPoints {
		offset := 13 + (i * 16)
		points[i] = decodePoint(byteOrder, data[offset:offset+8])
	}

	p.Points = points
	return nil
}

func (p LineString) Value() (driver.Value, error) {
	data := make([]byte, 13+len(p.Points)*16)
	// SRID
	data[0] = 0
	data[1] = 0
	data[2] = 0
	data[3] = 0

	// Byte order indicator (endianness)
	data[4] = 1 // Little endian

	byteOrder := binary.LittleEndian
	byteOrder.PutUint32(data[5:9], WKBTypeLineString)

	// Number of points
	byteOrder.PutUint32(data[9:13], uint32(len(p.Points)))

	// Points
	for i, p := range p.Points {
		offset := 13 + (i * 16)
		byteOrder.PutUint64(data[offset:offset+8], math.Float64bits(p.X))
		byteOrder.PutUint64(data[offset+8:offset+16], math.Float64bits(p.Y))
	}

	return data, nil
}
