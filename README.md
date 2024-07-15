# Go Downloader

Allows downloading files from multiple providers.

## Motivation

Needed a downloader for my HomeLab which can also download private Google Drive files.

## Note

- Only Google provider has been added for now. More providers will be added in the future.
- Currently, it only supports Google Drive.
- Only single user support is available. In the future, users will be able to add multiple accounts and change their sessions.

## Issues

1. Sometimes the session gets invalid and refreshing doesn't work.

## TODO

1. Add ability to add multiple accounts for the same provider (OAuth).
2. Change the session middleware to support multiple accounts.
3. Change the DB schema to store the added accounts which can be used to swap sessions.
4. Improve error handling for download errors and WebSocket progress errors (add a channel for errors).
5. Build the client (high priority).
6. Continue adding more providers.

## Installation

### Manual

- **DEV**: `air`
- **PROD**: `make`

### With Docker

Using docker compose:

```yml
---
services:
  go_downloader:
    image: ghcr.io/nilotpaul/go-downloader:1.0.0
    container_name: go_downloader
    ports:
      - "3000:3000" # If port 3000 is unavailable, change '3000:3000' to 'YOUR_PORT:3000'
    networks:
      - go_downloader_network
    environment:
      - GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
      - GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
      - SESSION_SECRET=some-secret # Random Secret, change this to something secure
      - APP_URL=${APP_URL} # Full URL with http or https
      - DOMAIN=${DOMAIN} # eg. yourdomain.com
      - DB_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@go_downloader_pg_db:5432/${POSTGRES_DB}?sslmode=disable
      - DEFAULT_DOWNLOAD_PATH=/media
      - PUID=1000 # Your user id
      - PGID=1000 # Your group id
    volumes:
      - /media:/media
    restart: unless-stopped

  go_downloader_pg_db:
    container_name: go_downloader_pg_db
    image: postgres:latest
    networks:
      - go_downloader_network
    environment:
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB}
    volumes:
      - go_downloader_pg_data:/var/lib/postgresql/data
    restart: unless-stopped

networks:
  go_downloader_network:

volumes:
  go_downloader_pg_data:
```

NOTE: **To make this work, you'll need a top-level domain or use it from your local machine via localhost as Google OAuth doesn't allow IP addresses. Later this limitation will be solved by using Google Service Account (upcoming).**

1. **Get Google Client ID and Secret**: Get your Google Client ID and Secret from [Google Cloud Console](https://console.cloud.google.com/). Follow [this tutorial](https://www.balbooa.com/help/gridbox-documentation/integrations/other/google-client-id) for guidance.

2. **App URL**: The `APP_URL` should be the full URL of your application. If you have a domain, use the full URL path (e.g., `https://yourdomain.com`). If not, you can use `http://localhost:3000`.

3. **Domain**: The `DOMAIN` should be your domain name (e.g., `yourdomain.com`). If running locally, use `localhost`.

4. **Default Media Path**: The `DEFAULT_DOWNLOAD_PATH` will be used as a fallback if you don't specify a specific path when starting a download.

5. **PUID and PGID**: You can find your PUID and PGID by running the following command on Linux or macOS:
   ```sh
   id $(whoami)
   ```
6. Mapping Correct System Path: To store the downloads in the correct paths or folders you want, you will need to map the correct system path inside the Docker container. For example, to map your system's `/media` directory to the Docker container's `/media` directory, use:
   ```yml
   volumes:
     - /media:/media
   ```

### With Docker CLI

**FOR DOWNLOADER**
```bash
docker run -d \
 --name go_downloader \
 -p 3000:3000 \
 --network go_downloader_network \
 -e GOOGLE_CLIENT_ID=yourclientid \
 -e GOOGLE_CLIENT_SECRET=yourclientsecret \
 -e SESSION_SECRET=some-secret \
 -e APP_URL=http://yourappurl \
 -e DOMAIN=yourdomain.com \
 -e DB_URL=postgres://yourpostgresuser:yourpostgrespassword@go_downloader_pg_db:5432/yourpostgresdb?sslmode=disable \
 -e DEFAULT_DOWNLOAD_PATH=/media \
 -e PUID=1000 \
 -e PGID=1000 \
 -v /media:/media \
 --restart unless-stopped \
 ghcr.io/nilotpaul/go-downloader:1.0.0
```

**FOR POSTGRES DB**
```bash
docker run -d \
 --name go_downloader_pg_db \
 --network go_downloader_network \
 -e POSTGRES_USER=yourpostgresuser \
 -e POSTGRES_PASSWORD=yourpostgrespassword \
 -e POSTGRES_DB=yourpostgresdb \
 -v go_downloader_pg_data:/var/lib/postgresql/data \
 --restart unless-stopped \
 postgres:latest
```

## Notes

I will add the feature for multiple accounts later. For now, will focus on improving error handling and building the client.

### Priority List:

1. Errors
2. Client
3. Multiple Accounts
4. More Providers

---

Sorry, commit history is a huge mess 😢
