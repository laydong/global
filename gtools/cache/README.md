## [单机cache使用介绍](https://github.com/patrickmn/go-cache)

## 引入cache包

```
import "github.com/layasugar/laya/gcache"
```

#### 初始化get和set

```
Mem := gcache.New(0, time.Duration(600)*time.Second)
Mem.Get()
Mem.Set()
```