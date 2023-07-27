package inventory

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
)

type InsufficientResourceError struct {
	Resource blueprints.ResourceName
}

func (e InsufficientResourceError) Error() string {
	return fmt.Sprintf("insufficient resource %s", e.Resource)
}

type InvalidResourceError struct {
	Resource blueprints.ResourceName
}

func (e InvalidResourceError) Error() string {
	return fmt.Sprintf("invalid resource %s", e.Resource)
}
