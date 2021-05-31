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

### RPC API

---

#### Notify

```protobuf
message ClientStream {
}
message ServerStream {
  bytes stream = 1;
}
service Notify {
	// 这个Notify用来监听服务端给出的通知信息
	// 例: 打牌时的实时数据
  rpc Notify(ClientStream) returns (stream ServerStream);
}
```

例:

```golang
notify := NewNotifyClient(conn)
notifyClient, err := notify.Notify(context.Background(), &ClientStream{})
if err == nil {
	go func() {
		for {
			in, err := notifyClient.Recv()
			if err != nil {
				break
			}
		}
	}()
}
```

#### FastTest

```protobuf
// 验证游戏
rpc authGame (ReqAuthGame) returns (ResAuthGame);
// 例:
// fast.AuthGame(context.Background(), &ReqAuthGame{
//	 // 登录时获取的AccountId
//	 AccountId: respLogin.GetAccountId(),
//	 // NotifyRoomGameStart 或 NotifyMatchGameStart 时获取的 ConnectToken
//	 Token:     ConnectToken,
//	 // 同上
//	 GameUuid:  GameUuid,
// })

// 确认观战信息
rpc authObserve (ReqAuthObserve) returns (ResCommon);
// 例:
// fast.AuthObserve(context.Background(), &ReqAuthObserve{
// 	Token: "",	// 暂时未知参数
// })

// 游戏内通知, 例如玩家发表情等
rpc broadcastInGame (ReqBroadcastInGame) returns (ResCommon);
// 例:
// fast.BroadcastInGame(context.Background(), &ReqBroadcastInGame{
//	 Content: "",				// 暂时未知参数
//	 ExceptSelf: false,	// 暂时未知参数
// })

// 检查网络延迟
rpc checkNetworkDelay (ReqCommon) returns (ResCommon);
// 例:
// fast.CheckNetworkDelay(context.Background(), &ReqCommon{})

// 清除???
rpc clearLeaving (ReqCommon) returns (ResCommon);
// 例:
// fast.ClearLeaving(context.Background(), &ReqCommon{})

// 小对局结束确认
rpc confirmNewRound (ReqCommon) returns (ResCommon);
// 例:
// fast.ConfirmNewRound(context.Background(), &ReqCommon{})

// 进入对局
rpc enterGame (ReqCommon) returns (ResEnterGame);
// 例:
// fast.EnterGame(context.Background(), &ReqCommon{})

// 获取对局玩家状态
rpc fetchGamePlayerState (ReqCommon) returns (ResGamePlayerState);
// 例:
// fast.EnterGame(context.Background(), &ReqCommon{})

// 结束重连同步
rpc finishSyncGame (ReqCommon) returns (ResCommon);
// 例:
// fast.FinishSyncGame(context.Background(), &ReqCommon{})

// 吃/碰/杠 操作
rpc inputChiPengGang (ReqChiPengGang) returns (ResCommon);
// 例:
// 例子为取消操作, 具体请根据实际情况选择
// 如果不懂请看 go demo 部分
// fast.InputChiPengGang(context.Background(), &ReqChiPengGang{
// 	 CancelOperation: true,
//	 Timeuse:         timeuse,
// })

// GM命令
rpc inputGameGMCommand (ReqGMCommandInGaming) returns (ResCommon);
// 例:
// fast.InputGameGMCommand(context.Background(), &ReqGMCommandInGaming{
// 	JsonData: "",	// 暂时未知参数
// })

// 普通操作, 例如: 打牌、立直等
rpc inputOperation (ReqSelfOperation) returns (ResCommon);
// 例:
// 例子为打牌操作, 具体请根据实际情况选择
// 如果不懂请看 go demo 部分
// fast.InputOperation(context.Background(), &ReqSelfOperation{
//	 Type: E_PlayOperation_Discard,
//	 Tile: tile,
//	 // !!! 真实情况中请根据是否是刚摸来的牌进行判断
//	 // !!! 如果是刚摸来的牌直接打出去则Moqie为true
//	 // !!! 请勿随便传, 否则会无法出牌
//	 // !!! 这里因为是Demo所以直接把摸来的牌摸切出去了
//	 Moqie:     moqie,
//	 Timeuse:   timeuse,
//	 TileState: 0,
// })

// 开始观战
rpc startObserve (ReqCommon) returns (ResStartObserve);
// 例:
// fast.StartObserve(context.Background(), &ReqCommon{})

// 停止观战
rpc stopObserve (ReqCommon) returns (ResCommon);
// 例:
// fast.StopObserve(context.Background(), &ReqCommon{})

// 开始重连
rpc syncGame (ReqSyncGame) returns (ResSyncGame);
// 例:
// fast.SyncGame(context.Background(), &ReqSyncGame{
// 	RoundId: "",	// 暂时未知参数
// 	Step:    0,		// 暂时未知参数
// })

// 结束游戏(仅友人房可以)
rpc terminateGame (ReqCommon) returns (ResCommon);
// 例:
// fast.TerminateGame(context.Background(), &ReqCommon{})

// 发起投票结束对局(友人房)
rpc voteGameEnd (ReqVoteGameEnd) returns (ResGameEndVote);
// 例:
// fast.VoteGameEnd(context.Background(), &ReqVoteGameEnd{
// 	Yes: true,	// 同意结束
// })
```

