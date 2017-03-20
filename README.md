# httptrottler

[![GoDoc](https://godoc.org/github.com/wind85/httptrottler?status.svg)](https://godoc.org/github.com/wind85/httptrottler)
[![Build Status](https://travis-ci.org/wind85/httptrottler.svg?branch=master)](https://travis-ci.org/wind85/httptrottler)
[![Coverage Status](https://coveralls.io/repos/github/wind85/httptrottler/badge.svg?branch=master)](https://coveralls.io/github/wind85/httptrottler?branch=master)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

### httptrottler package
This is a small package the provides a request limiter implemented using golang.org/x/time/rate,
it limits request by ip address, works behind a load balancer or reverse proxy by inspecting the
"X-Forwarded-For" and "X-Real-IP" request headers. It provides on middleware and the limiter 
itself.

- New accept two parameters max and ttl, max rapresent the maximun ammount of connections and
  ttl is the measure of max request in ttl time, meaning max=5 and ttl=1xtime.Minute it will 
  allow 5 request in one minute.
- Handler is the middleware that will perform the actual throttling on the requests. Takes a limiter
  instance and a next http.Handler returns an handler.

### How to use it

Pretty simple, there is only one method to create a new parser just call:
```
  limiter := httptrottling.New(20,1*time.Minute) 
  // limiter accepts 20 requests every one minute.
```
To use the middleware do it like so:
```
  httptrottler.Handler(limeter, http.Handler)
 // in this case http.Handler has been use as a placeholder example.
```
That's pretty much it.

#### Philosophy
This software is developed following the "mantra" keep it simple, stupid or better known as
KISS. Something so simple like a cache with auto eviction should not required over engineered 
solutions. Though it provides most of the functionality needed by generic configuration files, 
and most important of all meaning full error messages.

#### Disclaimer
This software in alpha quality, don't use it in a production environment, it's not even completed.

#### Thank You Notes
Massive thanks to https://github.com/didip/tollbooth since httptrottler is a fork of it.
