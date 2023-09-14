package guardian

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

func newContextKey(name string) contextKey {
	binData := md5.Sum([]byte(name + "_ctx"))
	key := hex.EncodeToString(binData[:])
	return contextKey(key)
}

func (man *Manager) newSessionID() string {
	binData := md5.Sum([]byte(man.name + "" + fmt.Sprint(time.Now().UnixNano())))
	key := hex.EncodeToString(binData[:])
	return (key)
}

var nameSpaces = make(map[string]struct{})

func ValidateNamespace(name string) error {
	_, ok := nameSpaces[name]
	if ok {
		// json.NewEncoder(os.Stdout).Encode(nameSpaces)
		return (fmt.Errorf("namespace %s already exists", name))
	}
	// json.NewEncoder(os.Stdout).Encode(nameSpaces)
	nameSpaces[name] = struct{}{}
	return nil
}
