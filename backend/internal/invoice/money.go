package invoice

type Money int64

func RoundedVAT(net Money, ratePercent int64) Money {
	return Money((int64(net)*ratePercent + 50) / 100)
}
