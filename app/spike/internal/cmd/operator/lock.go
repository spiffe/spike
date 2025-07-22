package operator

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike/internal/lock"

	"os"

	"github.com/spiffe/spike/app/spike/internal/env"

	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newOperatorLockCommand creates a new cobra command for locking
// SPIKE Nexus, disabling all operations until it is explicitly unlocked.
//
// This command can only be used by operators with the `watchdog` role.
//
// Parameters:
//   - source *workloadapi.X509Source: The X.509 source for SPIFFE authentication.
//   - spiffeId string: The SPIFFE ID of the caller for role-based access control.
//
// Returns:
//   - *cobra.Command: A cobra command that implements the lock functionality.
//
// The command performs the following operations:
//   - Verifies the caller has the 'lock' (watchdog) role, aborting otherwise.
//   - Authenticates the lock request.
//   - Sends the lock request to the SPIKE API.
//   - Prints confirmation if successful.
func newOperatorLockCommand(spiffeId string,
) *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Lock SPIKE Nexus and disable all operations",
		Run: func(cmd *cobra.Command, args []string) {
			if !spiffeid.IsPilotLock(env.TrustRoot(), spiffeId) {
				fmt.Println("")
				fmt.Println("  You need to have a `lock` role to use this command.")
				fmt.Println("  Please run ./hack/bare-metal/entry/spire-server-entry-lock-unlock-register.sh")
				fmt.Println("  with necessary privileges to assign this role.")
				fmt.Println("")
				os.Exit(1)
			}

			trust.AuthenticateLock(spiffeId)

			if lock.IsLocked() {
				fmt.Println("SPIKE is already locked. No action taken.")
				return
			}

			if err := lock.Lock(); err != nil {
				fmt.Println("  Failed to lock SPIKE:", err)
				os.Exit(1)
			}

			fmt.Println("SPIKE has been locked. All operations are now disabled.")
			fmt.Println("To unlock SPIKE, run `spike operator unlock`.")
		},
	}
}
