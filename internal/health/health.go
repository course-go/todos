package health

type Health string

const (
	OK    Health = "OK"
	WARN  Health = "WARN"
	ERROR Health = "ERROR"
)
