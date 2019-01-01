build:
	dep ensure
	go build -o gist
	./gist --version

release:
	dep ensure
	mkdir dist
	for i in darwin linux windows ; do \
    	echo $$i; \
		CGO_ENABLED=0 GOOS="$$i" GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o gist; \
		tar -czf dist/gist-"$$i"_amd64.tar.gz gist; \
		rm gist; \
	done
	shasum -a 256 dist/gist-darwin_amd64.tar.gz

clean:
	rm -rf vendor dist ./gist

git: clean
	dep ensure

scan:
	snyk test
	snyk monitor

.SILENT:
