# Crypto Alerts Streamers API

Crypto alerts streamers backend API for registering streamers, creating twitch widgets and donations.

# Development

Build inside docker container:

`ENV=dev docker-compose -f docker-compose.yml up --build --force-recreate`

Local testing, add env variables from file:

`export $(grep -v '^#' .env | xargs)`
