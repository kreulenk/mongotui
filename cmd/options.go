package cmd

import (
	"crypto/tls"
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

// Commented out options are options that exist in mongosh but have not been added to mongotui
type tlsOptions struct {
	tls                           bool
	tlsCertificateKeyFile         string
	tlsCertificateKeyFilePassword string
	tlsCAFile                     string
	tlsAllowInvalidHostnames      bool
	tlsAllowInvalidCertificates   bool
	//tlsCertificateSelector        string
	//tlsCRLFile                    string
	//tlsDisabledProtocols          string
	//tlsFIPSMode                   string
}

type apiVersionOptions struct {
	apiVersion           options.ServerAPIVersion
	apiStrict            bool
	apiDeprecationErrors bool
}

type fleOptions struct {
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsSessionToken    string
	keyVaultNamespace  string
	//kmsURL             string
}

type oidcOptions struct {
	oidcFlows                string
	oidcRedirectUri          string
	oidcTrustedEndpoint      bool
	oidcIdTokenAsAccessToken bool
	//oidcDumpTokens           string
	oidcNoNonce bool
}

type flagOptions struct {
	baseOptions           baseOptions
	authenticationOptions authenticationOptions
	tlsOptions            tlsOptions
	apiVersionOptions     apiVersionOptions
	fleOptions            fleOptions
	oidcOptions           oidcOptions
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
	if clientOps.Auth == nil && (flags.username != "" || flags.password != "" || flags.authenticationDatabase != "" ||
		flags.authenticationMechanism != "" || flags.awsIamSessionToken != "" || flags.gssApiServiceName != "" ||
		flags.sspiHostnameCanonicalization != "" || flags.sspiRealmOverride != "") {
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
	// TODO mongo docs imply that there are other options. Will need to investigate
	// https://www.mongodb.com/docs/drivers/node/v5.6/fundamentals/authentication/enterprise-mechanisms/
	if flags.sspiHostnameCanonicalization == "forward" {
		clientOps.Auth.AuthMechanismProperties["CANONICALIZE_HOST_NAME"] = "true"
	}
	if flags.sspiRealmOverride != "" {
		clientOps.Auth.AuthMechanismProperties["SERVICE_REALM"] = flags.sspiRealmOverride
	}
}

func applyTlsConfig(clientOps *options.ClientOptions, flags tlsOptions) error {
	if !flags.tls {
		return nil
	}
	if clientOps.TLSConfig == nil {
		clientOps.TLSConfig = &tls.Config{}
		if clientOps.Auth != nil {
			if clientOps.Auth.AuthMechanism != "" && clientOps.Auth.AuthMechanism != "MONGODB-X509" {
				return fmt.Errorf("--authenticationMechanism must be set to MONGODB-X509 when tls is enabled")
			}
			clientOps.Auth.AuthMechanism = "MONGODB-X509"
		} else {
			clientOps.Auth = &options.Credential{AuthMechanism: "MONGODB-X509"}
		}
	}
	tlsOpts := make(map[string]interface{})
	if flags.tlsCertificateKeyFile != "" {
		tlsOpts["tlsCertificateKeyFile"] = flags.tlsCertificateKeyFile
	}
	if flags.tlsCertificateKeyFilePassword != "" {
		tlsOpts["tlsCertificateKeyFilePassword"] = flags.tlsCertificateKeyFilePassword
	}
	if flags.tlsCAFile != "" {
		tlsOpts["tlsCAFile"] = flags.tlsCAFile
	}
	// We will use the BuildTLSConfig function from the options package to handle the certificate related configuration
	builtConf, err := options.BuildTLSConfig(tlsOpts)
	if err != nil {
		return fmt.Errorf("failed to generate tls configuration: %w", err)
	}

	// Assign the fields that were set during BuildTLSConfig
	clientOps.TLSConfig.MinVersion = builtConf.MinVersion
	clientOps.TLSConfig.Certificates = builtConf.Certificates
	clientOps.TLSConfig.RootCAs = builtConf.RootCAs

	// Not sure that both AllowInvalidHostnames and AllowInvalidCertificates these should map to InsecureSkipVerify
	if flags.tlsAllowInvalidHostnames {
		clientOps.TLSConfig.InsecureSkipVerify = true
	}
	if flags.tlsAllowInvalidCertificates {
		clientOps.TLSConfig.InsecureSkipVerify = true
	}
	return nil
}

func applyApiVersionConfig(clientOps *options.ClientOptions, flags apiVersionOptions) {
	if clientOps.ServerAPIOptions == nil {
		clientOps.ServerAPIOptions = options.ServerAPI(options.ServerAPIVersion1)
	}
	// Mongo currently only has v1 but supporting this flag to resemble mongosh.. Could be useful in the future
	if flags.apiVersion != "" {
		clientOps.ServerAPIOptions.ServerAPIVersion = flags.apiVersion
	}
	if flags.apiStrict {
		clientOps.ServerAPIOptions.Strict = &flags.apiStrict
	}
	if flags.apiDeprecationErrors {
		clientOps.ServerAPIOptions.DeprecationErrors = &flags.apiDeprecationErrors
	}
}

func applyFleConfig(clientOps *options.ClientOptions, flags fleOptions) error {
	if flags.awsAccessKeyId == "" && flags.awsSecretAccessKey == "" && flags.awsSessionToken == "" && flags.keyVaultNamespace == "" {
		return nil
	}
	if clientOps.Auth != nil {
		if clientOps.Auth.AuthMechanism != "" && clientOps.Auth.AuthMechanism != "MONGODB-AWS" {
			return fmt.Errorf("--authenticationMechanism must be set to MONGODB-AWS when FLE options are enabled")
		}
		clientOps.Auth.AuthMechanism = "MONGODB-AWS"
	} else {
		clientOps.Auth = &options.Credential{AuthMechanism: "MONGODB-AWS"}
	}

	if flags.awsAccessKeyId != "" {
		clientOps.Auth.Username = flags.awsAccessKeyId
	}
	if flags.awsSecretAccessKey != "" {
		clientOps.Auth.Password = flags.awsSecretAccessKey
	}
	if flags.awsSessionToken != "" {
		clientOps.Auth.AuthMechanismProperties["AWS_SESSION_TOKEN"] = flags.awsSessionToken
	}
	if flags.keyVaultNamespace != "" {
		clientOps.Auth.AuthSource = flags.keyVaultNamespace
	}

	return nil
}
