package errcode

func HCode(num int) int {
	if num > 999 || num < 1 {
		panic("num range must be between 0 to 1000")
	}
	return 200000 + num*100
}
