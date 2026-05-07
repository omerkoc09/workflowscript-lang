package checker

import "fmt"

type Type interface {
	typeString() string
}

type IntType struct{}
type FloatType struct{}
type BoolType struct{}
type VoidType struct{}
type TaskType struct{}

func (IntType) typeString() string   { return "int" }
func (FloatType) typeString() string { return "float" }
func (BoolType) typeString() string  { return "bool" }
func (VoidType) typeString() string  { return "void" }
func (TaskType) typeString() string  { return "task" }

var (
	TInt   Type = IntType{}
	TFloat Type = FloatType{}
	TBool  Type = BoolType{}
	TVoid  Type = VoidType{}
	TTask  Type = TaskType{}
)

// typesCompatible reports whether a value of type `a` can be used where type `b` is expected.
// int widens to float; no other implicit conversions.
func typesCompatible(a, b Type) bool {
	if a == b {
		return true
	}
	_, aIsInt := a.(IntType)
	_, bIsFloat := b.(FloatType)
	return aIsInt && bIsFloat
}

func typeFromLexeme(s string) (Type, error) {
	switch s {
	case "int":
		return TInt, nil
	case "float":
		return TFloat, nil
	case "bool":
		return TBool, nil
	case "void":
		return TVoid, nil
	case "task":
		return TTask, nil
	}
	return nil, fmt.Errorf("unknown type %q", s)
}
