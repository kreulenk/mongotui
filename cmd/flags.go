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

func applyAuthConfig(clientOps *options.ClientOptions, username, password, authDatabase, authMechanism, awsIamSessionToken, gssApiServiceName, sspiHostnameCanonicalization, sspiRealmOverride string) {
	if clientOps.Auth == nil && (username != "" || password != "" || authDatabase != "" || authMechanism != "") {
		clientOps.Auth = &options.Credential{}
	} else if clientOps.Auth == nil {
		return
	}
	if username != "" {
		clientOps.Auth.Username = username
	}
	if password != "" {
		clientOps.Auth.Password = password
	}
	if authDatabase != "" {
		clientOps.Auth.AuthMechanism = authMechanism
	}
	if authMechanism != "" {
		clientOps.Auth.AuthMechanism = authMechanism
	}

	if awsIamSessionToken != "" {
		clientOps.Auth.AuthMechanismProperties["AWS_SESSION_TOKEN"] = awsIamSessionToken
	}
	if gssApiServiceName != "" {
		clientOps.Auth.AuthMechanismProperties["SERVICE_NAME"] = gssApiServiceName
	}
	// flag description in mongosh reads 'Specify the SSPI hostname canonicalization (none or forward, available on Windows)'
	if sspiHostnameCanonicalization == "forward" {
		clientOps.Auth.AuthMechanismProperties["CANONICALIZE_HOST_NAME"] = "true"
	}
	if sspiRealmOverride != "" {
		clientOps.Auth.AuthMechanismProperties["SERVICE_REALM"] = sspiRealmOverride
	}

}
