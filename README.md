## wsx
websocket手脚架开放工具

## 版本公告
v1.2.7 追加了连接sessionid关键信息打印
v1.2.8 追加了pool.OfflineCtx方法,内部做了sessionId幂等。防止同用户多连接，早连接的池回调杀死新连接。

## 功能
- 同时支持messageID路由，和url路由
- 支持中间件
- 各种api，仿gin设计
- 支持用户池

## 依赖
- "github.com/gorilla/websocket"

## 声明
- "github.com/gorilla/websocket" 不支持对同一个conn并发读写。
- 本包将在vx-kf中孵化，完善后将基于MIT协议开源。

## 例子
./example

## 使用
```bash
git clone https://github.com/fwhezfwhez/wsx.git $GOPATH/src/wsx
```
