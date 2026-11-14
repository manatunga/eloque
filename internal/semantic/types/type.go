package types

// TypeID identifies a semantic type understood by the Eloque compiler.
//
// A TypeID represents the identity of a type during semantic analysis. It is
// used by semantic components to describe the types associated with
// expressions, declarations, and other compiler entities without storing
// type information directly in the immutable syntax tree.
//
// The zero value, InvalidTypeID, represents the absence of a valid type
// assignment. TypeError is a separate sentinel used when semantic analysis
// encounters an error and needs to continue processing without treating the
// affected expression as successfully typed.
//
// Built-in TypeIDs have fixed identities within the compiler. Their numeric
// values are implementation details: they are not source-language values,
// serialized type identifiers, runtime type tags, or Monowire register IDs.
//
// This initial representation covers only Eloque's built-in scalar types.
// Composite types, including optional and collection types, require an
// extended representation and are intentionally outside this definition.
type TypeID uint8

const (
	// InvalidTypeID represents the absence of a valid semantic type assignment.
	//
	// It is the zero value of TypeID and is used when a type has not yet been
	// determined or when no reliable type information is available. It is
	// distinct from TypeError, which explicitly represents a semantic error
	// during analysis.
	InvalidTypeID TypeID = iota

	// TypeInt8 identifies Eloque's signed 8-bit integer type.
	//
	// Values of this type have the signed 8-bit integer range. This identifier
	// establishes type identity only; range validation and overflow handling
	// are responsibilities of later numeric semantic analysis.
	TypeInt8

	// TypeInt16 identifies Eloque's signed 16-bit integer type.
	//
	// Values of this type have the signed 16-bit integer range. This identifier
	// establishes type identity only; range validation and overflow handling
	// are responsibilities of later numeric semantic analysis.
	TypeInt16

	// TypeInt32 identifies Eloque's signed 32-bit integer type.
	//
	// Values of this type have the signed 32-bit integer range. This identifier
	// establishes type identity only; range validation and overflow handling
	// are responsibilities of later numeric semantic analysis.
	TypeInt32

	// TypeInt identifies Eloque's default signed integer type.
	//
	// Its concrete width and range follow the language's definition of Int.
	// This identifier establishes type identity only; it does not itself
	// perform target-specific range validation or arithmetic.
	TypeInt

	// TypeFloat identifies Eloque's IEEE 754 binary64 floating-point type.
	//
	// It represents the language's built-in floating-point type. This
	// identifier establishes type identity only; floating-point parsing,
	// conversion, and arithmetic semantics are handled by their respective
	// compiler components.
	TypeFloat

	// TypeString identifies Eloque's immutable text type.
	//
	// It represents source-language string values. The identifier establishes
	// type identity only; string storage, representation, and operations are
	// outside this type identity definition.
	TypeString

	// TypeBool identifies Eloque's built-in Boolean type.
	//
	// It represents Boolean values used by source-language expressions and
	// declarations. The identifier establishes type identity only; Boolean
	// expression checking and operator semantics are handled separately.
	TypeBool

	// TypeError represents a semantic error type used during analysis recovery.
	//
	// It allows semantic analysis to continue after an earlier error without
	// treating the affected expression or declaration as successfully typed.
	// It is an internal compiler sentinel and cannot be declared as a
	// source-language type.
	TypeError
)

// String returns the stable human-readable name associated with id.
//
// Built-in source types are returned using their canonical Eloque spelling.
// InvalidTypeID is reported as "INVALID", TypeError as "<error>", and
// unrecognized identifiers as "UNKNOWN". These names are intended for
// diagnostics and debugging; they do not define source-language syntax or
// serialization formats.
func (id TypeID) String() string {
	switch id {
	case InvalidTypeID:
		return "INVALID"
	case TypeInt8:
		return "Int8"
	case TypeInt16:
		return "Int16"
	case TypeInt32:
		return "Int32"
	case TypeInt:
		return "Int"
	case TypeFloat:
		return "Float"
	case TypeString:
		return "String"
	case TypeBool:
		return "Bool"
	case TypeError:
		return "<error>"
	default:
		return "UNKNOWN"
	}
}

// IsInteger reports whether id identifies one of Eloque's built-in integer
// types.
//
// The recognized integer types are Int8, Int16, Int32, and Int. Float is
// numeric but is not an integer type. InvalidTypeID, TypeError, and unknown
// TypeID values are not classified as integers.
//
// This method classifies type identity only. It does not determine whether
// a particular integer value fits within the range of the identified type.
func (id TypeID) IsInteger() bool {
	switch id {
	case TypeInt8,
		TypeInt16,
		TypeInt32,
		TypeInt:
		return true
	default:
		return false
	}
}

// IsNumeric reports whether id identifies one of Eloque's built-in numeric
// types.
//
// The recognized numeric types are Int8, Int16, Int32, Int, and Float.
// String and Bool are not numeric types. InvalidTypeID, TypeError, and
// unknown TypeID values are not classified as numeric.
//
// This method classifies type identity only. It does not establish whether
// two numeric types are compatible or whether a conversion between them is
// permitted.
func (id TypeID) IsNumeric() bool {
	switch id {
	case TypeInt8,
		TypeInt16,
		TypeInt32,
		TypeInt,
		TypeFloat:
		return true
	default:
		return false
	}
}

// IsSourceType reports whether id identifies a type that source code may
// declare.
//
// The recognized source-declarable types are Int8, Int16, Int32, Int, Float,
// String, and Bool. InvalidTypeID and TypeError are internal sentinels rather
// than source-language types. Unknown TypeID values are also rejected.
//
// This classification covers only the built-in types represented by the
// current TypeID definition. It does not determine whether a declaration is
// otherwise valid in its surrounding scope or context.
func (id TypeID) IsSourceType() bool {
	switch id {
	case TypeInt8,
		TypeInt16,
		TypeInt32,
		TypeInt,
		TypeFloat,
		TypeString,
		TypeBool:
		return true
	default:
		return false
	}
}

// IsError reports whether id identifies the internal semantic error type.
//
// It returns true only for TypeError. InvalidTypeID represents missing or
// undetermined type information and is deliberately not treated as a
// semantic error type. Unknown TypeID values are not classified as errors.
//
// This distinction allows semantic analysis to differentiate an explicit
// error-recovery state from the absence of a type assignment.
func (id TypeID) IsError() bool {
	switch id {
	case TypeError:
		return true
	default:
		return false
	}
}
