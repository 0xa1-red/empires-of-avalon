package inventory

import (
	"fmt"
	"strings"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/exp/slog"
)

type ResourceRegister struct {
	mx *sync.Mutex

	Name       blueprints.ResourceName
	CapFormula string
	Cap        int
	Amount     int
	Reserved   int
}

// func newResourceRegister(name common.ResourceName, amount int) (*ResourceRegister, error) {
// 	rr, ok := common.Resources[name]
// 	if !ok {
// 		return nil, fmt.Errorf("invalid resource")
// 	}

// 	return &ResourceRegister{
// 		mx:         &sync.Mutex{},
// 		Name:       rr.Name,
// 		CapFormula: rr.CapFormula,
// 		Amount:     amount,
// 		Reserved:   0,
// 	}, nil
// }

func (rr *ResourceRegister) UpdateCap(resources map[blueprints.ResourceName]*ResourceRegister, buildings map[uuid.UUID]*BuildingRegister) error {
	if rr.CapFormula == "" {
		slog.Debug("no formula defined, skipping cap update", "resource", rr.Name)
		return nil
	}

	l := lua.NewState()
	defer l.Close()

	fn := fmt.Sprintf(`
function derive(buildings, resources)
	%s
end
`, rr.CapFormula)
	slog.Debug("registering lua function", "function", fn)

	resTbl := &lua.LTable{} // nolint

	for _, resource := range resources {
		slog.Debug("setting resource table item", "resource", rr.Name, "name", resource.Name, "amount", resource.Amount)
		resTbl.RawSetString(string(resource.Name), lua.LNumber(resource.Amount))
	}

	buildTbl := &lua.LTable{} // nolint

	for _, building := range buildings {
		slog.Debug("setting building table item", "resource", rr.Name, "name", building.Name, "amount", len(building.Completed))
		buildTbl.RawSetString(strings.ToLower(string(building.Name)), lua.LNumber(len(building.Completed)))
	}

	if err := l.DoString(fn); err != nil {
		return err
	}

	err := l.CallByParam(lua.P{ // nolint
		Fn:      l.GetGlobal("derive"),
		NRet:    1,
		Protect: true,
	}, buildTbl, resTbl)
	if err != nil {
		return err
	}

	if cap, ok := l.Get(1).(lua.LNumber); ok {
		slog.Debug("setting new cap", "resource", rr.Name, "cap", int(cap))
		rr.Cap = int(cap)
		rr.Update(0)
	}

	return nil
}

func getStartingResources() map[blueprints.ResourceName]*ResourceRegister {
	registers := make(map[blueprints.ResourceName]*ResourceRegister)

	for name, resource := range registry.GetResources() {
		registers[name] = &ResourceRegister{
			mx:         &sync.Mutex{},
			Name:       name,
			Amount:     resource.StartingAmount,
			Reserved:   0,
			Cap:        resource.StartingCap,
			CapFormula: resource.CapFormula,
		}
	}

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
