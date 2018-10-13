package main

import "context"
import "crypto/sha512"
import "encoding/base64"
import "encoding/json"
import "fmt"
import "net/http"
import "os"
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
    var hashedString [64]byte = sha512.Sum512([]byte(input))
    return base64.StdEncoding.EncodeToString(hashedString[:])
}

func main() {
    var wg sync.WaitGroup

    /*
     * *sigh* So, as it turns out there is a bug around timing (specifically how
     * durations and sleep work together, or don't) and as a result the times
     * being returned are... well, incorrect. What it comes down to is that the
     * time.Sleep appears to be getting excluded from time.Since(startTime),
     * which in turn means that all of the timing values are far smaller than
     * they should be.
     */
    go func() {
        startingTime := time.Now()
        time.Sleep(5 * time.Second)
        if time.Since(startingTime) < 4 * time.Second {
            os.Stderr.WriteString("Go timer bug #17696 (https://github.com/golang/go/issues/17696) encountered. Sleeps are discounted in timing calculations.\r\n\r\n")
        } else {
            fmt.Println("Timing seems to work properly.", time.Since(startingTime))
        }
    }()

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
