package controller_test

/*
DO NOT EDIT!
This code was generated automatically using github.com/gojuno/minimock v1.8
The original interface "Parser" can be found in github.com/fabric8-services/fabric8-toggles-service/vendor/github.com/fabric8-services/fabric8-auth/token
*/
import (
	context "context"
	rsa "crypto/rsa"
	"sync/atomic"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gojuno/minimock"

	testify_assert "github.com/stretchr/testify/assert"
)

//ParserMock implements github.com/fabric8-services/fabric8-toggles-service/vendor/github.com/fabric8-services/fabric8-auth/token.Parser
type ParserMock struct {
	t minimock.Tester

	ParseFunc       func(p context.Context, p1 string) (r *jwt.Token, r1 error)
	ParseCounter    uint64
	ParsePreCounter uint64
	ParseMock       mParserMockParse

	PublicKeysFunc       func() (r []*rsa.PublicKey)
	PublicKeysCounter    uint64
	PublicKeysPreCounter uint64
	PublicKeysMock       mParserMockPublicKeys
}

//NewParserMock returns a mock for github.com/fabric8-services/fabric8-toggles-service/vendor/github.com/fabric8-services/fabric8-auth/token.Parser
func NewParserMock(t minimock.Tester) *ParserMock {
	m := &ParserMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.ParseMock = mParserMockParse{mock: m}
	m.PublicKeysMock = mParserMockPublicKeys{mock: m}

	return m
}

type mParserMockParse struct {
	mock             *ParserMock
	mockExpectations *ParserMockParseParams
}

//ParserMockParseParams represents input parameters of the Parser.Parse
type ParserMockParseParams struct {
	p  context.Context
	p1 string
}

//Expect sets up expected params for the Parser.Parse
func (m *mParserMockParse) Expect(p context.Context, p1 string) *mParserMockParse {
	m.mockExpectations = &ParserMockParseParams{p, p1}
	return m
}

//Return sets up a mock for Parser.Parse to return Return's arguments
func (m *mParserMockParse) Return(r *jwt.Token, r1 error) *ParserMock {
	m.mock.ParseFunc = func(p context.Context, p1 string) (*jwt.Token, error) {
		return r, r1
	}
	return m.mock
}

//Set uses given function f as a mock of Parser.Parse method
func (m *mParserMockParse) Set(f func(p context.Context, p1 string) (r *jwt.Token, r1 error)) *ParserMock {
	m.mock.ParseFunc = f
	return m.mock
}

//Parse implements github.com/fabric8-services/fabric8-toggles-service/vendor/github.com/fabric8-services/fabric8-auth/token.Parser interface
func (m *ParserMock) Parse(p context.Context, p1 string) (r *jwt.Token, r1 error) {
	atomic.AddUint64(&m.ParsePreCounter, 1)
	defer atomic.AddUint64(&m.ParseCounter, 1)

	if m.ParseMock.mockExpectations != nil {
		testify_assert.Equal(m.t, *m.ParseMock.mockExpectations, ParserMockParseParams{p, p1},
			"Parser.Parse got unexpected parameters")

		if m.ParseFunc == nil {

			m.t.Fatal("No results are set for the ParserMock.Parse")

			return
		}
	}

	if m.ParseFunc == nil {
		m.t.Fatal("Unexpected call to ParserMock.Parse")
		return
	}

	return m.ParseFunc(p, p1)
}

//ParseMinimockCounter returns a count of ParserMock.ParseFunc invocations
func (m *ParserMock) ParseMinimockCounter() uint64 {
	return atomic.LoadUint64(&m.ParseCounter)
}

//ParseMinimockPreCounter returns the value of ParserMock.Parse invocations
func (m *ParserMock) ParseMinimockPreCounter() uint64 {
	return atomic.LoadUint64(&m.ParsePreCounter)
}

type mParserMockPublicKeys struct {
	mock *ParserMock
}

//Return sets up a mock for Parser.PublicKeys to return Return's arguments
func (m *mParserMockPublicKeys) Return(r []*rsa.PublicKey) *ParserMock {
	m.mock.PublicKeysFunc = func() []*rsa.PublicKey {
		return r
	}
	return m.mock
}

//Set uses given function f as a mock of Parser.PublicKeys method
func (m *mParserMockPublicKeys) Set(f func() (r []*rsa.PublicKey)) *ParserMock {
	m.mock.PublicKeysFunc = f
	return m.mock
}

