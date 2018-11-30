package keyvalue

import (
	"github.com/bmatsuo/lmdb-go/lmdb"
)

type LmdbMap struct {
	Env    *lmdb.Env
	Dbi    lmdb.DBI
	Locked bool
}

var _ Map = (*LmdbMap)(nil)

func (m *LmdbMap) update(fn lmdb.TxnOp) error {
	if m.Locked {
		return m.Env.UpdateLocked(fn)
	}
	return m.Env.Update(fn)
}

func (m *LmdbMap) GetN(k, v []string) ([]string, error) {
	if v == nil {
		v = make([]string, len(k))
	}

	v = v[:0]
	err := m.Env.View(func(txn *lmdb.Txn) error {
		for _, key := range k {
			b, err := txn.Get(m.Dbi, []byte(key))
			if lmdb.IsNotFound(err) {
				return ErrNotFound
			} else if err != nil {
				return err
			}
			v = append(v, string(b))
		}
		return nil
	})

	return v, err
}

func (m *LmdbMap) PutN(k, v []string) error {
	return m.update(func(txn *lmdb.Txn) error {
		for i, key := range k {
			if err := txn.Put(m.Dbi, []byte(key), []byte(v[i]), 0); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *LmdbMap) DelN(k []string) error {
	return m.update(func(txn *lmdb.Txn) error {
		for _, key := range k {
			if err := txn.Del(m.Dbi, []byte(key), nil); err != nil {
				return err
			}
		}
		return nil
	})
}
