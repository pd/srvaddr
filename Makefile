VERSION=`cat VERSION`

default: docker

srvaddr.linux.x64: srvaddr.go VERSION
	GOOS=linux go build -ldflags='-s' -o srvaddr.linux.x64

docker: srvaddr.linux.x64
	docker build --rm --tag "philodespotos/srvaddr:$(VERSION)" --tag philodespotos/srvaddr:latest .

clean:
	rm -f srvaddr srvaddr.linux.x64
