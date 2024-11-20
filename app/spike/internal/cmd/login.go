//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

// NewLoginCommand creates and returns a Cobra command that handles user
// authentication with SPIKE Nexus. The command prompts for an admin password
// and stores the resulting authentication token.
//
// The command performs the following steps:
//  1. Prompts the user for an admin password via secure password input
//  2. Attempts to authenticate with SPIKE Nexus using the provided X509Source
//  3. On successful authentication, saves the JWT token to a local file
//
// Parameters:
//   - source: An X509Source used for authenticating the connection to
//     SPIKE Nexus
//
// The command supports the following flags:
//   - --password: Optional flag to provide password
//     (not recommended for security reasons)
//
// The token is saved to the path specified by config.SpikePilotAdminTokenFile()
// with 0600 permissions.
//
// If any step fails (password input, authentication, or token storage),
// the command prints an error message and returns without saving the token.
//func NewLoginCommand(source *workloadapi.X509Source) *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "login",
//		Short: "Login to SPIKE Nexus",
//		Long:  `Login to SPIKE Nexus.`,
//		Run: func(cmd *cobra.Command, args []string) {
//			fmt.Print("Enter admin password: ")
//			password, err := term.ReadPassword(syscall.Stdin)
//			if err != nil {
//				fmt.Println("\nFailed to read password:")
//				fmt.Println(err.Error())
//				return
//			}
//			fmt.Println()
//
//			token, err := auth.Login(source, string(password))
//			if err != nil {
//				fmt.Println("Failed to login:")
//				fmt.Println(err.Error())
//				return
//			}
//
//			err = os.WriteFile(
//				config.SpikePilotAdminTokenFile(), []byte(token), 0600,
//			)
//			if err != nil {
//				fmt.Println("Failed to save token to file:")
//				fmt.Println(err.Error())
//				return
//			}
//
//			fmt.Println("Login successful.")
//		},
//	}
//
//	cmd.Flags().String("password", "", "Password to login to SPIKE Nexus")
//
//	return cmd
//}
