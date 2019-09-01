# wphpfpm (Windows PHP FPM) #

wphpfpm 是我初次練習 Go Lang 開發用來管理 Windows 下的 php-cgi

由於 php-cgi 一次只能服務一個客戶端，除非使用 apache 的 mod_fcgid，不然還真難管理

所以我就自己寫來玩玩，主要是因為用 caddy 來測試 php 只能啟動一隻 php-cgi 實在太不人道了，而且當我修改 php 的設定值，想要重啟 php-cgi，必須要自己手動砍掉 php-cgi 行程

目前有的功能如下

1. wphpfpm 是獨立的服務，類似 Linux 下的 php-fpm
2. 可以建立不同辦本的 php-cgi 來跑
3. php-cgi 可以設定最大啟動的數量
4. 可以安裝於 Windows Service，也可以命令列模式下跑
5. JSON 格式的設定檔

請直接下載 GO SDK (version 1.12+)後，執行以下命令，就可以得到 wphpfpm.exe

~~~bash
go build
~~~





## 設定檔說明 ##

以下是 json 範例

```json
{
    "Instances" : [
        {
            "Bind" : "127.0.0.1:8000",
            "ExecPath": "C:\\PHP7\\php-cgi.exe",
            "Args" : [],
            "Env": [
                "PHP_FCGI_MAX_REQUESTS=5000" ,
                "PHP_INI_SCAN_DIR=c:\\php7\\conf.d",
                "PHPIniDir=c:\\php7"
            ],
            "MaxProcesses" : 8,
            "MaxRequestsPerProcess": 5000,
            "Note" : "這是專門跑 PHP7 用的 , MaxProcesses 最大啟動 8 個 php-cgi"
        } ,

        {
            "Bind" : "127.0.0.1:8001",
            "ExecPath": "C:\\PHP5\\php-cgi.exe",
            "Args" : [],
            "Env": [
                "PHP_FCGI_MAX_REQUESTS=5000" ,
                "PHP_INI_SCAN_DIR=c:\\php5\\conf.d",
                "PHPIniDir=c:\\php5"
            ],
            "MaxProcesses" : 4,
            "MaxRequestsPerProcess": 5000,
            "Note" : "這是專門跑 PHP5 用的 , MaxProcesses 最大啟動 4 個 php-cgi"
        }
    ]
}
```

- Instances : 定義有多少種 php-cgi 要啟動，這可做為多版本之用
- Bind : 定義該 instance 要使用甚麼 IP 及 Port ，若針對多版本必須讓不同的 Instances 用不同的 Port 才有效
- ExecPath : php-cgi 真實路徑
- Args : 可以帶入 php-cgi 額外參數，**注意，不能使用 -b 的參數**
- Env : 可以額外加上環境變數
- MaxProcesses : 最大 php-cgi 執行數量
- MaxRequestsPerProcess : 每隻 php-cgi 行程，最多能處理幾次請求 , 這個數值必須與 Env 的環境變數 PHP_FCGI_MAX_REQUESTS 一致才不會出問題
- Note : 此欄位並無作用，只是用來註解的



## 使用方式 ##

### 在命令列模式下執行 ###

```
wphpfpm run --conf=config.json
```

### 安裝於 Windows Service ###

```
wphpfpm install --conf=c:\wphpfpm\config.json
```

注意，安裝為 Windows Service 模式運作時，必須使用管理者權限才能安裝

### 移除 wphpfpm service ###

```
wphpfpm uninstall
```



### 啟動及停止 wphpfpm service ###

```
wphpfpm start
wphpfpm stop
```

或者，在 Windows 的 **控制台\所有控制台項目\系統管理工具** 下的 **服務** 也可以進行啟動或停止 PHP FastCGI Manager for windows



## 尚未完成的功能 ##

- 更好的 Logger，目前都先隨便印訊息，能運作為先
- Log File
- Log Level
- backlog

## 資源清單 ##

- Windows Service 處理 : https://github.com/chai2010/winsvc
- 命令列處理 : https://gopkg.in/alecthomas/kingpin.v2
- Windows Named Pipe : https://github.com/natefinch/npipe
- 網路文章(https://blog.csdn.net/small_qch/article/details/19562661)