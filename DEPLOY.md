# Deploy runbook

Target host (existing deployment):

- SSH: `ssh 79749-9425@gate.infra.dewacloud.com -p 3022`
- nginx serves the customer PWA on `:80` from `/root/maremereso/customer`
  and the admin portal on `:8081` from `/root/maremereso/admin`
- Go backend on `127.0.0.1:8080`, binary at `/root/maremereso/backend/server`
- PostgreSQL on `:5433`

The steps below must run **in order**. Step 1 is a schema change and step 3
locks every staff account until step 4 is done, so plan a short maintenance
window.

---

## 1. Database migration

`internal/database/migrations/000004_production_hardening.up.sql` has not been
applied yet. It is additive apart from clearing the unusable seeded password
hashes, and it is safe to run on a live database.

Take a backup first:

```bash
ssh -p 3022 79749-9425@gate.infra.dewacloud.com \
  'pg_dump -h 127.0.0.1 -p 5433 -U <user> maremereso_olga | gzip > ~/olga-backup-$(date +%F-%H%M).sql.gz'
```

Copy and apply the migration:

```bash
scp -P 3022 maremeresto-backend/internal/database/migrations/000004_production_hardening.up.sql \
  79749-9425@gate.infra.dewacloud.com:/tmp/

ssh -p 3022 79749-9425@gate.infra.dewacloud.com \
  'psql -h 127.0.0.1 -p 5433 -U <user> -d maremereso_olga -v ON_ERROR_STOP=1 -f /tmp/000004_production_hardening.up.sql'
```

What it does:

- drops the placeholder driver defaults and nulls out the fabricated values
  (`Andi Pratama`, `B 1234 ABC`, `081234567890`)
- adds `orders.driver_assigned_at`, `orders.acknowledged_at`,
  `orders.acknowledged_by` and supporting indexes
- adds `users.email` (unique, case-insensitive) and backfills the four staff
  addresses, renaming the admins to match the Solo outlets
- adds promo validity/redemption columns
- **clears the seeded `password_hash` for every staff account** — see step 4

## 2. Configure the backend

The server now refuses to start without `JWT_SECRET` (min. 32 chars),
`DB_USER` and `MIDTRANS_SERVER_KEY`. Add the new keys to the existing
`/root/maremereso/backend/.env`:

```bash
ssh -p 3022 79749-9425@gate.infra.dewacloud.com
cd /root/maremereso/backend

# Set the environment explicitly: this tightens CORS and stops customer
# phone numbers being written to the log.
echo 'APP_ENV=production' >> .env

# Geocoding (address search). Nominatim asks for a contact address.
echo 'GEOCODER_URL=https://nominatim.openstreetmap.org' >> .env
echo 'GEOCODER_EMAIL=ops@yourdomain.example' >> .env

# Confirm JWT_SECRET is at least 32 characters; if not, replace it.
# Changing it signs out every customer and staff member.
awk -F= '/^JWT_SECRET=/{print "JWT_SECRET length:", length($2)}' .env
```

Also confirm `CUSTOMER_URL` and `ADMIN_URL` are the real public origins — they
are now the CORS **and** WebSocket origin allowlist, and `CUSTOMER_URL` is
where Midtrans returns the customer after payment.

## 3. Build and ship

From this repository:

```bash
# Backend (Linux binary)
cd maremeresto-backend
GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
GOOS=linux GOARCH=amd64 go build -o bin/admintool-linux ./cmd/admintool

# Frontends
cd ../frontend/customer && npm ci && npm run build
cd ../admin            && npm ci && npm run build
```

Upload and restart:

