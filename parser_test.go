//go:build !goplus_java

package participle

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	upstream "github.com/alecthomas/participle/v2"
)

type upstreamDocument struct {
	Assignments []upstreamAssignment `parser:"@@*"`
}

type upstreamAssignment struct {
	Name  string `parser:"@Ident '='"`
	Value int64  `parser:"@Int ';'?"`
}

var upstreamParser = upstream.MustBuild[upstreamDocument]()
var assignmentSink int

func checkedParser() Parser[[]Assignment] {
	grammar := AssignmentGrammar(17)
	first := AssignmentFirst(grammar)
	return BuildAssignments(grammar, first)
}

func unwrap(t *testing.T, outcome Outcome[[]Assignment]) []Assignment {
	t.Helper()
	switch value := outcome.(type) {
	case Parsed[[]Assignment]:
		return value.Value
	case Rejected[[]Assignment]:
		t.Fatalf("parse failed: %v", value.Failure)
	default:
		t.Fatal("impossible outcome")
	}
	return nil
}

func TestAssignmentsAndIndexedSpans(t *testing.T) {
	values := unwrap(t, Parse(checkedParser(), "alpha=1; beta = -2\ngamma=9223372036854775807"))
	if len(values) != 3 || values[0].Name != "alpha" || values[1].Value != -2 {
		t.Fatalf("unexpected values: %#v", values)
	}
	if got, want := SpanStart(values[1].Span), 9; got != want {
		t.Fatalf("span start=%d, want %d", got, want)
	}
	if got, want := SpanEnd(values[1].Span), 18; got != want {
		t.Fatalf("span end=%d, want %d", got, want)
	}
}

func TestMinInt64AndDiagnostics(t *testing.T) {
	values := unwrap(t, Parse(checkedParser(), "min=-9223372036854775808"))
	if values[0].Value != -1<<63 {
		t.Fatalf("value=%d", values[0].Value)
	}

	cases := []struct {
		input                string
		kind                 string
		offset, line, column int
	}{
		{"1=2", "expected identifier", 0, 1, 1},
		{"a 2", "expected '='", 2, 1, 3},
		{"a=", "expected integer", 2, 1, 3},
		{"a=9223372036854775808", "integer overflow", 2, 1, 3},
		{"a=1\nb=?", "expected integer", 6, 2, 3},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result, ok := Parse(checkedParser(), tc.input).(Rejected[[]Assignment])
			if !ok {
				t.Fatal("expected rejection")
			}
			if !strings.Contains(result.Failure.Error(), tc.kind) || result.Failure.Offset != tc.offset || result.Failure.Line != tc.line || result.Failure.Column != tc.column {
				t.Fatalf("failure=%+v (%v)", result.Failure, result.Failure.Error())
			}
		})
	}
}

func TestUpstreamBaseZeroIntegerSemantics(t *testing.T) {
	values := unwrap(t, Parse(checkedParser(), "octal=0010; negative=-0010"))
	if values[0].Value != 8 || values[1].Value != -8 {
		t.Fatalf("values=%#v", values)
	}
}

func TestErasedEvidenceGuards(t *testing.T) {
	grammar := AssignmentGrammar(1)
	otherGrammar := AssignmentGrammar(2)
	other := AssignmentFirst(otherGrammar)
	defer func() {
		if recover() == nil {
			t.Fatal("expected mismatched evidence panic")
		}
	}()
	BuildAssignments(grammar, other)
}

func TestAmbiguityIsExhaustiveFailure(t *testing.T) {
	grammar := AssignmentGrammar(3)
	first := firstSetValue{grammarID: 3, symbols: []string{"Identifier", "Identifier"}}
	result, ok := Parse(BuildAssignments(grammar, first), "a=1").(Rejected[[]Assignment])
	if !ok {
		t.Fatal("expected build rejection")
	}
	if _, ok := result.Failure.Kind.(AmbiguousFirstSet); !ok {
		t.Fatalf("kind=%T", result.Failure.Kind)
	}
}

