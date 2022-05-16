#!/bin/zsh
config=$(<firebase-credentials.json)
flyctl secrets set GOOGLE_APPLICATION_CREDENTIALS_JSON="$config"