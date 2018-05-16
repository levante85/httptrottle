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
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// New accept two parameters, max which is the number of maximum accepted requests
// and ttl the time frace in which max n requests are accepted.
func New(max int, ttl time.Duration) *Limiter {
	limiter := &Limiter{Max: max, Ttl: ttl, mtx: &sync.RWMutex{}}
	limiter.ContentType = "application/json"
	limiter.StatusCode = 429
	limiter.IpLookups = []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"}
	limiter.Trottler = make(map[string]*rate.Limiter)

	return limiter
}

// Limiter struct almost everything is customizable
type Limiter struct {
	ContentType string
	Max         int
	Ttl         time.Duration
	StatusCode  int
	IpLookups   []string
	Trottler    map[string]*rate.Limiter
	mtx         *sync.RWMutex
}

// LimitReached check if the given ip is allowed or not, this is a fairly low level
// interface and should not be called directly.
func (l *Limiter) LimitReached(ip string) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if _, found := l.Trottler[ip]; !found {
		l.Trottler[ip] = rate.NewLimiter(rate.Every(l.Ttl), l.Max)
	}

	return !l.Trottler[ip].Allow()
}

func getIPAdress(r *http.Request, headers []string) string {
	for _, h := range headers {
		if h == "RemoteAddr" {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			if isValidIp(strings.TrimSpace(ip)) {
				return ip
			}
		}

		addresses := strings.Split(r.Header.Get(h), ",")
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			if isValidIp(ip) {
				return ip
			}
		}

	}
	return ""
}

func isValidIp(ip string) bool {
	realIP := net.ParseIP(ip)
	isPrivate, _ := isPrivateSubnet(ip)
	if !realIP.IsGlobalUnicast() || isPrivate {
		return false
	}
	return true
}

func isPrivateSubnet(ip string) (bool, error) {
	var (
		first, second int
		err error
	)
	_, err = fmt.Sscanf(ip, "%d.%d", &first, &second)
	if err != nil {
		return false, err
	}

	if first == 10 {
		return true, nil
	}

	if first == 172 && second >= 16 && second <= 31  {
		return true, nil
	}

	if first == 192 && second == 168  {
		return true, nil
	}

	return false, nil
}

func limitByIp(l *Limiter, r *http.Request) error {
	if l.LimitReached(getIPAdress(r, l.IpLookups)) {
		return errors.New(`{"error":"Request limit reached for this ip address"}`)
	}

	return nil
}

// Handler is the actual rate limiter middleware takes a valid Limiter as first argument
// and the handler you want to rate limit.
func Handler(l *Limiter, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Rate-Limit-Limit", fmt.Sprintf("%s", l.Max))
		w.Header().Add("X-Rate-Limit-Duration", l.Ttl.String())

		err := limitByIp(l, r)
		if err != nil {
			w.Header().Add("Content-Type", l.ContentType)
			w.WriteHeader(l.StatusCode)
			w.Write([]byte(err.Error()))
			log.Printf("RPS limit reached for ip %s!\n", getIPAdress(r, l.IpLookups))
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(middle)
}
