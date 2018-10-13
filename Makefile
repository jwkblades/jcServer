EXE=jcAssignment

.PHONY: clean

${EXE}: main.go
	go build -o $@ $^

clean:
	rm -f ${EXE}
