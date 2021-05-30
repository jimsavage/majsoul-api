### 该项目有什么作用？

---

通过该 API 可以与雀魂服务器进行通信, 更多功能还在开发中

### Quick Start

---

```golang
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

	log.Println("登录成功")
```

### More

---

目前 API 还在整理中，但已全部实现雀魂所有功能

B 站 ID: [神崎·H·亚里亚](https://space.bilibili.com/898411/)  
B 站 ID: [关野萝可](https://space.bilibili.com/612462792/)  
QQ 交流群: [991568358](https://jq.qq.com/?_wv=1027&k=3gaKRwqg)

请作者喝一杯咖啡

<figure class="third">
    <img src="https://moxcomic.github.io/wechat.png" width=170>
    <img src="https://moxcomic.github.io/alipay.png" width=170>
    <img src="https://moxcomic.github.io/qq.png" width=170>
</figure>
