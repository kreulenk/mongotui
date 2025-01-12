package cmd

import (
	"fmt"
	"github.com/kreulenk/mtui/pkg/mongodata"
	"github.com/kreulenk/mtui/pkg/tui"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func genRootCmd() *cobra.Command {
	var hostFlag string
	var portFlag int
	var usernameFlag string
	var passwordFlag string
	var authMechanismFlag string

	var cmd = &cobra.Command{
		Use:   "mtui <db-address>",
		Short: "A MongoDB Terminal User Interface",
		Long: `mtui is a MongoDB Terminal User Interface that allows you to interact with MongoDB databases from the CLI
in a more intuitive way.`,
		Run: func(cmd *cobra.Command, args []string) {
			var dbAddress string
			if len(args) == 1 {
				dbAddress = args[0]
			}

			parsedConStr, err := getConnectionString(dbAddress, hostFlag, portFlag)
			cobra.CheckErr(err)
			clientOps, err := generateConnectionOptions(parsedConStr, usernameFlag, passwordFlag, authMechanismFlag)
			cobra.CheckErr(err)

			client, err := mongo.Connect(clientOps)
			cobra.CheckErr(err)
			tui.Initialize(client)
		},
	}

	cmd.Flags().StringVar(&hostFlag, "host", "", "Server to connect to")
	cmd.Flags().IntVar(&portFlag, "port", 0, "Port to connect to")
	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for authentication")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password for authentication")
	cmd.Flags().StringVar(&authMechanismFlag, "authenticationMechanism ", "", "Authentication mechanism to use") // TODO restrict this to a set of valid values

	return cmd
}

// generateConnectionOptions takes in the connectionString and any auth based flags and returns the clientOptions
// necessary to connect to mongodb.
// TODO more flags need to be added to fully support auth. Hardcoded to always use an auth database of admin for now
func generateConnectionOptions(connectionString, usernameFlag, passwordFlag, authMechanism string) (*options.ClientOptions, error) {
	clientOps := options.Client().ApplyURI(connectionString)
	clientOps.SetTimeout(mongodata.Timeout)

	if clientOps.Auth == nil && (usernameFlag != "" || passwordFlag != "") {
		clientOps.Auth = &options.Credential{
			AuthSource: "admin",
		}
	}

	// Overwrite the host and port from dbAddress if they are provided as flags
	if usernameFlag != "" {
		clientOps.Auth.Username = usernameFlag
	}
	if passwordFlag != "" {
		clientOps.Auth.Password = passwordFlag
	}
	if authMechanism != "" {
		clientOps.Auth.AuthMechanism = authMechanism
	}

	return clientOps, nil
}

// getConnectionString takes in the dbAddress that can contain the entire connection string as well as the connection based flags
// passed in from the UI to allow for overriding or setting of the different connection values.
func getConnectionString(dbAddress, hostFlag string, portFlag int) (string, error) {
	if !strings.Contains(dbAddress, "://") {
		dbAddress = "mongodb://" + dbAddress
	}
	parsedUrl, err := url.Parse(dbAddress)
	if err != nil {
		return "", err
	}

	parsedHost, parsedPort, err := net.SplitHostPort(parsedUrl.Host)
	if err != nil {
		parsedHost = parsedUrl.Host
	}
	if hostFlag != "" {
		parsedHost = hostFlag
	}
	if parsedHost == "" {
		return "", fmt.Errorf("no host provided") // TODO display help menu if this error is displayed
	}

	if portFlag == 0 && parsedPort == "" {
		parsedPort = "27017"
	} else if portFlag != 0 {
		parsedPort = fmt.Sprintf("%d", portFlag)
	}

	parsedUrl.Host = fmt.Sprintf("%s:%s", parsedHost, parsedPort)
	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = "mongodb"
	}

	return parsedUrl.String(), nil
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
