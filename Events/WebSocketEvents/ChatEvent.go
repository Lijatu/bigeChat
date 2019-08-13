package WebSocketEvents

import (
	"bigeChat/ChatModels"
	"bigeChat/Code/ConstantCode"
	"bigeChat/Conf"
	"bigeChat/Service"

	"github.com/buguang01/util"

	"github.com/buguang01/bige/event"
	"github.com/buguang01/bige/threads"
)

func WsEventChatPus(et event.JsonMap, wsmd *event.WebSocketModel, runobj *threads.ThreadGo) {
	jsdata := make(event.JsonMap)
	result := ConstantCode.Timeout
	threads.Try(
		func() {
			arr := et.GetArray("ChatLi")
			for _, name := range arr {
				chatmd := ChatModels.ChatEx.GetChat(util.ToString(name))
				chatmd.PusAdd(wsmd)
			}
			result = ConstantCode.Success
		},
		nil,
		func() {
			event.WebSocketReplyMsg(wsmd, et, result, jsdata)
		},
	)
}

func WsEventChatCancelPus(et event.JsonMap, wsmd *event.WebSocketModel, runobj *threads.ThreadGo) {
	jsdata := make(event.JsonMap)
	result := ConstantCode.Timeout
	threads.Try(
		func() {
			arr := et.GetArray("ChatLi")
			for _, name := range arr {
				chatmd := ChatModels.ChatEx.GetChat(util.ToString(name))
				chatmd.PusDal(wsmd)
			}
			result = ConstantCode.Success
		},
		nil,
		func() {
			event.WebSocketReplyMsg(wsmd, et, result, jsdata)
		},
	)
}

func WsEventChatDiDa(et event.JsonMap, wsmd *event.WebSocketModel, runobj *threads.ThreadGo) {
	event.WebSocketReplyMsg(wsmd, et, ConstantCode.Success, nil)
}

func WsEventChatSendMsg(et event.JsonMap, wsmd *event.WebSocketModel, runobj *threads.ThreadGo) {
	jsdata := make(event.JsonMap)
	result := ConstantCode.Timeout
	threads.Try(
		func() {
			redis := Service.RedisEx.GetConn()
			defer redis.Close()
			msg := new(ChatModels.ChatMessage)
			msg.UserInfo = wsmd.ConInfo.(string)
			msg.MemberID = wsmd.KeyID
			msg.ChatNode = Conf.FilterChack(util.ToString(et["ChatMsg"]))
			msg.CreateTime = util.GetCurrTimeSecond()
			name := util.ToString(et["ChatName"])
			chatmd := ChatModels.ChatEx.GetChat(name)
			chatmd.AddMsg(msg)
			result = ConstantCode.Success
		},
		nil,
		func() {
			event.WebSocketReplyMsg(wsmd, et, result, jsdata)
		},
	)
}