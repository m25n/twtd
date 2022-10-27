package testhelper

type MockLogger struct {
	WritingBodyErrs     []error
	FollowerLoggingErrs []error
	PostingStatusErrs   []error
	GettingTwtxtErrs    []error
}

func (l *MockLogger) GettingTwtxtErr(err error) {
	l.GettingTwtxtErrs = append(l.GettingTwtxtErrs, err)
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *MockLogger) WritingBodyErr(err error) {
	l.WritingBodyErrs = append(l.WritingBodyErrs, err)
}

func (l *MockLogger) FollowerLoggingErr(err error) {
	l.FollowerLoggingErrs = append(l.FollowerLoggingErrs, err)
}

func (l *MockLogger) PostingStatusErr(err error) {
	l.PostingStatusErrs = append(l.PostingStatusErrs, err)
}

type DummyLogger struct{}

func (d DummyLogger) GettingTwtxtErr(_ error) {}

func (d DummyLogger) WritingBodyErr(_ error) {}

func (d DummyLogger) FollowerLoggingErr(_ error) {}

func (d DummyLogger) PostingStatusErr(_ error) {}
