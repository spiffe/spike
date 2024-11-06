![SPIKE](../../assets/spike-banner.png)

## SPIKE Forced Root Key Reset

This will make all the stored secrets obsolete, so it should be done
as a last resort. This may be required in cases where the database has been
corrupted, or the admin user has lost access to their password manager
(we hope that they don't memorize passwords, and they have more trusted ways
of keeping random long-lived passwords elsewhere, like system keyring, or a
password manager).

```mermaid
sequenceDiagram
    participant P as Admin User
        Note over P: This is an out-of-band operation.<br>It will not touch SPIRE #semi;<br>however, it will reset SPIKE<br>to its day zero setting.
    alt skies have fallen apart
        Note over P: Total system crash.<br>Admin forgot their password.<br>Database corrupt.<br>Or the system is in a similar irrecoverable state.
        Note over P: Admin runs `./hack/reset.sh` to reset the system to day zero.
        Note over P: Prompt: Are you really sure? This will wipe all secrets. And it's irreversible.
        Note over P: Any key or token saved on the file system will be wiped out.
        Note over P: The database will be reset to the initial state.
        Note over P: SPIKE Pilot and SPIKE Nexus will be restarted.
        Note over P: System reset. Admin can re-run `spike init`.
        Note over P: Explicitly verify the system state after reset.
        Note over P: Confirm all component restarts.
        Note over P: Validate database reset.
    end
```