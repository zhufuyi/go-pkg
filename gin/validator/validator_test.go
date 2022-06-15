package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func init() {
	r := gin.Default()
	binding.Validator = Gin()

	r.POST("/hello", createHello)
	r.DELETE("/hello", deleteHello)
	r.PUT("/hello", updateHello)
	r.GET("/hello", getHello)
	r.GET("/hello/:id", getHello)
	r.GET("/hellos", getHellos)

	go func() {
		err := r.Run(":8080")
		if err != nil {
			panic(err)
		}
	}()
}

var (
	helloStr = "hello world"
	paramErr = "params is invalid"

	wantHello    = fmt.Sprintf(`"%s"`, helloStr)
	wantParamErr = fmt.Sprintf(`"%s"`, paramErr)
)

type postForm struct {
	Name  string `json:"name" form:"name" binding:"required"`
	Age   int    `json:"age" form:"age" binding:"gte=0,lte=150"`
	Email string `json:"email" form:"email" binding:"email,required"`
}

func createHello(c *gin.Context) {
	form := &postForm{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, paramErr)
		return
	}
	fmt.Printf("%+v\n", form)
	c.JSON(http.StatusOK, helloStr)
}

type deleteForm struct {
	IDS []uint64 `form:"ids" binding:"required,min=1"`
}

func deleteHello(c *gin.Context) {
	form := &deleteForm{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, paramErr)
		return
	}
	fmt.Printf("%+v\n", form)
	c.JSON(http.StatusOK, helloStr)
}

type updateForm struct {
	ID    uint64 `json:"id" form:"id" binding:"required,gt=0"`
	Age   int    `json:"age" form:"age" binding:"gte=0,lte=150"`
	Email string `json:"email" form:"email" binding:"email,required"`
}

func updateHello(c *gin.Context) {
	form := &updateForm{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, paramErr)
		return
	}
	fmt.Printf("%+v\n", form)
	c.JSON(http.StatusOK, helloStr)
}

type getForm struct {
	ID uint64 `form:"id" binding:"gt=0"`
}

func getHello(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 0)
	form := &getForm{ID: id}
	err := c.ShouldBindQuery(form)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, paramErr)
		return
	}
	fmt.Printf("%+v\n", form)
	c.JSON(http.StatusOK, helloStr)
}

type getsForm struct {
	Page int    `form:"page" binding:"gte=0"`
	Size int    `form:"size" binding:"gt=0"`
	Sort string `form:"sort" binding:"required,min=2"`
}

func getHellos(c *gin.Context) {
	form := &getsForm{}
	err := c.ShouldBindQuery(form)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, paramErr)
		return
	}
	fmt.Printf("%+v\n", form)
	c.JSON(http.StatusOK, helloStr)
}

// ------------------------------------------------------------------------------------------
// http client
// ------------------------------------------------------------------------------------------

func TestPostValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := do(http.MethodPost, "http://localhost:8080/hello", &postForm{
			Name:  "foo",
			Age:   10,
			Email: "bar@gmail.com",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantHello {
			t.Errorf("got: %s, want: %s", got, wantHello)
		}
	})

	t.Run("missing field error", func(t *testing.T) {
		got, err := do(http.MethodPost, "http://localhost:8080/hello", &postForm{
			Age:   10,
			Email: "bar@gmail.com",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})

	t.Run("field range  error", func(t *testing.T) {
		got, err := do(http.MethodPost, "http://localhost:8080/hello", &postForm{
			Name:  "foo",
			Age:   -1,
			Email: "bar@gmail.com",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})

	t.Run("email error", func(t *testing.T) {
		got, err := do(http.MethodPost, "http://localhost:8080/hello", &postForm{
			Name:  "foo",
			Age:   10,
			Email: "bar",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})
}

// ------------------------------------------------------------------------------------------

func TestDeleteValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := do(http.MethodDelete, "http://localhost:8080/hello", &deleteForm{
			IDS: []uint64{1, 2, 3},
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantHello {
			t.Errorf("got: %s, want: %s", got, wantHello)
		}
	})

	t.Run("missing field error", func(t *testing.T) {
		got, err := do(http.MethodDelete, "http://localhost:8080/hello", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})

	t.Run("ids  error", func(t *testing.T) {
		got, err := do(http.MethodDelete, "http://localhost:8080/hello", &deleteForm{IDS: []uint64{}})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})
}

// -------------------------------------------------------------------------------------------

func TestPutValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := do(http.MethodPut, "http://localhost:8080/hello", &updateForm{
			ID:    100,
			Age:   10,
			Email: "bar@gmail.com",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantHello {
			t.Errorf("got: %s, want: %s", got, wantHello)
		}
	})

	t.Run("missing field error", func(t *testing.T) {
		got, err := do(http.MethodPut, "http://localhost:8080/hello", &updateForm{
			Age:   10,
			Email: "bar@gmail.com",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})

	t.Run("email error", func(t *testing.T) {
		got, err := do(http.MethodPut, "http://localhost:8080/hello", &updateForm{
			ID:    101,
			Age:   10,
			Email: "bar",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})
}

// -------------------------------------------------------------------------------------------

func TestGetValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := do(http.MethodGet, "http://localhost:8080/hello?id=100", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantHello {
			t.Errorf("got: %s, want: %s", got, wantHello)
		}
	})

	t.Run("success2", func(t *testing.T) {
		got, err := do(http.MethodGet, "http://localhost:8080/hello/101", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantHello {
			t.Errorf("got: %s, want: %s", got, wantHello)
		}
	})

	t.Run("miss id error", func(t *testing.T) {
		got, err := do(http.MethodGet, "http://localhost:8080/hello", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})
}

// -------------------------------------------------------------------------------------------

func TestGetsValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := do(http.MethodGet, "http://localhost:8080/hellos?page=0&size=10&sort=-id", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantHello {
			t.Errorf("got: %s, want: %s", got, wantHello)
		}
	})

	t.Run("missing field error", func(t *testing.T) {
		got, err := do(http.MethodGet, "http://localhost:8080/hellos?page=0&size=10", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})

	t.Run("size error", func(t *testing.T) {
		got, err := do(http.MethodGet, "http://localhost:8080/hellos?page=0&size=0&sort=-id", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if string(got) != wantParamErr {
			t.Errorf("got: %s, want: %s", got, wantParamErr)
		}
	})
}

// ------------------------------------------------------------------------------------------

func do(method string, url string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body == nil {
		reader = nil
	} else {
		v, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(v)
	}

	method = strings.ToUpper(method)
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		req, err := http.NewRequest(method, url, reader)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)

	case http.MethodGet:
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)

	default:
		return nil, errors.New("unknown method")
	}
}
