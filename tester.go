package main

import "flag"
import "fmt"
import "math/rand"
import "os"
import "os/exec"
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-=[]{}\\|;'\":,./<>?~`"

func randomString(r *rand.Rand) string {
    n := r.Int()
    bytes := make([]byte, n)
    for i := range bytes {
        bytes[i] = letterBytes[r.Intn(len(letterBytes))]
    }
    return string(bytes)
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
    sp := launchSubProcess("./jcAssignment")

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

    webRequest := func(uri string, method int, fields *map[string]string) {
        fmt.Println("Nothing yet...")
        incrReqs()
    }

    wg.Add(*threads)
    for i := 0; i < *threads; i++ {
        go func(internalSeed int64) {
            defer func() {
                wg.Done()
            }()
            r := rand.New(rand.NewSource(internalSeed))

            for currentState != stopped {
                choice := r.Int()
                switch {
                    case choice < 10: // ~10 in 4 billion chance to stop the server.
                        fields := make(map[string]string)
                        fields["password"] = randomString(r)
                        webRequest("/hash", r.Intn(del + 1), &fields)
                    case choice < 2000000000: // ~50% chance
                }
            }

        }(rand.Int63())
    }

    sp.Wait()
    wg.Wait()
}
