package state

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/dgraph-io/badger/v4"
)

type Storage struct {
	badgerDB *badger.DB
}

func NewStorage(appLogger badger.Logger) (*Storage, error) {
	stateDirPath, err := xdg.ConfigFile("multibase/state")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	const indexCacheSize100MB = 100 << 20

	badgerDB, err := badger.Open(
		badger.
			DefaultOptions(stateDirPath).
			WithLogger(appLogger).
			WithEncryptionKey([]byte("8c755319-fd2a-4a89-b0d9-ae7b8d26")).
			WithIndexCacheSize(indexCacheSize100MB),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &Storage{
		badgerDB: badgerDB,
	}, nil
}

func (s *Storage) Close() error {
	if err := s.badgerDB.Close(); err != nil {
		return fmt.Errorf("failed to close badger db: %w", err)
	}

	return nil
}

func (s *Storage) Save(id string, jsonData any) error {
	key := []byte(id)

	data, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	err = s.badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, data)
		if err != nil {
			return fmt.Errorf("failed to set state: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

func (s *Storage) Load(id string, destination any) (bool, error) {
	key := []byte(id)

	var data []byte

	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(badger.ErrKeyNotFound, err) {
				return nil
			}

			return fmt.Errorf("failed to get state: %w", err)
		}

		data, err = item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to copy value: %w", err)
		}

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to view state: %w", err)
	}

	if len(data) == 0 {
		return false, nil
	}

	err = json.Unmarshal(data, destination)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return true, nil
}

func (s *Storage) Delete(id string) error {
	key := []byte(id)

	err := s.badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete state: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}