//PublicKeys implements github.com/fabric8-services/fabric8-toggles-service/vendor/github.com/fabric8-services/fabric8-auth/token.Parser interface
func (m *ParserMock) PublicKeys() (r []*rsa.PublicKey) {
	atomic.AddUint64(&m.PublicKeysPreCounter, 1)
	defer atomic.AddUint64(&m.PublicKeysCounter, 1)

	if m.PublicKeysFunc == nil {
		m.t.Fatal("Unexpected call to ParserMock.PublicKeys")
		return
	}

	return m.PublicKeysFunc()
}

//PublicKeysMinimockCounter returns a count of ParserMock.PublicKeysFunc invocations
func (m *ParserMock) PublicKeysMinimockCounter() uint64 {
	return atomic.LoadUint64(&m.PublicKeysCounter)
}

//PublicKeysMinimockPreCounter returns the value of ParserMock.PublicKeys invocations
func (m *ParserMock) PublicKeysMinimockPreCounter() uint64 {
	return atomic.LoadUint64(&m.PublicKeysPreCounter)
}

//ValidateCallCounters checks that all mocked methods of the interface have been called at least once
//Deprecated: please use MinimockFinish method or use Finish method of minimock.Controller
func (m *ParserMock) ValidateCallCounters() {

	if m.ParseFunc != nil && atomic.LoadUint64(&m.ParseCounter) == 0 {
		m.t.Fatal("Expected call to ParserMock.Parse")
	}

	if m.PublicKeysFunc != nil && atomic.LoadUint64(&m.PublicKeysCounter) == 0 {
		m.t.Fatal("Expected call to ParserMock.PublicKeys")
	}

}

//CheckMocksCalled checks that all mocked methods of the interface have been called at least once
//Deprecated: please use MinimockFinish method or use Finish method of minimock.Controller
func (m *ParserMock) CheckMocksCalled() {
	m.Finish()
}

//Finish checks that all mocked methods of the interface have been called at least once
//Deprecated: please use MinimockFinish or use Finish method of minimock.Controller
func (m *ParserMock) Finish() {
	m.MinimockFinish()
}

//MinimockFinish checks that all mocked methods of the interface have been called at least once
func (m *ParserMock) MinimockFinish() {

	if m.ParseFunc != nil && atomic.LoadUint64(&m.ParseCounter) == 0 {
		m.t.Fatal("Expected call to ParserMock.Parse")
	}

	if m.PublicKeysFunc != nil && atomic.LoadUint64(&m.PublicKeysCounter) == 0 {
		m.t.Fatal("Expected call to ParserMock.PublicKeys")
	}

}

//Wait waits for all mocked methods to be called at least once
//Deprecated: please use MinimockWait or use Wait method of minimock.Controller
func (m *ParserMock) Wait(timeout time.Duration) {
	m.MinimockWait(timeout)
}

//MinimockWait waits for all mocked methods to be called at least once
//this method is called by minimock.Controller
func (m *ParserMock) MinimockWait(timeout time.Duration) {
	timeoutCh := time.After(timeout)
	for {
		ok := true
		ok = ok && (m.ParseFunc == nil || atomic.LoadUint64(&m.ParseCounter) > 0)
		ok = ok && (m.PublicKeysFunc == nil || atomic.LoadUint64(&m.PublicKeysCounter) > 0)

		if ok {
			return
		}

		select {
		case <-timeoutCh:

			if m.ParseFunc != nil && atomic.LoadUint64(&m.ParseCounter) == 0 {
				m.t.Error("Expected call to ParserMock.Parse")
			}

			if m.PublicKeysFunc != nil && atomic.LoadUint64(&m.PublicKeysCounter) == 0 {
				m.t.Error("Expected call to ParserMock.PublicKeys")
			}

			m.t.Fatalf("Some mocks were not called on time: %s", timeout)
			return
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

//AllMocksCalled returns true if all mocked methods were called before the execution of AllMocksCalled,
//it can be used with assert/require, i.e. assert.True(mock.AllMocksCalled())
func (m *ParserMock) AllMocksCalled() bool {

	if m.ParseFunc != nil && atomic.LoadUint64(&m.ParseCounter) == 0 {
		return false
	}

	if m.PublicKeysFunc != nil && atomic.LoadUint64(&m.PublicKeysCounter) == 0 {
		return false
	}

	return true
}
