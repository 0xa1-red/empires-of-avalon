package common

import (
	"fmt"

	"github.com/google/uuid"
)

func GetInventoryID(userID uuid.UUID) uuid.UUID {
	label := fmt.Sprintf("%s-inventory", userID.String())
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(label))
}

func GetBuildingID(buildingName string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(buildingName))
}
