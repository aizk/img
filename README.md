# 图片上传七牛云

在 `configInit` 函数中配置自己的相关信息。

```
func configInit() {
	Config = QiNiuConfig{
		Access: "xxx",
		Secret: "xxx",
		Bucket: "xxx",
		Domain: "xxx",
	}
}
```

下载 `go` 源码后编译 `img` 上传工具，通过 `img 图片路径上传图片`。

```go
go install

img /path/to/img
```