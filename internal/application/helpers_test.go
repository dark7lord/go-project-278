package application

const (
	testExampleURL    = "https://example.com"
	testShortName     = "my-link"
	testGeneratedCode = "generated-code"
	testTargetName    = "target"
	testOKURL         = "https://ok.com"
	testShortCode     = "test-code"
)

// testRedirectStatus is not 302 on purpose: the service must record the status
// it is handed, not a literal 302.
const testRedirectStatus = int32(307)

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
