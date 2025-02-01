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
	var hostFlag string
	var portFlag int

	var usernameFlag string
	var passwordFlag string
	var authDatabaseFlag string
	var authMechanismFlag string

	var cmd = &cobra.Command{
		Use:   "mtui <db-address>",
		Short: "A MongoDB Terminal User Interface",
		Long:  `mongotui is a MongoDB Terminal User Interface`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Verify that a host has been provided
			if len(args) != 1 && hostFlag == "" {
				return fmt.Errorf("you must provide a valid hostname")
			}
			// Verify that authMechanism is a supported value if provided
			validAuthMechs := []string{"", "SCRAM-SHA-1", "SCRAM-SHA-256", "MONGODB-X509", "GSSAPI", "PLAIN"}
			if ok := slices.Contains(validAuthMechs, authMechanismFlag); !ok {
				return fmt.Errorf("invalid authenticationMechanism of %s provided. Must be one of %v", authMechanismFlag, validAuthMechs)
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

			applyHostConfig(clientOps, hostFlag, portFlag)
			applyAuthConfig(clientOps, usernameFlag, passwordFlag, authMechanismFlag, authDatabaseFlag)

			client, err := mongo.Connect(clientOps)
			cobra.CheckErr(err)
			tui.Initialize(client)
		},
	}

	// First group of flags when running mongosh --help
	regularFlags := pflag.NewFlagSet("regularFlags", pflag.ExitOnError)
	regularFlags.StringVar(&hostFlag, "host", "", "Server to connect to")
	regularFlags.IntVar(&portFlag, "port", 0, "Port to connect to")
	cmd.Flags().AddFlagSet(regularFlags)

	authenticationFlags := pflag.NewFlagSet("authenticationFlags", pflag.ExitOnError)
	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for authentication")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password for authentication")
	cmd.Flags().StringVar(&authDatabaseFlag, "authenticationDatabase", "", "User source (defaults to dbname)")
	cmd.Flags().StringVar(&authMechanismFlag, "authenticationMechanism", "", "Authentication mechanism to use")
	cmd.Flags().AddFlagSet(authenticationFlags)

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
