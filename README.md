# Hasher Server

A simple go application to hash incoming "password" requests and return the
base64 encoded result back to you.  

The server _only_ allows POST on `/hash`, and will respond with a base64
encoded sha512 of whatever was stored in the "password" POST form-data field.
It will ignore GET fields. Additionally, `/hash` will return a 405 (method not
allowed) if any method other than POST is used to attempt to access it.  

**NOTE** Do _not_ use this for actual password hashing, or at least if you do,
please use encrypted connections... this doesn't.  

# Building

On my system, at least, a "go build" will accomplish basic compilation of the
application and name it as `$(dirname ${PWD})`, however you could also choose
to run `go build -o main main.go` to have the resulting executable named
`main`. Further, and my preferred method, is simply to run `make`, which will
result in my expected executable name of `jcAssignment`; this is the name I
will be referring to the executable as in my automated test scripts as well as
in the rest of the documents.  
