package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomeHandler(t *testing.T) {
	deps := setupTestDependencies()
	app, err := NewAppWithDeps("", 80, deps)

	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	app.webEngine.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)
}
