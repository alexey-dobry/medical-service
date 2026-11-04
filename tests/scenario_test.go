//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"os"
	"strings"
	"testing"
	"time"
)

// 1x1 transparent PNG.
var pngPixel = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func env(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

var (
	authURL       = env("AUTH_URL", "http://localhost:8080")
	userURL       = env("USER_URL", "http://localhost:8081")
	adminEmail    = env("ADMIN_EMAIL", "admin@example.com")
	adminPassword = env("ADMIN_PASSWORD", "admin-dev-password")
)

// session is a browser-like client: cookies set by auth_service on :8080
// are sent to user_service on :8081 because cookies ignore the port.
type session struct {
	t      *testing.T
	client *http.Client
}

func newSession(t *testing.T) *session {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &session{t: t, client: &http.Client{Jar: jar, Timeout: 30 * time.Second}}
}

func (s *session) do(req *http.Request) (int, []byte) {
	s.t.Helper()
	resp, err := s.client.Do(req)
	if err != nil {
		s.t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}

func (s *session) json(method, url string, payload any) (int, []byte) {
	s.t.Helper()
	var body io.Reader
	if payload != nil {
		raw, _ := json.Marshal(payload)
		body = bytes.NewReader(raw)
	}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	return s.do(req)
}

func (s *session) multipart(url string, data any, photo []byte) (int, []byte) {
	s.t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	raw, _ := json.Marshal(data)
	_ = w.WriteField("data", string(raw))
	if photo != nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="photo"; filename="photo.png"`)
		h.Set("Content-Type", "image/png")
		part, _ := w.CreatePart(h)
		_, _ = part.Write(photo)
	}
	_ = w.Close()

	req, _ := http.NewRequest(http.MethodPost, url, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return s.do(req)
}

func (s *session) cookie(base, name string) string {
	s.t.Helper()
	req, _ := http.NewRequest(http.MethodGet, base+"/auth/refresh", nil)
	for _, c := range s.client.Jar.Cookies(req.URL) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// withToken sends a request with explicit cookies and no jar, to replay
// tokens that the jar already dropped after logout.
func withToken(t *testing.T, method, url string, cookies map[string]string) int {
	t.Helper()
	req, _ := http.NewRequest(method, url, nil)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func expect(t *testing.T, step string, want, got int, body []byte) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: want HTTP %d, got %d: %s", step, want, got, body)
	}
}

func TestPatientScenario(t *testing.T) {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	specialty := "card" + suffix[len(suffix)-6:]

	// Admin creates a doctor so that search has something to find.
	admin := newSession(t)
	code, body := admin.json(http.MethodPost, authURL+"/auth/login", map[string]string{"email": adminEmail, "password": adminPassword})
	expect(t, "admin login", http.StatusOK, code, body)

	code, body = admin.multipart(userURL+"/users/doctors", map[string]any{
		"first_name": "Anna", "middle_name": "Petrovna", "last_name": "Sidorova",
		"phone": "+79990000001", "email": "doctor" + suffix + "@example.com", "password": "doctor-password",
		"sex": "F", "birth_date": "1985-03-04", "specialty": specialty, "work_experience": "12",
		"description": "Experienced doctor", "services": []string{"consultation", "ecg"},
	}, pngPixel)
	expect(t, "create doctor", http.StatusCreated, code, body)
	var doctor struct{ UUID string }
	_ = json.Unmarshal(body, &doctor)
	t.Cleanup(func() {
		req, _ := http.NewRequest(http.MethodDelete, userURL+"/users/doctors/"+doctor.UUID, nil)
		if code, body := admin.do(req); code != http.StatusNoContent {
			t.Errorf("delete doctor: HTTP %d: %s", code, body)
		}
	})

	// 1. Registration with a profile photo.
	patient := newSession(t)
	email := "patient" + suffix + "@example.com"
	code, body = patient.multipart(userURL+"/users/register", map[string]string{
		"name": "Ivan", "surname": "Ivanov", "patronymic": "Ivanovich",
		"email": email, "phone_number": "+79990000002", "sex": "M",
		"password": "patient-password", "birth_date": "1990-01-02",
	}, pngPixel)
	expect(t, "register", http.StatusCreated, code, body)

	code, body = patient.multipart(userURL+"/users/register", map[string]string{
		"name": "Ivan", "surname": "Ivanov", "patronymic": "Ivanovich",
		"email": email, "phone_number": "+79990000002", "sex": "M",
		"password": "patient-password", "birth_date": "1990-01-02",
	}, nil)
	expect(t, "duplicate registration", http.StatusConflict, code, body)

	// 2. Login.
	code, body = patient.json(http.MethodPost, authURL+"/auth/login", map[string]string{"email": email, "password": "wrong-password"})
	expect(t, "login with wrong password", http.StatusUnauthorized, code, body)
	code, body = patient.json(http.MethodPost, authURL+"/auth/login", map[string]string{"email": email, "password": "patient-password"})
	expect(t, "login", http.StatusOK, code, body)

	// 3. Profile with photo.
	code, body = patient.json(http.MethodGet, userURL+"/users/profile/patient", nil)
	expect(t, "profile", http.StatusOK, code, body)
	var profile map[string]any
	_ = json.Unmarshal(body, &profile)
	if profile["surname"] != "Ivanov" || profile["patronymic"] != "Ivanovich" {
		t.Fatalf("profile fields mixed up: %s", body)
	}
	photoURL, _ := profile["profile_picture_url"].(string)
	if photoURL == "" {
		t.Fatalf("profile has no photo: %s", body)
	}
	resp, err := http.Get(photoURL)
	if err != nil {
		t.Fatalf("download photo: %v", err)
	}
	photo, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !bytes.Equal(photo, pngPixel) {
		t.Fatalf("presigned photo URL: HTTP %d, %d bytes", resp.StatusCode, len(photo))
	}

	// 4. Doctor search.
	code, body = patient.json(http.MethodGet, userURL+"/users/doctors/search?specialty="+specialty, nil)
	expect(t, "search", http.StatusOK, code, body)
	if !strings.Contains(string(body), doctor.UUID) {
		t.Fatalf("search did not return the doctor %s: %s", doctor.UUID, body)
	}
	code, body = patient.json(http.MethodGet, userURL+"/users/doctors/search?specialty=nonexistent"+suffix, nil)
	expect(t, "empty search", http.StatusOK, code, body)
	code, body = patient.json(http.MethodGet, userURL+"/users/doctors/"+doctor.UUID+"/details", nil)
	expect(t, "doctor details", http.StatusOK, code, body)
	code, body = patient.multipart(userURL+"/users/doctors", map[string]string{}, nil)
	expect(t, "patient cannot create doctors", http.StatusForbidden, code, body)

	// 5. Refresh rotates both tokens; the used refresh token is dead.
	oldAccess := patient.cookie(authURL, "accessToken")
	oldRefresh := patient.cookie(authURL, "refreshToken")
	code, body = patient.json(http.MethodPost, authURL+"/auth/refresh", nil)
	expect(t, "refresh", http.StatusOK, code, body)
	if patient.cookie(authURL, "accessToken") == oldAccess {
		t.Fatal("refresh did not issue a new access token")
	}
	if got := withToken(t, http.MethodPost, authURL+"/auth/refresh", map[string]string{"refreshToken": oldRefresh}); got != http.StatusUnauthorized {
		t.Fatalf("reusing a rotated refresh token: want 401, got %d", got)
	}
	if got := withToken(t, http.MethodGet, userURL+"/users/profile/patient", map[string]string{"accessToken": oldAccess}); got != http.StatusUnauthorized {
		t.Fatalf("access token replaced by refresh must be revoked: want 401, got %d", got)
	}
	code, body = patient.json(http.MethodGet, userURL+"/users/profile/patient", nil)
	expect(t, "profile after refresh", http.StatusOK, code, body)

	// 6. Logout really revokes the tokens.
	access := patient.cookie(authURL, "accessToken")
	refresh := patient.cookie(authURL, "refreshToken")
	code, body = patient.json(http.MethodPost, authURL+"/auth/logout", nil)
	expect(t, "logout", http.StatusOK, code, body)

	if got := withToken(t, http.MethodGet, userURL+"/users/profile/patient", map[string]string{"accessToken": access}); got != http.StatusUnauthorized {
		t.Fatalf("access token after logout: want 401, got %d", got)
	}
	if got := withToken(t, http.MethodPost, authURL+"/auth/refresh", map[string]string{"refreshToken": refresh}); got != http.StatusUnauthorized {
		t.Fatalf("refresh token after logout: want 401, got %d", got)
	}
}
