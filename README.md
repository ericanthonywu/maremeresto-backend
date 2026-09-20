# Mareme Group — Multi-Outlet Ordering System

Sistem pemesanan online untuk tiga outlet di Surakarta (Mareme Kerten, Mareme
Makamhaji, Mak Djan). Terdiri dari Customer PWA, Admin/Barista Portal, dan
backend Go dengan WebSocket real-time serta pembayaran Midtrans.

---

## Arsitektur

```
maremereso-olga/
├── maremeresto-backend/          # Go REST API + WebSocket
│   ├── cmd/
│   │   ├── server/               # Entry point, graceful shutdown
│   │   └── admintool/            # CLI: set staff passwords, list accounts
│   └── internal/
│       ├── apperror/             # Error types -> HTTP status + pesan pengguna
│       ├── config/               # Konfigurasi (wajib: JWT_SECRET, Midtrans key)
│       ├── controller/           # HTTP handler
│       ├── database/migrations/  # DDL + seed
│       ├── dto/                  # Request/response + validasi
│       ├── middleware/           # Auth, CORS, rate limit, logging, recovery
│       ├── model/                # Domain model
│       ├── repository/           # Query SQL, analytics, locking
│       ├── router/               # Routing + penerapan middleware
│       ├── service/              # Business logic
│       │   ├── delivery.go       # Haversine + tarif ongkir (sumber tunggal)
│       │   ├── geocode.go        # Proxy geocoder (cache + rate limit)
│       │   └── hours.go          # Jam operasional (zona Asia/Jakarta)
│       └── websocket/            # Hub dengan room ber-otorisasi
│
└── frontend/
    ├── customer/                 # React + Vite PWA (port 5173)
    └── admin/                    # React + Vite admin portal (port 5174)
```

---

## Menjalankan

### 1. Database

```bash
docker compose up -d          # PostgreSQL 16
```

Terapkan migrasi **berurutan**:

```bash
cd maremeresto-backend/internal/database/migrations
for f in 0000*.up.sql; do
  psql -h <host> -p <port> -U <user> -d maremereso_olga -v ON_ERROR_STOP=1 -f "$f"
done
```

### 2. Backend

```bash
cd maremeresto-backend
cp .env.example .env          # lalu isi JWT_SECRET dan kunci Midtrans
go run ./cmd/server           # http://localhost:8080
```

`JWT_SECRET` (min. 32 karakter), `DB_USER`, dan `MIDTRANS_SERVER_KEY` wajib ada.
Server menolak start bila salah satu kosong — ini disengaja agar kredensial
tidak pernah jatuh ke nilai default yang dapat diduga.

### 3. Password staf

Akun admin/owner tidak punya password bawaan. Tetapkan lewat CLI:

```bash
cd maremeresto-backend
go run ./cmd/admintool list                                  # lihat akun & status
go run ./cmd/admintool set-password owner@cafeolga.id        # password diminta interaktif
go run ./cmd/admintool set-password admin.kerten@cafeolga.id
go run ./cmd/admintool set-password admin.makamhaji@cafeolga.id
go run ./cmd/admintool set-password admin.makdjan@cafeolga.id
```

Password minimal 12 karakter dan harus memuat tiga dari empat kelas karakter
(huruf besar, huruf kecil, angka, simbol). Password tidak pernah diterima
sebagai argumen CLI agar tidak tercatat di shell history atau daftar proses.

### 4. Frontend

```bash
cd frontend/customer && npm install && npm run dev     # http://localhost:5173
cd frontend/admin    && npm install && npm run dev     # http://localhost:5174
```

Target proxy API dapat diarahkan ke backend lain:

```bash
echo "VITE_API_PROXY_TARGET=http://staging.internal:8080" > .env.local
```

---

## Fitur

### Customer PWA
- **Ongkir dari lokasi nyata.** Aplikasi meminta izin lokasi, lalu jarak dan
  tarif dihitung **di server** (`POST /delivery/quote`). Tidak ada koordinat
  default: tanpa lokasi, ongkir tidak ditampilkan sama sekali. Angka yang
  ditampilkan identik dengan yang ditagihkan karena keduanya berasal dari satu
  fungsi yang sama.
- **Pencarian alamat sungguhan.** Alamat yang diketik di-geocode (OSM Nominatim,
  diproksi backend), dan fix GPS di-reverse-geocode menjadi nama jalan.
