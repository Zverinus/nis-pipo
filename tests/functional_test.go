package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"nis-pipo/internal/db"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/participant"
	"nis-pipo/internal/repository/postgres"
	"nis-pipo/internal/transport"
	"nis-pipo/internal/user"
)

func setup(t *testing.T) (base string, client *http.Client, token string) {
	t.Helper()
	os.Setenv("JWT_SECRET", "test-secret")
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:1234@localhost:5432/pipo?sslmode=disable"
	}
	dbx, err := db.Connect(dsn)
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	t.Cleanup(func() { dbx.Close() })

	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join("..", "migrations")
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(dbx, migrationsDir); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	uRepo := postgres.NewUserRepo(dbx)
	uSvc := user.NewService(uRepo)
	mRepo := postgres.NewMeetingRepo(dbx)
	slotsRepo := postgres.NewParticipantSlotsRepo(dbx)
	mSvc := meeting.NewService(mRepo, slotsRepo)
	pRepo := postgres.NewParticipantRepo(dbx)
	pSvc := participant.NewService(pRepo, mRepo, slotsRepo)

	handler := transport.SetupRouter(uSvc, mSvc, pSvc, "test-secret")
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client = server.Client()
	base = server.URL

	regBody := `{"email":"func-test@test.local","password":"secret123"}`
	resp, err := client.Post(base+"/api/auth/register", "application/json", bytes.NewBufferString(regBody))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		t.Fatalf("register status %d", resp.StatusCode)
	}

	resp, err = client.Post(base+"/api/auth/login", "application/json", bytes.NewBufferString(regBody))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status %d", resp.StatusCode)
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("login decode: %v", err)
	}
	resp.Body.Close()
	if loginResp.Token == "" {
		t.Fatal("no token")
	}
	return base, client, loginResp.Token
}

func TestAuthRegisterLogin(t *testing.T) {
	base, client, token := setup(t)
	if token == "" {
		t.Fatal("no token")
	}

	req, _ := http.NewRequest("GET", base+"/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("auth/me status %d", resp.StatusCode)
	}
}

func TestAuth401WithoutToken(t *testing.T) {
	base, client, _ := setup(t)

	req, _ := http.NewRequest("GET", base+"/api/meetings", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}
}

func TestMeetingCRUD(t *testing.T) {
	base, client, token := setup(t)

	body := `{"title":"CRUD Meet","description":"D","date_start":"2025-04-01","date_end":"2025-04-02","slot_minutes":30}`
	req, _ := http.NewRequest("POST", base+"/api/meetings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", resp.StatusCode)
	}
	var meet struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meet); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()
	if meet.ID == "" {
		t.Fatal("no meeting id")
	}

	// list
	req, _ = http.NewRequest("GET", base+"/api/meetings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status %d", resp.StatusCode)
	}

	// get
	resp, err = client.Get(base + "/api/meetings/" + meet.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status %d", resp.StatusCode)
	}

	// update
	updBody := `{"title":"Updated","description":"D2"}`
	req, _ = http.NewRequest("PUT", base+"/api/meetings/"+meet.ID, bytes.NewBufferString(updBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status %d", resp.StatusCode)
	}

	// delete
	req, _ = http.NewRequest("DELETE", base+"/api/meetings/"+meet.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status %d", resp.StatusCode)
	}
}

func TestMeeting403Forbidden(t *testing.T) {
	base, client, token1 := setup(t)

	// create meeting as user1
	body := `{"title":"Owner Meet","description":"","date_start":"2025-04-01","date_end":"2025-04-02","slot_minutes":30}`
	req, _ := http.NewRequest("POST", base+"/api/meetings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", resp.StatusCode)
	}
	var meet struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meet); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()

	// register second user
	reg2 := `{"email":"other@test.local","password":"secret123"}`
	resp, err = client.Post(base+"/api/auth/register", "application/json", bytes.NewBufferString(reg2))
	if err != nil {
		t.Fatalf("register2: %v", err)
	}
	resp.Body.Close()
	resp, err = client.Post(base+"/api/auth/login", "application/json", bytes.NewBufferString(reg2))
	if err != nil {
		t.Fatalf("login2: %v", err)
	}
	var login2 struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&login2); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()

	// user2 tries to update user1 meeting
	updBody := `{"title":"Hack","description":""}`
	req, _ = http.NewRequest("PUT", base+"/api/meetings/"+meet.ID, bytes.NewBufferString(updBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login2.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for update, got %d", resp.StatusCode)
	}

	// user2 tries to delete user1 meeting
	req, _ = http.NewRequest("DELETE", base+"/api/meetings/"+meet.ID, nil)
	req.Header.Set("Authorization", "Bearer "+login2.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for delete, got %d", resp.StatusCode)
	}
}

func TestMeeting404NotFound(t *testing.T) {
	base, client, token := setup(t)

	req, _ := http.NewRequest("GET", base+"/api/meetings/00000000-0000-0000-0000-000000000000", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	req, _ = http.NewRequest("GET", base+"/api/meetings/00000000-0000-0000-0000-000000000000/results", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("results: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for results (non-existent meeting), got %d", resp.StatusCode)
	}
}

func TestMetrics(t *testing.T) {
	base, client, _ := setup(t)

	resp, err := client.Get(base + "/metrics")
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics status %d", resp.StatusCode)
	}
	if !bytes.Contains(body, []byte("http_requests_total")) {
		t.Fatal("metrics missing http_requests_total")
	}
}

func TestFullParticipantFlow(t *testing.T) {
	base, client, token := setup(t)

	body := `{"title":"Participant Meet","description":"","date_start":"2025-04-01","date_end":"2025-04-02","slot_minutes":30}`
	req, _ := http.NewRequest("POST", base+"/api/meetings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", resp.StatusCode)
	}
	var meet struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meet); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()

	partBody := `{"display_name":"Alice"}`
	resp, err = client.Post(base+"/api/meetings/"+meet.ID+"/participants", "application/json", bytes.NewBufferString(partBody))
	if err != nil {
		t.Fatalf("participant: %v", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("participant status %d", resp.StatusCode)
	}
	var partResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&partResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()
	if partResp.ID == "" {
		t.Fatal("no participant id")
	}

	slotsBody := `{"slot_indexes":[0,1,2]}`
	req, _ = http.NewRequest("PUT", base+"/api/meetings/"+meet.ID+"/participants/"+partResp.ID+"/slots", bytes.NewBufferString(slotsBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("slots: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("slots status %d", resp.StatusCode)
	}

	req, _ = http.NewRequest("GET", base+"/api/meetings/"+meet.ID+"/results", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("results: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("results status %d", resp.StatusCode)
	}
	var results []struct {
		SlotIndex int `json:"slot_index"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()
	if len(results) == 0 {
		t.Fatal("no results")
	}

	finalBody := `{"final_slot_index":0}`
	req, _ = http.NewRequest("PUT", base+"/api/meetings/"+meet.ID+"/finalize", bytes.NewBufferString(finalBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("finalize status %d", resp.StatusCode)
	}
}
