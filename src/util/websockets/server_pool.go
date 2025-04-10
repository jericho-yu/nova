package websockets

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/jericho-yu/nova/src/util/dict"

	"github.com/gorilla/websocket"
)

type (
	ServerPool struct {
		connections             *dict.AnyDict[string, *Server]
		addrToAuth              *dict.AnyDict[string, string]
		onConnectionFail        serverConnectionFailFn
		onConnectionSuccess     serverConnectionSuccessFn
		onSendMessageSuccess    serverSendMessageSuccessFn
		onSendMessageFail       serverSendMessageFailFn
		onReceiveMessageFail    serverReceiveMessageFailFn
		onReceiveMessageSuccess serverReceiveMessageSuccessFn
		onCloseCallback         serverCloseCallbackFn
	}
)

var (
	ServerPoolApp  ServerPool
	serverPoolOnce sync.Once
	serverPool     *ServerPool
)

// Once 单例化：websocket服务端
func (*ServerPool) Once(serverCallbackConfig ServerCallbackConfig) *ServerPool {
	return OnceServer(serverCallbackConfig)
}

// OnceServer 单例化：websocket服务端
//
//go:fix 推荐使用：Once方法
func OnceServer(serverCallbackConfig ServerCallbackConfig) *ServerPool {
	serverPoolOnce.Do(func() {
		serverPool = &ServerPool{
			connections:             dict.Make[string, *Server](),
			addrToAuth:              dict.Make[string, string](),
			onConnectionFail:        serverCallbackConfig.OnConnectionFail,
			onConnectionSuccess:     serverCallbackConfig.OnConnectionSuccess,
			onSendMessageSuccess:    serverCallbackConfig.OnSendMessageSuccess,
			onSendMessageFail:       serverCallbackConfig.OnSendMessageFail,
			onReceiveMessageFail:    serverCallbackConfig.OnReceiveMessageFail,
			onReceiveMessageSuccess: serverCallbackConfig.OnReceiveMessageSuccess,
			onCloseCallback:         serverCallbackConfig.OnCloseCallback,
		}
	})

	return serverPool
}

// appendConn 增加连接
func (*ServerPool) appendConn(authId *string, conn *websocket.Conn) (server *Server) {
	server = NewServer(conn)
	serverPool.addrToAuth.Set(conn.RemoteAddr().String(), *authId)
	serverPool.connections.Set(conn.RemoteAddr().String(), server)

	return
}

// removeConn 移除连接
func (*ServerPool) removeConn(addr *string) {
	serverPool.addrToAuth.RemoveByKey(*addr)
	serverPool.connections.RemoveByKey(*addr)
}

// SendMsgByAddr 发送消息：通过地址
func (my *ServerPool) SendMsgByAddr(addr *string, propMsg []byte) {
	serverPool.SendMessageByAddr(addr, propMsg)
}

// SendMessageByAddr 发送消息：通过地址
func (*ServerPool) SendMessageByAddr(addr *string, prototypeMessage []byte) {
	if server, ok := serverPool.connections.Get(*addr); ok {
		server.AsyncMessage(prototypeMessage, serverPool.onSendMessageSuccess, serverPool.onSendMessageFail)
	} else {
		if serverPool.onSendMessageFail != nil {
			serverPool.onSendMessageFail(fmt.Errorf("没有找到连接：%s", *addr))
		}
	}
}

// SendMsgByAuthId 发送消息：通过认证ID
func (my *ServerPool) SendMsgByAuthId(authId *string, propMsg []byte) {
	my.SendMsgByAddr(authId, propMsg)
}

// SendMessageByAuthId 发送消息：通过认证ID
func (*ServerPool) SendMessageByAuthId(authId *string, prototypeMessage []byte) {
	serverPool.connections.GetValuesByKeys(serverPool.addrToAuth.GetKeysByValues(*authId).ToSlice()...).Each(func(idx int, server *Server) {
		server.AsyncMessage(prototypeMessage, serverPool.onSendMessageSuccess, serverPool.onSendMessageFail)
	})
}

func (my *ServerPool) SetOnConnSuc(fn serverConnectionSuccessFn) *ServerPool {
	return my.SetOnConnectionSuccess(fn)
}

// SetOnConnectionSuccess 设置回调：当连接成功
func (*ServerPool) SetOnConnectionSuccess(onConnectionSuccess serverConnectionSuccessFn) *ServerPool {
	serverPool.onConnectionSuccess = onConnectionSuccess

	return serverPool
}

