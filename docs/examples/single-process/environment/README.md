# Environment Files

`.env` files in the `environment` directory and any sub-directories are loaded before executing services lifecycle scripts (`startup.sh`, `process.sh`, and `finish.sh`) or entrypoint lifecycle pre-commands.
The variables they contain are defined as environment variables in the container.

Files are loaded in alphabetical order, and variables are overwritten if they were already defined in a previous file.

**Container environment variables set at runtime override values defined in `.env` files.**

## Precedence and variable interpolation

Variables already present in the container environment, including those provided
with `docker run -e`, take precedence over environment files. Their values are
preserved exactly, including empty values, quotes, newlines, and `$` characters.

Within environment files, `$VAR` and `${VAR}` references are expanded using
gotenv. Existing container environment variables can be used to resolve these
references. File override order and interpolation follow the gotenv parser's rules.

To keep a reference literal, use single quotes or escape the dollar sign:

```dotenv
DOMAIN=example.org
URL="https://${DOMAIN}"           # https://example.org
LITERAL='${DOMAIN}'              # ${DOMAIN}
TEMPLATE="https://\${DOMAIN}"     # https://${DOMAIN}
```

Values already present in the environment are not expanded again when
environment files are loaded. To expand a template explicitly, use
`container envsubst`.

Your shell or Docker Compose may perform interpolation before passing a value
to the container. The preservation rule applies to the value received by the
container.
