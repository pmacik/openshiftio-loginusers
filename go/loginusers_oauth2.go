package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pmacik/loginusers-go/common"
	"github.com/pmacik/loginusers-go/loginusers"
)

func main() {
	authServerAddress := common.Getenv("AUTH_SERVER_ADDRESS", "http://localhost:8089")

	usersPropertiesFile := common.Getenv("USERS_PROPERTIES_FILE", "users.properties")

	userTokensFile := common.Getenv("USER_TOKENS_FILE", "user.tokens")
	userTokensIncludeUsername := strings.ToLower(common.Getenv("USER_TOKENS_INCLUDE_USERNAME", "false")) == "true"

	maxUsers := common.Getenv("MAX_USERS", "-1")

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
	common.CheckErr(err)
	defer tfile.Close()

	w := bufio.NewWriter(tfile)
	defer w.Flush()
	for index, userName := range userNames {
		maxUsersCount, err := strconv.Atoi(maxUsers)
		common.CheckErr(err)
		if maxUsersCount > 0 && index >= maxUsersCount {
			break
		}
		log.Printf("Loggin user %s in", userName)
		tokens, err := loginusers.LoginUsersOAuth2(authServerAddress, userName, userPasswords[index])
		common.CheckErr(err)
		tokenLine := fmt.Sprintf("%s;%s", tokens.AccessToken, tokens.RefreshToken)
		if userTokensIncludeUsername {
			tokenLine = fmt.Sprintf("%s;%s", tokenLine, userName)
		}

		_, err = w.WriteString(fmt.Sprintf("%s\n", tokenLine))
		common.CheckErr(err)
	}
}
