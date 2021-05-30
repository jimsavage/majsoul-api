package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// 操作枚举
const (
	E_PlayOperation_None         = iota
	E_PlayOperation_Discard      // 弃牌
	E_PlayOperation_Chi          // 吃
	E_PlayOperation_Pon          // 碰
	E_PlayOperation_Ankan        // 暗杠
	E_PlayOperation_Minkan       // 明杠
	E_PlayOperation_Kakan        // 加杠
	E_PlayOperation_RiiChi       // 立直
	E_PlayOperation_Tsumo        // 自摸
	E_PlayOperation_Ron          // 和
	E_PlayOperation_Kuku         // 九种九牌
	E_PlayOperation_Kita         // 拔北
	E_PlayOperation_HuanSanZhang // 换三张
	E_PlayOperation_DingQue      // 定缺
)

// 流局类型
const (
	E_LiuJu_None         = iota
	E_LiuJu_Kuku         // 九种九牌
	E_LiuJu_SiFengLianDa // 四风连打
	E_LiuJu_SiGangSanLe  // 四杠散了
	E_LiuJu_SiJiaLiZhi   // 四家立直
	E_LiuJu_SanJiaHuLe   // 三家和了
)

// 对局和牌类型
const (
	E_Round_Result_LiuJu        = iota // 流局
	E_Round_Result_ShaoJi              // ? 烧鸡 ?
	E_Round_Result_Tsumo               // 自摸
	E_Round_Result_Ron                 // 和
	E_Round_Result_Chong               // 放铳
	E_Round_Result_AnotherTsumo        // 被自摸
)

// 和牌类型
const (
	E_Hu_Type_Ron       = iota // 和
	E_Hu_Type_Tsumo            // 自摸
	E_Hu_Type_QiangGang        // 抢杠
)

// 打点类型
const (
	E_Dadian_Title_None       = iota
	E_Dadian_Title_ManGuan         // 满贯
	E_Dadian_Title_TiaoMan         // 跳满
	E_Dadian_Title_BeiMan          // 倍满
	E_Dadian_Title_SanBeiMan       // 三倍满
	E_Dadian_Title_YiMan           // 役满
	E_Dadian_Title_YiMan2          // 两倍役满
	E_Dadian_Title_YiMan3          // 三倍役满
	E_Dadian_Title_YiMan4          // 四倍役满
	E_Dadian_Title_YiMan5          // 五倍役满
	E_Dadian_Title_YiMan6          // 六倍役满
	E_Dadian_Title_LeiJiYiMan      // 累计役满
	E_Dadian_Title_LiuMan     = -1 // 流局满贯
)

type Waits map[int]int
type Improves map[int]Waits

type Hand13AnalysisResult struct {
	// 原手牌
	Tiles34 []int

	// 剩余牌
	LeftTiles34 []int

	// 是否已鸣牌（非门清状态）
	// 用于判断是否无役等
	IsNaki bool

	// 向听数
	Shanten int

	// 进张
	// 考虑了剩余枚数
	// 若某个进张牌 4 枚都可见，则该进张的 value 值为 0
	Waits Waits

	// 默听时的进张
	DamaWaits Waits

	// TODO: 鸣牌进张：他家打出这张牌，可以鸣牌，且能让向听数前进
	//MeldWaits Waits

	// map[进张牌]向听前进后的(最大)进张数
	NextShantenWaitsCountMap map[int]int

	// 向听前进后的(最大)进张数的加权均值
	AvgNextShantenWaitsCount float64

	// 综合了进张与向听前进后进张的评分
	MixedWaitsScore float64

	// 改良：摸到这张牌虽不能让向听数前进，但可以让进张变多
	// len(Improves) 即为改良的牌的种数
	Improves Improves

	// 改良情况数，这里计算的是有多少种使进张增加的摸牌-切牌方式
	ImproveWayCount int

	// 摸到非进张牌时的进张数的加权均值（非改良+改良。对于非改良牌，其进张数为 Waits.AllCount()）
	// 这里只考虑一巡的改良均值
	// TODO: 在考虑改良的情况下，如何计算向听前进所需要的摸牌次数的期望值？蒙特卡罗方法？
	AvgImproveWaitsCount float64

	// 听牌时的手牌和率
	// TODO: 未听牌时的和率？
	AvgAgariRate float64

	// 振听可能率（一向听和听牌时）
	FuritenRate float64

	// 役种
	YakuTypes map[int]struct{}

	// （鸣牌时）是否片听
	IsPartWait bool

	// 宝牌个数（手牌+副露）
	DoraCount int

	// 非立直状态下的打点期望（副露或默听）
	DamaPoint float64

	// 立直状态下的打点期望
	RiichiPoint float64

	// 局收支
	MixedRoundPoint float64

	// TODO: 赤牌改良提醒
}

