package httptrottle

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	l := New(20, 1*time.Minute)

	for i := 0; i < 22; i++ {
		ok := l.LimitReached("127.0.0.1")
		if i <= 19 && ok != false {
			t.Fatal("Something is wrong, i=", i)
		}
		if i >= 20 && ok == true {
			t.Log("As expected!, i=", i)
		}
	}
}

func TestGetIpAddress(t *testing.T) {
	req := httptest.NewRequest("GET", "http://12.168.0.1", nil)
	req.RemoteAddr = "12.168.0.1:4390"
	ip := getIPAdress(req, []string{"RemoteAddr"})
	if ip == "" {
		t.Fatal("RemoteAddr empty!")

	}
	if ip != "12.168.0.1" {
		t.Fatal("not local host!", ip)
	}

	req1 := httptest.NewRequest("GET", "http://12.168.0.1", nil)
	req1.Header.Add("X-Real-IP", "12.168.2.1")

	ip = getIPAdress(req1, []string{"X-Forwarded-For", "X-Real-IP"})
	if ip == "" {
		t.Fatal("forwarded, real-ip empty", ip)
	}
	if ip != "12.168.2.1" {
		t.Fatal("not the given address for custom headers!", ip)
	}

}

func TestLimitByIp(t *testing.T) {
	l := New(20, 1*time.Minute)
	r := httptest.NewRequest("GET", "http:/12.168.1", nil)
	for i := 0; i < 22; i++ {
		err := limitByIp(l, r)
		if i <= 19 && err != nil {
			t.Fatal("Something is wrong, i=", i)
		}
		if i >= 20 && err == nil {
			t.Log("As expected!, i=", i)
		}
	}
}

func TestLimitMiddleware(t *testing.T) {
	l := New(20, 1*time.Minute)
	h := Handler(l, http.NotFoundHandler())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://12.16.1.1.", nil)
	r.RemoteAddr = "12.16.1.1:8080"
	for i := 0; i < 22; i++ {
		h.ServeHTTP(w, r)
	}
}
