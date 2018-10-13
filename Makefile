EXE=jcAssignment
TESTER=jcTest

.PHONY: clean all

all: ${EXE} ${TESTER}

${EXE}: main.go
	go build -o $@ $^

${TESTER}: ${EXE} tester.go
	go build -o $@ tester.go

clean:
	rm -f ${EXE} ${TESTER}
