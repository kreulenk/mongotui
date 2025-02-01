package cmd

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type baseOptions struct {
	host string
	port int
}

type authenticationOptions struct {
	username                     string
	password                     string
	authenticationDatabase       string
	authenticationMechanism      string
	awsIamSessionToken           string
	gssApiServiceName            string
	sspiHostnameCanonicalization string
	sspiRealmOverride            string
}

type tlsOptions struct {
	tls                           bool
	tlsCertificateKeyFile         string
	tlsCertificateKeyFilePassword string
	tlsCAFile                     string
	tlsAllowInvalidHostnames      bool
	tlsAllowInvalidCertificates   bool
	tlsCertificateSelector        string
	tlsCRLFile                    string
	tlsDisabledProtocols          string
	tlsFIPSMode                   string
}

type flagOptions struct {
	baseOptions           baseOptions
	authenticationOptions authenticationOptions
	tlsOptions            tlsOptions
}

func applyHostConfig(clientOps *options.ClientOptions, flags baseOptions) {
	if flags.host != "" {
		if flags.port != 0 {
			clientOps.SetHosts([]string{fmt.Sprintf("%s:%d", flags.host, &flags.port)})
		}
		clientOps.SetHosts([]string{fmt.Sprintf("%s:27017", flags.host)})
	}
}

func applyAuthConfig(clientOps *options.ClientOptions, flags authenticationOptions) {
	if clientOps.Auth == nil && (flags.username != "" || flags.password != "" || flags.authenticationDatabase != "" || flags.authenticationMechanism != "") {
		clientOps.Auth = &options.Credential{}
	} else if clientOps.Auth == nil {
		return
	}
	if flags.username != "" {
		clientOps.Auth.Username = flags.username
	}
	if flags.password != "" {
		clientOps.Auth.Password = flags.password
	}
	if flags.authenticationDatabase != "" {
		clientOps.Auth.AuthMechanism = flags.authenticationMechanism
	}
	if flags.authenticationMechanism != "" {
		clientOps.Auth.AuthMechanism = flags.authenticationMechanism
	}
	if flags.awsIamSessionToken != "" {
		clientOps.Auth.AuthMechanismProperties["AWS_SESSION_TOKEN"] = flags.awsIamSessionToken
	}
	if flags.gssApiServiceName != "" {
		clientOps.Auth.AuthMechanismProperties["SERVICE_NAME"] = flags.gssApiServiceName
	}
	// flag description in mongosh reads 'Specify the SSPI hostname canonicalization (none or forward, available on Windows)'
	if flags.sspiHostnameCanonicalization == "forward" {
		clientOps.Auth.AuthMechanismProperties["CANONICALIZE_HOST_NAME"] = "true"
	}
	if flags.sspiRealmOverride != "" {
		clientOps.Auth.AuthMechanismProperties["SERVICE_REALM"] = flags.sspiRealmOverride
	}
}
