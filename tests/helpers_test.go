package tests

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"project/internal/api"
	"project/internal/db"
	"project/internal/models"

	_ "modernc.org/sqlite"
)

func newTestApi(t *testing.T) (*api.Api, *sql.DB) {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { conn.Close() })

	schema, err := os.ReadFile("../init.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	base := &db.Base{Conn: conn}

	return api.Init(base), conn
}

func doRequest(t *testing.T, s *api.Api, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}

	mux := http.NewServeMux()
	s.Init(mux)

	req := httptest.NewRequest(method, path, r)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func insertRoom(t *testing.T, conn *sql.DB, room models.Room) int64 {
	query := `INSERT INTO rooms (name) VALUES (?)`
	res, err := conn.Exec(query, room.Name)
	if err != nil {
		t.Fatal(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	return id
}
