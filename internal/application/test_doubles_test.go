package application

// fakeGenerator proposes the same short name every time and counts the calls.
type fakeGenerator struct {
	value string
	calls int
}

func (g *fakeGenerator) Generate() string {
	g.calls++

	return g.value
}

// fixedCode is a generator for tests that never rely on a generated name.
func fixedCode() string { return testShortCode }
