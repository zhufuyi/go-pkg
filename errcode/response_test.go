package errcode

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/zhufuyi/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func runHTTPServer(isFromRPC bool) string {
	serverAddr, requestAddr := utils.GetLocalHTTPAddrPairs()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	resp := NewResponse(isFromRPC)

	r.GET("/ping", func(c *gin.Context) {
		resp.Success(c, "ping")
	})
	r.GET("/err", func(c *gin.Context) {
		isIgnore := resp.Error(c, errors.New("unknown error"))
		fmt.Println("/err", isIgnore)
	})

	if isFromRPC {
		r.GET("/err1", func(c *gin.Context) {
			isIgnore := resp.Error(c, StatusServiceUnavailable.ToRPCErr())
			fmt.Println("/err1", isIgnore)
		})
		r.GET("/err2", func(c *gin.Context) {
			isIgnore := resp.Error(c, StatusInternalServerError.ToRPCErr())
			fmt.Println("/err2", isIgnore)
		})
		r.GET("/err3", func(c *gin.Context) {
			isIgnore := resp.Error(c, StatusNotFound.Err())
			fmt.Println("/err3", isIgnore)
		})
	} else {
		r.GET("/err4", func(c *gin.Context) {
			isIgnore := resp.Error(c, InternalServerError.Err())
			fmt.Println("/err4", isIgnore)
		})
		r.GET("/err5", func(c *gin.Context) {
			isIgnore := resp.Error(c, NotFound.Err())
			fmt.Println("/err5", isIgnore)
		})
	}

	go func() {
		err := r.Run(serverAddr)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond * 200)

	return requestAddr
}

func TestRPCResponse(t *testing.T) {
	requestAddr := runHTTPServer(true)

	result, err := http.Get(requestAddr + "/ping")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err1")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err2")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err3")
	assert.NoError(t, err)
	t.Log(result.StatusCode)
}

func TestHTTPResponse(t *testing.T) {
	requestAddr := runHTTPServer(false)

	result, err := http.Get(requestAddr + "/ping")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err4")
	assert.NoError(t, err)
	t.Log(result.StatusCode)

	result, err = http.Get(requestAddr + "/err5")
	assert.NoError(t, err)
	t.Log(result.StatusCode)
}
