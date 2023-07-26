package middleware

import "net/http"

// PrivacyPolicy adds a privacy policy header to the response.
func PrivacyPolicy(uri string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Privacy-Policy", uri)

		next.ServeHTTP(w, r)
	})
}

// TermsOfService adds a terms of service header to the response.
func TermsOfService(uri string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Terms-Of-Service", uri)

		next.ServeHTTP(w, r)
	})
}

// CORS adds CORS headers to the response.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}
