## Postgres Setup

> **Future Work**
>
> Postgres setup will be done in the future.
> Don't worry about it that much for now.

Here are steps to set up Postgres for Ubuntu Linux:

Install Postgres:

```bash
sudo apt install postgres
```

Configure Postgres to listen everywhere:

```bash 
sudo vim /etc/postgresql/$version/main/postgresql.conf
# change listen_address as follows:
# listen_address = '*'
```

Create database `spike`:

```bash
sudo -u postgres psql -c 'create database spike;';
```

Set a password for the postgres user:

```bash 
ALTER USER postgres with encrypted password 'your-password-here';
```

Enable SSL:

```bash
sudo vim /etc/postgresql/16/main/pg_hba.conf

# Update the file and set your IP range accordingly.
# hostssl spike postgres 10.211.55.1/24 scram-sha-256
```

That's it. Your database is configured for local development.