package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

func main() {
	var (
		flagSet = flag.NewFlagSet("", flag.ContinueOnError)
		addr    = flagSet.String(
			"addr", "localhost:8000",
			"HTTP server address to listen (host or ip and port are required)",
		)
		fbClientID     = flagSet.String("id", "", "Facebook app client ID")
		fbClientSecret = flagSet.String("secret", "", "Facebook app client sercret")
		fbPermissions  = flagSet.String(
			"permissions", "email,public_profile,user_friends",
			"Facebook requested user permissions. Provide them with a comma separated list",
		)
	)

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Invalid option.\n%+v", err)
		os.Exit(1)
	}

	if *fbClientID == "" || *fbClientSecret == "" {
		fmt.Println(`"id" and "secret" application parameters are required`)
		flagSet.PrintDefaults()
		os.Exit(1)
	}

	var (
		fbConfig = oauth2.Config{
			ClientID:     *fbClientID,
			ClientSecret: *fbClientSecret,
			RedirectURL:  "http://" + *addr + "/facebook",
			Scopes:       strings.Split(*fbPermissions, ","),
			Endpoint:     facebook.Endpoint,
		}
	)
	http.Handle("/", homeHandler(fbConfig))
	http.Handle("/facebook", facebookHandler(fbConfig))

	log.Printf("Facebook login will be redirected to: %s", fbConfig.RedirectURL)
	log.Printf("Server listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func homeHandler(fbConfig oauth2.Config) http.Handler {
	const bodyHTML = `<html>
	<head>
		<title>Facebook login example</title>	
	</head>
	<body>
		<h1>This a simple basic example of login with Facebook</h1>
		<p>It's implemented in Go using
			<a href="https://godoc.org/golang.org/x/oauth2">golang.org/x/oauth2</a>
			package
		</p>
		<p>Facebook doesn't do anything different if it's used
			<prea<cod>oauth2.AccessTypeOnline</code></pre>  or
			<prea<cod>oauth2.AccessTypeOffline</code></pre>, take yourself a look
			<br>
			<a href="%[1]s">Login with Facebook (access type <b>online</b>)</a>
			<br>
			<a href="%[2]s">Login with Facebook (access type <b>offline</b>)</a>
		</p>
	</body>
</html>
`

	var csrfToken = "should-be-secure-token-for-csrf-attakcs-protection"
	var h = func(w http.ResponseWriter, r *http.Request) {
		var loginOnlineURL = fbConfig.AuthCodeURL(csrfToken, oauth2.ApprovalForce)
		var loginOfflineURL = fbConfig.AuthCodeURL(
			csrfToken, oauth2.AccessTypeOffline, oauth2.ApprovalForce,
		)

		if _, err := w.Write(
			[]byte(
				fmt.Sprintf(bodyHTML, loginOnlineURL, loginOfflineURL),
			),
		); err != nil {
			log.Printf("Error when writing Home page response body. Err= %+v", err)
		}
	}

	return http.HandlerFunc(h)
}

func facebookHandler(fbConfig oauth2.Config) http.Handler {
	type bodyData struct {
		RequestURL     string
		RequestHeaders http.Header
		AuthCode       string
		Token          *oauth2.Token
	}

	var bodyHTMLTpl = template.Must(template.New("pageBody").Parse(`<html>
	<head>
		<title>Logged with Facebook</title>	
	</head>
	<body>
		<h1>You have logged with Facebook</h1>
		<p>Facebook sent the following information</p>
		<ul>
			<li>Request URL: <pre>{{.RequestURL}}</pre></li>
			<li>Request Headers:{{range $name, $values := .RequestHeaders}}
					<pre>Header: {{$name}}
{{ range $v := $values}}{{$v}}
{{end}}</pre>
				{{end}}
			</li>
			<li>Token generated from Code ({{.AuthCode}}):
				<ul>
					<li>Access Token: {{.Token.AccessToken}}</li>
					<li>TokenType: {{.Token.TokenType}}</li>
					<li>Refresh Token: {{.Token.RefreshToken}}</li>
					<li>Expiry: {{.Token.Expiry}}</li>
				</ul>
			</li>
		</ul>
	</body>
</html>
`))

	var errorBodyHTML = `<html>
	<head>
		<title>Logged with Facebook</title>	
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>%+v</p>
	</body>
</html>
`

	h := func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			pageBody bytes.Buffer
			fbCode   = r.FormValue("code")
			token    *oauth2.Token
		)

		if token, err = fbConfig.Exchange(r.Context(), fbCode); err != nil {
			if _, err := w.Write([]byte(fmt.Sprintf(errorBodyHTML, err))); err != nil {
				log.Printf("Error when writing page response body. Err= %+v", err)
			}
			return
		}

		bodyHTMLTpl.Execute(&pageBody, bodyData{
			RequestURL:     r.URL.String(),
			RequestHeaders: r.Header,
			AuthCode:       fbCode,
			Token:          token,
		})

		if _, err := w.Write(pageBody.Bytes()); err != nil {
			log.Printf("Error when writing page response body. Err= %+v", err)
		}
	}

	return http.HandlerFunc(h)
}
