package cmd

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func applyHostConfig(clientOps *options.ClientOptions, hostFlag string, portFlag int) {
	if hostFlag != "" {
		if portFlag != 0 {
			clientOps.SetHosts([]string{fmt.Sprintf("%s:%d", hostFlag, portFlag)})
		}
		clientOps.SetHosts([]string{fmt.Sprintf("%s:27017", hostFlag)})
	}
}

func applyAuthConfig(clientOps *options.ClientOptions, usernameFlag, passwordFlag, authMechanism, authDatabaseFlag string) {
	if clientOps.Auth == nil && (usernameFlag != "" || passwordFlag != "" || authMechanism != "" || authDatabaseFlag != "") {
		clientOps.Auth = &options.Credential{}
	} else if clientOps.Auth == nil {
		return
	}
	if usernameFlag != "" {
		clientOps.Auth.Username = usernameFlag
	}
	if passwordFlag != "" {
		clientOps.Auth.Password = passwordFlag
	}
	if authMechanism != "" {
		clientOps.Auth.AuthMechanism = authMechanism
	}
	if authDatabaseFlag != "" {
		clientOps.Auth.AuthMechanism = authMechanism
	}
}
