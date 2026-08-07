# participle

`goforge.dev/participle` is a Go+-authored, grammar-indexed parser core. The
first compatibility target is the MIT-licensed
[`github.com/alecthomas/participle/v2`](https://github.com/alecthomas/participle)
v2.1.4 assignment-language workload.

The package uses Go+ abstractions that ordinary Go cannot state:

- `Expression[n]` is a GADT indexed by minimum token consumption. Sequence
  increments the index, lookahead consumes zero, commit preserves it, and
  alternatives must agree.
- `Grammar[g]`, `FirstSet[g]`, and `Parser[g,T]` share a grammar identity.
  Go+ rejects evidence or parsers from another grammar before generation.
- `Span[n]` records a source range covering exactly `n` grammar tokens.
- `SomeGrammar` is the checked dynamic-loading boundary; `BindAssignments`
  reintroduces a static identity after validating its retained runtime ID.
- parse failures are an exhaustive algebraic type with stable byte offset,
  line, column, unexpected input, and expected symbols.

```go
grammar := participle.AssignmentGrammar(41)
first := participle.AssignmentFirst(grammar)
parser := participle.BuildAssignments(grammar, first)
result := participle.Parse(parser, "answer=42")
```

The `nat` witness arguments are inferred and erased. Generated Go exposes a
conventional generic API and retains runtime identity guards at Go boundaries.

## Java 25 and Maven

Version 0.2 adds a Java 25 library target generated from the same Go+ semantic
source. Build and test it with Go+ v0.142.0 or newer:

```sh
goplus build --target java ./...
goplus test --target java ./...
```

The resulting Java module and JAR are both named `dev.goforge.participle`.
After its first Maven Central publication, Java projects can depend on
`dev.goforge:participle:0.2.0`. The Central release includes the binary,
generated sources, documentation, POM, checksums, and OpenPGP signatures.
The standalone consumer under `testdata/java-consumer` is compiled against only
that finished JAR during release verification.

## Verification

```sh
goplus gen -check .
go test -race ./...
goplus test --target java ./...
go test -run '^$' -bench BenchmarkAssignments -benchmem -count=5
```

Compile fixtures under `testdata/compile` prove that valid programs generate,
while a foreign FIRST witness and a two-token span passed as `Span[3]` fail
with `dependent index mismatch`.

The parser currently targets the surface in [COMPATIBILITY.md](COMPATIBILITY.md);
it is not presented as a drop-in replacement for Participle's complete API.

## Independent consumer

Agile Frontier v0.2 uses this package inside its Go+/WASM planning engine to
parse compact capacity overrides. That separate module exercises inferred
grammar identity, FIRST evidence, exhaustive outcomes, generated-Go interop,
and the browser WASM boundary rather than treating the parser as an isolated
example.
