package postgres

import (
	"database/sql"
	"database/sql/driver"

	"encoding/json"
	"errors"
	"fmt"
	"net"

	_ "github.com/lib/pq"
	"github.com/lib/pq/hstore"
	"gopkg.in/netaddr.v1"
)

type Hstore map[string]*string

// Value get value of Hstore
func (h Hstore) Value() (driver.Value, error) {
	hstore := hstore.Hstore{Map: map[string]sql.NullString{}}
	if len(h) == 0 {
		return nil, nil
	}

	for key, value := range h {
		var s sql.NullString
		if value != nil {
			s.String = *value
			s.Valid = true
		}
		hstore.Map[key] = s
	}
	return hstore.Value()
}

// Scan scan value into Hstore
func (h *Hstore) Scan(value interface{}) error {
	hstore := hstore.Hstore{}

	if err := hstore.Scan(value); err != nil {
		return err
	}

	if len(hstore.Map) == 0 {
		return nil
	}

	*h = Hstore{}
	for k := range hstore.Map {
		if hstore.Map[k].Valid {
			s := hstore.Map[k].String
			(*h)[k] = &s
		} else {
			(*h)[k] = nil
		}
	}

	return nil
}

// Jsonb Postgresql's JSONB data type
type Jsonb struct {
	json.RawMessage
}

// Value get value of Jsonb
func (j Jsonb) Value() (driver.Value, error) {
	if len(j.RawMessage) == 0 {
		return nil, nil
	}
	return j.MarshalJSON()
}

// Scan scan value into Jsonb
func (j *Jsonb) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, j)
}

type Inet net.IP

func (ip Inet) Value() (driver.Value, error) {
	switch len(ip) {
		case 4, 16:
			return net.IP(ip).String(), nil
		default:
			return nil, errors.New(fmt.Sprint("Failed to marshal INET value: %+v", ip))
	}
}

func (ip *Inet) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal INET value:", value))
	}

	*ip = Inet(netaddr.ParseIP(str))

	return nil
}
