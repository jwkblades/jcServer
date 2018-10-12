package main

import "fmt"
import "crypto/sha512"
import "encoding/base64"
import "net/http"
import "time"
import "sync"
import "context"
import "strings"

func sha512Base64(input string) string {
    var hashedString [64]byte = sha512.Sum512([]byte(input))
    return base64.StdEncoding.EncodeToString(hashedString[:])
}


func main() {
    fmt.Println(sha512Base64("angryMonkey"))
    var wg sync.WaitGroup

    var server *http.Server = &http.Server{
        Addr: ":28080",
        Handler: nil,
        ReadTimeout: 30 * time.Second,
        WriteTimeout: 30 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    http.HandleFunc("/hash", func(response http.ResponseWriter, request *http.Request) {
        wg.Add(1)
        defer wg.Done()
        method := strings.ToLower(request.Method)
        if method == "post" {
            fmt.Fprintf(response, "%s", sha512Base64(request.PostFormValue("password")))
            time.Sleep(5 * time.Second)
        } else {
            http.Error(response, "Only POST is acceptable.", 405)
        }
    })

    http.HandleFunc("/shutdown", func(response http.ResponseWriter, request *http.Request) {
        defer server.Shutdown(context.Background())
        fmt.Println("Shutting down server.")
        fmt.Fprintf(response, "Shutting down.")
    });

    server.ListenAndServe()
    wg.Wait()
}