#### Lobby

```protobuf
service Lobby {
	// 软登出(不销毁AccessToken)
	rpc softLogout (ReqLogout) returns (ResLogout);
	// 领取月卡
	rpc takeMonthTicket (ReqCommon) returns (ResPayMonthTicket);

	rpc addCollectedGameRecord (ReqAddCollectedGameRecord) returns (ResAddCollectedGameRecord);
	rpc addFinishedEnding (ReqFinishedEnding) returns (ResCommon);
	rpc applyFriend (ReqApplyFriend) returns (ResCommon);
	rpc bindAccount (ReqBindAccount) returns (ResCommon);
	rpc bindEmail (ReqBindEmail) returns (ResCommon);
	rpc bindPhoneNumber (ReqBindPhoneNumber) returns (ResCommon);
	rpc buyFromChestShop (ReqBuyFromChestShop) returns (ResBuyFromChestShop);
	rpc buyFromShop (ReqBuyFromShop) returns (ResBuyFromShop);
	rpc buyFromZHP (ReqBuyFromZHP) returns (ResCommon);
	rpc buyInABMatch (ReqBuyInABMatch) returns (ResCommon);
	rpc buyShiLian (ReqBuyShiLian) returns (ResCommon);
	rpc cancelGooglePlayOrder (ReqCancelGooglePlayOrder) returns (ResCommon);
	rpc cancelMatch (ReqCancelMatchQueue) returns (ResCommon);
	rpc cancelUnifiedMatch (ReqCancelUnifiedMatch) returns (ResCommon);
	rpc changeAvatar (ReqChangeAvatar) returns (ResCommon);
	rpc changeCharacterSkin (ReqChangeCharacterSkin) returns (ResCommon);
	rpc changeCharacterView (ReqChangeCharacterView) returns (ResCommon);
	rpc changeCollectedGameRecordRemarks (ReqChangeCollectedGameRecordRemarks) returns (ResChangeCollectedGameRecordRemarks);
	rpc changeCommonView (ReqChangeCommonView) returns (ResCommon);
	rpc changeMainCharacter (ReqChangeMainCharacter) returns (ResCommon);
	rpc clientMessage (ReqClientMessage) returns (ResCommon);
	rpc completeActivityFlipTask (ReqCompleteActivityTask) returns (ResCommon);
	rpc completeActivityTask (ReqCompleteActivityTask) returns (ResCommon);
	rpc completePeriodActivityTask (ReqCompleteActivityTask) returns (ResCommon);
	rpc completeRandomActivityTask (ReqCompleteActivityTask) returns (ResCommon);
	rpc composeShard (ReqComposeShard) returns (ResCommon);
	rpc createAlipayAppOrder (ReqCreateAlipayAppOrder) returns (ResCreateAlipayAppOrder);
	rpc createAlipayOrder (ReqCreateAlipayOrder) returns (ResCreateAlipayOrder);
	rpc createAlipayScanOrder (ReqCreateAlipayScanOrder) returns (ResCreateAlipayScanOrder);
	rpc createBillingOrder (ReqCreateBillingOrder) returns (ResCreateBillingOrder);
	rpc createDMMOrder (ReqCreateDMMOrder) returns (ResCreateDmmOrder);
	rpc createENAlipayOrder (ReqCreateENAlipayOrder) returns (ResCreateENAlipayOrder);
	rpc createENJCBOrder (ReqCreateENJCBOrder) returns (ResCreateENJCBOrder);
	rpc createENMasterCardOrder (ReqCreateENMasterCardOrder) returns (ResCreateENMasterCardOrder);
	rpc createENPaypalOrder (ReqCreateENPaypalOrder) returns (ResCreateENPaypalOrder);
	rpc createENVisaOrder (ReqCreateENVisaOrder) returns (ResCreateENVisaOrder);
	rpc createEmailVerifyCode (ReqCreateEmailVerifyCode) returns (ResCommon);
	rpc createGameObserveAuth (ReqCreateGameObserveAuth) returns (ResCreateGameObserveAuth);
	rpc createIAPOrder (ReqCreateIAPOrder) returns (ResCreateIAPOrder);
	rpc createJPAuOrder (ReqCreateJPAuOrder) returns (ResCreateJPAuOrder);
	rpc createJPCreditCardOrder (ReqCreateJPCreditCardOrder) returns (ResCreateJPCreditCardOrder);
	rpc createJPDocomoOrder (ReqCreateJPDocomoOrder) returns (ResCreateJPDocomoOrder);
	rpc createJPPaypalOrder (ReqCreateJPPaypalOrder) returns (ResCreateJPPaypalOrder);
	rpc createJPSoftbankOrder (ReqCreateJPSoftbankOrder) returns (ResCreateJPSoftbankOrder);
	rpc createJPWebMoneyOrder (ReqCreateJPWebMoneyOrder) returns (ResCreateJPWebMoneyOrder);
	rpc createMyCardAndroidOrder (ReqCreateMyCardOrder) returns (ResCreateMyCardOrder);
	rpc createMyCardWebOrder (ReqCreateMyCardOrder) returns (ResCreateMyCardOrder);
	rpc createNickname (ReqCreateNickname) returns (ResCommon);
	rpc createPaypalOrder (ReqCreatePaypalOrder) returns (ResCreatePaypalOrder);
	rpc createPhoneLoginBind (ReqCreatePhoneLoginBind) returns (ResCommon);
	rpc createPhoneVerifyCode (ReqCreatePhoneVerifyCode) returns (ResCommon);
	rpc createRoom (ReqCreateRoom) returns (ResCreateRoom);
	rpc createSteamOrder (ReqCreateSteamOrder) returns (ResCreateSteamOrder);
	rpc createWechatAppOrder (ReqCreateWechatAppOrder) returns (ResCreateWechatAppOrder);
	rpc createWechatNativeOrder (ReqCreateWechatNativeOrder) returns (ResCreateWechatNativeOrder);
	rpc createXsollaOrder (ReqCreateXsollaOrder) returns (ResCreateXsollaOrder);
	rpc createYostarSDKOrder (ReqCreateYostarOrder) returns (ResCreateYostarOrder);
	rpc deleteComment (ReqDeleteComment) returns (ResCommon);
	rpc deleteMail (ReqDeleteMail) returns (ResCommon);
	rpc dmmPreLogin (ReqDMMPreLogin) returns (ResDMMPreLogin);
	rpc doActivitySignIn (ReqDoActivitySignIn) returns (ResDoActivitySignIn);
	rpc doDailySignIn (ReqCommon) returns (ResCommon);
	rpc dressingStatus (ReqRoomDressing) returns (ResCommon);
	rpc emailLogin (ReqEmailLogin) returns (ResLogin);
	rpc enterCustomizedContest (ReqEnterCustomizedContest) returns (ResEnterCustomizedContest);
	rpc exchangeActivityItem (ReqExchangeActivityItem) returns (ResExchangeActivityItem);
	rpc exchangeChestStone (ReqExchangeCurrency) returns (ResCommon);
	rpc exchangeCurrency (ReqExchangeCurrency) returns (ResCommon);
	rpc exchangeDiamond (ReqExchangeCurrency) returns (ResCommon);
	rpc fetchABMatchInfo (ReqCommon) returns (ResFetchABMatch);
	rpc fetchAccountActivityData (ReqCommon) returns (ResAccountActivityData);
	rpc fetchAccountChallengeRankInfo (ReqAccountInfo) returns (ResAccountChallengeRankInfo);
	rpc fetchAccountCharacterInfo (ReqCommon) returns (ResAccountCharacterInfo);
	rpc fetchAccountInfo (ReqAccountInfo) returns (ResAccountInfo);
	rpc fetchAccountSettings (ReqCommon) returns (ResAccountSettings);
	rpc fetchAccountState (ReqAccountList) returns (ResAccountStates);
	rpc fetchAccountStatisticInfo (ReqAccountStatisticInfo) returns (ResAccountStatisticInfo);
	rpc fetchAchievement (ReqCommon) returns (ResAchievement);
	rpc fetchAchievementRate (ReqCommon) returns (ResFetchAchievementRate);
	rpc fetchActivityBuff (ReqCommon) returns (ResActivityBuff);
	rpc fetchActivityFlipInfo (ReqFetchActivityFlipInfo) returns (ResFetchActivityFlipInfo);
	rpc fetchActivityList (ReqCommon) returns (ResActivityList);
	rpc fetchAllCommonViews (ReqCommon) returns (ResAllcommonViews);
	rpc fetchAnnouncement (ReqFetchAnnouncement) returns (ResAnnouncement);
	rpc fetchBagInfo (ReqCommon) returns (ResBagInfo);
	rpc fetchChallengeInfo (ReqCommon) returns (ResFetchChallengeInfo);
	rpc fetchChallengeLeaderboard (ReqChallangeLeaderboard) returns (ResChallengeLeaderboard);
	rpc fetchChallengeSeason (ReqCommon) returns (ResChallengeSeasonInfo);
	rpc fetchCharacterInfo (ReqCommon) returns (ResCharacterInfo);
	rpc fetchClientValue (ReqCommon) returns (ResClientValue);
	rpc fetchCollectedGameRecordList (ReqCommon) returns (ResCollectedGameRecordList);
	rpc fetchCommentContent (ReqFetchCommentContent) returns (ResFetchCommentContent);
	rpc fetchCommentList (ReqFetchCommentList) returns (ResFetchCommentList);
	rpc fetchCommentSetting (ReqCommon) returns (ResCommentSetting);
	rpc fetchCommonView (ReqCommon) returns (ResCommonView);
	rpc fetchCommonViews (ReqCommonViews) returns (ResCommonViews);
	rpc fetchConnectionInfo (ReqCommon) returns (ResConnectionInfo);
	rpc fetchCurrentMatchInfo (ReqCurrentMatchInfo) returns (ResCurrentMatchInfo);
	rpc fetchCustomizedContestAuthInfo (ReqFetchCustomizedContestAuthInfo) returns (ResFetchCustomizedContestAuthInfo);
	rpc fetchCustomizedContestByContestId (ReqFetchCustomizedContestByContestId) returns (ResFetchCustomizedContestByContestId);
	rpc fetchCustomizedContestExtendInfo (ReqFetchCustomizedContestExtendInfo) returns (ResFetchCustomizedContestExtendInfo);
	rpc fetchCustomizedContestGameLiveList (ReqFetchCustomizedContestGameLiveList) returns (ResFetchCustomizedContestGameLiveList);
	rpc fetchCustomizedContestGameRecords (ReqFetchCustomizedContestGameRecords) returns (ResFetchCustomizedContestGameRecords);
	rpc fetchCustomizedContestList (ReqFetchCustomizedContestList) returns (ResFetchCustomizedContestList);
	rpc fetchCustomizedContestOnlineInfo (ReqFetchCustomizedContestOnlineInfo) returns (ResFetchCustomizedContestOnlineInfo);
	rpc fetchDailySignInInfo (ReqCommon) returns (ResDailySignInInfo);
	rpc fetchDailyTask (ReqCommon) returns (ResDailyTask);
	rpc fetchFriendApplyList (ReqCommon) returns (ResFriendApplyList);
	rpc fetchFriendList (ReqCommon) returns (ResFriendList);
	rpc fetchGameLiveInfo (ReqGameLiveInfo) returns (ResGameLiveInfo);
	rpc fetchGameLiveLeftSegment (ReqGameLiveLeftSegment) returns (ResGameLiveLeftSegment);
	rpc fetchGameLiveList (ReqGameLiveList) returns (ResGameLiveList);
	rpc fetchGamePointRank (ReqGamePointRank) returns (ResGamePointRank);
	rpc fetchGameRecord (ReqGameRecord) returns (ResGameRecord);
	rpc fetchGameRecordList (ReqGameRecordList) returns (ResGameRecordList);
	rpc fetchGameRecordsDetail (ReqGameRecordsDetail) returns (ResGameRecordsDetail);
	rpc fetchIDCardInfo (ReqCommon) returns (ResIDCardInfo);
	rpc fetchLevelLeaderboard (ReqLevelLeaderboard) returns (ResLevelLeaderboard);
	rpc fetchMailInfo (ReqCommon) returns (ResMailInfo);
	rpc fetchMisc (ReqCommon) returns (ResMisc);
	rpc fetchModNicknameTime (ReqCommon) returns (ResModNicknameTime);
	rpc fetchMonthTicketInfo (ReqCommon) returns (ResMonthTicketInfo);
	rpc fetchMultiAccountBrief (ReqMultiAccountId) returns (ResMultiAccountBrief);
	rpc fetchMutiChallengeLevel (ReqMutiChallengeLevel) returns (ResMutiChallengeLevel);
	rpc fetchPhoneLoginBind (ReqCommon) returns (ResFetchPhoneLoginBind);
	rpc fetchPlatformProducts (ReqPlatformBillingProducts) returns (ResPlatformBillingProducts);
	rpc fetchRankPointLeaderboard (ReqFetchRankPointLeaderboard) returns (ResFetchRankPointLeaderboard);
	rpc fetchRefundOrder (ReqCommon) returns (ResFetchRefundOrder);
	rpc fetchReviveCoinInfo (ReqCommon) returns (ResReviveCoinInfo);
	rpc fetchRollingNotice (ReqCommon) returns (ReqRollingNotice);
	rpc fetchRoom (ReqCommon) returns (ResSelfRoom);
	rpc fetchSelfGamePointRank (ReqGamePointRank) returns (ResFetchSelfGamePointRank);
	rpc fetchServerSettings (ReqCommon) returns (ResServerSettings);
	rpc fetchServerTime (ReqCommon) returns (ResServerTime);
	rpc fetchShopInfo (ReqCommon) returns (ResShopInfo);
	rpc fetchTitleList (ReqCommon) returns (ResTitleList);
	rpc fetchVipReward (ReqCommon) returns (ResVipReward);
	rpc followCustomizedContest (ReqTargetCustomizedContest) returns (ResCommon);
	rpc forceCompleteChallengeTask (ReqForceCompleteChallengeTask) returns (ResCommon);
	rpc gainAccumulatedPointActivityReward (ReqGainAccumulatedPointActivityReward) returns (ResCommon);
	rpc gainMultiPointActivityReward (ReqGainMultiPointActivityReward) returns (ResCommon);
	rpc gainRankPointReward (ReqGainRankPointReward) returns (ResCommon);
	rpc gainReviveCoin (ReqCommon) returns (ResCommon);
	rpc gainVipReward (ReqGainVipReward) returns (ResCommon);
	rpc gameMasterCommand (ReqGMCommand) returns (ResCommon);
	rpc goNextShiLian (ReqCommon) returns (ResCommon);
	rpc handleFriendApply (ReqHandleFriendApply) returns (ResCommon);
	rpc heatbeat (ReqHeatBeat) returns (ResCommon);
	rpc joinCustomizedContestChatRoom (ReqJoinCustomizedContestChatRoom) returns (ResJoinCustomizedContestChatRoom);
	rpc joinRoom (ReqJoinRoom) returns (ResJoinRoom);
	rpc kickPlayer (ReqRoomKick) returns (ResCommon);
	rpc leaveComment (ReqLeaveComment) returns (ResCommon);
	rpc leaveCustomizedContest (ReqCommon) returns (ResCommon);
	rpc leaveCustomizedContestChatRoom (ReqCommon) returns (ResCommon);
	rpc leaveRoom (ReqCommon) returns (ResCommon);
	rpc likeSNS (ReqLikeSNS) returns (ResLikeSNS);
	rpc login (ReqLogin) returns (ResLogin);
	rpc loginBeat (ReqLoginBeat) returns (ResCommon);
	rpc loginSuccess (ReqCommon) returns (ResCommon);
	rpc logout (ReqLogout) returns (ResLogout);
	rpc matchGame (ReqJoinMatchQueue) returns (ResCommon);
	rpc matchShiLian (ReqCommon) returns (ResCommon);
	rpc modifyBirthday (ReqModifyBirthday) returns (ResCommon);
	rpc modifyNickname (ReqModifyNickname) returns (ResCommon);
	rpc modifyPassword (ReqModifyPassword) returns (ResCommon);
	rpc modifyRoom (ReqModifyRoom) returns (ResCommon);
	rpc modifySignature (ReqModifySignature) returns (ResCommon);
	rpc oauth2Auth (ReqOauth2Auth) returns (ResOauth2Auth);
	rpc oauth2Check (ReqOauth2Check) returns (ResOauth2Check);
	rpc oauth2Login (ReqOauth2Login) returns (ResLogin);
	rpc oauth2Signup (ReqOauth2Signup) returns (ResOauth2Signup);
	rpc openAllRewardItem (ReqOpenAllRewardItem) returns (ResOpenAllRewardItem);
	rpc openChest (ReqOpenChest) returns (ResOpenChest);
	rpc openManualItem (ReqOpenManualItem) returns (ResCommon);
	rpc openRandomRewardItem (ReqOpenRandomRewardItem) returns (ResOpenRandomRewardItem);
	rpc payMonthTicket (ReqPayMonthTicket) returns (ResPayMonthTicket);
	rpc quitABMatch (ReqCommon) returns (ResCommon);
	rpc readAnnouncement (ReqReadAnnouncement) returns (ResCommon);
	rpc readGameRecord (ReqGameRecord) returns (ResCommon);
	rpc readMail (ReqReadMail) returns (ResCommon);
	rpc readSNS (ReqReadSNS) returns (ResReadSNS);
	rpc readyPlay (ReqRoomReady) returns (ResCommon);
	rpc receiveABMatchReward (ReqCommon) returns (ResCommon);
	rpc receiveAchievementGroupReward (ReqReceiveAchievementGroupReward) returns (ResReceiveAchievementGroupReward);
	rpc receiveAchievementReward (ReqReceiveAchievementReward) returns (ResReceiveAchievementReward);
	rpc receiveActivityFlipTask (ReqReceiveActivityFlipTask) returns (ResReceiveActivityFlipTask);
	rpc receiveChallengeRankReward (ReqReceiveChallengeRankReward) returns (ResReceiveChallengeRankReward);
	rpc receiveEndingReward (ReqFinishedEnding) returns (ResCommon);
	rpc receiveVersionReward (ReqCommon) returns (ResCommon);
	rpc refreshChallenge (ReqCommon) returns (ResRefreshChallenge);
	rpc refreshDailyTask (ReqRefreshDailyTask) returns (ResRefreshDailyTask);
	rpc refreshGameObserveAuth (ReqRefreshGameObserveAuth) returns (ResRefreshGameObserveAuth);
	rpc refreshZHPShop (ReqReshZHPShop) returns (ResRefreshZHPShop);
	rpc removeCollectedGameRecord (ReqRemoveCollectedGameRecord) returns (ResRemoveCollectedGameRecord);
	rpc removeFriend (ReqRemoveFriend) returns (ResCommon);
	rpc replySNS (ReqReplySNS) returns (ResReplySNS);
	rpc richmanAcitivitySpecialMove (ReqRichmanSpecialMove) returns (ResRichmanNextMove);
	rpc richmanActivityChestInfo (ReqRichmanChestInfo) returns (ResRichmanChestInfo);
	rpc richmanActivityNextMove (ReqRichmanNextMove) returns (ResRichmanNextMove);
	rpc saveCommonViews (ReqSaveCommonViews) returns (ResCommon);
	rpc sayChatMessage (ReqSayChatMessage) returns (ResCommon);
	rpc searchAccountById (ReqSearchAccountById) returns (ResSearchAccountById);
	rpc searchAccountByPattern (ReqSearchAccountByPattern) returns (ResSearchAccountByPattern);
	rpc sellItem (ReqSellItem) returns (ResCommon);
	rpc sendClientMessage (ReqSendClientMessage) returns (ResCommon);
	rpc sendGiftToCharacter (ReqSendGiftToCharacter) returns (ResSendGiftToCharacter);
	rpc shopPurchase (ReqShopPurchase) returns (ResShopPurchase);
	rpc signup (ReqSignupAccount) returns (ResSignupAccount);
	rpc solveGooglePayOrderV3 (ReqSolveGooglePlayOrderV3) returns (ResCommon);
	rpc solveGooglePlayOrder (ReqSolveGooglePlayOrder) returns (ResCommon);
	rpc startCustomizedContest (ReqStartCustomizedContest) returns (ResCommon);
	rpc startRoom (ReqRoomStart) returns (ResCommon);
	rpc startUnifiedMatch (ReqStartUnifiedMatch) returns (ResCommon);
	rpc stopCustomizedContest (ReqCommon) returns (ResCommon);
	rpc takeAttachmentFromMail (ReqTakeAttachment) returns (ResCommon);
	rpc unbindPhoneNumber (ReqUnbindPhoneNumber) returns (ResCommon);
	rpc unfollowCustomizedContest (ReqTargetCustomizedContest) returns (ResCommon);
	rpc updateAccountSettings (ReqUpdateAccountSettings) returns (ResCommon);
	rpc updateCharacterSort (ReqUpdateCharacterSort) returns (ResCommon);
	rpc updateClientValue (ReqUpdateClientValue) returns (ResCommon);
	rpc updateCommentSetting (ReqUpdateCommentSetting) returns (ResCommon);
	rpc updateIDCardInfo (ReqUpdateIDCardInfo) returns (ResCommon);
	rpc updateReadComment (ReqUpdateReadComment) returns (ResCommon);
	rpc upgradeActivityBuff (ReqUpgradeActivityBuff) returns (ResActivityBuff);
	rpc upgradeChallenge (ReqCommon) returns (ResUpgradeChallenge);
	rpc upgradeCharacter (ReqUpgradeCharacter) returns (ResUpgradeCharacter);
	rpc useBagItem (ReqUseBagItem) returns (ResCommon);
	rpc useCommonView (ReqUseCommonView) returns (ResCommon);
	rpc useGiftCode (ReqUseGiftCode) returns (ResUseGiftCode);
	rpc useSpecialGiftCode (ReqUseGiftCode) returns (ResUseSpecialGiftCode);
	rpc useTitle (ReqUseTitle) returns (ResCommon);
	rpc userComplain (ReqUserComplain) returns (ResCommon);
	rpc verfifyCodeForSecure (ReqVerifyCodeForSecure) returns (ResVerfiyCodeForSecure);
	rpc verificationIAPOrder (ReqVerificationIAPOrder) returns (ResVerificationIAPOrder);
	rpc verifyMyCardOrder (ReqVerifyMyCardOrder) returns (ResCommon);
	rpc verifySteamOrder (ReqVerifySteamOrder) returns (ResCommon);
}
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
