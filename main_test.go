package main

import (
	. "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func Test_BasicEndToEnd(t *testing.T) {
	originalArgs := os.Args

	secret := "theSecret"
	os.Args = []string{"mailigo", "-jwt-secret", secret, "-host=localhost", "-port=3000"}
	defer func() { os.Args = originalArgs }()

	go main()

	time.Sleep(time.Second)

	// success
	r, err := http.Get("http://localhost:3000/health")
	NoError(t, err)

	Equal(t, 200, r.StatusCode)

	_, err = ioutil.ReadAll(r.Body)
	NoError(t, err)
}
