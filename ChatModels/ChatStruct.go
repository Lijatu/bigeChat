package ChatModels

import (
	"bigeChat/Code/ActionCode"
	"bigeChat/Code/ConstantCode"
	"bigeChat/Service"
	"context"
	"fmt"
	"time"

	"github.com/buguang01/Logger"

	"github.com/buguang01/bige/event"
	"github.com/buguang01/bige/json"
	"github.com/buguang01/bige/threads"

	"github.com/buguang01/util"
)

//频道结构
type ChatMD struct {
	ChatName   string                        //频道名字
	TypeChat   int                           //频道类型
	pusList    map[int]*event.WebSocketModel //消费用户列表
	queue      *ChatQueue                    //频道内消息
	CreateTime time.Time                     //创建时间，用来判断是不是从0开始读
	msgChan    chan *ChatMessage
	pusChan    chan PusSetMsg //加消费用户
	threadgo   *threads.ThreadGo
	UpTime     time.Time //更新时间
}

func NewChatMD(name string, ctype int) *ChatMD {
	result := new(ChatMD)
	result.ChatName = name
	result.TypeChat = ctype
	result.pusList = make(map[int]*event.WebSocketModel)
	result.queue = NewQueue()
	result.CreateTime = util.GetCurrTimeSecond()
	result.msgChan = make(chan *ChatMessage, 30)
	result.pusChan = make(chan PusSetMsg, 30)
	return result
}

//第一次拿信息
//写入信息
//广播信息

func (this *ChatMD) AutoHander(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
		case msg := <-this.pusChan:
			if msg.Ptype == 1 {
				this.pusList[msg.Conn.KeyID] = msg.Conn
				this.firstmag(msg.Conn)
			} else if msg.Ptype == 2 {
				delete(this.pusList, msg.Conn.KeyID)
			}
		case msg := <-this.msgChan:
			arr := make([]*ChatMessage, 1, 30)
			arr[0] = msg
		readall:
			for len(arr) < 30 {
				select {
				case msg := <-this.msgChan:
					arr = append(arr, msg)
				default:
					break readall
				}
			}
			for _, v := range arr {
				if this.queue.IsFull() {
					this.queue.Poll()
				}
				this.queue.Offer(v)
			}
			//通知消费者
			js := make(event.JsonMap)
			js["ACTION"] = ActionCode.Ws_Chat_Notice
			js["ACTIONCOM"] = ConstantCode.Success
			{
				jsdata := make(event.JsonMap)
				jd := make(event.JsonMap)
				jd["Msgs"] = arr
				jd["ChanName"] = this.ChatName
				jsdata["Chat"] = jd
				js["JSDATA"] = jsdata
			}
			jsbuf, _ := json.Marshal(&js)
			for k, conn := range this.pusList {
				if _, err := conn.Write(jsbuf); err != nil {
					delete(this.pusList, k)
				}
			}
		}
	}
}

func (this *ChatMD) firstmag(conn *event.WebSocketModel) {
	arr, _ := this.queue.GetArray(0)
	jsdata := make(event.JsonMap)
	jd := make(event.JsonMap)
	jd["Msgs"] = arr
	jd["ChanName"] = this.ChatName
	jsdata["Chat"] = jd
	event.WebSocketSendMsg(conn, ConstantCode.Success, jsdata)
}

func (this *ChatMD) PusAdd(conn *event.WebSocketModel) {
	this.pusChan <- PusSetMsg{conn, 1}
}

func (this *ChatMD) PusDal(conn *event.WebSocketModel) {
	this.pusChan <- PusSetMsg{conn, 2}
}

type PusSetMsg struct {
	Conn  *event.WebSocketModel
	Ptype int //1为添加；2为删除；
}

func (this *ChatMD) AddMsg(msg *ChatMessage) {
	this.msgChan <- msg
}

//放入管理器的KEY
func (this *ChatMD) GetKey() string {
	return fmt.Sprintf("%s,%d", this.ChatName, this.CreateTime.Unix())
}

//确认加入到了管理器中后，用来开启，这个数据的一些自动任务
//如果用这个方法本自来启动任务，就可以用对应的这些方法来关闭自动任务
func (this *ChatMD) RunAutoEvents() {
	if this.threadgo != nil {
		this.threadgo.CloseWait()
	}
	this.threadgo = threads.NewThreadGoBySub(Service.GoTreandEx.Ctx)
	this.threadgo.Go(this.AutoHander)
	ChatEx.GetChat(this.ChatName)

	// Player_AutoInit(this)
}

//时间到时，运行的方法,如果发出了委托，就返回true
func (this *ChatMD) UnloadRun() bool {
	// if this.player.ClientInfo == nil {
	// Logger.PInfo("auto closeing")
	// this.AutoTasks.CloseWait()
	// if util.GetCurrTimeSecond().Sub(this.UpTime) > time.Duration(Service.Sconf.MemoryConf.RunTime)*time.Second {
	if ChatEx.DelChat(this.ChatName) {
		this.threadgo.CloseWait()
		Logger.PInfo("auto closeed")
		return true
	}
	return false
	// }
}

//当服务关闭时，运行的方法，这个时候可能就不清内存了，只是关一些自动任务
func (this *ChatMD) DoneRun() {
	Logger.PInfo("DoneRun auto closeed")
}
