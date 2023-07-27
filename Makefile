user := alfreddobradi
app := avalond
BUILD=`date +%FT%T%z`

docker:
	docker buildx build --push \
		-o type=image \
		--platform=linux/amd64 \
		--tag ${user}/${app}:latest \
		--tag ${user}/${app}:${GAMED_VERSION} \
		--tag registry.0x42.in/${user}/${app}:latest \
		--tag registry.0x42.in/${user}/${app}:${GAMED_VERSION} .

build:
	go build -ldflags "-X github.com/0xa1-red/empires-of-avalon/version.Tag=`git describe --tags --abbrev=0` -X github.com/0xa1-red/empires-of-avalon/version.Revision=`git rev-parse HEAD` -X 'github.com/0xa1-red/empires-of-avalon/version.BuildTime=${BUILD}'" -o ./target/ ./cmd/...

lint:
	golangci-lint run