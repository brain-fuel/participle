package compilefixture

import participle "goforge.dev/participle"

func Rejected() {
	grammar := participle.AssignmentGrammar(41)
	wrongFirst := participle.AssignmentFirstFor(42)
	_ = participle.BuildAssignments(grammar, wrongFirst)
}
