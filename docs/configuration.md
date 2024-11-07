![SPIKE](assets/spike-banner.png)

## Configuring SPIKE

You can use environment variables to configure the **SPIKE** components. 

The following table lists the environment variables that you can use to 
configure the SPIKE components:

| Component    | Environment Variable        | Description                                                    | Default Value |
|--------------|-----------------------------|----------------------------------------------------------------|---------------|
| SPIKE Nexus  | `SPIKE_NEXUS_POLL_INTERVAL` | The interval SPIKE Nexus syncs up its state with SPIKE Keeper. | `"5m"`        |
| SPIKE Nexus  | `SPIKE_NEXUS_TLS_PORT`      | The TLS port SPIKE Nexus listens on.                           | `":8553"`     |
| SPIKE Keeper | `SPIKE_KEEPER_TLS_PORT`     | The TLS port SPIKE Keeper listens on.                          | `":8443"`     |

We'll add more configuration options in the future. Stay tuned!