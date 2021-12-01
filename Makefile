build:
	go build -o kudos

test:
	sudo docker rm -f kudos-redis
	sudo docker run --name kudos-redis -d -p 6379:6379 redis
	go test
	# sudo docker rm -f kudos-redis
