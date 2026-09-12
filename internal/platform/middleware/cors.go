package middleware

import "net/http"

// CORS lets a browser-based client (a different origin from this
// API — e.g. the app's own web target) actually reach it. Without
// this, every browser fetch() call is silently blocked before it
// even leaves the browser; curl and the native mobile app's fetch
// aren't affected at all, since CORS is a browser-only enforcement
// mechanism — this is exactly what made the gap easy to miss until
// actually testing the web target.
//
// Allow-Origin: * is safe here specifically because auth is a Bearer
// token a client must deliberately read and attach, not a cookie a
// browser sends automatically on every request — there's no ambient
// credential for a different origin to silently ride along on.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		// A CORS "preflight" — the browser sends this automatically
		// before certain cross-origin requests, with no Authorization
		// header of its own, purely to check whether the real request
		// would be allowed. Must be answered here, before Auth ever
		// runs — a preflight can never carry a token, so reaching Auth
		// would always mean a 401 and a broken CORS handshake.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
