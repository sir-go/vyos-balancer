package main

func eh(err error, msg ...string) {
	if err == nil {
		return
	}
	LOG.Panic("panic", "err", err, "msg", msg)
}