// SetOnConnFail 设置回调：当连接失败
func (my *ServerPool) SetOnConnFail(fn serverConnectionFailFn) *ServerPool {
	return my.SetOnConnectionFail(fn)
}

// SetOnConnectionFail 设置回调：当连接失败
func (*ServerPool) SetOnConnectionFail(onConnectionFail serverConnectionFailFn) *ServerPool {
	serverPool.onConnectionFail = onConnectionFail

	return serverPool
}

// SetOnSendMsgSuc 设置回调：当发送消息成功
func (my *ServerPool) SetOnSendMsgSuc(fn serverSendMessageSuccessFn) *ServerPool {
	return my.SetOnSendMessageSuccess(fn)
}

// SetOnSendMessageSuccess 设置回调：当发送消息成功
func (*ServerPool) SetOnSendMessageSuccess(onSendMessageSuccess serverSendMessageSuccessFn) *ServerPool {
	serverPool.onSendMessageSuccess = onSendMessageSuccess

	return serverPool
}

// SetOnSendMsgFail 设置回调：当发送消息失败
func (my *ServerPool) SetOnSendMsgFail(fn serverSendMessageFailFn) *ServerPool {
	return my.SetOnSendMessageFail(fn)
}

// SetOnSendMessageFail 设置回调：当发送消息失败
func (*ServerPool) SetOnSendMessageFail(onSendMessageFail serverSendMessageFailFn) *ServerPool {
	serverPool.onSendMessageFail = onSendMessageFail

	return serverPool
}

// SetOnRecMsgSuc 设置回调：当接收消息成功
func (my *ServerPool) SetOnRecMsgSuc(fn serverReceiveMessageSuccessFn) *ServerPool {
	return my.SetOnReceiveMessageSuccess(fn)
}

// SetOnReceiveMessageSuccess 设置回调：当接收消息成功
func (*ServerPool) SetOnReceiveMessageSuccess(onReceiveMessageSuccess serverReceiveMessageSuccessFn) *ServerPool {
	serverPool.onReceiveMessageSuccess = onReceiveMessageSuccess

	return serverPool
}

// SetOnRecMsgFail 设置回调：当接收消息失败
func (my *ServerPool) SetOnRecMsgFail(fn serverReceiveMessageFailFn) *ServerPool {
	return my.SetOnReceiveMessageFail(fn)
}

// SetOnReceiveMessageFail 设置回调：当接收消息失败
func (*ServerPool) SetOnReceiveMessageFail(onReceiveMessageFail serverReceiveMessageFailFn) *ServerPool {
	serverPool.onReceiveMessageFail = onReceiveMessageFail

	return serverPool
}

// SetOnClsCb 设置回调：关闭时回调
func (my *ServerPool) SetOnClsCb(fn serverCloseCallbackFn) *ServerPool {
	return my.SetOnCloseCallback(fn)
}

// SetOnCloseCallback 设置回调：关闭时回调
func (*ServerPool) SetOnCloseCallback(onCloseCallback serverCloseCallbackFn) *ServerPool {
	serverPool.onCloseCallback = onCloseCallback

	return serverPool
}

// Handle 消息处理
func (*ServerPool) Handle(
	writer http.ResponseWriter,
	req *http.Request,
	header http.Header,
	condition serverConnectionCheckFn,
) *ServerPool {
	var (
		err  error
		conn *websocket.Conn
	)

	if condition == nil {
		serverPool.onConnectionFail(errors.New("验证方法不能为空"))
		return serverPool
	}

	// 升级协议
	conn, err = upgrader.Upgrade(writer, req, header)
	if err != nil {
		if serverPool.onConnectionFail != nil {
			serverPool.onConnectionFail(err)
		}
	}

	// 验证连接
	identity, err := condition(header)
	if err != nil && serverPool.onConnectionFail != nil {
		serverPool.onConnectionFail(err)
		return serverPool
	}

	// 加入连接池
	server := serverPool.appendConn(&identity, conn)

	// 开启接收消息
	if err = server.Boot(
		serverPool.onReceiveMessageSuccess,
		serverPool.onReceiveMessageFail,
		serverPool.onSendMessageFail,
		serverPool.onCloseCallback,
	); err != nil {
		if serverPool.onConnectionFail != nil {
			serverPool.onConnectionFail(err)
		}

		server.Close()
		serverPool.removeConn(&server.addr)
		server = nil
	}

	return serverPool
}
