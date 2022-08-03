package middleware

import (
	"fmt"
	"github.com/zhufuyi/pkg/gin/render"
	"github.com/zhufuyi/pkg/gohttp"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zhufuyi/pkg/jwt"
)

var uid = "123"

func initServer2() {
	jwt.Init()

	addr := getAddr()
	r := gin.Default()

	tokenFun := func(c *gin.Context) {
		token, _ := jwt.GenerateToken(uid)
		fmt.Println("token =", token)
		render.Success(c, token)
	}

	userFun := func(c *gin.Context) {
		render.Success(c, "hello "+uid)
	}

	r.GET("/token", tokenFun)
	r.GET("/user/:id", Auth(), userFun) // 需要鉴权

	go func() {
		err := r.Run(addr)
		if err != nil {
			panic(err)
		}
	}()
}

func TestAuth(t *testing.T) {
	initServer2()

	// 获取token
	result := &gohttp.StdResult{}
	err := gohttp.Get(result, requestAddr+"/token")
	if err != nil {
		t.Fatal(err)
	}
	token := result.Data.(string)

	// 使用访问
	authorization := fmt.Sprintf("Bearer %s", token)
	val, err := getUser(authorization)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(val)
}

func getUser(authorization string) (string, error) {
	client := &http.Client{}
	url := requestAddr + "/user/" + uid
	reqest, err := http.NewRequest("GET", url, nil)
	reqest.Header.Add("Authorization", authorization)
	if err != nil {
		return "", err
	}
	response, _ := client.Do(reqest)
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
