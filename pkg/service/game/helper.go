package game

import "github.com/0xa1-red/empires-of-avalon/pkg/model"

func GetBuildingAmount(r model.BuildRequest) int64 {
	if !queueBuildings {
		return 1
	}

	amt := int64(r.Amount)

	if amt > maximumBuildingRequest {
		return maximumBuildingRequest
	}

	return amt
}
