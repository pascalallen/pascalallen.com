#!/bin/bash

# Script that runs when Postgres container spins up.
#
# Only runs if you start the container with a data directory that is empty.

#set -e

#psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
#  ALTER USER $POSTGRES_USER PASSWORD '$POSTGRES_PASSWORD';
#  GRANT ALL PRIVILEGES ON DATABASE $POSTGRES_DB TO $POSTGRES_USER;
#EOSQL
