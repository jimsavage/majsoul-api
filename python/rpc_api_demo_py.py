# !!! 请先安装以下内容 !!!
# pip install grpcio
# pip install protobuf
# pip install grpcio_tools

# 编译 proto 文件
# python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. *.proto

# proto 文件编译出来的两个文件
import ex_pb2
import ex_pb2_grpc
import grpc


def run():
    with open("./cer/client.key", "rb") as f:
        key = f.read()
    with open("./cer/client.pem", "rb") as f:
        cert = f.read()
    with open("./cer/ca.crt", "rb") as f:
        root_ca = f.read()
    creds = grpc.ssl_channel_credentials(
        root_certificates=root_ca, private_key=key, certificate_chain=cert)
    conn = grpc.secure_channel("majserver.sykj.site:20009", creds)
    lobby = ex_pb2_grpc.LobbyStub(conn)
    # # 普通账号密码登录
    respLogin = lobby.login(ex_pb2.ReqLogin(
        account="账号", password="密码"))
    # 账号密码登录, 附加Server Chan通知
    # lobby.login(ReqLogin(account="账号", password="密码", server_chan=ServerChan(type=1, send_key="Server Chan SendKey")))
    # AccessToken登录
    # lobby.oauth2Login(ReqOauth2Login(access_token="Token"))
    print(respLogin.access_token)

    # 创建身份验证模块
    md = [("access_token", respLogin.access_token)]

    # 领取月卡
    respTakeMonthTicket = lobby.takeMonthTicket(
        ex_pb2.ReqCommon(),
        metadata=md,
    )
    print(respTakeMonthTicket)

    # 软登出
    respLogout = lobby.softLogout(
        ex_pb2.ReqCommon(),
        metadata=md,
    )
    print(respLogout)


if __name__ == "__main__":
    run()
