package main

import "context"
import "crypto/sha512"
import "encoding/base64"
import "encoding/json"
import "fmt"
import "net/http"
import "strings"
import "sync"
import "time"

type Statistics struct {
    Average   int64         `json:"average"`
    Count     int64         `json:"total"`
    Durations int64         `json:"-"`
    guard     *sync.Mutex   `json:"-"`
}

func (stats *Statistics) Increase(deltaDuration time.Duration) {
    stats.guard.Lock()
    defer func() {
        stats.guard.Unlock()
    }()
    stats.Count++
    stats.Durations += deltaDuration.Nanoseconds() / int64(time.Microsecond) // Need to convert ns to us; 1ns == 1000us
    stats.Average = stats.Durations / stats.Count
}

func sha512Base64(input string) string {
    fmt.Println("Encoding: '", input, "'")
    var hashedString [64]byte = sha512.Sum512([]byte(input))
    return base64.StdEncoding.EncodeToString(hashedString[:])
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
    // Create return string
    var request []string
    // Add the request string
    url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
    request = append(request, url)
    // Add the host
    request = append(request, fmt.Sprintf("Host: %v", r.Host))
    // Loop through headers
    for name, headers := range r.Header {
        name = strings.ToLower(name)
        for _, h := range headers {
            request = append(request, fmt.Sprintf("%v: %v", name, h))
        }
    }

    // If this is a POST, add post data
    if r.Method == "POST" || r.Method == "post" {
        r.ParseForm()
        request = append(request, "\n")
        request = append(request, r.Form.Encode())
    }
    // Return the request as a string
    return strings.Join(request, "\n")
}

func main() {
    var wg sync.WaitGroup

    hashStatistics := &Statistics{
        Count: 0,
        Durations: 0,
        Average: 0,
        guard: &sync.Mutex{},
    }

    var server *http.Server = &http.Server{
        Addr: ":28080",
        Handler: nil,
        ReadTimeout: 30 * time.Second,
        WriteTimeout: 30 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    http.HandleFunc("/hash", func(response http.ResponseWriter, request *http.Request) {
        wg.Add(1)
        defer func() {
            wg.Done()
        }()
        fmt.Println("Request: ", formatRequest(request))
        var startTime time.Time = time.Now()
        /* NOTE: We can't simply defer the duration increase here because if we
         * do, we hit something that looks like #17696. So instead we wrap the
         * function in a lambda and it magically does the correct thing then.
         * This likely has to do with the way golang is processing the deferred
         * statement and it is evalating time.Since far too early in the
         * function body.
         * Bug reference: https://github.com/golang/go/issues/17696
         */
        defer func() {
            hashStatistics.Increase(time.Since(startTime))
        }()
        method := strings.ToLower(request.Method)
        if method == "post" {
            fmt.Fprintf(response, "%s", sha512Base64(request.PostFormValue("password")))
            time.Sleep(5 * time.Second)
        } else {
            http.Error(response, "Only POST is acceptable.", 405)
        }
    })

    http.HandleFunc("/shutdown", func(response http.ResponseWriter, request *http.Request) {
        defer func() {
            server.Shutdown(context.Background())
        }()
        fmt.Println("Shutting down server.")
        fmt.Fprintf(response, "Shutting down.")
    });

    http.HandleFunc("/stats", func(response http.ResponseWriter, request *http.Request) {
        hashStatistics.guard.Lock()
        defer func() {
            hashStatistics.guard.Unlock()
        }()
        jsonResponse, _ := json.Marshal(*hashStatistics)
        response.Header().Set("content-type", "application/json; charset=utf-8")
        fmt.Fprintf(response, "%s", jsonResponse)
    })

    server.ListenAndServe()
    wg.Wait()
}
