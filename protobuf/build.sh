protoc --go_out=. --go_opt=paths=source_relative --proto_path=. *.proto
protoc -I=. -I=$GOPATH/src --gograinv2_out=. *.proto

goimports -w .