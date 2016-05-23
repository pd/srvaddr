VERSION=`cat VERSION`

default: osx linux docker

srvaddr_linux_amd64: srvaddr.go VERSION
	GOOS=linux go build -ldflags='-s' -o srvaddr_linux_amd64

srvaddr_darwin_amd64: srvaddr.go VERSION
	GOOS=darwin go build -ldflags='-s' -o srvaddr_darwin_amd64

linux: srvaddr_linux_amd64
osx: srvaddr_darwin_amd64

docker: srvaddr_linux_amd64
	docker build --rm --tag "philodespotos/srvaddr:$(VERSION)" .
	docker build --rm --tag "philodespotos/srvaddr:latest" .

clean:
	rm -f srvaddr srvaddr_darwin_amd64 srvaddr_linux_amd64
