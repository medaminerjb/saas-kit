package provider

import (
	"html/template"
	"net/http"

	"github.com/zitadel/oidc/v3/pkg/op"
)

// loginTmpl is a minimal, production-replaceable login page.
var loginTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sign In — SaaSKit</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Inter', system-ui, -apple-system, sans-serif;
    background: #0f172a;
    color: #e2e8f0;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .card {
    background: #1e293b;
    border: 1px solid #334155;
    border-radius: 16px;
    padding: 2.5rem;
    width: 100%;
    max-width: 420px;
    box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5);
  }
  .logo {
    text-align: center;
    margin-bottom: 1.5rem;
  }
  .logo h1 {
    font-size: 1.5rem;
    font-weight: 700;
    background: linear-gradient(135deg, #6366f1, #8b5cf6, #a78bfa);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .logo p { color: #94a3b8; font-size: 0.875rem; margin-top: 0.25rem; }
  .form-group { margin-bottom: 1.25rem; }
  label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    margin-bottom: 0.5rem;
    color: #cbd5e1;
  }
  input[type="email"],
  input[type="password"] {
    width: 100%;
    padding: 0.75rem 1rem;
    border: 1px solid #475569;
    border-radius: 8px;
    background: #0f172a;
    color: #f1f5f9;
    font-size: 0.95rem;
    outline: none;
    transition: border-color 0.2s;
  }
  input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.15); }
  .btn {
    width: 100%;
    padding: 0.75rem;
    border: none;
    border-radius: 8px;
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    color: #fff;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.2s;
  }
  .btn:hover { opacity: 0.9; }
  .error {
    background: #450a0a;
    border: 1px solid #7f1d1d;
    color: #fca5a5;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    margin-bottom: 1rem;
    font-size: 0.875rem;
  }
  .footer {
    text-align: center;
    margin-top: 1.5rem;
    font-size: 0.75rem;
    color: #64748b;
  }
</style>
</head>
<body>
<div class="card">
  <div class="logo">
    <h1>SaaSKit</h1>
    <p>Sign in to continue</p>
  </div>
  {{if .Error}}
  <div class="error">{{.Error}}</div>
  {{end}}
  <form method="POST" action="/oidc/login?authRequestID={{.AuthRequestID}}">
    <input type="hidden" name="authRequestID" value="{{.AuthRequestID}}">
    <div class="form-group">
      <label for="email">Email</label>
      <input type="email" id="email" name="email" required autocomplete="email" placeholder="you@example.com" {{if .LoginHint}}value="{{.LoginHint}}"{{end}}>
    </div>
    <div class="form-group">
      <label for="password">Password</label>
      <input type="password" id="password" name="password" required autocomplete="current-password" placeholder="••••••••">
    </div>
    <button type="submit" class="btn">Sign in</button>
  </form>
  <div class="footer">Powered by SaaSKit Identity Platform</div>
</div>
</body>
</html>`))

type loginData struct {
	AuthRequestID string
	LoginHint     string
	Error         string
}

// LoginHandler renders the login form.
func LoginHandler(storage *Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRequestID := r.URL.Query().Get("authRequestID")
		if authRequestID == "" {
			http.Error(w, "missing authRequestID", http.StatusBadRequest)
			return
		}

		data := loginData{
			AuthRequestID: authRequestID,
			LoginHint:     r.URL.Query().Get("login_hint"),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginTmpl.Execute(w, data)
	}
}

// LoginSubmitHandler processes the login form submission.
func LoginSubmitHandler(storage *Storage, provider *op.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		authRequestID := r.FormValue("authRequestID")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if err := storage.CheckUsernamePassword(r.Context(), email, password, authRequestID); err != nil {
			data := loginData{
				AuthRequestID: authRequestID,
				LoginHint:     email,
				Error:         "Invalid email or password",
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_ = loginTmpl.Execute(w, data)
			return
		}

		// Authentication succeeded — redirect back to the OIDC callback
		http.Redirect(w, r, "/oidc/authorize/callback?id="+authRequestID, http.StatusFound)
	}
}
