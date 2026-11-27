package arborist_test

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/uc-cdis/arborist/arborist"
)

// With WithUsersReadOnly(true) the user-writing endpoints answer 403; reads and grants are unchanged.
func TestUsersReadOnly(t *testing.T) {
	db, err := sqlx.Open("postgres", os.Getenv("ARBORIST_TEST_DB"))
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		t.Fatal(err)
	}
	server, err := arborist.NewServer().
		WithLogger(log.New(io.Discard, "", 0)).
		WithJWTApp(&mockJWTApp{}).
		WithDB(db).
		WithUsersReadOnly(true).
		Init()
	if err != nil {
		t.Fatal(err)
	}
	handler := server.MakeRouter(io.Discard)
	call := func(method, path, body string) int {
		w := httptest.NewRecorder()
		var reader io.Reader
		if body != "" {
			reader = bytes.NewBufferString(body)
		}
		handler.ServeHTTP(w, httptest.NewRequest(method, path, reader))
		return w.Code
	}
	assert.Equal(t, http.StatusForbidden, call("POST", "/user", `{"name": "ro-user"}`))
	assert.Equal(t, http.StatusForbidden, call("PATCH", "/user/ro-user", `{"email": "x@y.z"}`))
	assert.Equal(t, http.StatusForbidden, call("DELETE", "/user/ro-user", ""))
	assert.Equal(t, http.StatusNotFound, call("GET", "/user/ro-user", ""))
	assert.Equal(t, http.StatusOK, call("GET", "/user", ""))
}
