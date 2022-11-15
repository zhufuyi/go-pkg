package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTokenStandard(t *testing.T) {
	opt = nil
	token, err := GenerateTokenStandard()
	assert.Error(t, err)

	Init()
	token, err = GenerateTokenStandard()
	assert.NoError(t, err)
	t.Log(token)
}

func TestVerifyTokenStandard(t *testing.T) {
	opt = nil
	err := VerifyTokenStandard("token")
	assert.Error(t, err)

	Init(WithSigningKey("123456"))

	// 正常验证
	token, err := GenerateTokenStandard()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(token)
	err = VerifyTokenStandard(token)
	if err != nil {
		t.Fatal(err)
	}

	// 无效token格式
	token2 := "xxx.xxx.xxx"
	err = VerifyTokenStandard(token2)
	assert.Equal(t, err, errFormat)

	// 签名失败
	token3 := token + "xxx"
	err = VerifyTokenStandard(token3)
	assert.Equal(t, err, errSignature)

	// token已过期
	Init(
		WithSigningKey("123456"),
		WithExpire(time.Second),
	)
	token, err = GenerateTokenStandard()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)
	err = VerifyTokenStandard(token)
	assert.Equal(t, err, errExpired)
}
