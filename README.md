# Hasher Server

A simple go application to hash incoming "password" requests and return the
base64 encoded result back to you.  

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

# Running

Once you have built the assignment (using `make`), you can run it by running
`./jcAssignment`. This will start up a simply web server on `localhost:80` by
default, which will listen for non-encrypted connections and respond on
`/hash`, `/stats`, and `/shutdown`.  

You may change the port that the server is being hosted on by sending
`-port=###` as a flag to `jcAssignment`.  

## `/hash`

Hash accepts POST methods and will hash the `password` field from the form
data. If the password field is empty, you will be returned the base64 encoded
sha512 of an empty string. Any other method used to access this page will
result in a 405 (method not allowed) error being returned.  

Hash sleeps for 5 seconds on all valid requests to simulate having a lot of
work to do, however if the method used to access it is not POST (and as such
would result in a 405), then hash returns immediately and does not sleep.  

## `/shutdown`

Shutdown requests the server to finish up and stop running. After a shutdown
has been requested, all new incoming requests should see that there is no
server available (connection refused), and once all outstanding requests have
been completed the server application should exit.  

## `/stats`

Stats returns a JSON object containing 2 fields: "average" and "total".
"average" contains the average number of micro seconds that a request to
`/hash` takes to complete (including the 5-second nap it takes) while "total"
is the total number of requests that have been sent to `/hash`, including those
that were invalid and as such resulted in a 405 being returned.  

# Tests

An additional executable `jcTest` is available to run pseudo-random tests (PRT)
on the assignment server.  

**NOTE** `jcTest` does require that the server is not currently running (and
that port 28080 is free) as it launches the server in a sub-process itself.  

The PRT uses a simple state machine to determine if the server is supposed to
be running at the moment or not. Based off of that, it sends data to the server
one one of the 3 available pages, using a randomized method and checks the
response that it gets back. This is primarily important for the POST method on
`/hash`, in which case we verify that the returned hash is what we expect it to
be (by running the same hash on our end and comparing them).  

The `/stats` endpoint is simply printed, because while we keep track of the
number of requests that have been sent, the PRT doesn't have an easy way to
synchronize the threads to ensure that no requests to `/hash` are made while a
`/stats` request is being made - this means that in theory the expected and
actual number of requests that have been made to the server could change in the
middle of fetching the statistics; resulting in an unexpected value and a test
failure.  

# Assumptions

I did make a few assumptions with the assignment as things went:  
1. The statistics should also track the time that is spent sleeping for the
   5-second "hard work" nap.  
2. Invalid requests to hash should return ASAP and not take the 5-second nap.  
3. People are comfortable with `make`, or at least with running it.  
4. The default go directory structure is primarily useful for creating
   libraries instead of applications.  
