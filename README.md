# chinaproxy
chinaproxy 是一个运行在本地的代理服务，旨在解决DNS污染和SNI阻断的问题

这是个 Demo，请勿用于生产场景和非法用途
大量代码片段来自互联网和 AI 生成，核心功能为人工编写实现，但未做足够的输入检查
```
Usage of chinaproxy:
  -addr string
        监听地址，默认0.0.0.0 (default "0.0.0.0")
  -port string
        监听端口，默认8080 (default "8080")
```

aria2 使用示例
```
aria2c --all-proxy=http://127.0.0.1:8080 -x 16 -s 16 https://huggingface.co/openai/gpt-oss-120b/resolve/main/model-00000-of-00014.safetensors?download=true 
```

输出
```
08/06 14:30:41 [NOTICE] Downloading 1 item(s)
省略大量输出内容
[#5e84ee 4.2GiB/4.3GiB(99%) CN:9 DL:38MiB]
08/06 14:33:03 [NOTICE] Download complete: /mnt/d/code/test_proxy/a0757e755eddceb5f2b25f78d9967ee1caad26bbec8aa0b1418fce084041335a

Download Results:
gid   |stat|avg speed  |path/URI
======+====+===========+=======================================================
5e84ee|OK  |    39MiB/s|/mnt/d/code/test_proxy/a0757e755eddceb5f2b25f78d9967ee1caad26bbec8aa0b1418fce084041335a

Status Legend:
(OK):download completed.

```