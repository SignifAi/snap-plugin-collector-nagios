go:
				glide up
				go build

all:
				go
clean:
				rm -rf snap-plugin-publisher-signifai

test:
				go test -v $$(glide novendor)
