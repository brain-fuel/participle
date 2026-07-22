// Package participle provides grammar-indexed parser construction with
// immutable spans, explicit commit semantics, and typed AST output.
package participle

// Span[n] covers exactly n grammar tokens. Offsets are byte offsets and the
// end is exclusive.
//goplus:derive off
type Span[n nat] enum {
	spanValue(Start int, End int, Tokens int) Span[n]
}

type Assignment struct {
	Name string
	Value int64
	Span Span[3]
}

// Expression[n] is a grammar AST indexed by its minimum token consumption.
// Lookahead consumes zero, commit preserves consumption, alternatives must
// agree, sequence adds, and repetition has a zero-token lower bound.
type Expression[n nat] enum {
	Empty() Expression[0]
	Token(Symbol string) Expression[1]
	Sequence(Prefix Expression[n], Last Expression[1]) Expression[n+1]
	Alternative(Left Expression[n], Right Expression[n]) Expression[n]
	Lookahead(Body Expression[n]) Expression[0]
	Guarded(Guard Expression[0], Body Expression[n]) Expression[n]
	Commit(Body Expression[n]) Expression[n]
	Many(Body Expression[n]) Expression[0]
}

// Grammar[g] is an immutable grammar identity and its FIRST set.
//goplus:derive off
type Grammar[g nat] enum {
	grammarValue(ID int, Name string, Root Expression[0], First []string) Grammar[g]
}

// SomeGrammar is the dynamic-loading boundary. It hides the static grammar
// identity until BindAssignments checks and reintroduces a requested witness.
type SomeGrammar enum {
	LoadedAssignments(ID int, Root Expression[0], First []string)
}

// FirstSet[g] is evidence computed for exactly Grammar[g].
//goplus:derive off
type FirstSet[g nat] enum {
	firstSetValue(GrammarID int, Symbols []string) FirstSet[g]
}

// Parser[g,T] can only produce T for Grammar[g].
type parserRunner[T any] func(string) (T, *Failure)

//goplus:derive off
type Parser[g nat, T any] enum {
	parserValue(GrammarID int, Run parserRunner[T]) Parser[g, T]
	parserInvalid(Failure Failure) Parser[g, T]
}

type FailureKind enum {
	UnexpectedByte
	ExpectedIdentifier
	ExpectedEquals
	ExpectedInteger
	IntegerOverflow
	AmbiguousFirstSet
}

type Failure struct {
	Kind FailureKind
	Offset int
	Line int
	Column int
	Unexpected string
	Expected []string
}

func (failure Failure) Error() string { return failureMessage(&failure) }

type Outcome[T any] enum {
	Parsed(Value T)
	Rejected(Failure Failure)
}

// AssignmentGrammar is: (Identifier '=' Integer Separator*)* EOF.
func AssignmentGrammar(id nat) Grammar[id] {
	assignment := Sequence(Sequence(Token("Identifier"), Token("=")), Token("Integer"))
	root := Many(Guarded(Lookahead(Token("Identifier")), Commit(assignment)))
	return grammarValue(int(id), "assignments", root, []string{"Identifier", "EOF"})
}

func AssignmentFirst(0 id nat, grammar Grammar[id]) FirstSet[id] {
	match grammar { case grammarValue(grammarID, _, _, first): return firstSetValue(grammarID, append([]string(nil), first...)) }
}

// AssignmentFirstFor is the explicit witness constructor used when grammar
// and FIRST evidence are loaded independently across a package boundary.
func AssignmentFirstFor(id nat) FirstSet[id] {
	return firstSetValue(int(id), []string{"Identifier", "EOF"})
}

func LoadAssignments(id int) SomeGrammar {
	assignment := Sequence(Sequence(Token("Identifier"), Token("=")), Token("Integer"))
	return LoadedAssignments(id, Many(Guarded(Lookahead(Token("Identifier")), Commit(assignment))), []string{"Identifier", "EOF"})
}

func BindAssignments(id nat, loaded SomeGrammar) Grammar[id] {
	match loaded {
	case LoadedAssignments(loadedID, root, first):
		if loadedID != int(id) { panic("participle: dynamic grammar identity mismatch") }
		return grammarValue(loadedID, "assignments", root, append([]string(nil), first...))
	}
}

func NewSpan(tokens nat, start int, end int) Span[tokens] {
	if tokens < 0 || start < 0 || end < start { panic("participle: invalid span") }
	return spanValue(start, end, int(tokens))
}

func SpanStart(0 n nat, span Span[n]) int { match span { case spanValue(start, _, _): return start } }
func SpanEnd(0 n nat, span Span[n]) int { match span { case spanValue(_, end, _): return end } }

func BuildAssignments(0 g nat, grammar Grammar[g], first FirstSet[g]) Parser[g, []Assignment] {
	failure := validateFirst(grammar, first)
	if failure != nil { return parserInvalid[[]Assignment](*failure) }
	match grammar { case grammarValue(grammarID, _, _, _): return parserValue(grammarID, parseAssignments) }
}

func validateFirst(0 g nat, grammar Grammar[g], first FirstSet[g]) *Failure {
	match grammar {
	case grammarValue(grammarID, _, _, _):
		match first {
		case firstSetValue(evidenceID, symbols):
			if grammarID != evidenceID { panic("participle: FIRST evidence belongs to another grammar") }
			if hasDuplicate(symbols) { return &Failure{Kind: AmbiguousFirstSet(), Expected: append([]string(nil), symbols...)} }
			return nil
		}
	}
}

func Parse[T any](0 g nat, parser Parser[g, T], input string) Outcome[T] {
	match parser {
	case parserValue(_, run):
		value, failure := run(input)
		if failure != nil { return Rejected[T](*failure) }
		return Parsed(value)
	case parserInvalid(failure): return Rejected[T](failure)
	}
}

func hasDuplicate(symbols []string) bool {
	seen := make(map[string]bool, len(symbols))
	for _, symbol := range symbols {
		if seen[symbol] { return true }
		seen[symbol] = true
	}
	return false
}
