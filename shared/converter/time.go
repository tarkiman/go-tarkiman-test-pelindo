package converter

func TimeDaytoSecond(day int) int {
	return day * 24 * 60 * 60
}
func TimeSecondtoDay(second int) int {
	return second / 24 / 60 / 60
}
