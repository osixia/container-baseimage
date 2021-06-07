# Environment Files

`.env` files in this directory and any sub-directories are loaded before executing services lifeycle script(s) (`startup.sh`, `process.sh` and `finish.sh`) or entrypoint lifecycle pre-commands.
The variables they contain are defined as environment variables in the container.

Files are loaded in alphabetical order, and variables are overwrite if already defined in a previous file.

**Container environment variables set at run time will overwrite value defined in `.env` files.**
