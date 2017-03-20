// Package httptrottle provides a simple configurable http rate limiter, limits
// are imposed by ip, given a ttl and max number of request par second in the
// constructor arguments, when the configured limit is reached further requests
// get denied and an 429 status is returned paired with an json error message.
// Both of them are customizable, by default the places where it looks for an
// ip address are the headers usually added by reverse proxy and load balancer
// as well as the RemoteAddr that is present in every http.Request.

package httptrottle

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func New(max int, ttl time.Duration) *Limiter {
	limiter := &Limiter{Max: max, Ttl: ttl, mtx: &sync.RWMutex{}}
	limiter.ContentType = "application/json"
	limiter.StatusCode = 429
	limiter.IpLookups = []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"}
	limiter.Trottler = make(map[string]*rate.Limiter)

	return limiter
}

type Limiter struct {
	ContentType string
	Max         int
	Ttl         time.Duration
	StatusCode  int
	IpLookups   []string
	Trottler    map[string]*rate.Limiter
	mtx         *sync.RWMutex
}

func (l *Limiter) LimitReached(ip string) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if _, found := l.Trottler[ip]; !found {
		l.Trottler[ip] = rate.NewLimiter(rate.Every(l.Ttl), l.Max)
	}

	return !l.Trottler[ip].Allow()
}

func getIPAdress(r *http.Request, headers []string) string {
	var ip string

	for _, h := range headers {
		addresses := strings.Split(r.Header.Get(h), ",")
		for i := len(addresses) - 1; i >= 0; i-- {
			ip = strings.TrimSpace(addresses[i])
			if isValidIp(ip) {
				return ip
			}
		}

		if h == "RemoteAddr" {
			ip = strings.Split(h, ":")[0]
			if isValidIp(ip) {
				return ip
			}
		}

	}
	return ""
}

func isValidIp(ip string) bool {
	realIP := net.ParseIP(ip)
	if !realIP.IsGlobalUnicast() || isPrivateSubnet(ip) {
		return false
	}

	return true
}

func isPrivateSubnet(ip string) bool {
	first := strings.Split(ip, ".")[0]

	if first == "10" || first == "192" || first == "176" {
		return true
	}

	return false
}

func limitByIp(l *Limiter, r *http.Request) error {
	if l.LimitReached(getIPAdress(r, l.IpLookups)) {
		return errors.New(`{"error":"Request limit reached for this ip address"}`)
	}

	return nil
}

func Handler(l *Limiter, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Rate-Limit-Limit", fmt.Sprintf("%s", l.Max))
		w.Header().Add("X-Rate-Limit-Duration", l.Ttl.String())

		err := limitByIp(l, r)
		if err != nil {
			w.Header().Add("Content-Type", l.ContentType)
			w.WriteHeader(l.StatusCode)
			w.Write([]byte(err.Error()))
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(middle)
}
