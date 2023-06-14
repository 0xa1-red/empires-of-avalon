package inventory

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/common"
)

type InsufficientResourceError struct {
	Resource common.ResourceName
}

func (e InsufficientResourceError) Error() string {
	return fmt.Sprintf("insufficient resource %s", e.Resource)
}

type InvalidResourceError struct {
	Resource common.ResourceName
}

func (e InvalidResourceError) Error() string {
	return fmt.Sprintf("invalid resource %s", e.Resource)
}
