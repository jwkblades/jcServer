package main

import "crypto/sha512"
import "encoding/base64"
import "flag"
import "fmt"
import "io/ioutil"
import "math/rand"
import "net/http"
import "net/url"
//import "os"
//import "os/exec"
import "strconv"
import "strings"
import "sync"
import "time"

type State int

const (
    running State = iota
    stopped       = iota
)

const (
    get = iota
    post = iota
    put = iota
    del = iota
)

func methodFromInt(method int) string {
    switch {
    case method == get:
        return "GET"
    case method == post:
        return "POST"
    case method == put:
        return "PUT"
    case method == del:
        return "DELETE"
    default:
        return "HEAD"
    }
}


const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-=[]{}\\|;'\":,./<>?~`"

func randomString(r *rand.Rand) string {
    n := r.Intn(64) // 1MB max
    bytes := make([]byte, n)
    for i := range bytes {
        bytes[i] = letterBytes[r.Intn(len(letterBytes))]
    }
    return string(bytes)
}

//func launchSubProcess(program string, args ...string) *exec.Cmd {
//    cmd := exec.Command(program, args...)
//    cmd.Stdin = os.Stdin
//    cmd.Stderr = os.Stderr
//    cmd.Stdout = os.Stdout
//
//    e := cmd.Start()
//    if e != nil {
//        fmt.Printf("Encountered error for %s: %v\n", program, e)
//    }
//    return cmd
//}

func sha512Base64(input string) string {
    var hashedString [64]byte = sha512.Sum512([]byte(input))
    return base64.StdEncoding.EncodeToString(hashedString[:])
}

func main() {
    threads := flag.Int("threads", 1, "The number of threads to run")
    seed := flag.Int("seed", int(time.Now().Unix()), "The seed for the PRT")
    flag.Parse()

    rand.Seed(int64(*seed))
    //sp := launchSubProcess("./jcAssignment")

    var wg sync.WaitGroup
    var requests int = 0
    var reqGuard *sync.Mutex = &sync.Mutex{}
    var currentState State = running

    incrReqs := func() {
        reqGuard.Lock()
        defer func() {
            reqGuard.Unlock()
        }()
        requests++
    }

    webRequest := func(path string, method int, fields *map[string]string) (int, string) {
        fmt.Println("Nothing yet...")
        defer func() {
            incrReqs()
        }()
        uri, _ := url.ParseRequestURI("http://localhost:28080")
        uri.Path = path

        data := url.Values{}
        for k, v := range *fields {
            data.Set(k, v)
        }

        encodedData := data.Encode()
        stringData := strings.NewReader(encodedData)
        fmt.Println("Sending: ", stringData, "  --- Originally: ", encodedData)
        request, _ := http.NewRequest(methodFromInt(method), uri.String(), stringData)
        request.Header.Add("content-type", "application/x-www-form-urlencoded")
        request.Header.Add("content-length", strconv.Itoa(len(encodedData)))

        client := &http.Client{}
        response, _ := client.Do(request)
        body, _ := ioutil.ReadAll(response.Body)
        return response.StatusCode, string(body)
    }

    fmt.Printf("Starting up %d threads, initial seed: %d\r\n", *threads, *seed)
    wg.Add(*threads)
    for i := 0; i < *threads; i++ {
        go func(internalSeed int64) {
            defer func() {
                wg.Done()
            }()
            r := rand.New(rand.NewSource(internalSeed))

            for currentState != stopped {
                choice := r.Uint32()
                fmt.Println("Choice: ", choice)
                switch {
                    case choice < 10: // ~10 in 4 billion chance to stop the server.
                        status, _ := webRequest("/shutdown", r.Intn(del + 1), nil)
                        if status != 200 {
                            panic("Something went contrary to expected, and shutting down the server returned a non-good status!")
                        }
                    case choice < 3000000000: // ~75% chance
                        fields := make(map[string]string)
                        fields["password"] = randomString(r)
                        method := r.Intn(del + 1)
                        status, body := webRequest("/hash", method, &fields)
                        if method != post {
                            if status != 405 {
                                panic("Got an unexpected status from /hash with non-POST method!")
                            }
                        } else {
                            expected := sha512Base64(fields["password"])
                            fmt.Printf("Original: %s\nHash:     %s\nExpected: %s\n\n", fields["password"], body, expected)
                            if expected != body {
                                panic("Recieved unexpected hash!")
                            }
                        }
                    default:
                        continue
                }
            }

        }(rand.Int63())
    }

    //sp.Wait()
    wg.Wait()
}
