package main

import "crypto/sha512"
import "encoding/base64"
import "flag"
import "fmt"
import "io/ioutil"
import "math/rand"
import "net/http"
import "net/url"
import "os"
import "os/exec"
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

func sha512Base64(input string) string {
    var hashedString [64]byte = sha512.Sum512([]byte(input))
    return base64.StdEncoding.EncodeToString(hashedString[:])
}

func launchSubProcess(program string, args ...string) *exec.Cmd {
    cmd := exec.Command(program, args...)
    cmd.Stdin = os.Stdin
    cmd.Stderr = os.Stderr
    cmd.Stdout = os.Stdout

    e := cmd.Start()
    if e != nil {
        fmt.Printf("Encountered error for %s: %v\n", program, e)
    }
    return cmd
}


func main() {
    threads := flag.Int("threads", 1, "The number of threads to run")
    seed := flag.Int("seed", int(time.Now().Unix()), "The seed for the PRT")
    flag.Parse()

    rand.Seed(int64(*seed))

    var wg sync.WaitGroup
    var requests int = 0
    var reqGuard *sync.Mutex = &sync.Mutex{}
    var currentState State = running

    sp := launchSubProcess("./jcAssignment")
    if sp == nil {
        panic("Unable to start web server. Aborting.")
    }

    incrReqs := func() {
        reqGuard.Lock()
        defer func() {
            reqGuard.Unlock()
        }()
        requests++
    }

    webRequest := func(path string, method int, fields *map[string]string) (int, string) {
        defer func() {
            incrReqs()
        }()
        uri, _ := url.ParseRequestURI("http://localhost:28080")
        uri.Path = path

        data := url.Values{}
        if fields != nil {
            for k, v := range *fields {
                data.Set(k, v)
            }
        }

        encodedData := data.Encode()
        stringData := strings.NewReader(encodedData)
        request, _ := http.NewRequest(methodFromInt(method), uri.String(), stringData)
        request.Header.Add("content-type", "application/x-www-form-urlencoded")
        request.Header.Add("content-length", strconv.Itoa(len(encodedData)))

        client := &http.Client{}
        response, err1 := client.Do(request)
        if err1 != nil {
            fmt.Fprintf(os.Stderr, "%v\n", err1)
        }

        if response != nil {
            body, err2 := ioutil.ReadAll(response.Body)
            if err2 != nil {
                fmt.Fprintf(os.Stderr, "%v\n", err2)
            }
            return response.StatusCode, string(body)
        }
        return -1, "Test error, see stderr."
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
                switch {
                    case choice < 100000: // ~100000 in 4 billion chance to stop the server.
                        status, _ := webRequest("/shutdown", r.Intn(del + 1), nil)
                        currentState = stopped
                        if status == -1 {
                            break
                        } else if status != 200 {
                            panic("Something went contrary to expected, and shutting down the server returned a non-good status!")
                        }
                    case choice < 3000100000: // ~75% chance
                        fields := make(map[string]string)
                        fields["password"] = randomString(r)
                        method := r.Intn(del + 1)
                        status, body := webRequest("/hash", method, &fields)
                        if status == -1 {
                            break
                        }

                        if method != post {
                            if status != 405 {
                                panic("Got an unexpected status from /hash with non-POST method!")
                            }
                        } else {
                            expected := sha512Base64(fields["password"])
                            if expected != body {
                                panic("Recieved unexpected hash!")
                            }
                        }
                    case choice < 4000000000:
                        method := r.Intn(del + 1)
                        status, body := webRequest("/stats", method, nil)
                        if status == -1 {
                            break
                        }
                        fmt.Println(body)
                    default:
                        continue
                }
            }

        }(rand.Int63())
    }

    wg.Wait()
    sp.Wait()
}
