# Snip
A simple service for shortening URLs.

## Quick start

### Prerequisites
- **[Docker](https://docs.docker.com/engine/install/)** and **[Docker Compose](https://docs.docker.com/compose/install/)** must be installed on your machine.
- **[Git](https://git-scm.com/)** must be installed on your machine.
- **[migrate](https://github.com/golang-migrate/migrate)** binary must be available from cli on your machine.

### Initial local setup
1. Clone this repository: `git clone https://github.com/aboyadzhiev/snip.git`
2. Enter the project directory: `cd snip`
3. Configure the environment variables:
   1. Create local `.env` file by executing: `cp .env.dist .env`
   2. Enter appropriate values for the variables:
      1. `SNIP_HOSTNAME` holds the hostname/domain that will be used by snip when shortening/resolving URLs.
      An example value is `https://snip.local`, please note that in this case you'll have to update and your `host` file.
      2. `VALKEY_HOSTS` holds comma separated list of `valkey` cluster hosts represented by hostname and port number e.g.
      `valkey:6379`. Single value is acceptable.
      3. `POSTGRES_HOST` holds `PostgreSQL` hostname and port number e.g. `db:5432`
      4. `POSTGRES_USER`, `POSTGRES_DB`, `POSTGRES_PASSWORD` are self-explanatory.
      5. `URLHAUS_API_ENDPOINT` holds static value of `https://urlhaus.abuse.ch/downloads/json_online/` used for detection
      of malicious URLs.
4. Start the docker compose stack: `docker compose up`
5. Define `POSTGRESQL_URL` environment e.g. `export POSTGRESQL_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable`
6. Execute the database migrations: `migrate -database ${POSTGRESQL_URL} -path server/db/migrations up`
7. (Optional) Import Caddy's root certificate into your web browser to allow https access to `SNIP_HOSTNAME` e.g. for `https://snip.local`
   1. Determinate the location(`Mountpoint`) of Caddy's root certificate: `docker volume inspect snip_caddy_data`
   2. Copy the Caddy's root certificate to your host machine: `sudo cp /var/lib/docker/volumes/snip_caddy_data/_data/caddy/pki/authorities/local/root.crt ~/`
   3. Trust the Caddy's root certificate in your web browser.
8. Done. Navigate your web browser to the value defined in `SNIP_HOSTNAME`.

## Screenshots

### Homepage
![Homepage](screenshots/Screenshot_1.png)

### Shortening URL
![Shortening URL](screenshots/Screenshot_2.png)

### Visiting shortened URL
![Visiting shortened URL](screenshots/Screenshot_3.png)

### Attempting to shorten malicious URL
![Attempting to shorten malicious URL](screenshots/Screenshot_4.png)
