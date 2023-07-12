package dummy

import (
	"github.com/0xa1-red/empires-of-avalon/persistence/contract"
)

type Persister struct {
}

func (p *Persister) Persist(item contract.Persistable) (int, error) {
	raw, err := item.Encode()
	if err != nil {
		return 0, err
	}

	return len(raw), nil
}

func (p *Persister) Restore(key string) ([]byte, error) {
	return nil, nil
}
