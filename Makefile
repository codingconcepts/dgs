validate_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

release: validate_version
	- mkdir releases

	# linux
	GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o dgs ;\
	tar -zcvf ./releases/dgs_${VERSION}_linux.tar.gz ./dgs ;\

	# macos (arm)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o dgs ;\
	tar -zcvf ./releases/dgs_${VERSION}_macos_arm64.tar.gz ./dgs ;\

	# macos (amd)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o dgs ;\
	tar -zcvf ./releases/dgs_${VERSION}_macos_amd64.tar.gz ./dgs ;\

	# windows
	GOOS=windows go build -ldflags "-X main.version=${VERSION}" -o dgs ;\
	tar -zcvf ./releases/dgs_${VERSION}_windows.tar.gz ./dgs ;\

	rm ./dgs


build:
	go build dgs.go

test:
	go test ./... -v -cover

cluster:
	docker compose -f compose.yml up -d
	docker exec -it node1 cockroach init --insecure
	docker exec -it node1 cockroach sql --insecure

clean:
	cockroach sql --insecure -e "TRUNCATE pet; TRUNCATE person CASCADE;"

teardown:
	- docker ps -aq | xargs docker rm -f
	- pkill cockroach