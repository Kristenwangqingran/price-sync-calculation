package miscellaneousinner

import (
	"net/http"
	"time"
)

const (
	StrClose         = "close"
	StrKeepAlive     = "keep-alive"
	HeaderConnection = "Connection"
)

func IsConnectionClose(req *http.Request) bool {
	return req.Header.Get(HeaderConnection) == StrClose
}

func ProcessHeader(header http.Header, connectionClose bool) {
	if header.Get("Date") == "" {
		var dateBuf [len(http.TimeFormat)]byte

		appendTime(dateBuf[:0], time.Now())

		header.Set("Date", string(dateBuf[:]))
	}

	if connectionClose {
		header.Set(HeaderConnection, StrClose)
	} else {
		header.Set(HeaderConnection, StrKeepAlive)
	}
}

// appendTime is a non-allocating version of []byte(t.UTC().Format(TimeFormat))
func appendTime(b []byte, t time.Time) []byte {
	const days = "SunMonTueWedThuFriSat"

	const months = "JanFebMarAprMayJunJulAugSepOctNovDec"

	t = t.UTC()
	yy, mm, dd := t.Date()
	hh, mn, ss := t.Clock()
	day := days[3*t.Weekday():]
	mon := months[3*(mm-1):]

	return append(b,
		day[0], day[1], day[2], ',', ' ',
		byte('0'+dd/10), byte('0'+dd%10), ' ',
		mon[0], mon[1], mon[2], ' ',
		byte('0'+yy/1000), byte('0'+(yy/100)%10), byte('0'+(yy/10)%10), byte('0'+yy%10), ' ',
		byte('0'+hh/10), byte('0'+hh%10), ':',
		byte('0'+mn/10), byte('0'+mn%10), ':',
		byte('0'+ss/10), byte('0'+ss%10), ' ',
		'G', 'M', 'T')
}
