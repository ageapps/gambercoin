
# Space separated patterns of packages to skip in list, test, format.
IGNORED_PACKAGES := vendor test pkg

V := 1 # When V is set, print commands and build progress.
Q := $(if $V,,@)

# cd into the GOPATH to workaround ./... not following symlinks
_allpackages = $(shell ( go list ./... 2>&1 1>&3 | \
    grep -v -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)) 1>&2 ) 3>&1 | \
    grep -v -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)))

# memoize allpackages, so that it's executed only once and only if used
allpackages = $(if $(__allpackages),,$(eval __allpackages := $$(_allpackages)))$(__allpackages)



n = 4 # node number
f = test.png
d = nodeA
s = test

list:
	@echo $(allpackages)


build:
	@for MOD in $(allpackages); do \
		package=$$(echo "$${MOD##*/}"); \
	 	echo "... building package $$package ..."; \
		go build  -o ./bin/$$package $$MOD; \
	done

clean:
	rm ./bin/*

.PHONY: test
test:
	$Q go test $(if $V,-v) -race ./test # install -race libs to speed up next run

vet:
	$Q go vet $(allpackages)


run:
	@sum=`expr $(n) - 1`; echo $$sum; go run --race . -UIPort=1000$(n) -nodepAddr=127.0.0.1:500$(n) -name=node$(n) -peers=127.0.0.1:500$$sum

run1:
	go run --race ./cmd/node_headless -UIPort=10000 -nodepAddr=127.0.0.1:5000 -name=nodeA

run2:
	go run --race ./cmd/node_headless -UIPort=10001 -nodepAddr=127.0.0.1:5001 -name=nodeB -peers=127.0.0.1:5000 

run3:
	go run --race ./cmd/node_headless -UIPort=10002 -nodepAddr=127.0.0.1:5002 -name=nodeC -peers=127.0.0.1:5001 

send1:
	go run --race ./cmd/client -UIPort=10000 -msg=Hello
send2:
	go run --race ./cmd/client -UIPort=10001 -msg=Hello
send3:
	go run --race ./cmd/client -UIPort=10002 -msg=Hello

send:
	go run --race ./cmd/client -UIPort=10001 -msg=Hello -Dest=$(d) -file=$(f) -request=$(h)

search:
	go run --race ./cmd/client -UIPort=1000$(n) -keywords=$(s)

serve:
	go run --race ./cmd/node_server

private:
	go run --race ./cmd/client -UIPort=10002 -msg=Hello -Dest=$(d)
	
front:	
	location=~/git/gambercoin-app; \
	current=$(shell pwd) && cd $$location && npm run build && cd $$current; \
	bash -c "rm -r web/*"; \
	cp -R $$location/dist/* ./web 

test1:
	sh test/test_1_ring.sh

test2:
	sh test/test_2_ring.sh

test3:
	sh test/test_3.sh

cchunks:
	rm ./._Chunks/*
	rm ./._Metafiles/*

lint:
	golint ./...

kill:
	kill $(lsof -nP -t -i4TCP:8080)

deps:
	dep ensure

show-deps:
	dep status -dot | dot -T png | open -f -a /Applications/Preview.app
