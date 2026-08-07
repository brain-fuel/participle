# Compatibility and scope

Pinned upstream: `github.com/alecthomas/participle/v2` v2.1.4, commit
`bcbb39153e17f8018257f17aba8eac628d396b64`. Its MIT notice is preserved in
`UPSTREAM_LICENSE`.

The differential corpus covers repeated `identifier = signed-int64`
assignments, optional semicolons, whitespace, exact values, and acceptance vs.
rejection. GoForge additionally accepts newline-separated assignments and
returns immutable three-token source spans. It reports explicit identifier,
equals, integer, overflow, unexpected-byte, and FIRST-set failures.
Integer capture follows upstream's Go base-zero behavior, including octal
leading-zero forms such as `0010`.

Not yet compatible: struct-tag grammar compilation, custom mapper/capture
hooks, arbitrary lexer definitions, EBNF rendering, elision, union types, and
Participle's full error API. These stay outside v0.2 rather than being claimed
as compatible.

Version 0.2 adds a Java 25 output target for this same documented semantic
surface. The Java artifact is generated from this package's Go+ enums, indexed
grammar structures, and parser implementation, with Java-specific conformance
tests for successful and rejected parses.

The generated parser is tied to the assignment grammar AST and its FIRST-set
witness. Lookahead and commit are explicit nodes rather than implicit control
flow. The runtime scanner is specialized after that checked construction
boundary.

## Standard-library decision

No `std/parsec` API was changed in this release. Its rune-oriented input
positions and combinator consumption model differ from this package's byte
offsets and token-count-indexed spans. Promoting a shared span today would
either weaken `Span[n]` or force an unrelated parsec migration with only one
proven consumer. Promotion becomes prudent after a second package demonstrates
the same immutable token-span/error primitive.
