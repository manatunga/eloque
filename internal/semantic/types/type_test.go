package types

import "testing"

func TestTypeIDString(t *testing.T) {
	tests := []struct {
		name string
		id   TypeID
		want string
	}{
		{"invalid", InvalidTypeID, "INVALID"},
		{"int8", TypeInt8, "Int8"},
		{"int16", TypeInt16, "Int16"},
		{"int32", TypeInt32, "Int32"},
		{"int", TypeInt, "Int"},
		{"float", TypeFloat, "Float"},
		{"string", TypeString, "String"},
		{"bool", TypeBool, "Bool"},
		{"error", TypeError, "<error>"},
		{"unknown", TypeID(255), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeIDClassification(t *testing.T) {
	tests := []struct {
		name       string
		id         TypeID
		integer    bool
		numeric    bool
		sourceType bool
		errorType  bool
	}{
		{"invalid", InvalidTypeID, false, false, false, false},
		{"int8", TypeInt8, true, true, true, false},
		{"int16", TypeInt16, true, true, true, false},
		{"int32", TypeInt32, true, true, true, false},
		{"int", TypeInt, true, true, true, false},
		{"float", TypeFloat, false, true, true, false},
		{"string", TypeString, false, false, true, false},
		{"bool", TypeBool, false, false, true, false},
		{"error", TypeError, false, false, false, true},
		{"unknown", TypeID(255), false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsInteger(); got != tt.integer {
				t.Errorf("IsInteger() = %t, want %t", got, tt.integer)
			}

			if got := tt.id.IsNumeric(); got != tt.numeric {
				t.Errorf("IsNumeric() = %t, want %t", got, tt.numeric)
			}

			if got := tt.id.IsSourceType(); got != tt.sourceType {
				t.Errorf("IsSourceType() = %t, want %t", got, tt.sourceType)
			}

			if got := tt.id.IsError(); got != tt.errorType {
				t.Errorf("IsError() = %t, want %t", got, tt.errorType)
			}
		})
	}
}

func TestTypeIDZeroValueIsInvalid(t *testing.T) {
	var id TypeID

	if id != InvalidTypeID {
		t.Fatalf("zero-value TypeID = %d, want InvalidTypeID (%d)", id, InvalidTypeID)
	}

	if id.IsSourceType() {
		t.Error("zero-value TypeID must not be classified as a source type")
	}

	if id.IsError() {
		t.Error("zero-value TypeID must not be classified as an error type")
	}
}

func TestBuiltinTypeIDsAreDistinct(t *testing.T) {
	ids := []TypeID{
		TypeInt8,
		TypeInt16,
		TypeInt32,
		TypeInt,
		TypeFloat,
		TypeString,
		TypeBool,
	}

	seen := make(map[TypeID]struct{}, len(ids))

	for _, id := range ids {
		if id == InvalidTypeID {
			t.Fatalf("built-in source type %s has InvalidTypeID", id)
		}

		if id == TypeError {
			t.Fatalf("built-in source type %s has TypeError", id)
		}

		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate built-in type identity: %s (%d)", id, id)
		}

		seen[id] = struct{}{}
	}
}

func TestInvalidAndErrorTypeIDsAreDistinct(t *testing.T) {
	if InvalidTypeID == TypeError {
		t.Fatal("InvalidTypeID and TypeError must have distinct identities")
	}

	if InvalidTypeID.IsError() {
		t.Error("InvalidTypeID must not be classified as an error type")
	}

	if !TypeError.IsError() {
		t.Error("TypeError must be classified as an error type")
	}

	if InvalidTypeID.IsSourceType() {
		t.Error("InvalidTypeID must not be source-declarable")
	}

	if TypeError.IsSourceType() {
		t.Error("TypeError must not be source-declarable")
	}
}
