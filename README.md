# go-ojol

Gojek Simplified (ceunah). A customer app, a driver app, and three Go services behind one gateway. You book a motorcycle or a car, nearby standby drivers get a 30-second offer, and after accept the two phones share GPS until the driver marks the trip complete.

This is not a payment product yet. Complete writes `paid_at` and both apps treat that as done.

How the pieces talk to each other lives in [ARCHITECTURE.md](./ARCHITECTURE.md). This file is how to run it.

## Repo

```
backend/gateway   path reverse proxy. no database.
backend/auth      register, login, sessions, JWKS, users, Casbin
backend/trip      quotes, matching, WebSockets, trip lifecycle, saved addresses
frontend/goojol-cst  Expo customer app
frontend/goojol-drv  Expo driver app
tests/http           Kulala / IntelliJ HTTP fixtures
```

Each backend service is its own Go module. There is no `go.work` at the root.

## What you need

- Go 1.26
- PostgreSQL with `uuid-ossp`
- Redis
- Node 20+ and pnpm (or npm) for the Expo apps
- A P-256 private key for JWT signing
- Optional: Docker, if you want the local Indonesia OSRM graph

Postgres and Redis are expected on the host. There is no root `docker-compose.yml`.

## Backend



### 1. JWT key

```bash
mkdir -p backend/auth/keys # This is already ignored in .gitignore, its up to you btw
cd backend/auth/keys
openssl ecparam -name prime256v1 -genkey -noout -out jwt-ec-p256.pem
```

Point `JWT_PRIVATE_KEY_PATH` at that file in Auth's `.env`.

### 2. Env per service

Copy the idea, not a shared file. Auth and Trip should use the **same** `DB_`* values. Gateway does not need a database.

Typical local ports, matching `tests/http/http-client.env.json`:


| Process | `GOLANG_PORT` |
| ------- | ------------- |
| Gateway | `8001`        |
| Auth    | `8002`        |
| Trip    | `8003`        |

I'm literally use that on my computer.


Auth `.env` (minimum):

```
APP_ENV=localhost
GOLANG_PORT=8002
CORS_ALLOWED_ORIGINS=http://localhost:8081
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=goojol
JWT_PRIVATE_KEY_PATH=keys/jwt-ec-p256.pem
JWT_ISSUER=go-ojol-auth
UPLOAD_THING_TOKEN=
```

Trip `.env` (minimum):

```
APP_ENV=localhost
GOLANG_PORT=8003
CORS_ALLOWED_ORIGINS=http://localhost:8081
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=goojol
JWT_ISSUER=go-ojol-auth
AUTH_JWKS_URL=http://127.0.0.1:8002/.well-known/jwks.json
REDIS_ADDR=127.0.0.1:6379
```

Gateway `.env`:

```
APP_ENV=localhost
GOLANG_PORT=8001
CORS_ALLOWED_ORIGINS=http://localhost:8081
AUTH_SERVICE_URL=http://127.0.0.1:8002
TRIP_SERVICE_URL=http://127.0.0.1:8003
```

Create the database and enable the extension once:

```bash
createdb goojol
psql -d goojol -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";' # or run this directly into DBEaver. Ask you GPT for step by step instruction please!
```



### 3. Migrate and seed

Auth first. That creates users, vehicles, Casbin, and sessions.

```bash
cd backend/auth
make migrate-seed
```

Trip next, so transaction enums and location columns exist.

```bash
cd backend/trip
make migrate
```

Seed accounts (password `hehe1234` for all three):


| Email                 | Role     |
| --------------------- | -------- |
| `cst.adics@gmail.com` | customer |
| `drv.adics@gmail.com` | driver   |
| `adm.adics@gmail.com` | admin    |


If you add Casbin resources, keep `backend/auth/script/casbin_seed.go` and the `casbin_*.json` files in sync, then `make seed` from Auth. The CLI is `go run cmd/main.go --script:casbin_seed` with no extra flags.

### 4. Run the three processes

```bash
cd backend/auth && make run
cd backend/trip && make run
cd backend/gateway && make run
```
Yeah i know, we can setup docker later.

"Smoke test" the gateway:

```bash
curl -s http://127.0.0.1:8001/.well-known/jwks.json
```

I'm using [mistweaverco/kulala.nvim](https://github.com/mistweaverco/kulala.nvim) to test the API without Postman bloat.

HTTP walkthroughs, in order: `tests/http/01-calculate-argo.http`, `02-driver-standby.http`, `03-find-nearby-driver.http`, `04-trip-lifecycle.http`, plus `04-auth-sessions.http` for refresh/logout.

### Tests

```bash
cd backend/auth && make test-all
cd backend/trip && make test-all
cd backend/gateway && make test
```



### Local OSRM (Good luck setup your own Self-Host)

Quotes call OSRM. Out of the box that is the public demo at `router.project-osrm.org`, which is fine until it rate-limits you. A self-hosted Indonesia extract (car profile, first build is slow and RAM-heavy) is:

```bash
cd backend/trip/docker/osrm
docker compose up --build # Just need some piece of your ram *em-dash 16GB
```

It listens on `127.0.0.1:5001`. Trip does not read an env var for this yet. `backend/trip/providers/core.go` currently passes `""` into `NewDispatchService`, which selects the public default. Change that argument to `http://127.0.0.1:5001` if you want the local graph.

## Frontends

Both apps are Expo 56 with a dev client (MapLibre is native). `expo start` on web will not give you the map.

```bash
cd frontend/goojol-cst
# or frontend/goojol-drv
```

Create `.env`:

```
EXPO_PUBLIC_API_URL=http://127.0.0.1:8001
EXPO_PUBLIC_WS_URL=ws://127.0.0.1:8001
```

On a physical phone, use your machine's LAN IP instead of `127.0.0.1`. Then:

```bash
yarn install
yarn ios
# or yarn android
```

Customer path: login as `cst.adics@gmail.com` → book → pin pickup and destination → quote → find driver. Driver path: login as `drv.adics@gmail.com` → go online on the home map → accept the offer → start → complete.

If the customer app is killed mid-ride it calls `GET /api/trip/transactions/active` on launch and jumps back to the trip screen.

## Useful make targets

From `backend/auth` or `backend/trip`:

```
make run
make migrate
make seed
make migrate-status
make module name=<feature>
```

Trip also has `make test-dispatch` and `make test-trip`.

# Credit
- [Caknoooo/go-gin-clean-starter](https://github.com/Caknoooo/go-gin-clean-starter)
- [Kenney's Pixel Vehicle Pack](https://kenney.nl/assets/pixel-vehicle-pack)