func TestDynamicGrammarBinding(t *testing.T) {
	grammar := BindAssignments(9, LoadAssignments(9))
	values := unwrap(t, Parse(BuildAssignments(grammar, AssignmentFirst(grammar)), "dynamic=9"))
	if len(values) != 1 || values[0].Name != "dynamic" {
		t.Fatalf("values=%#v", values)
	}
}

func TestParserLaws(t *testing.T) {
	parser := checkedParser()
	input := "alpha=1; beta=-2; gamma=3"
	first := Parse(parser, input)
	second := Parse(parser, input)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("determinism law failed")
	}

	semicolon := unwrap(t, first)
	newline := unwrap(t, Parse(parser, "alpha=1\nbeta=-2\ngamma=3"))
	if len(semicolon) != len(newline) {
		t.Fatal("separator law changed result length")
	}
	for i := range semicolon {
		if semicolon[i].Name != newline[i].Name || semicolon[i].Value != newline[i].Value {
			t.Fatalf("separator law failed at %d", i)
		}
		start, end := SpanStart(semicolon[i].Span), SpanEnd(semicolon[i].Span)
		if start < 0 || end > len(input) || start >= end || !strings.HasPrefix(input[start:end], semicolon[i].Name) {
			t.Fatalf("span law failed for %#v", semicolon[i])
		}
	}

	dynamic := BindAssignments(17, LoadAssignments(17))
	dynamicResult := Parse(BuildAssignments(dynamic, AssignmentFirst(dynamic)), input)
	if !reflect.DeepEqual(first, dynamicResult) {
		t.Fatal("dynamic/static binding law failed")
	}
}

func TestAllocationReductionGate(t *testing.T) {
	input := benchmarkInput(256)
	parser := checkedParser()
	ours := testing.AllocsPerRun(20, func() { assignmentSink = len(unwrapB(Parse(parser, input))) })
	up := testing.AllocsPerRun(20, func() {
		value, err := upstreamParser.ParseString("allocation-gate", input)
		if err != nil {
			panic(err)
		}
		assignmentSink = len(value.Assignments)
	})
	if ours*2 > up {
		t.Fatalf("allocation gate: ours %.0f, upstream %.0f", ours, up)
	}
}

func FuzzDifferentialAssignments(f *testing.F) {
	for _, seed := range []string{"", "a=1", "a=-1;b=2", "alpha=9223372036854775807;", "bad=?"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		ours, oursOK := Parse(checkedParser(), input).(Parsed[[]Assignment])
		up, err := upstreamParser.ParseString("", input)
		if oursOK != (err == nil) {
			return
		} // Grammars intentionally differ on whitespace/newlines.
		if err == nil {
			if len(ours.Value) != len(up.Assignments) {
				t.Fatalf("length mismatch")
			}
			for i := range ours.Value {
				if ours.Value[i].Name != up.Assignments[i].Name || ours.Value[i].Value != up.Assignments[i].Value {
					t.Fatalf("value mismatch")
				}
			}
		}
	})
}

func benchmarkInput(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "field%d=%d;", i, i)
	}
	return b.String()
}

func BenchmarkAssignments(b *testing.B) {
	input := benchmarkInput(256)
	parser := checkedParser()
	b.Run("GoForge", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(input)))
		for i := 0; i < b.N; i++ {
			assignmentSink = len(unwrapB(Parse(parser, input)))
		}
	})
	b.Run("Upstream", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(input)))
		for i := 0; i < b.N; i++ {
			value, err := upstreamParser.ParseString(strconv.Itoa(i), input)
			if err != nil {
				b.Fatal(err)
			}
			assignmentSink = len(value.Assignments)
		}
	})
}

func unwrapB(outcome Outcome[[]Assignment]) []Assignment {
	switch result := outcome.(type) {
	case Parsed[[]Assignment]:
		return result.Value
	case Rejected[[]Assignment]:
		panic(result.Failure)
	default:
		panic("impossible outcome")
	}
}
