## Secrets management

The backend does not keep real settings (passwords, keys, ports) in the repo.
They are in a `.env` file stored in MongoDB, and the server downloads that file when it
starts. This ensures that everyone shares one config and no secrets are committed to git.

The MongoDB database can be either:

- **Atlas (cloud):** Shared team database used for normal development.
- **Local Docker container:** Private database on your own box used for experimentation or offline work (like a sandbox).

> **Note:** Run `make` commands from the repo root and `./server/backend` commands from the `backend/` directory.

> **Rule: never commit a real secret.** Check `git status` before every commit (our gitignore should catch this, but look anyway).

### How settings are chosen

For each setting, the server uses the first of these that has a value:

| Priority | Source | Used for |
|---|---|---|
| 1 (highest) | Real environment variables, e.g. `SERVER_PORT=7070 ./server serve ...` | One-off changes, CI |
| 2 | `backend/.env.override` (own local file) | Personal overrides of the team config |
| 3 | The `.env` stored in MongoDB | Shared team config |

### Files

All of these live in `backend/`.

| File | Committed? | Description |
|---|---|---|
| `.env.example` | Yes | Template showing which settings exist |
| `secrets.env.example` | Yes | Template for connecting to local Docker database |
| `secrets.atlas.env.example` | Yes | Template for connecting to Atlas |
| `.env` | **No** | Real settings that get uploaded to MongoDB |
| `secrets.local.env` | **No** | Connection details for local Docker database |
| `secrets.atlas.env` | **No** | Connection details for Atlas (read only user) |
| `secrets.atlas-writer.env` | **No** | Connection details for Atlas (write user) |
| `.env.override` | **No** | Optional personal overrides |

### Using Atlas database

1. Get Atlas cluster address and `secrets-reader` password.
   Don't paste them anywhere like Jira or GitHub.
2. From repo root, create connection file:
   ```
   cd backend
   cp secrets.atlas.env.example secrets.atlas.env
   chmod 600 secrets.atlas.env
   cd ..
   ```
3. In `backend/secrets.atlas.env`, replace `<db_username>`, `<db_password>`, and `<cluster_address>`.
4. From repo root, build and run:
   ```
   make server
   cd backend
   ./server/backend serve --secrets-file secrets.atlas.env
   ```
5. The output should include a log line containing `"msg":"Secrets loaded"`. Press Ctrl+C to stop.

If it fails, see [Troubleshooting](#troubleshooting).

### Using local Docker database

Your local database starts empty, so upload a `.env` into it once.

1. Start MongoDB from repo root:
   ```
   docker compose up -d
   docker ps
   ```
   `docker ps` should list a container named `mongo` with a status starting with `Up`.
2. Create connection file and `.env`:
   ```
   cd backend
   cp secrets.env.example secrets.local.env
   cp .env.example .env
   cd ..
   ```
   `.env.example` has placeholder values.
3. Build server and upload `.env` into local database:
   ```
   make server
   cd backend
   ./server/backend secrets-push --secrets-file secrets.local.env --file .env --env dev
   ```
   Expect: `Uploaded .env to the secrets store.`
4. Run the server:
   ```
   ./server/backend serve --secrets-file secrets.local.env
   ```
   The output should include a log line containing `"msg":"Secrets loaded"`. Press Ctrl+C to stop, then:
   ```
   cd ..
   ```
5. Optional check that document exists (does not print contents):
   ```
   docker exec -it mongo mongosh -u root -p localdevpassword --authenticationDatabase admin --eval 'db.getSiblingDB("secrets").env_files.find({}, {dotenv: 0})'
   ```

### Overrides

To change a setting just for yourself, put it in `backend/.env.override`:
```
echo "EXAMPLE_SECRET=my-local-value" > backend/.env.override
```
Or for a single run, from `backend/`:
```
EXAMPLE_SECRET=my-local-value ./server/backend serve --secrets-file secrets.atlas.env
```
Delete the file (or line) to go back to the shared value.

### Changing shared config

Use `secrets-writer` Atlas creds.

1. Create `backend/secrets.atlas-writer.env` from `secrets.atlas.env.example` and use the
   `secrets-writer` user and password.
2. Edit `backend/.env` with new values.
3. Upload it:
   ```
   make server
   cd backend
   ./server/backend secrets-push --secrets-file secrets.atlas-writer.env --file .env --env dev
   cd ..
   ```
4. Restart server (rebuild) to see change.

`--env` selects which stored document to write, so multiple can be kept (like
`dev`, `ci`, `prod`). The server reads the one named by `SECRETS_ENV` in the connection file.

### Troubleshooting

| Observe | Cause/Fix |
|---|---|
| `SECRETS_MONGO_URI is not set` | You are not in `backend/`, or `--secrets-file` name is misspelled |
| `authentication failed` / `unauthorized` (Atlas) | Wrong username or password in your file, or lack of permission |
| `authentication failed` (Docker) | The database was first created with a different password, run `docker compose down -v`, then `docker compose up -d` and upload again |
| Timeout or `server selection` error (Atlas) | Atlas Network Access does not allow your IP |
| Timeout or `connection refused` (Docker) | The container is not running, run `docker compose up -d` |
| `no secrets stored for environment dev` | Nothing has been uploaded to that database yet, or you used a different `--env` name |
| `port is already allocated` when starting Docker | Another MongoDB is using port 27017. Stop it, or change the port in `docker-compose.yml` and in `secrets.local.env` |
