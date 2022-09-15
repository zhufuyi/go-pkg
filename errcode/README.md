## errcode

错误码通常包括系统级错误码和服务级错误码，一共6位十进制数字组成，例如200101

| 10 | 01 | 01 |
| :------ | :------ | :------ |
| 对于http错误码，20表示服务级错误(10为系统级错误) | 服务模块代码 | 具体错误代码 |
| 对于grpc错误码，40表示服务级错误(30为系统级错误) | 服务模块代码 | 具体错误代码 |

- 错误级别占2位数：10(http)和30(grpc)表示系统级错误，20(http)和40(grpc)表示服务级错误，通常是由用户非法操作引起的。
- 服务模块占两位数：一个大型系统的服务模块通常不超过两位数，如果超过，说明这个系统该拆分了。
- 错误码占两位数：防止一个模块定制过多的错误码，后期不好维护。

<br>

### 安装

> go get -u github.com/zhufuyi/pkg/errcode

<br>

### 使用示例

### http错误码使用示例

```go
    // 定义错误码
    var ErrLogin = errcode.NewError(200101, "用户名或密码错误")

    // 请求返回
    response.Error(c, errcode.LoginErr)
```

<br>

### grpc错误码使用示例

```go
    // 定义错误码
    var ErrLogin = NewRPCStatus(400101, "用户名或密码错误")

    // 返回错误
    errcode.ErrLogin.Err()
    // 返回附带错误详情信息
    errcode.ErrLogin.Err(errcode.Any("err", err))
```