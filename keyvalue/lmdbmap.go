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

func (m *LmdbMap) Get(key string) (res string, err error) {
	err = m.Env.View(func(txn *lmdb.Txn) error {
		b, err := txn.Get(m.Dbi, []byte(key))
		if lmdb.IsNotFound(err) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		res = string(b)
		return nil
	})

	return
}

func (m *LmdbMap) Put(key, value string) error {
	return m.update(func(txn *lmdb.Txn) error {
		if err := txn.Put(m.Dbi, []byte(key), []byte(value), 0); err != nil {
			return err
		}
		return nil
	})
}

func (m *LmdbMap) Del(key string) error {
	return m.update(func(txn *lmdb.Txn) error {
		if err := txn.Del(m.Dbi, []byte(key), nil); err != nil {
			return err
		}
		return nil
	})
}

func (m *LmdbMap) Sample(keys []string) (int, error) {
	// TODO
	err := m.Env.View(func(txn *lmdb.Txn) error {
		return nil
	})

	return 0, err
}
