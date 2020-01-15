# lovegaudi

## 先安装最新版本的 go
```
add-apt-repository ppa:longsleep/golang-backports
apt-get update
apt-get install golang-go
```

## 配置 goproxy
```
export GO111MODULE=on
export GOPROXY=https://goproxy.cn
```

## 进入项目目录直接运行
```
git clone https://github.com/taojy123/lovegaudi
cd lovegaudi
go run main.go
```

