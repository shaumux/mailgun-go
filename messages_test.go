package mailgun

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/facebookgo/ensure"
	"github.com/mailgun/mailgun-go/events"
)

const (
	fromUser       = "=?utf-8?q?Katie_Brewer=2C_CFP=C2=AE?= <joe@example.com>"
	exampleSubject = "Mailgun-go Example Subject"
	exampleText    = "Testing some Mailgun awesomeness!"
	exampleHtml    = "<html><head /><body><p>Testing some <a href=\"http://google.com?q=abc&r=def&s=ghi\">Mailgun HTML awesomeness!</a> at www.kc5tja@yahoo.com</p></body></html>"

	exampleMime = `Content-Type: text/plain; charset="ascii"
Subject: Joe's Example Subject
From: Joe Example <joe@example.com>
To: BARGLEGARF <sam.falvo@rackspace.com>
Content-Transfer-Encoding: 7bit
Date: Thu, 6 Mar 2014 00:37:52 +0000

Testing some Mailgun MIME awesomeness!
`
	templateText  = "Greetings %recipient.name%!  Your reserved seat is at table %recipient.table%."
	exampleDomain = "testDomain"
	exampleAPIKey = "testAPIKey"
)

func TestGetStoredMessage(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		id, err := findStoredMessageID(mg) // somehow...
		if err != nil {
			t.Log(err)
			return
		}

		ctx := context.Background()
		// First, get our stored message.
		msg, err := mg.GetStoredMessage(ctx, id)
		ensure.Nil(t, err)

		fields := map[string]string{
			"       From": msg.From,
			"     Sender": msg.Sender,
			"    Subject": msg.Subject,
			"Attachments": fmt.Sprintf("%d", len(msg.Attachments)),
			"    Headers": fmt.Sprintf("%d", len(msg.MessageHeaders)),
		}
		for k, v := range fields {
			fmt.Printf("%13s: %s\n", k, v)
		}

		// We're done with it; now delete it.
		ensure.Nil(t, mg.DeleteStoredMessage(ctx, id))
	})
}

// Tries to locate the first stored event type, returning the associated stored message key.
func findStoredMessageID(mg Mailgun) (string, error) {
	it := mg.ListEvents(nil)

	var page []Event
	for it.Next(context.Background(), &page) {
		for _, event := range page {
			if event.GetName() == events.EventStored {
				return event.(*events.Stored).Storage.Key, nil
			}
		}
	}
	if it.Err() != nil {
		return "", it.Err()
	}
	return "", fmt.Errorf("No stored messages found.  Try changing MG_EMAIL_TO to an address that stores messages and try again.")
}

func TestSendMGPlain(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendPlain:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGPlainWithTracking(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
		m.SetTracking(true)
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendPlainWithTracking:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGPlainAt(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
		m.SetDeliveryTime(time.Now().Add(5 * time.Minute))
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendPlainAt:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGHtml(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
		m.SetHtml(exampleHtml)
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendHtml:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGTracking(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, exampleText+"Tracking!\n", toUser)
		m.SetTracking(false)
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendTracking:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGTag(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, exampleText+"Tags Galore!\n", toUser)
		m.AddTag("FooTag")
		m.AddTag("BarTag")
		m.AddTag("BlortTag")
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendTag:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGMIME(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMIMEMessage(ioutil.NopCloser(strings.NewReader(exampleMime)), toUser)
		msg, id, err := mg.Send(ctx, m)
		ensure.Nil(t, err)
		t.Log("TestSendMIME:MSG(" + msg + "),ID(" + id + ")")
	})
}

func TestSendMGBatchFailRecipients(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		m := mg.NewMessage(fromUser, exampleSubject, exampleText+"Batch\n")
		for i := 0; i < MaxNumberOfRecipients; i++ {
			m.AddRecipient("") // We expect this to indicate a failure at the API
		}
		err = m.AddRecipientAndVariables(toUser, nil)
		// In case of error the SDK didn't send the message,
		// OR the API didn't check for empty To: headers.
		ensure.NotNil(t, err)
	})
}

func TestSendMGBatchRecipientVariables(t *testing.T) {
	if reason := SkipNetworkTest(); reason != "" {
		t.Skip(reason)
	}

	spendMoney(t, func() {
		toUser := os.Getenv("MG_EMAIL_TO")
		mg, err := NewMailgunFromEnv()
		ensure.Nil(t, err)

		ctx := context.Background()
		m := mg.NewMessage(fromUser, exampleSubject, templateText)
		err = m.AddRecipientAndVariables(toUser, map[string]interface{}{
			"name":  "Joe Cool Example",
			"table": 42,
		})
		ensure.Nil(t, err)
		_, _, err = mg.Send(ctx, m)
		ensure.Nil(t, err)
	})
}

