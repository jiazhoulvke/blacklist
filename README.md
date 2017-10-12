# blacklist #

echo的黑名单中间件，也可以脱离echo单独使用。

### 示例 ###

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/jiazhoulvke/blacklist"
	"github.com/labstack/echo"
)

func main() {
	blacklist.BlackListFunc = func(c echo.Context) error {
		return c.JSON(200, "你的IP已被禁用")
	}
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.HTML(200, c.RealIP())
	}, blacklist.BlackList) //首页使用blacklist中间件
	e.GET("/del", func(c echo.Context) error {
		err := blacklist.Del(c.RealIP())
		msg := "删除成功"
		if err != nil {
			msg = fmt.Sprintf("删除失败: %v", err)
		}
		return c.JSON(200, msg)
	})
	e.GET("/add", func(c echo.Context) error {
		t, err := strconv.ParseInt(c.FormValue("t"), 10, 64)
		if err != nil {
      t = 0
		}
		err = blacklist.Add(c.RealIP(), t)
		msg := "添加成功"
		if err != nil {
			msg = fmt.Sprintf("添加失败: %v", err)
		}
		return c.JSON(200, msg)
	})
	e.GET("/list", func(c echo.Context) error {
		return c.JSON(200, blacklist.List())
	})
	e.Start(":6060")
}
```

- 先访问首页 http://localhost:6060/ ,可以看到你当前的IP地址
- 再访问 http://localhost:6060/add ,会把当前的IP禁用
- 再次访问首页 http://localhost:6060/ ,会提示“你的IP已被禁用”
- 访问 http://localhost:6060/list ，可以查看目前封禁的IP
- 访问 http://localhost:6060/del ,会把当前IP从黑名单中去掉，于是又可以正常访问首页了

