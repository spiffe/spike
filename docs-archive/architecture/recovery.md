## Recovery Mechanisms

### SPIKE Nexus Crash Recovery

1. **SPIKE Nexus** crashes.
2. New **SPIKE Nexus** instance starts.
3. **SPIKE Nexus** authenticates to **SPIKE Keeper** using [**SPIFFE** 
   mTLS][spiffe].
4. **SPIKE Keeper** provides the root key over **mTLS**.
5. **SPIKE Nexus** resumes normal operation.

[spiffe]: https://spiffe.io/

### Complete System Recovery

1. Both **SPIKE Nexus**, **SPIKE Keeper** are unavailable, or the system is
   in on other irrecoverable state.
2. Admin executes `spike recover`.
3. Admin provides their **password**.
4. The encrypted **root key** is fetched from the database and injected to
   the memory of **SPIKE Nexus**.
5. **SPIKE Nexus** syncs the **root key** with **SPIKE Keeper**.
6. The system resumes normal operation.

### System Reset

* Both **SPIKE Nexus** and all **SPIKE Keeper** instances have crashed, there
  is no way to fetch the root key from **SPIKE Keeper**(s).
* Admin has lost access to their password.
* There is no backup administrator that could have avoided the "bus factor".
* No one is using a password management software so that this would have been
  a non-issue from day zero.
* Everyone have learned their lessons, and now it's time to reset the system 
  and conduct an extensive "what went wrong / what should have been done" analysis. 
* An administrator with adequate privileges manually runs a reset shell script.
* **SPIKE** is reset to its factory defaults.
* A new admin user initializes **SPIKE** with a new secure password.
* All former secrets are lost.
* All former policies and configuration is lost.
* This is a complete system reset.

## Data Consistency

* If the root key changes during recovery, the old encrypted data becomes
  invalid.
* The root key is backed up in the backing store (*Postgres*); however
  securely backing up the backing store is also crucial.
* The root key cannot be recovered without the admin password. Therefore
  securely backing up the admin password in a out-of-band password manager
  is also crucial.