func TestSendMGOffline(t *testing.T) {
	const (
		exampleDomain  = "testDomain"
		exampleAPIKey  = "testAPIKey"
		toUser         = "test@test.com"
		exampleMessage = "Queue. Thank you"
		exampleID      = "<20111114174239.25659.5817@samples.mailgun.org>"
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ensure.DeepEqual(t, req.Method, http.MethodPost)
		ensure.DeepEqual(t, req.URL.Path, fmt.Sprintf("/%s/messages", exampleDomain))
		ensure.DeepEqual(t, req.FormValue("from"), fromUser)
		ensure.DeepEqual(t, req.FormValue("subject"), exampleSubject)
		ensure.DeepEqual(t, req.FormValue("text"), exampleText)
		ensure.DeepEqual(t, req.FormValue("to"), toUser)
		rsp := fmt.Sprintf(`{"message":"%s", "id":"%s"}`, exampleMessage, exampleID)
		fmt.Fprint(w, rsp)
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)
	ctx := context.Background()

	m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
	msg, id, err := mg.Send(ctx, m)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, msg, exampleMessage)
	ensure.DeepEqual(t, id, exampleID)
}

func TestSendMGSeparateDomain(t *testing.T) {
	const (
		exampleDomain = "testDomain"
		signingDomain = "signingDomain"

		exampleAPIKey  = "testAPIKey"
		toUser         = "test@test.com"
		exampleMessage = "Queue. Thank you"
		exampleID      = "<20111114174239.25659.5817@samples.mailgun.org>"
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ensure.DeepEqual(t, req.Method, http.MethodPost)
		ensure.DeepEqual(t, req.URL.Path, fmt.Sprintf("/%s/messages", signingDomain))
		ensure.DeepEqual(t, req.FormValue("from"), fromUser)
		ensure.DeepEqual(t, req.FormValue("subject"), exampleSubject)
		ensure.DeepEqual(t, req.FormValue("text"), exampleText)
		ensure.DeepEqual(t, req.FormValue("to"), toUser)
		rsp := fmt.Sprintf(`{"message":"%s", "id":"%s"}`, exampleMessage, exampleID)
		fmt.Fprint(w, rsp)
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)

	ctx := context.Background()
	m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
	m.AddDomain(signingDomain)

	msg, id, err := mg.Send(ctx, m)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, msg, exampleMessage)
	ensure.DeepEqual(t, id, exampleID)
}

func TestSendMGMessageVariables(t *testing.T) {
	const (
		exampleDomain       = "testDomain"
		exampleAPIKey       = "testAPIKey"
		toUser              = "test@test.com"
		exampleMessage      = "Queue. Thank you"
		exampleID           = "<20111114174239.25659.5820@samples.mailgun.org>"
		exampleStrVarKey    = "test-str-key"
		exampleStrVarVal    = "test-str-val"
		exampleBoolVarKey   = "test-bool-key"
		exampleBoolVarVal   = "false"
		exampleMapVarKey    = "test-map-key"
		exampleMapVarStrVal = `{"test":"123"}`
	)
	var (
		exampleMapVarVal = map[string]string{"test": "123"}
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ensure.DeepEqual(t, req.Method, http.MethodPost)
		ensure.DeepEqual(t, req.URL.Path, fmt.Sprintf("/%s/messages", exampleDomain))

		ensure.DeepEqual(t, req.FormValue("from"), fromUser)
		ensure.DeepEqual(t, req.FormValue("subject"), exampleSubject)
		ensure.DeepEqual(t, req.FormValue("text"), exampleText)
		ensure.DeepEqual(t, req.FormValue("to"), toUser)
		ensure.DeepEqual(t, req.FormValue("v:"+exampleMapVarKey), exampleMapVarStrVal)
		ensure.DeepEqual(t, req.FormValue("v:"+exampleBoolVarKey), exampleBoolVarVal)
		ensure.DeepEqual(t, req.FormValue("v:"+exampleStrVarKey), exampleStrVarVal)
		rsp := fmt.Sprintf(`{"message":"%s", "id":"%s"}`, exampleMessage, exampleID)
		fmt.Fprint(w, rsp)
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)

	m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
	m.AddVariable(exampleStrVarKey, exampleStrVarVal)
	m.AddVariable(exampleBoolVarKey, false)
	m.AddVariable(exampleMapVarKey, exampleMapVarVal)

	msg, id, err := mg.Send(context.Background(), m)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, msg, exampleMessage)
	ensure.DeepEqual(t, id, exampleID)
}

func TestAddRecipientsError(t *testing.T) {

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	m := mg.NewMessage(fromUser, exampleSubject, exampleText)

	for i := 0; i < 1000; i++ {
		recipient := fmt.Sprintf("recipient_%d@example.com", i)
		ensure.Nil(t, m.AddRecipient(recipient))
	}

	err := m.AddRecipient("recipient_1001@example.com")
	ensure.NotNil(t, err)
	ensure.DeepEqual(t, err.Error(), "recipient limit exceeded (max 1000)")
}

