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
		redirectURL       = "http://" + *addr + "/facebook"
		fbPermissionsList = strings.Split(*fbPermissions, ",")
	)
	http.Handle(
		"/", homeHandler(*fbClientID, *fbClientSecret, redirectURL, fbPermissionsList),
	)
	http.Handle("/facebook", facebookHandler())

	log.Printf("Facebook login will be redirected to: %s", redirectURL)
	log.Printf("Server listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func homeHandler(
	fbClientID, fbClientSecret, redirectURL string, fbPermissions []string,
) http.Handler {
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
		<a href="%s">Login with Facebook</a>
	</body>
</html>
`

	var (
		fbConfig = oauth2.Config{
			ClientID:     fbClientID,
			ClientSecret: fbClientSecret,
			RedirectURL:  redirectURL,
			Scopes:       fbPermissions,
			Endpoint:     facebook.Endpoint,
		}
		csrfToken = "should-be-secure-token-for-csrf-attakcs-protection"
	)

	var h = func(w http.ResponseWriter, r *http.Request) {
		var loginURL = fbConfig.AuthCodeURL(csrfToken)

		if _, err := w.Write([]byte(fmt.Sprintf(bodyHTML, loginURL))); err != nil {
			log.Printf("Error when writing Home page response body. Err= %+v", err)
		}
	}

	return http.HandlerFunc(h)
}

func facebookHandler() http.Handler {
	type bodyData struct {
		RequestURL     string
		RequestHeaders http.Header
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
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(fmt.Sprintf(errorBodyHTML, err))); err != nil {
				log.Printf("Error when writing page response body. Err= %+v", err)
			}
			return
		}

		bodyHTMLTpl.Execute(&pageBody, bodyData{
			RequestURL:     r.URL.String(),
			RequestHeaders: r.Header,
		})

		if _, err := w.Write(pageBody.Bytes()); err != nil {
			log.Printf("Error when writing page response body. Err= %+v", err)
		}
	}

	return http.HandlerFunc(h)
}
