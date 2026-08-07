//go:build goplus_java

package participle

import "testing"

func TestJavaAssignments(t *testing.T) {
	grammar := AssignmentGrammar(17)
	parser := BuildAssignments(grammar, AssignmentFirst(grammar))
	match Parse[[]Assignment](parser, "workers=3; sprintDays=10") {
	case Rejected(failure): t.Fatal(failure.Error())
	case Parsed(values):
		if len(values) != 2 { t.Fatal("want two assignments") }
		if values[0].Name != "workers" || values[0].Value != 3 { t.Fatal("wrong first assignment") }
		if values[1].Name != "sprintDays" || values[1].Value != 10 { t.Fatal("wrong second assignment") }
	}
}

func TestJavaFailurePosition(t *testing.T) {
	grammar := AssignmentGrammar(18)
	parser := BuildAssignments(grammar, AssignmentFirst(grammar))
	match Parse[[]Assignment](parser, "workers=no") {
	case Parsed(_): t.Fatal("invalid integer parsed")
	case Rejected(failure):
		if failure.Line != 1 || failure.Column != 9 { t.Fatal("wrong failure position") }
	}
}
