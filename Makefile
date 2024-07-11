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