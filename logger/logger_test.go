package logger

import (
	"fmt"
	"testing"
)

func printInfo() {
	defer func() {
		recover()
	}()

	Debug("this is debug")
	Info("this is info")
	Warn("this is warn")
	Error("this is error")

	type people struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	p := &people{"张三", 11}
	ps := []people{{"张三", 11}, {"李四", 12}}
	pMap := map[string]*people{"123": p, "456": p}
	Info("this is debug object", Any("object1", p), Any("object2", ps), Any("object3", pMap)) // 使用debug不能打印这一句

	Panic("this is panic")
}

func TestInit(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "terminal console debug",
			args:    args{},
			wantErr: false,
		},
		{
			name: "terminal json debug",
			args: args{[]Option{
				WithFormat("json"),
			}},
			wantErr: false,
		},
		{
			name: "terminal json warn",
			args: args{[]Option{
				WithFormat("json"), WithLevel("warn"),
			}},
			wantErr: false,
		},
		{
			name:    "file console debug",
			args:    args{[]Option{WithSave(true)}},
			wantErr: false,
		},
		{
			name: "file json debug",
			args: args{[]Option{
				WithFormat("json"),
				WithSave(
					true,
					WithFileName("my.log"),
					WithFileMaxSize(5),
					WithFileMaxBackups(5),
					WithFileMaxAge(10),
					WithFileIsCompression(true),
				),
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Init(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			printInfo()
		})
	}
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info("this is info", String("string", "hello golang"))
	}
}

func BenchmarkInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info("benchmark type int", Int("int", i))
	}
}

func BenchmarkAny(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info("benchmark type any", Any(fmt.Sprintf("object_%d", i), map[string]int{"张三": 11}))
	}
}
