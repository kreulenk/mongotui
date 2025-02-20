package cmd

import (
	"fmt"
	"github.com/kreulenk/mongotui/internal/build"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/kreulenk/mongotui/pkg/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
	"runtime"
	"slices"
	"strings"
)

func genRootCmd() *cobra.Command {
	flags := flagOptions{
		baseOptions:           baseOptions{},
		authenticationOptions: authenticationOptions{},
		tlsOptions:            tlsOptions{},
		apiVersionOptions:     apiVersionOptions{},
		fleOptions:            fleOptions{},
		//oidcOptions:           oidcOptions{},
	}

	var cmd = &cobra.Command{
		Use:   "mongotui <connection-string>",
		Short: "A MongoDB Terminal User Interface",
		Long:  `mongotui is a MongoDB Terminal User Interface`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Verify that a host has been provided
			if len(args) != 1 && flags.baseOptions.host == "" && !flags.baseOptions.version {
				return fmt.Errorf("you must provide a valid hostname")
			}
			// Verify that authenticationMechanism is a supported value if provided
			validAuthMechs := []string{"", "SCRAM-SHA-1", "SCRAM-SHA-256", "MONGODB-X509", "GSSAPI", "PLAIN", "MONGODB-OIDC", "MONGODB-AWS"}
			if ok := slices.Contains(validAuthMechs, flags.authenticationOptions.authenticationMechanism); !ok {
				return fmt.Errorf("invalid authenticationMechanism of %s provided. Must be one of %v", flags.authenticationOptions.authenticationMechanism, validAuthMechs[1:])
			}

			validSspiHostnameCanonicalization := []string{"", "forward", "none"}
			if ok := slices.Contains(validSspiHostnameCanonicalization, flags.authenticationOptions.sspiHostnameCanonicalization); !ok {
				return fmt.Errorf("invalid --validSspiHostnameCanonicalization of %s provided. Must be one of %v",
					flags.authenticationOptions.sspiHostnameCanonicalization, validSspiHostnameCanonicalization[1:],
				)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if flags.baseOptions.version {
				fmt.Printf("%s\n%s_%s\n%s\n", build.Version, runtime.GOOS, runtime.GOARCH, build.SHA)
				return
			}

			clientOps := options.Client()
			clientOps.SetTimeout(mongoengine.Timeout)
			if len(args) == 1 {
				connectionString := args[0]
				if !strings.Contains(connectionString, "://") {
					connectionString = "mongodb://" + connectionString
				}
				clientOps.ApplyURI(connectionString) // May or may not be set
			}

			applyHostConfig(clientOps, flags.baseOptions)
			applyAuthConfig(clientOps, flags.authenticationOptions)
			err := applyTlsConfig(clientOps, flags.tlsOptions)
			cobra.CheckErr(err)
			applyApiVersionConfig(clientOps, flags.apiVersionOptions)
			err = applyFleConfig(clientOps, flags.fleOptions)
			cobra.CheckErr(err)

			client, err := mongo.Connect(clientOps)
			cobra.CheckErr(err)
			tui.Initialize(client)
		},
	}

	var flagSets []namedFlagSet
	// First group of flags when running mongosh --help
	baseFlags := pflag.NewFlagSet("base", pflag.ExitOnError)
	baseFlags.StringVar(&flags.baseOptions.host, "host", "", "Server to connect to")
	baseFlags.IntVar(&flags.baseOptions.port, "port", 0, "Port to connect to")
	baseFlags.BoolVar(&flags.baseOptions.version, "version", false, "Show version information")
	flagSets = append(flagSets, namedFlagSet{name: "Options", flagset: baseFlags})

	authenticationFlags := pflag.NewFlagSet("authentication", pflag.ExitOnError)
	authenticationFlags.StringVarP(&flags.authenticationOptions.username, "username", "u", "", "Username for authentication")
	authenticationFlags.StringVarP(&flags.authenticationOptions.password, "password", "p", "", "Password for authentication")
	authenticationFlags.StringVar(&flags.authenticationOptions.authenticationDatabase, "authenticationDatabase", "", "User source (defaults to dbname)")
	authenticationFlags.StringVar(&flags.authenticationOptions.authenticationMechanism, "authenticationMechanism", "", "Authentication mechanism to use")
	authenticationFlags.StringVar(&flags.authenticationOptions.awsIamSessionToken, "awsIamSessionToken", "", "AWS IAM Temporary Session Token ID")
	authenticationFlags.StringVar(&flags.authenticationOptions.gssApiServiceName, "gssapiServiceName", "", "Service name to use when authenticating using GSSAPI/Kerberos")
	authenticationFlags.StringVar(&flags.authenticationOptions.sspiHostnameCanonicalization, "sspiHostnameCanonicalization", "", "Specify the SSPI hostname canonicalization (none or forward, available on Windows)")
	authenticationFlags.StringVar(&flags.authenticationOptions.sspiRealmOverride, "sspiRealmOverride", "", "Specify the SSPI server realm (available on Windows)")
	flagSets = append(flagSets, namedFlagSet{name: "Authentication Options", flagset: authenticationFlags})

	tlsFlags := pflag.NewFlagSet("tls", pflag.ExitOnError)
	tlsFlags.BoolVar(&flags.tlsOptions.tls, "tls", false, "Use TLS for all connections")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCertificateKeyFile, "tlsCertificateKeyFile", "", "PEM certificate/key file for TLS")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCertificateKeyFilePassword, "tlsCertificateKeyFilePassword", "", "Password for key in PEM file for TLS")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCAFile, "tlsCAFile", "", "Certificate Authority file for TLS")
	tlsFlags.BoolVar(&flags.tlsOptions.tlsAllowInvalidHostnames, "tlsAllowInvalidHostnames", false, "Allow connections to servers with non-matching hostnames")
	tlsFlags.BoolVar(&flags.tlsOptions.tlsAllowInvalidCertificates, "tlsAllowInvalidCertificates", false, "Allow connections to servers with invalid certificates")
	//tlsFlags.StringVar(&flags.tlsOptions.tlsCertificateSelector, "tlsCertificateSelector", "", "TLS Certificate in system store (Windows and macOS only)")
	//tlsFlags.StringVar(&flags.tlsOptions.tlsCRLFile, "tlsCRLFile", "", "Specifies the .pem file that contains the Certificate Revocation List")
	//tlsFlags.StringVar(&flags.tlsOptions.tlsDisabledProtocols, "tlsDisabledProtocols", "", "Comma separated list of TLS protocols to disable [TLS1_0,TLS1_1,TLS1_2]")
	//tlsFlags.StringVar(&flags.tlsOptions.tlsFIPSMode, "tlsFIPSMode", "", "Enable the system TLS library's FIPS mode")
	//flagSets = append(flagSets, namedFlagSet{name: "TLS Options", flagset: tlsFlags})
	flagSets = append(flagSets, namedFlagSet{name: "TLS Options", flagset: tlsFlags})

	apiVersionFlags := pflag.NewFlagSet("apiVersion", pflag.ExitOnError)
	apiVersionFlags.StringVar((*string)(&flags.apiVersionOptions.apiVersion), "apiVersion", "", "Specifies the API version to connect with")
	apiVersionFlags.BoolVar(&flags.apiVersionOptions.apiStrict, "apiStrict", false, "Use strict API version mode")
	apiVersionFlags.BoolVar(&flags.apiVersionOptions.apiDeprecationErrors, "apiDeprecationErrors", false, "Fail deprecated commands for the specified API version")
	flagSets = append(flagSets, namedFlagSet{name: "API version Options", flagset: apiVersionFlags})

	fleFlags := pflag.NewFlagSet("fle", pflag.ExitOnError)
	fleFlags.StringVar(&flags.fleOptions.awsAccessKeyId, "awsAccessKeyId", "", "AWS Access Key for FLE Amazon KMS")
	fleFlags.StringVar(&flags.fleOptions.awsSecretAccessKey, "awsSecretAccessKey", "", "AWS Secret Key for FLE Amazon KMS")
	fleFlags.StringVar(&flags.fleOptions.awsSessionToken, "awsSessionToken", "", "Optional AWS Session Token ID")
	fleFlags.StringVar(&flags.fleOptions.keyVaultNamespace, "keyVaultNamespace", "", "database.collection to store encrypted FLE parameters")
	//fleFlags.StringVar(&flags.fleOptions.kmsURL, "kmsURL", "", "Test parameter to override the URL of the KMS endpoint")
	flagSets = append(flagSets, namedFlagSet{name: "FLE Options", flagset: fleFlags})

	//oidcFlags := pflag.NewFlagSet("oidc", pflag.ExitOnError)
	//oidcFlags.StringVar(&flags.oidcOptions.oidcFlows, "oidcFlows", "", "Supported OIDC auth flows [auth-code,device-auth]")
	//oidcFlags.StringVar(&flags.oidcOptions.oidcRedirectUri, "oidcRedirectUri", "http://localhost:27097/redirect", "Local auth code flow redirect URL")
	//oidcFlags.BoolVar(&flags.oidcOptions.oidcTrustedEndpoint, "oidcTrustedEndpoint", false, "Treat the cluster/database mongosh as a trusted endpoint")
	//oidcFlags.BoolVar(&flags.oidcOptions.oidcIdTokenAsAccessToken, "oidcIdTokenAsAccessToken", false, "Use ID tokens in place of access tokens for auth")
	//oidcFlags.StringVar(&flags.oidcOptions.oidcDumpTokens, "oidcDumpTokens", "", "Debug OIDC by printing tokens to mongosh's output [full|include-secrets]")
	//oidcFlags.BoolVar(&flags.oidcOptions.oidcNoNonce, "oidcNoNonce", false, "Don't send a nonce argument in the OIDC auth request")
	//flagSets = append(flagSets, namedFlagSet{name: "OIDC auth options:", flagset: oidcFlags})

	addFlagsAndSetHelpMenu(cmd, flagSets)
	return cmd
}

func Execute() {
	rootCmd := genRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
