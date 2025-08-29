package schema

import (
	"testing"
)

func TestParseVectorElementType(t *testing.T) {
	sg := &SchemaGenerator{}

	tests := []struct {
		vectorType string
		expected   string
	}{
		{"vector(128,float)", "float"},
		{"vector(256,double)", "double"},
		{"vector(512,int)", "int"},
		{"vector(1024,bigint)", "bigint"},
		{"VECTOR(128,FLOAT)", "float"},
		{"Vector(256,Double)", "double"},
		{"vector(128)", "float"},           // Default when no element type (MariaDB default)
		{"vector(1024)", "float"},          // Real MariaDB format - dimension only
		{"vector", "float"},               // Default for invalid format
		{"not_a_vector", "float"},         // Default for non-vector type
		{"vector(128, float )", "float"},   // With spaces
		{"vector(256, double)", "double"},  // With spaces
	}

	for _, test := range tests {
		result := sg.parseVectorElementType(test.vectorType)
		if result != test.expected {
			t.Errorf("parseVectorElementType(%q) = %q, expected %q", 
				test.vectorType, result, test.expected)
		}
	}
}

func TestMysqlTypeToGoType_Boolean(t *testing.T) {
	sg := &SchemaGenerator{}

	tests := []struct {
		mysqlType string
		nullable  bool
		expected  string
	}{
		{"tinyint(1)", false, "bool"},
		{"tinyint(1)", true, "sql.NullBool"},
		{"TINYINT(1)", false, "bool"},
		{"TINYINT(1)", true, "sql.NullBool"},
		{"tinyint(2)", false, "int32"}, // Not a boolean, should be int32
		{"tinyint", false, "int32"},    // Not a boolean, should be int32
		{"bool", false, "bool"},        // Legacy support
		{"boolean", false, "bool"},     // Legacy support
		{"bit", false, "bool"},         // Legacy support
	}

	for _, test := range tests {
		result := sg.mysqlTypeToGoType(test.mysqlType, test.nullable, false, "test_table", "test_column")
		if result != test.expected {
			t.Errorf("mysqlTypeToGoType(%q, nullable=%t) = %q, expected %q", 
				test.mysqlType, test.nullable, result, test.expected)
		}
	}
}

func TestMysqlTypeToGoType_Vector(t *testing.T) {
	sg := &SchemaGenerator{}

	tests := []struct {
		mysqlType string
		expected  string
	}{
		{"vector(128,float)", "types.Vector[float32]"},
		{"vector(256,double)", "types.Vector[float64]"},
		{"vector(512,int)", "types.Vector[int32]"},
		{"vector(1024,bigint)", "types.Vector[int64]"},
		{"VECTOR(128,FLOAT)", "types.Vector[float32]"},
		{"vector(256)", "types.Vector[float32]"}, // Default to float32 (MariaDB default)
		{"vector(1024)", "types.Vector[float32]"}, // Real MariaDB format
		{"vector(128,unknown)", "types.Vector[float64]"}, // Default to float64 for unknown types
	}

	for _, test := range tests {
		result := sg.mysqlTypeToGoType(test.mysqlType, false, false, "test_table", "test_column")
		if result != test.expected {
			t.Errorf("mysqlTypeToGoType(%q) = %q, expected %q", 
				test.mysqlType, result, test.expected)
		}
	}
}

func TestToColumnTypeName(t *testing.T) {
	sg := &SchemaGenerator{}

	tests := []struct {
		tableName  string
		columnName string
		expected   string
	}{
		{"users", "id", "Users_Id"},
		{"user_profiles", "user_id", "UserProfiles_UserId"},
		{"order_items", "created_at", "OrderItems_CreatedAt"},
		{"test_table", "test_column", "TestTable_TestColumn"},
		{"USERS", "EMAIL", "USERS_EMAIL"},
		{"my_table", "my_field", "MyTable_MyField"},
	}

	for _, test := range tests {
		result := sg.toColumnTypeName(test.tableName, test.columnName)
		if result != test.expected {
			t.Errorf("toColumnTypeName(%q, %q) = %q, expected %q", 
				test.tableName, test.columnName, result, test.expected)
		}
	}
}
