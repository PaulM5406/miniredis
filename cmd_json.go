// Commands from https://redis.io/commands#json

package miniredis

import (
	"encoding/json"
	"reflect"

	"github.com/PaesslerAG/jsonpath"
	"github.com/alicebob/miniredis/v2/server"
)

// commandsJson handles JSON commands
func commandsJson(m *Miniredis) {
	m.srv.Register("JSON.SET", m.cmdJsonSet)
	m.srv.Register("JSON.GET", m.cmdJsonGet)
}

var msgInvalidJson = "EOF while parsing a string at line 1 column 1"

// JSON.SET
func (m *Miniredis) cmdJsonSet(c *server.Peer, cmd string, args []string) {

	// Validate arguments
	nArgs := len(args)
	if nArgs < 3 {
		c.WriteError(errWrongNumber(cmd))
		return
	} else if nArgs > 4 {
		c.WriteError(msgSyntaxError)
		return
	}

	var option string
	if nArgs == 4 {
		option = args[3]
	}
	if option != "" && option != "NX" && option != "XX" {
		c.WriteError(msgSyntaxError)
		return
	}

	key, path, value := args[0], args[1], args[2]

	// Validate that given value starts at least with a valid json
	valid_value := ""
	for i := len(value); i > 0; i-- {
		if json.Valid([]byte(value[0:i])) {
			valid_value = value[0:i]
			break
		}
	}
	if valid_value == "" {
		c.WriteError(msgInvalidJson)
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		isKey := db.exists(key)

		if option == "NX" && isKey {
			c.WriteNull()
			return
		}

		if option == "XX" && !isKey {
			c.WriteNull()
			return
		}

		if path != "$" && !isKey {
			c.WriteError(msgRootToCreateObject)
			return
		}

		if path == "$" {
			db.stringSet(key, valid_value)
			c.WriteOK()
			return
		}

		existingValue := db.stringGet(key)
		updatedValue, err := setJSONPath(path, existingValue, value)
		if err != nil {
			c.WriteError(err.Error())
			return
		}
		db.stringSet(key, updatedValue)
		c.WriteOK()
	})
}

func setJSONPath(path string, existingValue string, value string) (string, error) {
	// fmt.Println(value)
	// var valueStruct interface{}
	// err := json.Unmarshal([]byte(value), &valueStruct)
	// if err != nil {
	// 	return "", err
	// }

	// updatedValue, err := setValueAtPath(jsonData, path, value)
	// if err != nil {
	// 	return "", fmt.Errorf("Failed to set value at JSON path: %v", err)
	// }

	// updatedJSON, err := json.Marshal(updatedValue)
	// if err != nil {
	// 	return "", err
	// }

	// return string(updatedJSON), nil
	return value, nil
}

// JSON.GET
func (m *Miniredis) cmdJsonGet(c *server.Peer, cmd string, args []string) {

	// Validate arguments
	nArgs := len(args)
	if nArgs == 0 {
		c.WriteError(errWrongNumber(cmd))
		return
	}

	key := args[0]

	var paths []string
	if nArgs > 1 {
		paths = args[1:]
	}

	// Retrieve stored json value
	var value string
	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)
		if !db.exists(key) {
			c.WriteNull()
			return
		}
		value = db.stringGet(key)

		// No path given
		if nArgs == 1 {
			c.WriteInline(value)
			return
		}

		// Retrieve values at paths
		var unMarshaledValue interface{}
		json.Unmarshal([]byte(value), &unMarshaledValue)

		valuesAtPaths := make(map[string][]interface{})
		for _, path := range paths {
			values := make([]interface{}, 0)
			valuesAtPath, _ := jsonpath.Get(path, unMarshaledValue)
			if valuesAtPath != nil {
				if reflect.ValueOf(valuesAtPath).Kind() == reflect.Slice {
					values = append(values, valuesAtPath.([]interface{})...)
				} else {
					values = append(values, valuesAtPath)
				}
			}

			// fmt.Println(value, valuesAtPath, key, paths)
			valuesAtPaths[path] = values
		}

		var valueToUnMarshal interface{}
		if len(paths) == 1 {
			valueToUnMarshal = valuesAtPaths[paths[0]]
		} else {
			valueToUnMarshal = valuesAtPaths
		}
		byte_value, _ := json.Marshal(valueToUnMarshal)
		c.WriteInline(string(byte_value))

	})
}
