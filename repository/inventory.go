package repository

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/gamecluster"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
)

func Inventory(userID uuid.UUID) (*protobuf.DescribeInventoryResponse, error) {
	inventory := protobuf.GetInventoryGrainClient(gamecluster.GetC(), common.GetInventoryID(userID).String())

	res, err := inventory.Describe(&protobuf.DescribeInventoryRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory")
	}

	return res, err
}
