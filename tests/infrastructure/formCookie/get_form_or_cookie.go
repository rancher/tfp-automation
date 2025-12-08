package formCookie

import "net/http"

// GetFormOrCookie returns the value from the form if available. Otherwise, retrieves from the cookie
func GetFormOrCookie(r *http.Request, name string) string {
	val := r.FormValue(name)

	if val == "" {
		c, err := r.Cookie(name)
		if err == nil {
			val = c.Value
		}
	}

	return val
}
