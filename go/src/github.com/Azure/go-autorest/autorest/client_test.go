package autorest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/mocks"
)

func TestLoggingInspectorWithInspection(t *testing.T) {
	b := bytes.Buffer{}
	c := Client{}
	li := LoggingInspector{Logger: log.New(&b, "", 0)}
	c.RequestInspector = li.WithInspection()

	Prepare(mocks.NewRequestWithContent("Content"),
		c.WithInspection())

	if len(b.String()) <= 0 {
		t.Error("autorest: LoggingInspector#WithInspection did not record Request to the log")
	}
}

func TestLoggingInspectorWithInspectionEmitsErrors(t *testing.T) {
	b := bytes.Buffer{}
	c := Client{}
	r := mocks.NewRequestWithContent("Content")
	li := LoggingInspector{Logger: log.New(&b, "", 0)}
	c.RequestInspector = li.WithInspection()

	r.Body.Close()
	Prepare(r,
		c.WithInspection())

	if len(b.String()) <= 0 {
		t.Error("autorest: LoggingInspector#WithInspection did not record Request to the log")
	}
}

func TestLoggingInspectorWithInspectionRestoresBody(t *testing.T) {
	b := bytes.Buffer{}
	c := Client{}
	r := mocks.NewRequestWithContent("Content")
	li := LoggingInspector{Logger: log.New(&b, "", 0)}
	c.RequestInspector = li.WithInspection()

	Prepare(r,
		c.WithInspection())

	s, _ := ioutil.ReadAll(r.Body)
	if len(s) <= 0 {
		t.Error("autorest: LoggingInspector#WithInspection did not restore the Request body")
	}
}

func TestLoggingInspectorByInspecting(t *testing.T) {
	b := bytes.Buffer{}
	c := Client{}
	li := LoggingInspector{Logger: log.New(&b, "", 0)}
	c.ResponseInspector = li.ByInspecting()

	Respond(mocks.NewResponseWithContent("Content"),
		c.ByInspecting())

	if len(b.String()) <= 0 {
		t.Error("autorest: LoggingInspector#ByInspection did not record Response to the log")
	}
}

func TestLoggingInspectorByInspectingEmitsErrors(t *testing.T) {
	b := bytes.Buffer{}
	c := Client{}
	r := mocks.NewResponseWithContent("Content")
	li := LoggingInspector{Logger: log.New(&b, "", 0)}
	c.ResponseInspector = li.ByInspecting()

	r.Body.Close()
	Respond(r,
		c.ByInspecting())

	if len(b.String()) <= 0 {
		t.Error("autorest: LoggingInspector#ByInspection did not record Response to the log")
	}
}

func TestLoggingInspectorByInspectingRestoresBody(t *testing.T) {
	b := bytes.Buffer{}
	c := Client{}
	r := mocks.NewResponseWithContent("Content")
	li := LoggingInspector{Logger: log.New(&b, "", 0)}
	c.ResponseInspector = li.ByInspecting()

	Respond(r,
		c.ByInspecting())

	s, _ := ioutil.ReadAll(r.Body)
	if len(s) <= 0 {
		t.Error("autorest: LoggingInspector#ByInspecting did not restore the Response body")
	}
}

func TestNewClientWithUserAgent(t *testing.T) {
	ua := "UserAgent"
	c := NewClientWithUserAgent(ua)

	if c.UserAgent != ua {
		t.Errorf("autorest: NewClientWithUserAgent failed to set the UserAgent -- expected %s, received %s",
			ua, c.UserAgent)
	}
}

func TestClientSenderReturnsHttpClientByDefault(t *testing.T) {
	c := Client{}

	if fmt.Sprintf("%T", c.sender()) != "*http.Client" {
		t.Error("autorest: Client#sender failed to return http.Client by default")
	}
}

func TestClientSenderReturnsSetSender(t *testing.T) {
	c := Client{}

	s := mocks.NewSender()
	c.Sender = s

	if c.sender() != s {
		t.Error("autorest: Client#sender failed to return set Sender")
	}
}

func TestClientDoInvokesSender(t *testing.T) {
	c := Client{}

	s := mocks.NewSender()
	c.Sender = s

	c.Do(&http.Request{})
	if s.Attempts() != 1 {
		t.Error("autorest: Client#Do failed to invoke the Sender")
	}
}

func TestClientDoSetsUserAgent(t *testing.T) {
	ua := "UserAgent"
	c := Client{UserAgent: ua}
	r := mocks.NewRequest()

	c.Do(r)

	if r.UserAgent() != ua {
		t.Errorf("autorest: Client#Do failed to correctly set User-Agent header: %s=%s",
			http.CanonicalHeaderKey(headerUserAgent), r.UserAgent())
	}
}

