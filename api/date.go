// package handler

// import (
// 	"fmt"
// 	"net/http"
// 	"time"
// )

// func Date(w http.ResponseWriter, r *http.Request) {
// 	currentTime := time.Now().Format(time.RFC850)
// 	fmt.Fprint(w, currentTime)
// }


package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func GetFileContent(w http.ResponseWriter, r *http.Request) {

	proxyURL, err := url.Parse("http://x:x@14.188.187.234:10953")
	if err != nil {
		http.Error(w, "Invalid proxy URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", "https://1fichier.com/?bopqkt2nz13mv3fvk3n0", nil)
	if err != nil {
		fmt.Fprint(w, "Error creating request: ", err)
		return
	}

	req.Header.Set("Cookie", "show_cm=no; SID=EesWwrKekPA0Koh14igcheGxfEqNSZQwrrcwVxMpotRpQt9=uWK9eX3l4o2KIaym")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprint(w, "Error sending request: ", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, "Error reading response: ", err)
		return
	}

	fmt.Fprint(w, string(body))
}

