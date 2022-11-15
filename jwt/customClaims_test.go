package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	opt = nil
	token, err := GenerateToken("123")
	assert.Error(t, err)

	Init()
	token, err = GenerateToken("123")
	assert.NoError(t, err)
	t.Log(token)
}

func TestVerifyToken(t *testing.T) {
	opt = nil
	v, err := VerifyToken("token")
	assert.Error(t, err)

	uid := "123"
	role := "admin"

	Init(
		WithSigningKey("123456"),
		WithExpire(time.Millisecond*500),
		WithSigningMethod(HS512),
	)

	// 正常验证
	token, err := GenerateToken(uid, role)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(token)
	v, err = VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v)

	// 无效token格式
	token2 := "xxx.xxx.xxx"
	v, err = VerifyToken(token2)
	assert.Equal(t, err, errFormat)

	// 签名失败
	token3 := token + "xxx"
	v, err = VerifyToken(token3)
	assert.Equal(t, err, errSignature)

	// token已过期
	token, err = GenerateToken(uid, role)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)
	v, err = VerifyToken(token)
	assert.Equal(t, err, errExpired)
}