type MajsoulExAnalysisResult struct {
	DiscardTile                 string
	IsDiscardDoraTile           bool
	DiscardTileValue            float64
	IsIsolatedYaochuDiscardTile bool
	Result13                    *Hand13AnalysisResult
	Result13String              string
	DiscardHonorTileRisk        int
	LeftDrawTilesCount          int
	OpenTiles                   []string
	IsOpen                      bool
}

var (
	LastHelperResult *[]*MajsoulExAnalysisResult
	LastDeal         string
)

var (
	fast  FastTestClient
	Tiles = make([]string, 0)
)

// !!! helper 不会输出 [赤牌] 所以需要结合自己判断手里有没有
func GetHandTile(t string) string {
	if len(Tiles) == 0 || t[0:1] != "5" {
		return t
	}
	i := -1
	for e, n := range Tiles {
		if n == t {
			i = e
		}
	}
	if i != -1 {
		return t
	}

	return "0" + t[1:]
}

func StartWebSocketServer() {
	http.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		var err error
		for {
			var msg string
			if err = websocket.Message.Receive(conn, &msg); err != nil {
				break
			}
			json.Unmarshal([]byte(msg), &LastHelperResult)
		}
	}))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 20001), nil))
}

var tr = http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	DisableKeepAlives: false,
}

var client = http.Client{
	Timeout:   10 * time.Second,
	Transport: &tr,
}

func PostToHelper(t interface{}) {
	if data, err := json.Marshal(t); err == nil {
		if string(data) == "{}" {
			return
		}
		// log.Println("post:", string(data))
		// log.Println(client.Post("https://localhost:12121", "Content-Type: application/json", bytes.NewBuffer(data)))
		client.Post("https://localhost:12121", "Content-Type: application/json", bytes.NewBuffer(data))
	}
}