```bash
HOST=79749-9425@gate.infra.dewacloud.com

scp -P 3022 maremeresto-backend/bin/server-linux    $HOST:/root/maremereso/backend/server.new
scp -P 3022 maremeresto-backend/bin/admintool-linux $HOST:/root/maremereso/backend/admintool
rsync -avz -e 'ssh -p 3022' --delete frontend/customer/dist/ $HOST:/root/maremereso/customer/
rsync -avz -e 'ssh -p 3022' --delete frontend/admin/dist/    $HOST:/root/maremereso/admin/

ssh -p 3022 $HOST '
  cd /root/maremereso/backend
  pkill -f "maremereso/backend/server" || true
  sleep 1
  mv server server.old && mv server.new server && chmod +x server admintool
  APP_ENV=production nohup ./server > /var/log/maremereso-backend.log 2>&1 &
  sleep 3
  curl -fsS http://127.0.0.1:8080/api/v1/health && echo " backend healthy"
'
```

`--delete` on the frontend rsync is intentional: stale hashed asset bundles
would otherwise accumulate. It does not touch `/root/maremereso/backend/uploads`.

## 4. Set staff passwords

Step 1 locked every staff account. Nobody can sign in to the admin portal until
this is done:

```bash
ssh -p 3022 79749-9425@gate.infra.dewacloud.com
cd /root/maremereso/backend

./admintool list        # every account should read "no - locked"

./admintool set-password owner@cafeolga.id
./admintool set-password admin.kerten@cafeolga.id
./admintool set-password admin.makamhaji@cafeolga.id
./admintool set-password admin.makdjan@cafeolga.id

./admintool list        # each should now read "yes"
```

Distribute the passwords out of band; do not put them in chat or a ticket.

## 5. Midtrans

In the Midtrans dashboard (Settings -> Configuration), set the payment
notification URL to:

```
https://<customer-domain>/api/v1/payments/notification
```

The signature is now verified in every environment, so `MIDTRANS_SERVER_KEY`
must match the dashboard exactly or all notifications will be rejected and
paid orders will stay stuck at "menunggu pembayaran".

Verify with a sandbox transaction before switching `MIDTRANS_IS_PROD=true`.

## 6. nginx

The existing config already proxies `/api/` with `Upgrade`/`Connection`
headers, so WebSockets work as-is. Two things worth adding:

```nginx
# Inside both server blocks, so a long-lived order feed is not cut off:
location /api/v1/ws {
    proxy_pass http://maremereso_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
}
```

Reload with `nginx -t && nginx -s reload`.

## 7. Post-deploy checks

```bash
# Public reads
curl -s https://<customer-domain>/api/v1/health
curl -s https://<customer-domain>/api/v1/branches | head -c 300

# Delivery quote from a real Solo coordinate
curl -s -X POST https://<customer-domain>/api/v1/delivery/quote \
  -H 'Content-Type: application/json' \
  -d '{"lat":-7.5532,"lon":110.8061,"subtotal":50000,"order_type":"delivery"}'

# The old auth bypass must fail
curl -s -X POST https://<customer-domain>/api/v1/auth/admin-login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"owner@cafeolga.id","password":"password"}'
# expected: {"error":"unauthorized","success":false}

# Reading an order without a token must fail
curl -s -o /dev/null -w '%{http_code}\n' https://<customer-domain>/api/v1/orders/<any-uuid>
# expected: 401
```

Then, in the browser:

1. Open the customer PWA. It should ask for location permission, and show no
   delivery fee until you grant it or search an address.
2. Confirm the basket starts **empty**.
3. Place an order and confirm you are redirected to the Midtrans payment page.
4. Open the admin portal in a second window; the order should appear with a
   chime and a toast without reloading.
5. Advance it to "Diantar" and confirm it demands a courier first.

## Rollback

```bash
ssh -p 3022 79749-9425@gate.infra.dewacloud.com '
  cd /root/maremereso/backend
  pkill -f "maremereso/backend/server" || true
  mv server server.failed && mv server.old server
  nohup ./server > /var/log/maremereso-backend.log 2>&1 &
'
```

Note that the previous build cannot sign staff in either, because step 1
cleared the password hashes it relied on the removed bypass to accept. A full
rollback therefore means restoring the database backup from step 1 as well.