- **Radius antar dipatuhi.** Alamat di luar `max_delivery_radius_km` outlet
  ditolak, dengan saran ambil sendiri atau pindah outlet.
- **Pembayaran nyata.** Checkout membuat transaksi Midtrans Snap lalu
  **mengalihkan pelanggan ke halaman pembayaran**. Status pesanan hanya maju
  setelah webhook Midtrans terverifikasi masuk.
- **Pelacakan real-time** melalui WebSocket per-pesanan, dengan polling sebagai
  jaring pengaman, serta pembatalan mandiri dalam 5 menit pertama.

### Admin Portal
- **CRUD menu penuh**: tambah, ubah, hapus, atur ketersediaan, unggah foto
  (otomatis dikompresi ke 800px JPEG), kategori, label, dan urutan tampil.
- **Notifikasi pesanan baru** di setiap halaman: toast yang menetap sampai
  ditutup, chime Web Audio, notifikasi desktop (opsional, izin diminta
  eksplisit), badge di sidebar dan judul tab. Hitungan belum-ditangani
  disimpan di database sehingga tidak hilang saat halaman dimuat ulang.
- **Alur pesanan** dengan state machine tervalidasi, optimistic locking, dan
  penugasan kurir wajib sebelum status "Diantar".
- **Dashboard** dari data nyata: omset per jam, perbandingan hari sebelumnya
  (ditampilkan kosong bila belum ada pembanding, bukan angka karangan), dan
  rincian per outlet untuk owner.
- **Pengaturan outlet**: jam operasional, tarif ongkir bertingkat, biaya
  layanan, minimum order, ambang gratis ongkir, kontak WhatsApp.

### Backend
- **Harga ditentukan server.** Subtotal, ongkir, biaya layanan, dan diskon
  dihitung ulang dari database saat pesanan dibuat; nilai dari klien diabaikan.
- **Isolasi cabang** ditegakkan pada setiap operasi baca dan tulis.
- **WebSocket ber-otorisasi.** Setiap room diperiksa: pelanggan hanya menerima
  pesanannya sendiri, admin cabang hanya outletnya, owner seluruh jaringan.
  Tidak ada broadcast ke semua socket.
- **Pencegahan bayar ganda**: `pg_advisory_xact_lock` per pesanan + header
  `Idempotency-Key`.
- **Verifikasi webhook Midtrans**: `SHA512(order_id + status_code +
  gross_amount + server_key)` dibandingkan secara constant-time di **semua**
  environment, ditambah pemeriksaan kecocokan nominal.
- **Rate limiting** per IP pada login, pembuatan pesanan, pembayaran, dan
  geocoding.

---

## Catatan keamanan

Beberapa hal yang perlu diketahui saat mengoperasikan sistem ini:

- **Tidak ada login demo.** Tombol "Quick Login" dan penerimaan password
  literal `"password"` sudah dihapus. Satu-satunya cara membuat kredensial
  adalah `admintool set-password`.
- **`.env` jangan di-commit.** Sudah masuk `.gitignore`. Bila `JWT_SECRET`
  pernah bocor, ganti dan seluruh sesi akan otomatis tidak valid.
- **Setel `APP_ENV=production`** di produksi: CORS dibatasi ke
  `CUSTOMER_URL`/`ADMIN_URL`, dan level log dinaikkan agar nomor telepon
  pelanggan tidak tercatat di log.
- **Webhook Midtrans** harus diarahkan ke `POST /api/v1/payments/notification`
  di dashboard Midtrans. Tanpa itu pesanan akan tetap berstatus menunggu
  pembayaran meski pelanggan sudah membayar.
- **Portal admin sebaiknya tidak terbuka ke internet publik** tanpa pembatasan
  tambahan (VPN, IP allowlist, atau autentikasi di reverse proxy).

---

## Pengujian

```bash
cd maremeresto-backend
go test ./...          # unit test: jarak, tarif, jam operasional, validasi
go vet ./...

cd frontend/customer && npm run build && npm run lint
cd frontend/admin    && npm run build && npm run lint
```

---

## Akun staf

| Role | Email |
|---|---|
| Owner HQ | `owner@cafeolga.id` |
| Admin Kerten | `admin.kerten@cafeolga.id` |
| Admin Makamhaji | `admin.makamhaji@cafeolga.id` |
| Admin Mak Djan | `admin.makdjan@cafeolga.id` |

Password ditetapkan oleh operator dengan `admintool set-password`.
