package core

import (
	"strconv"

	"github.com/rkapps/go_finance/store"
	"github.com/rkapps/go_finance/utils"
)

type qRSI struct {
	period int
	sgain  float64
	sloss  float64
	again  float64
	aloss  float64
	rs     float64
	rsi    float64
}

type qSMA struct {
	period int
	sprice float64
	sma    float64
}

type qEMA struct {
	period     int
	multiplier float64
	sprice     float64
	ema        float64
}

func updateRSI(tha []*store.TickerHistory) {

	// log.Println(len(tha))
	var ph *store.TickerHistory
	var diff, gain, loss float64

	var qrs []*qRSI
	for _, period := range store.RSIPeriods {
		qrs = append(qrs, &qRSI{period, 0, 0, 0, 0, 0, 0})
	}

	for x, th := range tha {

		if x == 0 {
			continue
		}
		ph = tha[x-1]
		diff = th.Close - ph.Close

		if diff > 0 {
			gain = diff
			loss = 0
		} else {
			gain = 0
			loss = diff * -1
		}

		th.RSI = make(map[string]float64)
		// if x < 6 {
		// 	log.Printf("x: %d close: %f pclose: %f diff: %f", x, th.Close, ph.Close, diff)
		// }

		for _, qr := range qrs {

			var period float64
			period = float64(qr.period)
			// if x < 3 {
			// 	log.Printf("period: %f sgain: %f sloss: %f", period, rcalc.sgain, rcalc.sloss)
			// }
			if x < qr.period {
				qr.sgain += gain
				qr.sloss += loss
			} else {
				if x == qr.period {
					qr.again = qr.sgain / period
					qr.aloss = qr.sloss / period
				} else {
					qr.again = ((qr.again * (period - 1)) + gain) / period
					qr.aloss = ((qr.aloss * (period - 1)) + loss) / period
				}

				qr.rs = qr.again / qr.aloss
				if qr.aloss == 0 {
					qr.rsi = 100
				} else {
					qr.rsi = 100 - (100 / (1 + qr.rs))
				}

				th.RSI[strconv.Itoa(qr.period)] = utils.RoundUp(qr.rsi)
			}

			// if x == len(tha)-1 {
			// 	log.Printf("period: %d gain: %f loss : %f sgain: %f sloss: %f rsi: %f", qr.period, gain, loss, qr.sgain, qr.sloss, qr.rsi)
			// 	log.Printf("RSI: %v", th.RSI)
			// }
		}

	}

}

func updateMAs(tha []*store.TickerHistory) {

	var qss []*qSMA
	for _, period := range store.SMAPeriods {
		qss = append(qss, &qSMA{period, 0, 0})
	}

	var qes []*qEMA
	for _, period := range store.EMAPeriods {
		qes = append(qes, &qEMA{period, 0, 0, 0})
	}

	for _, qe := range qes {
		var period float64
		period = float64(qe.period)
		qe.multiplier = 2 / (period + 1)
	}

	for x, th := range tha {

		th.SMA = make(map[string]float64)
		th.EMA = make(map[string]float64)

		for _, qs := range qss {

			var period float64
			period = float64(qs.period)

			if x < (qs.period - 1) {
				qs.sprice += th.Close
			} else {
				qs.sprice += th.Close
				qs.sma = qs.sprice / period
				qs.sprice = qs.sprice - tha[x-qs.period+1].Close
				th.SMA[strconv.Itoa(qs.period)] = formatDec(qs.sma)
			}
		}

		for _, qe := range qes {
			var speriod = strconv.Itoa(qe.period)
			var period float64
			period = float64(qe.period)
			if x < (qe.period - 1) {
				qe.sprice += th.Close
			} else if x == (qe.period - 1) {
				qe.sprice += th.Close
				qe.ema = qe.sprice / period
				th.EMA[speriod] = formatDec(qe.ema)
			} else {
				var pema = tha[x-1].EMA[speriod]
				qe.ema = (th.Close-pema)*qe.multiplier + pema
				th.EMA[speriod] = formatDec(qe.ema)
			}

		}

		// if x == len(tha)-1 {
		// 	log.Printf("SMA: %v", th.SMA)
		// 	log.Printf("EMA: %v", th.EMA)
		// }
	}
}

func formatDec(value float64) float64 {
	if value < 1 {
		return utils.ToFixed(value, 4)
	} else {
		return utils.ToFixed(value, 2)
	}
}
