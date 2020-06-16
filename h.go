package wsx

type H map[string]interface{}

func HURLPattern(urlPattern string) *H {
	var h = H(make(map[string]interface{}))
	return h.UseURLPattern(urlPattern)
}

func (h *H) UseURLPattern(urlPattern string) *H {
	(*h)["Router-Type"] = "URL_PATTERN"
	(*h)["URL-Pattern-Value"] = urlPattern
	return h
}
