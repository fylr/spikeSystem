package main

import (
	"net/http"
	"os"
	"runtime"
	"spikeSystem/localSpike"
	"spikeSystem/remoteSpike"
	"spikeSystem/util"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

var (
	lSpike    localSpike.LocalSpike
	rSpike    remoteSpike.RemoteSpikeKeys
	redisPool *redis.Pool
	logFile   *os.File
	done      chan int
)

//初始化要使用的结构体和redis连接池
func init() {
	lSpike = localSpike.LocalSpike{
		LocalInStock:     150,
		LocalSalesVolume: 0,
	}
	rSpike = remoteSpike.RemoteSpikeKeys{
		SpikeOrderHashKey:  "ticket_hash_key",
		TotalInventoryKey:  "ticket_total_nums",
		QuantityOfOrderKey: "ticket_sold_nums",
	}
	redisPool = remoteSpike.NewPool()
	done = make(chan int, runtime.NumCPU())
	done <- 1
}

func main() {
	fd, err := os.OpenFile("./stat_3001.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err.Error())
	} else {
		logFile = fd
		defer logFile.Close()
	}

	http.HandleFunc("/buy/ticket", handleReq)
	http.ListenAndServe(":3001", nil)
}

//处理请求函数,根据请求将响应结果信息写入日志
func handleReq(w http.ResponseWriter, r *http.Request) {
	redisConn := redisPool.Get()
	success_flag := false
	//全局读写锁
	<-done
	if lSpike.LocalDeductionStock() && rSpike.RemoteDeductionStock(redisConn) {
		success_flag = true
	}
	done <- 1
	//将抢票状态写入到log中
	LogMsg := ""
	if success_flag {
		util.RespJson(w, 1, "抢票成功", nil)
		LogMsg += "result:1, localSales:" + strconv.FormatInt(lSpike.LocalSalesVolume, 10) + "\n"
	} else {
		util.RespJson(w, -1, "已售罄", nil)
		LogMsg += "result:0, localSales:" + strconv.FormatInt(lSpike.LocalSalesVolume, 10) + "\n"
	}
	logFile.WriteString(LogMsg)
}
