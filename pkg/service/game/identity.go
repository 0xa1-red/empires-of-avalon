package game

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func GetBuildingID(buildingName string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(strings.ToLower(buildingName)))
}

func GetInventoryID(userID uuid.UUID) uuid.UUID {
	label := fmt.Sprintf("%s-inventory", userID.String())
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(label))
}
