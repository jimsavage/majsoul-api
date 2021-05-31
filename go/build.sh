go build -ldflags "-w -s" -a -o majsoulex_asura_mac && upx -9 majsoulex_asura_mac
CGO_ENABLED=0 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ GOOS=windows GOARCH=386 go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_win32.exe && upx -9 majsoulex_asura_win32.exe
CGO_ENABLED=0 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_win64.exe && upx -9 majsoulex_asura_win64.exe
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_mac_arm64 && upx -9 majsoulex_asura_mac_arm64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_linux_amd64 && upx -9 majsoulex_asura_linux_amd64
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_linux_386 && upx -9 majsoulex_asura_linux_386
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_linux_arm64 && upx -9 majsoulex_asura_linux_arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_linux_arm && upx -9 majsoulex_asura_linux_arm

CGO_ENABLED=0 GOOS=windows GOARCH=arm go build -x -v -ldflags "-s -w" -a -o majsoulex_asura_windows_arm