func doOp(op *OptionalOperationList, tile string) {
	time.Sleep(2 * time.Second)

	var (
		timeuse = uint32(2)
		moqie   = true
	)

	if LastHelperResult != nil && len(*LastHelperResult) > 0 {
		info := (*LastHelperResult)[0]
		moqie = GetHandTile(info.DiscardTile) == LastDeal
		tile = GetHandTile(info.DiscardTile)
		LastHelperResult = nil
	}

	canRiichi := func() bool {
		for _, o := range op.GetOperationList() {
			if o.GetType() == E_PlayOperation_RiiChi {
				return true
			}
		}
		return false
	}

	if canRiichi() {
		fast.InputOperation(context.Background(), &ReqSelfOperation{
			Type: E_PlayOperation_RiiChi,
			Tile: tile,
			// !!! 真实情况中请根据是否是刚摸来的牌进行判断
			// !!! 如果是刚摸来的牌直接打出去则Moqie为true
			// !!! 请勿随便传, 否则会无法出牌
			// !!! 这里因为是Demo所以直接把摸来的牌摸切出去了
			Moqie:     moqie,
			Timeuse:   timeuse,
			TileState: 0,
		})
		return
	}

	for _, o := range op.GetOperationList() {
		switch o.GetType() {
		// 鸣牌操作时直接取消
		case E_PlayOperation_Chi, E_PlayOperation_Pon, E_PlayOperation_Ankan, E_PlayOperation_Minkan, E_PlayOperation_Kakan:
			// 取消鸣牌操作
			fast.InputChiPengGang(context.Background(), &ReqChiPengGang{
				CancelOperation: true,
				Timeuse:         timeuse,
			})
			// fast.InputOperation(context.Background(), &ReqSelfOperation{
			// 	CancelOperation: true,
			// 	Timeuse:         1,
			// })
		// 荣和
		case E_PlayOperation_Ron:
			fast.InputChiPengGang(context.Background(), &ReqChiPengGang{
				Type:  E_PlayOperation_Ron,
				Index: 0,
			})
		// 自摸
		case E_PlayOperation_Tsumo:
			fast.InputOperation(context.Background(), &ReqSelfOperation{
				Type:  E_PlayOperation_Tsumo,
				Index: 0,
			})
		// 九种九牌(想什么国土无双呢?)
		case E_PlayOperation_Kuku:
			fast.InputOperation(context.Background(), &ReqSelfOperation{
				Type:    E_PlayOperation_Kuku,
				Index:   0,
				Timeuse: timeuse,
			})
		// 立直
		case E_PlayOperation_RiiChi:
			fast.InputOperation(context.Background(), &ReqSelfOperation{
				Type: E_PlayOperation_RiiChi,
				Tile: tile,
				// !!! 真实情况中请根据是否是刚摸来的牌进行判断
				// !!! 如果是刚摸来的牌直接打出去则Moqie为true
				// !!! 请勿随便传, 否则会无法出牌
				// !!! 这里因为是Demo所以直接把摸来的牌摸切出去了
				Moqie:     moqie,
				Timeuse:   timeuse,
				TileState: 0,
			})
		// 出牌
		case E_PlayOperation_Discard:
			respDiscard, err := fast.InputOperation(context.Background(), &ReqSelfOperation{
				Type: E_PlayOperation_Discard,
				Tile: tile,
				// !!! 真实情况中请根据是否是刚摸来的牌进行判断
				// !!! 如果是刚摸来的牌直接打出去则Moqie为true
				// !!! 请勿随便传, 否则会无法出牌
				// !!! 这里因为是Demo所以直接把摸来的牌摸切出去了
				Moqie:     moqie,
				Timeuse:   timeuse,
				TileState: 0,
			})
			log.Println(tile, respDiscard, err)
		// 拔北
		// case E_PlayOperation_Kita:
		// 	fast.InputOperation(context.Background(), &ReqSelfOperation{
		// 		Type: E_PlayOperation_Kita,
		// 		// !!! 真实情况中请根据是否是刚摸来的牌进行判断
		// 		// !!! 如果是刚摸来的牌直接打出去则Moqie为true
		// 		// !!! 请勿随便传, 否则会无法出牌
		// 		// !!! 这里因为是Demo所以直接把摸来的牌摸切出去了
		// 		Moqie:   !!!,
		// 		Timeuse: timeuse,
		// 	})
		// 换三张
		case E_PlayOperation_HuanSanZhang:
			fast.InputOperation(context.Background(), &ReqSelfOperation{
				Type:        E_PlayOperation_HuanSanZhang,
				ChangeTiles: o.GetChangeTiles(),
				TileStates:  o.GetChangeTileStates(),
				Timeuse:     timeuse,
			})
		}
	}
}

// 重要！！！
// 身份认证部分，必须存在
type Authentication struct {
	AccessToken string
}

func (t *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"access_token": t.AccessToken}, nil
}

func (t *Authentication) RequireTransportSecurity() bool {
	return false
}

