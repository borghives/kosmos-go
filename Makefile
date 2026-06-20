export GOPRIVATE=git.mypierian.com

.PHONY: all sync update tag clean commit stage submit


all: 
	@echo "Please specify a command: make init, make update, etc."

sync:
	git pull forge main; git pull

update:
	go get -u ; go mod tidy

tag:
	@uptag-patch
	@echo $$(getorigin)@$$(gettag) > tag

clean:
	rm -f tag

stage:
	git add .

commit:
	gca && git push

test:
	go test ./...

submit: sync update test stage commit tag 