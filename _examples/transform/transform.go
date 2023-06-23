package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	uuid := "dea9155e-747d-4b14-857c-5bbb29048e89"
	c := http.DefaultClient
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "http://localhost:8080/inventory", nil)
	req.Header.Set("Authorization", uuid)
	if err != nil {
		panic(err)
	}
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	r := &protobuf.DescribeInventoryResponse{}
	decoder := json.NewDecoder(res.Body)

	if err := decoder.Decode(&r); err != nil {
		panic(err)
	}

	spew.Dump(r)
}
