# leetocde题解下载器
一个leetcode题解下载器，欢迎提issue和pr

# 使用
1. 配置config.json
```json
{
    "output_dir": "./output", // 下载文件的输出目录
    "cookie": "***", // 登录leetcode的cookie
    "day": 100 // 下载最近100天的题解，可以自己设定
}
```
2. 执行
```shell
go run main.go
```

# TODO
- [ ] 利用Golang的并发能力提高下载速度
- [ ] 容器化操作，定时爬去题解，并推送到仓库中
- [ ] 增加超时重试机制，如果爬取失败，则重试