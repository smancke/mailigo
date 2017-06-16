package mail

import (
	. "github.com/stretchr/testify/assert"
	"io/ioutil"
	_ "os"
	"testing"
)

func Test_StoreBasicOperations(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	//defer os.RemoveAll(dir)
	println(dir)

	store := NewStore()
	NoError(t, store.Open(dir))

	m := DBMailing{
		ID:           "theId",
		TemplateName: "theTemplate",
		GlobalData:   "theGlobalData",
	}
	store.CreateMailing(m)
}
