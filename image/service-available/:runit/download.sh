#!/bin/bash -e

# download runit from apt-get
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends runit
