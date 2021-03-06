package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"server/settings"
)

type Secret struct {
	User string `json:"user"`
	Pass string `json:"password"`
}

func GetAccounts() []string {
	path := settings.Path
	buf, err := ioutil.ReadFile(filepath.Join(path, "accs.db"))
	if err != nil {
		return nil
	}
	secret := []Secret{}
	var ret []string
	err = json.Unmarshal(buf, &secret)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	for _, t := range secret {
		ret = append(ret, t.User, t.Pass)
	}
	return ret
}
