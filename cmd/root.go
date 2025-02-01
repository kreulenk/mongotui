package cmd

import (
	"fmt"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/kreulenk/mongotui/pkg/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
	"slices"
	"strings"
)

func genRootCmd() *cobra.Command {
	flags := flagOptions{
		baseOptions:           baseOptions{},
		authenticationOptions: authenticationOptions{},
		tlsOptions:            tlsOptions{},
	}

	var cmd = &cobra.Command{
		Use:   "mtui <db-address>",
		Short: "A MongoDB Terminal User Interface",
		Long:  `mongotui is a MongoDB Terminal User Interface`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Verify that a host has been provided
			if len(args) != 1 && flags.baseOptions.host == "" {
				return fmt.Errorf("you must provide a valid hostname")
			}
			// Verify that authenticationMechanism is a supported value if provided
			validAuthMechs := []string{"", "SCRAM-SHA-1", "SCRAM-SHA-256", "MONGODB-X509", "GSSAPI", "PLAIN"}
			if ok := slices.Contains(validAuthMechs, flags.authenticationOptions.authenticationMechanism); !ok {
				return fmt.Errorf("invalid authenticationMechanism of %s provided. Must be one of %v", flags.authenticationOptions.authenticationMechanism, validAuthMechs[1:])
			}

			validSspiHostnameCanonicalization := []string{"", "forward", "none"}
			if ok := slices.Contains(validSspiHostnameCanonicalization, flags.authenticationOptions.sspiHostnameCanonicalization); !ok {
				return fmt.Errorf("invalid validSspiHostnameCanonicalization of %s provided. Must be one of %v",
					flags.authenticationOptions.sspiHostnameCanonicalization, validSspiHostnameCanonicalization[1:],
				)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			clientOps := options.Client()
			clientOps.SetTimeout(mongoengine.Timeout)
			if len(args) == 1 {
				dbAddress := args[0]
				if !strings.Contains(dbAddress, "://") {
					dbAddress = "mongodb://" + dbAddress
					clientOps.ApplyURI(dbAddress) // May or may not be set
				}
			}

			applyHostConfig(clientOps, flags.baseOptions)
			applyAuthConfig(clientOps, flags.authenticationOptions)

			client, err := mongo.Connect(clientOps)
			cobra.CheckErr(err)
			tui.Initialize(client)
		},
	}

	var flagSets []namedFlagSet
	// First group of flags when running mongosh --help
	regularFlags := pflag.NewFlagSet("regularFlags", pflag.ExitOnError)
	regularFlags.StringVar(&flags.baseOptions.host, "host", "", "Server to connect to")
	regularFlags.IntVar(&flags.baseOptions.port, "port", 0, "Port to connect to")
	flagSets = append(flagSets, namedFlagSet{name: "Options", flagset: regularFlags})

	authenticationFlags := pflag.NewFlagSet("authenticationFlags", pflag.ExitOnError)
	authenticationFlags.StringVarP(&flags.authenticationOptions.username, "username", "u", "", "Username for authentication")
	authenticationFlags.StringVarP(&flags.authenticationOptions.password, "password", "p", "", "Password for authentication")
	authenticationFlags.StringVar(&flags.authenticationOptions.authenticationDatabase, "authenticationDatabase", "", "User source (defaults to dbname)")
	authenticationFlags.StringVar(&flags.authenticationOptions.authenticationMechanism, "authenticationMechanism", "", "Authentication mechanism to use")
	authenticationFlags.StringVar(&flags.authenticationOptions.awsIamSessionToken, "awsIamSessionToken", "", "AWS IAM Temporary Session Token ID")
	authenticationFlags.StringVar(&flags.authenticationOptions.gssApiServiceName, "gssapiServiceName", "", "Service name to use when authenticating using GSSAPI/Kerberos")
	authenticationFlags.StringVar(&flags.authenticationOptions.sspiHostnameCanonicalization, "sspiHostnameCanonicalization", "", "Specify the SSPI hostname canonicalization (none or forward, available on Windows)")
	authenticationFlags.StringVar(&flags.authenticationOptions.sspiRealmOverride, "sspiRealmOverride", "", "Specify the SSPI server realm (available on Windows)")
	flagSets = append(flagSets, namedFlagSet{name: "Authentication Options", flagset: authenticationFlags})

	tlsFlags := pflag.NewFlagSet("tlsFlags", pflag.ExitOnError)
	tlsFlags.BoolVar(&flags.tlsOptions.tls, "tls", false, "Use TLS for all connections")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCertificateKeyFile, "tlsCertificateKeyFile", "", "PEM certificate/key file for TLS")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCertificateKeyFilePassword, "tlsCertificateKeyFilePassword", "", "Password for key in PEM file for TLS")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCAFile, "tlsCAFile", "", "Certificate Authority file for TLS")
	tlsFlags.BoolVar(&flags.tlsOptions.tlsAllowInvalidHostnames, "tlsAllowInvalidHostnames", false, "Allow connections to servers with non-matching hostnames")
	tlsFlags.BoolVar(&flags.tlsOptions.tlsAllowInvalidCertificates, "tlsAllowInvalidCertificates", false, "Allow connections to servers with invalid certificates")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCertificateSelector, "tlsCertificateSelector", "", "TLS Certificate in system store (Windows and macOS only)")
	tlsFlags.StringVar(&flags.tlsOptions.tlsCRLFile, "tlsCRLFile", "", "Specifies the .pem file that contains the Certificate Revocation List")
	tlsFlags.StringVar(&flags.tlsOptions.tlsDisabledProtocols, "tlsDisabledProtocols", "", "Comma separated list of TLS protocols to disable [TLS1_0,TLS1_1,TLS1_2]")
	tlsFlags.StringVar(&flags.tlsOptions.tlsFIPSMode, "tlsFIPSMode", "", "Enable the system TLS library's FIPS mode")
	//flagSets = append(flagSets, namedFlagSet{name: "TLS Options", flagset: tlsFlags})

	addFlagsAndSetHelpMenu(cmd, flagSets)

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the genRootCmd.
func Execute() {
	rootCmd := genRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
