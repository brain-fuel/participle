package compilefixture

import participle "goforge.dev/participle"

func Accepted() {
	grammar := participle.AssignmentGrammar(41)
	first := participle.AssignmentFirst(grammar)
	parser := participle.BuildAssignments(grammar, first)
	_ = participle.Parse(parser, "answer=42")
}
