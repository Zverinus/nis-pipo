package tests

import (
	"bytes"
	"encoding/json"
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

func TestFullScenario(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:1234@localhost:5432/pipo?sslmode=disable"
	}
	dbx, err := db.Connect(dsn)
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	defer dbx.Close()

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

	os.Setenv("JWT_SECRET", "test-secret")
	handler := transport.SetupRouter(uSvc, mSvc, pSvc, "test-secret")
	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	base := server.URL

	regBody := `{"email":"e2e@test.local","password":"secret123"}`
	resp, err := client.Post(base+"/api/auth/register", "application/json", bytes.NewBufferString(regBody))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		t.Fatalf("register status %d", resp.StatusCode)
	}
	resp.Body.Close()

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

	meetBody := `{"title":"Test Meet","description":"Desc","date_start":"2025-04-01","date_end":"2025-04-02","slot_minutes":30}`
	req, _ := http.NewRequest("POST", base+"/api/meetings", bytes.NewBufferString(meetBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("create meeting: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create meeting status %d", resp.StatusCode)
	}
	var meet struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meet); err != nil {
		t.Fatalf("meeting decode: %v", err)
	}
	resp.Body.Close()
	if meet.ID == "" {
		t.Fatal("no meeting id")
	}

	resp, err = client.Get(base + "/api/meetings/" + meet.ID)
	if err != nil {
		t.Fatalf("get meeting: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get meeting status %d", resp.StatusCode)
	}
	resp.Body.Close()

	partBody := `{"display_name":"Alice"}`
	resp, err = client.Post(base+"/api/meetings/"+meet.ID+"/participants", "application/json", bytes.NewBufferString(partBody))
	if err != nil {
		t.Fatalf("create participant: %v", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("create participant status %d", resp.StatusCode)
	}
	var partResp struct {
		ID    string `json:"id"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&partResp); err != nil {
		t.Fatalf("participant decode: %v", err)
	}
	resp.Body.Close()
	if partResp.Token == "" {
		partResp.Token = partResp.ID
	}
	if partResp.Token == "" {
		t.Fatal("no participant token")
	}

	slotsBody := `{"slot_indexes":[0,1,2]}`
	req, _ = http.NewRequest("PUT", base+"/api/meetings/"+meet.ID+"/participants/"+partResp.Token+"/slots", bytes.NewBufferString(slotsBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("set slots: %v", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("set slots status %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest("GET", base+"/api/meetings/"+meet.ID+"/results", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("get results: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get results status %d", resp.StatusCode)
	}
	var results []struct {
		SlotIndex        int      `json:"slot_index"`
		Count            int      `json:"count"`
		ParticipantNames []string `json:"participant_names"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("results decode: %v", err)
	}
	resp.Body.Close()
	if len(results) == 0 {
		t.Fatal("no results")
	}

	finalBody := `{"final_slot_index":0}`
	req, _ = http.NewRequest("PUT", base+"/api/meetings/"+meet.ID+"/finalize", bytes.NewBufferString(finalBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("finalize status %d", resp.StatusCode)
	}
	resp.Body.Close()
}
