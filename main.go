package main

import "fmt"
import "crypto/sha512"
import "encoding/base64"

func sha512Base64(input string) string {
    var hashedString [64]byte = sha512.Sum512([]byte(input))
    return base64.StdEncoding.EncodeToString(hashedString[:])
}

func main() {
    fmt.Println(sha512Base64("angryMonkey"))
}
