kiteagent项目由ytzhan0828@qq.com创建。该项目使用Apache-2.0 license发布。
kiteagent项目是是flying-kite的子项目，是kitemanager的客户端(agent)。

暂时文档放在了 https://yt.wavesxa.com/blog/readBlog/24.html

#### 一. Flying-Kite 是类salt stack的计算资源节点管理工具。
可实现远程指令调用下发、部署执行，并作为节点agent等功能；可视为云计算的基础设施。
同时又支持普通的内网穿透应用，如http代理等。

#### 二. 特点：
```
	（1）跨云计算overlay网络管理机器节点
	（2）客户端由go语言编写，体积小巧，性能强悍；服务端由java编写，方便企业级扩展。
	（3）客户端服务端通过https（wss）加密传输指令，确保信息安全传输。
	（4）系统支持同步执行和异步执行两种指令操作。
	（5）支持linux,windows,mac系统应用。
	（6）可作为http代理。
```

#### 三.目前支持命令集

| 命令类型  |  说明 |
| ------------ | ------------ |
| cmd.rum  | 通用命令 |
| proxy.http | http代理命令 |

#### 四. 体验
详见 https://yt.wavesxa.com/blog/readBlog/27.html 位较低版本示例，最新版本应用说明待续。

kitemanager端示例图类似如下
列表
![](https://yt.wavesxa.com/blog/attachment/24-kitemanager-index-v0.7.0.jpg?id=19)
命令执行
![](https://yt.wavesxa.com/blog/attachment/24-kite-exec-v0.7.0.jpg?id=20)