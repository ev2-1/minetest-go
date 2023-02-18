// This file contains how data is saved in the database
// For how data is represented while loaded see client_data.go
package minetest

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.

	"bytes"
	"errors"
	"log"
	"sync"
)

var (
	clientDB *sql.DB
)

var (
	DBLog *log.Logger
)

var (
	ErrUUIDFormat = errors.New("uuid not 16 bytes!")
)

var (
	stmtPlayerSet     *sql.Stmt
	stmtPlayerGetName *sql.Stmt
	stmtPlayerGetUUID *sql.Stmt

	stmtPlayerSetData    *sql.Stmt
	stmtPlayerDelallData *sql.Stmt
	stmtPlayerDelData    *sql.Stmt
	stmtPlayerGetData    *sql.Stmt
)

// PlayerPut updates a players name / adds a player
func DB_PlayerSet(uuid UUID, name string) (err error) {
	_, err = stmtPlayerSet.Exec(uuid[:], name)

	return
}

// PlayerGetByName returns the UUID of a name
func DB_PlayerGetByName(name string) (uuid UUID, err error) {
	var r *sql.Rows

	r, err = stmtPlayerGetName.Query(name)
	if err != nil {
		return UUIDNil, err
	}

	defer r.Close()

	if r.Next() {
		var uid []byte

		err = r.Scan(&uid)
		if err != nil {
			return UUIDNil, err
		}

		if len(uid) != 16 {
			return UUIDNil, ErrUUIDFormat
		}

		copy(uuid[:], uid)
	}

	return
}

// PlayerGetByUUID returns the name corosponding to a given UUID
func DB_PlayerGetByUUID(uuid UUID) (name string, err error) {
	var r *sql.Rows

	r, err = stmtPlayerGetUUID.Query(uuid[:])
	if err != nil {
		return "", err
	}

	defer r.Close()

	if r.Next() {
		err = r.Scan(uuid[:])
		if err != nil {
			return "", err
		}
	}

	return
}

// PlayerGetData returns all client data of a given UUID
func DB_PlayerGetData(uuid UUID) (cd map[string]ClientData, bytes int, err error) {
	cd = make(map[string]ClientData)
	var r *sql.Rows

	r, err = stmtPlayerGetData.Query(uuid[:])
	if err != nil {
		return
	}

	defer r.Close()

	for r.Next() {
		var data []byte
		var name string

		if err := r.Scan(&name, &data); err != nil {
			return nil, 0, err
		}

		bytes += len(data)

		cd[name] = &ClientDataSaved{data}
	}

	return cd, bytes, err
}

// PlayerSetData adds or updates some datafiled for a given UUID
func DB_PlayerSetData(uuid UUID, name string, data []byte) (err error) {
	_, err = stmtPlayerSetData.Exec(uuid[:], name, data)

	return
}

// PlayerDelData removes a data field
func DB_PlayerDelData(uuid UUID, name string) (err error) {
	_, err = stmtPlayerDelData.Exec(uuid[:], name)

	return
}

func mustPrepare(q string) (stmt *sql.Stmt) {
	stmt, err := clientDB.Prepare(q)
	if err != nil {
		DBLog.Fatalf("Can't prepare statement '%s': %s", q, err)
	}

	return
}

func initClientDataDB() (err error) {
	DBLog = log.New(log.Writer(), "[DB] ", log.Flags())

	configuredPath, _ := GetConfig("playerdb", "players.sqlite")

	clientDB, err = sql.Open("sqlite3", configuredPath)
	if err != nil {
		return err
	}

	_, err = clientDB.Exec("CREATE TABLE IF NOT EXISTS `players` ( `uuid` [16]BYTE PRIMARY KEY, `name` TEXT KEY UNIQUE );")
	if err != nil {
		return err
	}

	_, err = clientDB.Exec("CREATE TABLE IF NOT EXISTS `playerdata` ( `uuid` [16]BYTE, `name` TEXT, `data` BLOB, FOREIGN KEY(\"uuid\") REFERENCES \"players\" ON DELETE CASCADE, PRIMARY KEY(\"uuid\", \"name\") );")
	if err != nil {
		return err
	}

	stmtPlayerSet = mustPrepare("INSERT OR REPLACE INTO `players` VALUES(?, ?)")
	stmtPlayerGetName = mustPrepare("SELECT uuid FROM `players` WHERE name = ?")
	stmtPlayerGetUUID = mustPrepare("SELECT name FROM `players` WHERE uuid = ?")

	stmtPlayerDelData = mustPrepare("DELETE FROM `playerdata` WHERE (uuid = ? AND name = ?)")
	stmtPlayerDelallData = mustPrepare("DELETE FROM `playerdata` WHERE (uuid = ?)")
	stmtPlayerSetData = mustPrepare("INSERT OR REPLACE INTO `playerdata` VALUES(?, ?, ?)")
	stmtPlayerGetData = mustPrepare("SELECT name, data FROM `playerdata` WHERE (uuid = ?)")

	return nil
}

// SyncPlayerData, syncronizes all data for a given client
func SyncPlayerData(c *Client) {
	c.Log("Saving client data!")

	// Serialize:
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()

	wg := &sync.WaitGroup{}

	for k := range c.data {
		wg.Add(1)

		go func(k string) {
			defer wg.Done()

			d, ok := c.data[k].(ClientDataSerialize)
			if !ok {
				if c.data[k] == nil {
					c.Logf("Field '%s' will be deleted!", k)

					err := DB_PlayerDelData(c.UUID, k)
					if err != nil {
						c.Logf("Error deleting '%s': %s", k, err)
					}

					return
				}

				c.Logf("Type can't be saved '%s': can't be casted into ClientDataSerialize; type is '%T'\n", k, c.data[k])
				return
			}

			buf := &bytes.Buffer{}

			d.Serialize(buf)
			b := buf.Bytes()

			c.Logf("Serializing field '%s' len: %d", k, len(b))
			if err := DB_PlayerSetData(c.UUID, k, b); err != nil {
				c.Logf("Error setting field '%s': %s", k, err)
			}
		}(k)
	}

	wg.Wait()
}
