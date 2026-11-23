package utils

import "time"

// START OF DAY RETURNS 00:00:00 OF t IN t'S LOCATION
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// END OF DAY EXCLUSIVE RETURNS THE START OF THE NEXT DAY, USE AS UPPER BOUND WITH "<"
func EndOfDayExclusive(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
}
