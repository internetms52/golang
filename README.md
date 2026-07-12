# golang
golang excercise

## 基本指令

### 初始化專案
```bash
go mod init <module-name>
```
初始化一個新的 Go module，會建立 `go.mod` 檔案來管理相依套件。

### 編譯程式
```bash
go build
```
編譯當前目錄的 Go 程式，會產生可執行檔。

```bash
go build -o <output-name>
```
指定輸出檔案名稱。

### Makefile 用法
建立 `Makefile` 來簡化常用指令：

```makefile
build:
	go build -o bin/app

run:
	go run main.go

clean:
	rm -rf bin/
```

使用方式：
```bash
make build   # 編譯程式
make run     # 執行程式
make clean   # 清理編譯檔案
```

# 安裝gopls
```
//install
go install golang.org/x/tools/gopls@latest

//config
echo 'export PATH=$PATH:/Users/luyuhsin/go/bin' >> ~/.zshrc
source ~/.zshrc

//verify
which gopls

```