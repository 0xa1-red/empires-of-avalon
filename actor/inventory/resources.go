package inventory

import (
	"sync"

	"github.com/0xa1-red/empires-of-avalon/common"
	"golang.org/x/exp/slog"
)

type ResourceRegister struct {
	mx *sync.Mutex

	Name     common.ResourceName
	Cap      int
	Amount   int
	Reserved int
}

func newResourceRegister(name common.ResourceName, amount int) *ResourceRegister {
	return &ResourceRegister{
		mx:       &sync.Mutex{},
		Name:     name,
		Amount:   amount,
		Reserved: 0,
	}
}

func getStartingResources() map[common.ResourceName]*ResourceRegister {
	registers := make(map[common.ResourceName]*ResourceRegister)

	for name, resource := range common.Resources {
		registers[name] = &ResourceRegister{
			mx:       &sync.Mutex{},
			Name:     name,
			Amount:   0,
			Reserved: 0,
			Cap:      resource.StartingCap,
		}
	}

	registers[common.Population] = newResourceRegister(common.Population, 5)
	registers[common.Wood] = newResourceRegister(common.Wood, 100)

	return registers
}

func (rr *ResourceRegister) Update(amount int) {
	rr.mx.Lock()
	defer rr.mx.Unlock()

	currAmount := rr.Amount + rr.Reserved

	newAmount := currAmount + amount
	if rr.Cap > 0 {
		if currAmount == rr.Cap {
			slog.Debug(
				"resource already at cap",
				"name", rr.Name,
				"amount_generated", amount,
				"amount_added", 0,
				"amount", rr.Amount,
				"reserved", rr.Reserved,
				"cap", rr.Cap,
			)
			return
		}
		if newAmount > rr.Cap {
			newAmount = rr.Cap
		}
	}

	slog.Debug(
		"resource updated",
		"name", rr.Name,
		"amount_generated", amount,
		"amount_added", newAmount-currAmount,
		"amount", currAmount,
		"reserved", rr.Reserved,
		"cap", rr.Cap,
	)

	rr.Amount = newAmount
}