func main() {
	go StartWebSocketServer()

	var (
		Account  = "账号"
		Password = "密码"
		_        = "Token"
		URL      = "majserver.sykj.site"
	)

	// 从雀魂Ex官方获取Client端证书
	cert, err := tls.LoadX509KeyPair("./cer/client.pem", "./cer/client.key")
	certPool := x509.NewCertPool()
	ca, _ := ioutil.ReadFile("./cer/ca.crt")
	certPool.AppendCertsFromPEM(ca)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "majserver.sykj.site",
		RootCAs:      certPool,
	})

	conn, err := grpc.Dial(URL+":20009", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("[登录失败]: ", err)
	}
	defer conn.Close()

	// 登录获取AccessToken, 后续作为身份验证使用
	lobby := NewLobbyClient(conn)
	// 账号密码登录
	// 普通账号密码登录
	respLogin, err := lobby.Login(context.Background(), &ReqLogin{Account: Account, Password: Password})
	// 账号密码登录, 附加Server Chan通知
	// Type 0 => 旧版本
	// Type 1 => Turbo
	// Server Chan只需要登录时提交一次即可
	// respLogin, err := lobby.Login(context.Background(), &ReqLogin{Account: Account, Password: Password, ServerChan: &ServerChan{Type: 1, Sendkey: "Server Chan SendKey"}})
	// AccessToken登录
	// respLogin, err := lobby.Oauth2Login(context.Background(), &ReqOauth2Login{AccessToken: AccessToken})
	if err != nil {
		log.Fatal("[登录失败]: ", err)
	}

	PostToHelper(respLogin)

	log.Println("登录成功")
	// log.Println(respLogin)

	// 创建登录验证模块再次连接rpc
	auth := Authentication{
		AccessToken: respLogin.GetAccessToken(),
	}

	// 重连并附带AccessToken进行认证
	// !!! 重要, 不附带AccessToken将无法执行除登录外其他任何操作
	conn, err = grpc.Dial(URL+":20009", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatal("[连接失败]: ", err)
	}
	defer conn.Close()
	log.Println("二次连接成功")

	// 重新定义rpc client
	lobby = NewLobbyClient(conn)

	// ? 发送登录成功心跳 ?
	lobby.LoginBeat(context.Background(), &ReqLoginBeat{Contract: "DF2vkXCnfeXp4WoGrBGNcJBufZiMN3uP"})

	// 从杂货铺用铜币购买一个猫粮小本子
	// lobby.BuyFromZHP(context.Background(), &ReqBuyFromZHP{
	// 	GoodsId: 7,
	// 	Count:   1,
	// })

	// 领取复活币
	// lobby.GainReviveCoin(context.Background(), &ReqCommon{})

	// 赠送礼物给五十岚阳菜
	// lobby.SendGiftToCharacter(context.Background(), &ReqSendGiftToCharacter{
	// 	CharacterId: 200020,
	// 	Gifts: []*ReqSendGiftToCharacter_Gift{
	// 		{ItemId: 303062, Count: 1},
	// 	},
	// })

	// 进行匹配
	// lobby.MatchGame(context.Background(), &ReqJoinMatchQueue{
	// 	MatchMode: 40, // !!! 40 修罗之战 | 不知道请勿瞎填
	// })

	// FastClient 出牌操作
	// 对局内使用
	fast = NewFastTestClient(conn)

	var (
		ConnectToken string
		GameUuid     string
	)

	// 服务端推送回来的消息从这里收取
	// 所有打牌有关的数据也从这里收取
	notify := NewNotifyClient(conn)
	notifyClient, err := notify.Notify(context.Background(), &ClientStream{})
	if err == nil {
		go func() {
			for {
				in, err := notifyClient.Recv()
				if err != nil {
					break
				}
				// 收到的是 byte 数据
				// 解析方法:
				// 使用 proto 里的 Wrapper 进行一次解析
				// Wrapper { Name: "", Data: []byte }
				// 获取 name 后 将 data 解析为对应的 proto 数据
				// log.Println("收到服务端推送消息:", in.GetStream())

				wrapper := &Wrapper{}
				err = proto.Unmarshal(in.GetStream(), wrapper)
				if err != nil {
					log.Println("解析 Wrapper 数据错误:", err, "data:", in.GetStream())
					continue
				}

				if msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName("ex." + wrapper.GetName())); err == nil {
					msg := proto.MessageV1(msgType.New())
					if err = proto.Unmarshal(wrapper.GetData(), msg); err == nil {
						PostToHelper(msg)
					}
				}

				switch wrapper.GetName() {
				case "NotifyRoomGameStart":
					msg := &NotifyRoomGameStart{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					if err != nil {
						log.Println("解析 Wrapper.NotifyRoomGameStart 数据错误:", err, "data:", wrapper.GetData())
						continue
					}
					ConnectToken = msg.GetConnectToken()
					GameUuid = msg.GetGameUuid()
					log.Println("Notify Wrapper.NotifyRoomGameStart:", msg)
				case "NotifyMatchGameStart":
					msg := &NotifyMatchGameStart{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					if err != nil {
						log.Println("解析 Wrapper.NotifyRoomGameStart 数据错误:", err, "data:", wrapper.GetData())
						continue
					}
					ConnectToken = msg.GetConnectToken()
					GameUuid = msg.GetGameUuid()
					log.Println("Notify Wrapper.NotifyRoomGameStart:", msg)
				case "NotifyGameConnect": // 收到此消息说明已经连接上对局服务器, 服务端自动连接不用操心
					if ConnectToken == "" || GameUuid == "" {
						log.Println("未能获取到连接对局服务器信息")
						continue
					}
					// 验证对局信息
					respAuthGame, err := fast.AuthGame(context.Background(), &ReqAuthGame{
						AccountId: respLogin.GetAccountId(),
						Token:     ConnectToken,
						GameUuid:  GameUuid,
					})
					PostToHelper(respAuthGame)
					log.Println("AuthGame", respAuthGame, err)
					// 进入对局
					respEnterGame, err := fast.EnterGame(context.Background(), &ReqCommon{})
					PostToHelper(respEnterGame)
					log.Println("EnterGame", respEnterGame, err)
					log.Println("进入对局...")
				case "NotifyGameEndResult": // 对局结束
					// 进行匹配
					// lobby.MatchGame(context.Background(), &ReqJoinMatchQueue{
					// 	MatchMode: 40, // !!! 40 修罗之战 | 不知道请勿瞎填
					// })
					continue
				case "ActionNewRound":
					msg := &ActionNewRound{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					if err != nil {
						log.Println("未能解析对局信息")
						continue
					}
					Tiles = msg.GetTiles()
					op := msg.GetOperation()
					if op != nil {
						// !!! 这里需要等待, 防止服务器还没有开始导致无法出牌
						// !!! 雀魂发牌需要1200毫秒
						// !!! 如果是修罗需要额外再加[修罗之战]动画的等待时间
						// !!! 如果不等待则会出牌失败
						time.Sleep(1500 * time.Millisecond)
						doOp(op, msg.GetTiles()[len(msg.GetTiles())-1])
					}
				case "ActionDealTile":
					msg := &ActionDealTile{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					if err != nil {
						log.Println("未能解析摸牌信息")
						continue
					}
					Tiles = append(Tiles, msg.GetTile())
					op := msg.GetOperation()
					if op != nil {
						LastDeal = msg.GetTile()
						doOp(op, msg.GetTile())
					}
				case "ActionDiscardTile":
					msg := &ActionDiscardTile{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					if err != nil {
						log.Println("未能解析NoTile信息")
						continue
					}
					op := msg.GetOperation()
					if op != nil {
						doOp(op, msg.GetTile())
					}
				case "ActionChangeTile": // 换三张
					msg := &ActionChangeTile{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					if err != nil {
						log.Println("未能解析NoTile信息")
						continue
					}
					indexOf := func(t string) int {
						for i, e := range Tiles {
							if t == e {
								return i
							}
						}
						return -1
					}
					removeFromHand := func(out []string) {
						for _, t := range out {
							index := indexOf(t)
							if index == -1 {
								continue
							}
							Tiles = append(Tiles[:index], Tiles[index+1:]...)
						}
					}
					removeFromHand(msg.GetOutTiles())
					Tiles = append(Tiles, msg.GetInTiles()...)
					op := msg.GetOperation()
					if op != nil {
						doOp(op, msg.GetInTiles()[len(msg.GetInTiles())-1])
					}
				case "ActionNoTile":
					msg := &ActionNoTile{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					fast.ConfirmNewRound(context.Background(), &ReqCommon{})
				case "ActionHuleXueZhanEnd":
					msg := &ActionHuleXueZhanEnd{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					fast.ConfirmNewRound(context.Background(), &ReqCommon{})
				case "ActionHule":
					msg := &ActionHule{}
					err = proto.Unmarshal(wrapper.GetData(), msg)
					fast.ConfirmNewRound(context.Background(), &ReqCommon{})
				case "NotifyEndGameVote": // 方便测试, 收到投票结束立即同意投票结束对局
					fast.VoteGameEnd(context.Background(), &ReqVoteGameEnd{Yes: true})
				// and more case ...
				default:
					log.Println("未知 Wrapper 数据:", wrapper.GetName(), "data:", wrapper.GetData())
				}
			}
		}()
	}

	// time.Sleep(5 * time.Second)
	// log.Println("领取月卡")
	// log.Println(lobby.TakeMonthTicket(context.Background(), &ReqCommon{}))

	// time.Sleep(5 * time.Second)
	// lobby.JoinRoom(context.Background(), &ReqJoinRoom{RoomId: 21449})
	// time.Sleep(5 * time.Second)
	// lobby.ReadyPlay(context.Background(), &ReqRoomReady{Ready: true})
	// time.Sleep(5 * time.Second)
	// lobby.LeaveRoom(context.Background(), &ReqCommon{})

	// time.Sleep(5 * time.Second)
	// log.Println("登出")
	// 软登出
	// log.Println(lobby.SoftLogout(context.Background(), &ReqLogout{}))
	// !!! 区别, 硬登出会使AccessToken失效, 下一次无法再使用
	// 硬登出
	// log.Println(lobby.Logout(context.Background(), &ReqLogout{}))
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println(lobby.SoftLogout(context.Background(), &ReqLogout{}))
			break
		}
		if strings.Contains(line, "joinroom") {
			arr := strings.Split(strings.Trim(line, "\n"), " ")
			if len(arr) != 2 {
				log.Println("参数错误")
				continue
			}
			roomid, err := strconv.Atoi(strings.TrimSpace(strings.Trim(arr[1], "\n")))
			if err != nil {
				log.Println("输入参数非数字")
				continue
			}
			lobby.JoinRoom(context.Background(), &ReqJoinRoom{RoomId: uint32(roomid)})
			lobby.ReadyPlay(context.Background(), &ReqRoomReady{Ready: true})
		}
		if strings.Contains(line, "ready") {
			lobby.ReadyPlay(context.Background(), &ReqRoomReady{Ready: true})
		}
		if strings.Contains(line, "discard") {
			arr := strings.Split(strings.Trim(line, "\n"), " ")
			if len(arr) != 3 {
				log.Println("参数错误")
				continue
			}
			moqie, err := strconv.ParseBool(strings.TrimSpace(strings.Trim(arr[2], "\n")))
			if err != nil {
				log.Println("参数错误")
				continue
			}
			respDiscard, err := fast.InputOperation(context.Background(), &ReqSelfOperation{
				Type: E_PlayOperation_Discard,
				Tile: GetHandTile(strings.TrimSpace(strings.Trim(arr[1], "\n"))),
				// !!! 真实情况中请根据是否是刚摸来的牌进行判断
				// !!! 如果是刚摸来的牌直接打出去则Moqie为true
				// !!! 请勿随便传, 否则会无法出牌
				// !!! 这里因为是Demo所以直接把摸来的牌摸切出去了
				Moqie:     moqie,
				Timeuse:   1,
				TileState: 0,
			})
			log.Println(arr[1], respDiscard, err)
		}
		if strings.Contains(line, "exit") {
			log.Println(lobby.SoftLogout(context.Background(), &ReqLogout{}))
			break
		}
	}
}
