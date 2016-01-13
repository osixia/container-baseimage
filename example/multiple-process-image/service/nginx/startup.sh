#!/bin/bash -e
FIRST_START_DONE="${CONTAINER_STATE_DIR}/nginx-first-start-done"

# container first start
if [ ! -e "$FIRST_START_DONE" ]; then
  echo ${WHO_AM_I}  >> /var/www/html/index.html
  touch $FIRST_START_DONE
fi

echo "The secret is: $FIRST_START_SETUP_ONLY_SECRET"

exit 0
