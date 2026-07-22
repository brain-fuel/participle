package compilefixture

import participle "goforge.dev/participle"

func NeedsThree(span participle.Span[3]) {}

func Rejected() {
	NeedsThree(participle.NewSpan(2, 0, 2))
}