func TestClientDoSetsAuthorization(t *testing.T) {
	r := mocks.NewRequest()
	s := mocks.NewSender()
	c := Client{Authorizer: mockAuthorizer{}, Sender: s}

	c.Do(r)
	if len(r.Header.Get(http.CanonicalHeaderKey(headerAuthorization))) <= 0 {
		t.Errorf("autorest: Client#Send failed to set Authorization header -- %s=%s",
			http.CanonicalHeaderKey(headerAuthorization),
			r.Header.Get(http.CanonicalHeaderKey(headerAuthorization)))
	}
}

func TestClientDoInvokesRequestInspector(t *testing.T) {
	r := mocks.NewRequest()
	s := mocks.NewSender()
	i := &mockInspector{}
	c := Client{RequestInspector: i.WithInspection(), Sender: s}

	c.Do(r)
	if !i.wasInvoked {
		t.Error("autorest: Client#Send failed to invoke the RequestInspector")
	}
}

func TestClientDoInvokesResponseInspector(t *testing.T) {
	r := mocks.NewRequest()
	s := mocks.NewSender()
	i := &mockInspector{}
	c := Client{ResponseInspector: i.ByInspecting(), Sender: s}

	c.Do(r)
	if !i.wasInvoked {
		t.Error("autorest: Client#Send failed to invoke the ResponseInspector")
	}
}

func TestClientDoReturnsErrorIfPrepareFails(t *testing.T) {
	c := Client{}
	s := mocks.NewSender()
	c.Authorizer = mockFailingAuthorizer{}
	c.Sender = s

	_, err := c.Do(&http.Request{})
	if err == nil {
		t.Errorf("autorest: Client#Do failed to return an error when Prepare failed")
	}
}

func TestClientDoDoesNotSendIfPrepareFails(t *testing.T) {
	c := Client{}
	s := mocks.NewSender()
	c.Authorizer = mockFailingAuthorizer{}
	c.Sender = s

	c.Do(&http.Request{})
	if s.Attempts() > 0 {
		t.Error("autorest: Client#Do failed to invoke the Sender")
	}
}

func TestClientAuthorizerReturnsNullAuthorizerByDefault(t *testing.T) {
	c := Client{}

	if fmt.Sprintf("%T", c.authorizer()) != "autorest.NullAuthorizer" {
		t.Error("autorest: Client#authorizer failed to return the NullAuthorizer by default")
	}
}

func TestClientAuthorizerReturnsSetAuthorizer(t *testing.T) {
	c := Client{}
	c.Authorizer = mockAuthorizer{}

	if fmt.Sprintf("%T", c.authorizer()) != "autorest.mockAuthorizer" {
		t.Error("autorest: Client#authorizer failed to return the set Authorizer")
	}
}

func TestClientWithAuthorizer(t *testing.T) {
	c := Client{}
	c.Authorizer = mockAuthorizer{}

	req, _ := Prepare(&http.Request{},
		c.WithAuthorization())

	if req.Header.Get(headerAuthorization) == "" {
		t.Error("autorest: Client#WithAuthorizer failed to return the WithAuthorizer from the active Authorizer")
	}
}

func TestClientWithInspection(t *testing.T) {
	c := Client{}
	r := &mockInspector{}
	c.RequestInspector = r.WithInspection()

	Prepare(&http.Request{},
		c.WithInspection())

	if !r.wasInvoked {
		t.Error("autorest: Client#WithInspection failed to invoke RequestInspector")
	}
}

func TestClientWithInspectionSetsDefault(t *testing.T) {
	c := Client{}

	r1 := &http.Request{}
	r2, _ := Prepare(r1,
		c.WithInspection())

	if !reflect.DeepEqual(r1, r2) {
		t.Error("autorest: Client#WithInspection failed to provide a default RequestInspector")
	}
}

func TestClientByInspecting(t *testing.T) {
	c := Client{}
	r := &mockInspector{}
	c.ResponseInspector = r.ByInspecting()

	Respond(&http.Response{},
		c.ByInspecting())

	if !r.wasInvoked {
		t.Error("autorest: Client#ByInspecting failed to invoke ResponseInspector")
	}
}

func TestClientByInspectingSetsDefault(t *testing.T) {
	c := Client{}

	r := &http.Response{}
	Respond(r,
		c.ByInspecting())

	if !reflect.DeepEqual(r, &http.Response{}) {
		t.Error("autorest: Client#ByInspecting failed to provide a default ResponseInspector")
	}
}

func randomString(n int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	s := make([]byte, n)
	for i := range s {
		s[i] = chars[r.Intn(len(chars))]
	}
	return string(s)
}
