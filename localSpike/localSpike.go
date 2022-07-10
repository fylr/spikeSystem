package localSpike

import "time"

type LocalSpike struct {
	LocalInStock     int64
	LocalSalesVolume int64
}

//本地扣库存,返回bool值
func (spike *LocalSpike) LocalDeductionStock(done chan int) bool {
	//全局写锁
	<-done
	spike.LocalSalesVolume = spike.LocalSalesVolume + 1
	time.Sleep(time.Nanosecond)
	done <- 1
	return spike.LocalSalesVolume <= spike.LocalInStock
}
