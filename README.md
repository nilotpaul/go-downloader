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

## Run

- **DEV**: `air`
- **PROD**: `make`

## Notes

I will add the feature for multiple accounts later. For now, I will focus on improving error handling and building the client.

### Priority List:

1. Errors
2. Client
3. Multiple Accounts
4. More Providers
