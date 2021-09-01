package dbactions

import (
	"github.com/xujiajun/nutsdb"
	"loggicat.com/publicwatcher/internal/app/pkg/util"
)

func Test(db *nutsdb.DB) {
	err := Set(db, "testkey", "testval")
	if err != nil {
		util.PrintRedFatal("failed to connect to nutsdb, err : " + err.Error())
	}
	_, err = Get(db, "testkey")
	if err != nil {
		util.PrintRedFatal("failed to connect to nutsdb, err : " + err.Error())
	}
}

func Get(db *nutsdb.DB, key string) (string, error) {
	out := ""
	err := db.View(func(tx *nutsdb.Tx) error {
		key := []byte(key)
		if e, err := tx.Get("", key); err != nil {
			if err.Error() == "key not found" {
				return nil
			}
			util.PrintRed("failed to get nutsdb value, err : " + err.Error())
			return err
		} else {
			out = string(e.Value)
			return nil
		}
	})
	if err != nil {
		return out, err
	}
	return out, nil
}

func Set(db *nutsdb.DB, key string, val string) error {
	err := db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Put("", []byte(key), []byte(val), 0); err != nil {
			util.PrintRed("failed to set nutsdb value, err : " + err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