func TestAddRecipientAndVariablesError(t *testing.T) {
	var err error

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	m := mg.NewMessage(fromUser, exampleSubject, exampleText)

	for i := 0; i < 1000; i++ {
		recipient := fmt.Sprintf("recipient_%d@example.com", i)
		err = m.AddRecipientAndVariables(recipient, map[string]interface{}{"id": i})
		ensure.Nil(t, err)
	}

	err = m.AddRecipientAndVariables("recipient_1001@example.com", map[string]interface{}{"id": 1001})
	ensure.DeepEqual(t, err.Error(), "recipient limit exceeded (max 1000)")
}

func TestSendEOFError(t *testing.T) {
	const (
		exampleDomain = "testDomain"
		exampleAPIKey = "testAPIKey"
		toUser        = "test@test.com"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("")
		return
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)

	m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
	_, _, err := mg.Send(context.Background(), m)
	ensure.NotNil(t, err)
	ensure.StringContains(t, err.Error(), "remote server prematurely closed connection: Post ")
	ensure.StringContains(t, err.Error(), "/messages: EOF")
}

func TestHasRecipient(t *testing.T) {
	const (
		exampleDomain = "testDomain"
		exampleAPIKey = "testAPIKey"
		recipient     = "test@test.com"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ensure.DeepEqual(t, req.Method, http.MethodPost)
		ensure.DeepEqual(t, req.URL.Path, fmt.Sprintf("/%s/messages", exampleDomain))
		fmt.Fprint(w, `{"message":"Queued, Thank you", "id":"<20111114174239.25659.5820@samples.mailgun.org>"}`)
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)

	// No recipient
	m := mg.NewMessage(fromUser, exampleSubject, exampleText)
	_, _, err := mg.Send(context.Background(), m)
	ensure.NotNil(t, err)
	ensure.DeepEqual(t, err.Error(), "message not valid")

	// Provided Bcc
	m = mg.NewMessage(fromUser, exampleSubject, exampleText)
	m.AddBCC(recipient)
	_, _, err = mg.Send(context.Background(), m)
	ensure.Nil(t, err)

	// Provided cc
	m = mg.NewMessage(fromUser, exampleSubject, exampleText)
	m.AddCC(recipient)
	_, _, err = mg.Send(context.Background(), m)
	ensure.Nil(t, err)
}

func TestResendStored(t *testing.T) {
	const (
		exampleDomain  = "testDomain"
		exampleAPIKey  = "testAPIKey"
		toUser         = "test@test.com"
		exampleMessage = "Queue. Thank you"
		exampleID      = "<20111114174239.25659.5820@samples.mailgun.org>"
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ensure.DeepEqual(t, req.Method, http.MethodPost)
		ensure.DeepEqual(t, req.URL.Path, fmt.Sprintf("/domains/%s/messages/some-url", exampleDomain))
		ensure.DeepEqual(t, req.FormValue("to"), toUser)

		rsp := fmt.Sprintf(`{"message":"%s", "id":"%s"}`, exampleMessage, exampleID)
		fmt.Fprint(w, rsp)
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)

	msg, id, err := mg.ReSend(context.Background(), "some-url")
	ensure.NotNil(t, err)
	ensure.DeepEqual(t, err.Error(), "must provide at least one recipient")

	msg, id, err = mg.ReSend(context.Background(), "some-url", toUser)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, msg, exampleMessage)
	ensure.DeepEqual(t, id, exampleID)
}

func TestSendTLSOptions(t *testing.T) {
	const (
		exampleDomain  = "testDomain"
		exampleAPIKey  = "testAPIKey"
		toUser         = "test@test.com"
		exampleMessage = "Queue. Thank you"
		exampleID      = "<20111114174239.25659.5817@samples.mailgun.org>"
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ensure.DeepEqual(t, req.Method, http.MethodPost)
		ensure.DeepEqual(t, req.URL.Path, fmt.Sprintf("/%s/messages", exampleDomain))
		ensure.DeepEqual(t, req.FormValue("from"), fromUser)
		ensure.DeepEqual(t, req.FormValue("subject"), exampleSubject)
		ensure.DeepEqual(t, req.FormValue("text"), exampleText)
		ensure.DeepEqual(t, req.FormValue("to"), toUser)
		ensure.DeepEqual(t, req.FormValue("o:require-tls"), "true")
		ensure.DeepEqual(t, req.FormValue("o:skip-verification"), "true")
		rsp := fmt.Sprintf(`{"message":"%s", "id":"%s"}`, exampleMessage, exampleID)
		fmt.Fprint(w, rsp)
	}))
	defer srv.Close()

	mg := NewMailgun(exampleDomain, exampleAPIKey)
	mg.SetAPIBase(srv.URL)
	ctx := context.Background()

	m := mg.NewMessage(fromUser, exampleSubject, exampleText, toUser)
	m.SetRequireTLS(true)
	m.SetSkipVerification(true)

	msg, id, err := mg.Send(ctx, m)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, msg, exampleMessage)
	ensure.DeepEqual(t, id, exampleID)
}
