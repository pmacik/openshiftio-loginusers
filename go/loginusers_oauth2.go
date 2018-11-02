package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tebeka/selenium"
)

func main() {
	authServerAddress := getenv("AUTH_SERVER_ADDRESS", "http://localhost:8089")

	usersPropertiesFile := getenv("USERS_PROPERTIES_FILE", "users.properties")

	userTokensFile := getenv("USER_TOKENS_FILE", "user.tokens")
	userTokensIncludeUsername := strings.ToLower(getenv("USER_TOKENS_INCLUDE_USERNAME", "false")) == "true"

	maxUsers := getenv("MAX_USERS", "-1")

	log.SetOutput(os.Stdout)

	ufile, err := os.Open(usersPropertiesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer ufile.Close()

	scanner := bufio.NewScanner(ufile)

	var userNames []string
	var userPasswords []string

	for scanner.Scan() {
		line := scanner.Text()
		credentials := strings.Split(line, "=")
		userNames = append(userNames, credentials[0])
		userPasswords = append(userPasswords, credentials[1])
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(userTokensFile); os.IsExist(err) {
		os.Remove(userTokensFile)
	}
	tfile, err := os.Create(userTokensFile)
	checkErr(err)
	defer tfile.Close()

	w := bufio.NewWriter(tfile)
	defer w.Flush()
	for index, userName := range userNames {
		maxUsersCount, err := strconv.Atoi(maxUsers)
		checkErr(err)
		if maxUsersCount > 0 && index >= maxUsersCount {
			break
		}
		log.Printf("Loggin user %s in", userName)
		tokens, err := LoginUsersOAuth2(authServerAddress, userName, userPasswords[index])
		checkErr(err)
		tokenLine := fmt.Sprintf("%s;%s", tokens.AccessToken, tokens.RefreshToken)
		if userTokensIncludeUsername {
			tokenLine = fmt.Sprintf("%s;%s", tokenLine, userName)
		}

		_, err = w.WriteString(fmt.Sprintf("%s\n", tokenLine))
		checkErr(err)
	}
	//write tokens to user.tokens file

}

// LoginUsersOAuth2 attempts to login into CodeReady Toolchain (former Openshift.io)
func LoginUsersOAuth2(authServerAddress string, userName string, userPassword string) (*Tokens, error) {
	wd, service := initSelenium()

	defer service.Stop()
	defer wd.Quit()

	clientID := getenv("AUTH_CLIENT_ID", "740650a2-9c44-4db5-b067-a3d1b2cd2d01")
	redirectURL := fmt.Sprintf("%s/api/status", authServerAddress)
	state, _ := uuid.NewUUID()

	startURL := fmt.Sprintf("%s/api/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", authServerAddress, clientID, redirectURL, state.String())

	log.Printf("open-login-page...")
	if err := wd.Get(startURL); err != nil {
		return nil, fmt.Errorf("failed to open URL: '%s'", startURL)
	}

	findElementBy(wd, selenium.ByID, "kc-login")

	log.Printf("get-code...")
	sendKeysToElementBy(wd, selenium.ByID, "username", userName)
	elem := findElementBy(wd, selenium.ByID, "password")
	sendKeysToElement(elem, userPassword)
	submitElement(elem)

	wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		currentURL, _ := wd.CurrentURL()
		return strings.Contains(currentURL, state.String()), nil
	}, 10000)

	currentURL, _ := wd.CurrentURL()
	u, err := url.Parse(currentURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: '%s'", currentURL)
	}
	code := u.Query()["code"]
	log.Printf("get-token...")
	resp, err := http.PostForm(
		fmt.Sprintf("%s/api/token", authServerAddress),
		url.Values{
			"grant_type":   {"authorization_code"},
			"client_id":    {clientID},
			"code":         code,
			"redirect_uri": {redirectURL},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %s", err)
	}

	var tokens Tokens

	json.Unmarshal(body, &tokens)
	log.Printf("done...")
	return &tokens, nil
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func getenv(key string, defaultValue string) string {
	value, found := os.LookupEnv(key)
	if found {
		return value
	}
	return defaultValue
}

func findElementBy(wd selenium.WebDriver, by string, selector string) selenium.WebElement {
	elem, err := wd.FindElement(by, selector)
	checkErr(err)
	return elem
}

func sendKeysToElementBy(wd selenium.WebDriver, by string, selector string, keys string) {
	err := findElementBy(wd, by, selector).SendKeys(keys)
	checkErr(err)
}

func sendKeysToElement(element selenium.WebElement, keys string) {
	err := element.SendKeys(keys)
	checkErr(err)
}

func submitElement(element selenium.WebElement) {
	err := element.Submit()
	checkErr(err)
}

func initSelenium() (selenium.WebDriver, *selenium.Service) {
	chromeDriverPath := getenv("CHROMEDRIVER_BINARY", "chromedriver")
	chromeDriverPort := getenv("CHROMEDRIVER_PORT", "9515")

	port, err := strconv.Atoi(chromeDriverPort)
	checkErr(err)

	service, err := selenium.NewChromeDriverService(chromeDriverPath, port)
	checkErr(err)

	chromeOptions := map[string]interface{}{
		"args": []string{
			"--headless",
			"--window-size=1920,1080",
			"--window-position=0,0",
		},
	}

	caps := selenium.Capabilities{
		"browserName":   "chrome",
		"chromeOptions": chromeOptions,
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	checkErr(err)
	return wd, service
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
