# See README.txt.

.PHONY: all clean

all: go

go:     protoc_middleman

clean:
	rm -f protoc_middleman *.pb.go
	rm -f *.pyc

protoc_middleman: basesocial.proto 
	protoc --go_out=. basesocial.proto 
	@touch protoc_middleman


