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

func TestPrivateClassA(t *testing.T) {
	ip := "10.0.0.1"
	if ok, err := isPrivateSubnet(ip); !ok || err != nil {
		t.Fatalf("Expected private subnet for %s", ip)
	}
	ip = "10.255.255.255"
	if ok, err := isPrivateSubnet(ip); !ok || err != nil  {
		t.Fatalf("Expected private subnet for %s", ip)
	}
	ip = "9.255.255.255"
	if ok, _ := isPrivateSubnet(ip); ok {
		t.Fatalf("Unexpected private subnet for %s", ip)
	}
	ip = "11.0.0.0"
	if ok, _ := isPrivateSubnet(ip); ok {
		t.Fatalf("Unexpected private subnet for %s", ip)
	}
}

func TestPrivateClassB(t *testing.T) {
	ip := "172.16.0.0"
	if ok, err := isPrivateSubnet(ip); !ok || err != nil  {
		t.Fatalf("Expected private subnet for %s", ip)
	}
	ip = "172.31.255.255"
	if ok, err := isPrivateSubnet(ip); !ok || err != nil  {
		t.Fatalf("Expected private subnet for %s", ip)
	}
	ip = "172.15.0.0"
	if ok, _ := isPrivateSubnet(ip); ok {
		t.Fatalf("Unexpected private subnet for %s", ip)
	}
	ip = "172.32.0.0"
	if ok, _ := isPrivateSubnet(ip); ok {
		t.Fatalf("Unexpected private subnet for %s", ip)
	}
}

func TestPrivateClassC(t *testing.T) {
	ip := "192.168.0.0"
	if ok, err := isPrivateSubnet(ip); !ok || err != nil  {
		t.Fatalf("Expected private subnet for %s", ip)
	}
	ip = "192.168.255.255"
	if ok, err := isPrivateSubnet(ip); !ok || err != nil  {
		t.Fatalf("Expected private subnet for %s", ip)
	}
	ip = "192.167.255.255"
	if ok, _ :=  isPrivateSubnet(ip); ok {
		t.Fatalf("Unexpected private subnet for %s", ip)
	}
	ip = "192.1689.0.0"
	if ok, _ := isPrivateSubnet(ip); ok {
		t.Fatalf("Unexpected private subnet for %s", ip)
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
