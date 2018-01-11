package text

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jrmsdev/go-jcms/lib/internal/context/appctx"
	"github.com/jrmsdev/go-jcms/lib/internal/doctype"
	"github.com/jrmsdev/go-jcms/lib/internal/doctype/testing/testeng"
)

const engTestName = "text"

func TestEngine(t *testing.T) {
	e, err := doctype.GetEngine(engTestName)
	if err != nil {
		t.Fatal(err)
	}
	if e.Type() != engTestName {
		t.Error(".Type !=", engTestName)
	}
}

func TestHandle(t *testing.T) {
	r := testeng.Handle(t, engTestName, &testeng.Query{})
	if appctx.Failed(r.Ctx) {
		t.Error("handle context should not fail:", r.Resp.Error())
	}
	status := r.Resp.Status()
	if status != http.StatusOK {
		t.Error("invalid resp status:", status)
	}
	body := strings.TrimSpace(string(r.Resp.Body()))
	if body != "testing" {
		t.Error("invalid resp body:", body)
	}
}

func TestHandleNotFound(t *testing.T) {
	r := testeng.Handle(t, engTestName, &testeng.Query{Path: "/invaliduri"})
	if !appctx.Failed(r.Ctx) {
		t.Error("handle context should fail")
	}
	status := r.Resp.Status()
	if status != http.StatusNotFound {
		t.Error("invalid resp status:", status)
	}
	errmsg := r.Resp.Error()
	if errmsg != "file not found" {
		t.Error("invalid error message:", errmsg)
	}
}
