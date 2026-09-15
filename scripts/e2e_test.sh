#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
PGPASSWORD="${DB_PASSWORD:-postgres}"
export PGPASSWORD
PGUSER="${DB_USER:-postgres}"
PGDATABASE="${DB_NAME:-maremereso_olga}"
PGHOST="${DB_HOST:-localhost}"
PGPORT="${DB_PORT:-5432}"
SERVER_KEY="${MIDTRANS_SERVER_KEY:-dummy}"

db_exec() {
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -t -A -c "$1"
}

echo "=== E2E TESTING START ==="

# 1. Health check
echo "--> Testing GET /api/v1/health"
RES=$(curl -s "$BASE_URL/api/v1/health")
echo "$RES" | jq .
HEALTH_STATUS=$(echo "$RES" | jq -r .status)
if [ "$HEALTH_STATUS" != "ok" ]; then
    echo "ERROR: Health check failed"
    exit 1
fi

# 2. Public Branches
echo "--> Testing GET /api/v1/branches"
BRANCHES_RES=$(curl -s "$BASE_URL/api/v1/branches")
echo "$BRANCHES_RES" | jq .
BRANCH_ID=$(echo "$BRANCHES_RES" | jq -r '.data[0].id')
BRANCH_SLUG=$(echo "$BRANCHES_RES" | jq -r '.data[0].slug')

echo "--> Testing GET /api/v1/branches/$BRANCH_SLUG"
BRANCH_RES=$(curl -s "$BASE_URL/api/v1/branches/$BRANCH_SLUG")
echo "$BRANCH_RES" | jq .

# 3. Public Categories
echo "--> Testing GET /api/v1/categories"
CATS_RES=$(curl -s "$BASE_URL/api/v1/categories")
echo "$CATS_RES" | jq .
CAT_ID=$(echo "$CATS_RES" | jq -r '.data[0].id')

# 4. Public Menu by Branch
echo "--> Testing GET /api/v1/branches/$BRANCH_SLUG/menu"
MENU_RES=$(curl -s "$BASE_URL/api/v1/branches/$BRANCH_SLUG/menu")
echo "$MENU_RES" | jq .
MENU_ITEM_ID=$(echo "$MENU_RES" | jq -r '.data[0].id')
MENU_ITEM_NAME=$(echo "$MENU_RES" | jq -r '.data[0].name')
MENU_ITEM_PRICE=$(echo "$MENU_RES" | jq -r '.data[0].price')

# 5. Geocode & Quote & Promo
echo "--> Testing GET /api/v1/geocode/search"
GEO_SEARCH=$(curl -s "$BASE_URL/api/v1/geocode/search?q=Kerten")
echo "$GEO_SEARCH" | jq .

echo "--> Testing GET /api/v1/geocode/reverse"
GEO_REV=$(curl -s "$BASE_URL/api/v1/geocode/reverse?lat=-7.5597&lon=110.7942")
echo "$GEO_REV" | jq .

echo "--> Testing POST /api/v1/delivery/quote"
QUOTE_RES=$(curl -s -X POST "$BASE_URL/api/v1/delivery/quote" \
  -H "Content-Type: application/json" \
  -d "{\"branch_id\":\"$BRANCH_ID\",\"delivery_lat\":-7.5600,\"delivery_lon\":110.7950,\"subtotal\":50000}")
echo "$QUOTE_RES" | jq .

echo "--> Testing POST /api/v1/promos/validate"
PROMO_RES=$(curl -s -X POST "$BASE_URL/api/v1/promos/validate" \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"OLGACOFFEE\",\"subtotal\":50000,\"branch_id\":\"$BRANCH_ID\"}")
echo "$PROMO_RES" | jq .

# 6. Auth
echo "--> Testing Customer Login"
CUST_LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/customer-login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"081299998888\",\"name\":\"Test Customer\"}")
echo "$CUST_LOGIN" | jq .
CUST_TOKEN=$(echo "$CUST_LOGIN" | jq -r '.data.token')

echo "--> Testing GET /api/v1/auth/me (Customer)"
CUST_ME=$(curl -s -H "Authorization: Bearer $CUST_TOKEN" "$BASE_URL/api/v1/auth/me")
echo "$CUST_ME" | jq .

echo "--> Testing PUT /api/v1/auth/customer-profile"
CUST_PROF=$(curl -s -X PUT "$BASE_URL/api/v1/auth/customer-profile" \
  -H "Authorization: Bearer $CUST_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Test Customer Updated\",\"address\":\"Jl. Test No 123\",\"latitude\":-7.5600,\"longitude\":110.7950}")
echo "$CUST_PROF" | jq .

# Check DB for customer
DB_CUST_ADDR=$(db_exec "SELECT address FROM users WHERE phone = '+6281299998888';")
echo "DB Check Customer Address: $DB_CUST_ADDR"

echo "--> Testing Admin Login (Branch Admin Kerten)"
ADMIN_LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/admin-login" \
  -H "Content-Type: application/json" \
  -d "{\"identifier\":\"admin.kerten@cafeolga.id\",\"password\":\"Password123!\"}")
echo "$ADMIN_LOGIN" | jq .
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN" | jq -r '.data.token')

echo "--> Testing Admin Login (Owner HQ)"
OWNER_LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/admin-login" \
  -H "Content-Type: application/json" \
  -d "{\"identifier\":\"owner@cafeolga.id\",\"password\":\"Password123!\"}")
echo "$OWNER_LOGIN" | jq .
OWNER_TOKEN=$(echo "$OWNER_LOGIN" | jq -r '.data.token')

# 7. Create Order
echo "--> Testing POST /api/v1/orders (Auth customer)"
ORDER_REQ=$(cat <<EOF
{
  "branch_id": "$BRANCH_ID",
  "order_type": "delivery",
  "customer_name": "Test Customer Updated",
  "customer_phone": "081299998888",
  "delivery_address": "Jl. Test No 123",
  "delivery_notes": "Tolong antar di gerbang depan",
  "delivery_lat": -7.5600,
  "delivery_lon": 110.7950,
  "promo_code": "OLGACOFFEE",
  "items": [
    {
      "menu_item_id": "$MENU_ITEM_ID",
      "quantity": 2,
      "notes": "Extra hot"
    }
  ]
}
EOF
)
CREATE_ORDER_RES=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
  -H "Authorization: Bearer $CUST_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ORDER_REQ")
echo "$CREATE_ORDER_RES" | jq .
ORDER_ID=$(echo "$CREATE_ORDER_RES" | jq -r '.data.id')
ORDER_NUM=$(echo "$CREATE_ORDER_RES" | jq -r '.data.order_number')
GRAND_TOTAL=$(echo "$CREATE_ORDER_RES" | jq -r '.data.grand_total')

# Check DB for created order
DB_ORDER_STATUS=$(db_exec "SELECT status FROM orders WHERE id = '$ORDER_ID';")
echo "DB Order status: $DB_ORDER_STATUS"

echo "--> Testing GET /api/v1/orders/$ORDER_ID (Customer)"
GET_ORDER_RES=$(curl -s -H "Authorization: Bearer $CUST_TOKEN" "$BASE_URL/api/v1/orders/$ORDER_ID")
echo "$GET_ORDER_RES" | jq .

echo "--> Testing GET /api/v1/orders (Customer list)"
MY_ORDERS=$(curl -s -H "Authorization: Bearer $CUST_TOKEN" "$BASE_URL/api/v1/orders")
echo "$MY_ORDERS" | jq .

# 8. Create Payment
echo "--> Testing POST /api/v1/payments"
PAYMENT_REQ=$(cat <<EOF
{
  "order_id": "$ORDER_ID",
  "payment_method": "qris",
  "idempotency_key": "idempotency-key-test-1234"
}
EOF
)
CREATE_PAYMENT_RES=$(curl -s -X POST "$BASE_URL/api/v1/payments" \
  -H "Authorization: Bearer $CUST_TOKEN" \
  -H "Idempotency-Key: idempotency-key-test-1234" \
  -H "Content-Type: application/json" \
  -d "$PAYMENT_REQ")
echo "$CREATE_PAYMENT_RES" | jq .
PAYMENT_ID=$(echo "$CREATE_PAYMENT_RES" | jq -r '.data.id')

# Test payment idempotency
echo "--> Testing POST /api/v1/payments (Idempotency duplicate test)"
DUP_PAYMENT_RES=$(curl -s -X POST "$BASE_URL/api/v1/payments" \
  -H "Authorization: Bearer $CUST_TOKEN" \
  -H "Idempotency-Key: idempotency-key-test-1234" \
  -H "Content-Type: application/json" \
  -d "$PAYMENT_REQ")
echo "$DUP_PAYMENT_RES" | jq .

echo "--> Testing GET /api/v1/orders/$ORDER_ID/payment"
GET_PAY_STATUS=$(curl -s -H "Authorization: Bearer $CUST_TOKEN" "$BASE_URL/api/v1/orders/$ORDER_ID/payment")
echo "$GET_PAY_STATUS" | jq .

# Check DB payment
DB_PAY_STATUS=$(db_exec "SELECT status FROM payments WHERE order_id = '$ORDER_ID';")
echo "DB Payment Status: $DB_PAY_STATUS"

# 9. Midtrans Notification Webhook
echo "--> Testing POST /api/v1/payments/notification (Simulating Midtrans settlement webhook)"
GROSS_AMT_STR="${GRAND_TOTAL}.00"
STATUS_CODE="200"
RAW_SIG_INPUT="${ORDER_NUM}${STATUS_CODE}${GROSS_AMT_STR}${SERVER_KEY}"
SIG_KEY=$(python3 -c "import hashlib; print(hashlib.sha512('$RAW_SIG_INPUT'.encode()).hexdigest())")

MIDTRANS_NOTIF=$(cat <<EOF
{
  "order_id": "$ORDER_NUM",
  "status_code": "$STATUS_CODE",
  "gross_amount": "$GROSS_AMT_STR",
  "signature_key": "$SIG_KEY",
  "transaction_status": "settlement",
  "fraud_status": "accept",
  "payment_type": "qris",
  "transaction_id": "midtrans-tx-12345"
}
EOF
)
WEBHOOK_RES=$(curl -s -X POST "$BASE_URL/api/v1/payments/notification" \
  -H "Content-Type: application/json" \
  -d "$MIDTRANS_NOTIF")
echo "$WEBHOOK_RES" | jq .

# DB Check after webhook
DB_ORDER_STATUS2=$(db_exec "SELECT status FROM orders WHERE id = '$ORDER_ID';")
echo "DB Order status after payment webhook: $DB_ORDER_STATUS2"

# 10. Admin Endpoints
echo "--> Testing GET /api/v1/admin/dashboard"
ADMIN_DASH=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/v1/admin/dashboard")
echo "$ADMIN_DASH" | jq .

echo "--> Testing GET /api/v1/admin/orders"
ADMIN_ORDERS=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/v1/admin/orders")
echo "$ADMIN_ORDERS" | jq .

echo "--> Testing GET /api/v1/admin/orders/$ORDER_ID"
ADMIN_GET_ORDER=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/v1/admin/orders/$ORDER_ID")
echo "$ADMIN_GET_ORDER" | jq .

echo "--> Testing POST /api/v1/admin/orders/acknowledge"
ACK_RES=$(curl -s -X POST "$BASE_URL/api/v1/admin/orders/acknowledge" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"order_ids\":[\"$ORDER_ID\"]}")
echo "$ACK_RES" | jq .

echo "--> Testing PUT /api/v1/admin/orders/$ORDER_ID/status (Update status to completed)"
# Fetch latest version from DB
CURRENT_VER=$(db_exec "SELECT version FROM orders WHERE id = '$ORDER_ID';")
UPDATE_STATUS_RES=$(curl -s -X PUT "$BASE_URL/api/v1/admin/orders/$ORDER_ID/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"status\":\"completed\",\"expected_version\":$CURRENT_VER}")
echo "$UPDATE_STATUS_RES" | jq .

# Check DB Order Status
DB_ORDER_STATUS3=$(db_exec "SELECT status FROM orders WHERE id = '$ORDER_ID';")
echo "DB Order status after admin update: $DB_ORDER_STATUS3"

# 11. Customer Feedback
echo "--> Testing PUT /api/v1/orders/$ORDER_ID/feedback"
FEEDBACK_RES=$(curl -s -X PUT "$BASE_URL/api/v1/orders/$ORDER_ID/feedback" \
  -H "Authorization: Bearer $CUST_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"rating\":5,\"comment\":\"Great coffee!\"}")
echo "$FEEDBACK_RES" | jq .

# Check DB Feedback
DB_FB_RATING=$(db_exec "SELECT rating FROM order_feedback WHERE order_id = '$ORDER_ID';")
echo "DB Feedback Rating: $DB_FB_RATING"

# 12. Create another order to test Cancel
echo "--> Creating second order for Cancel test"
ORDER2_RES=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
  -H "Authorization: Bearer $CUST_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ORDER_REQ")
ORDER2_ID=$(echo "$ORDER2_RES" | jq -r '.data.id')

echo "--> Testing POST /api/v1/orders/$ORDER2_ID/cancel"
CANCEL_RES=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER2_ID/cancel" \
  -H "Authorization: Bearer $CUST_TOKEN")
echo "$CANCEL_RES" | jq .

DB_ORDER2_STATUS=$(db_exec "SELECT status FROM orders WHERE id = '$ORDER2_ID';")
echo "DB Order 2 Status: $DB_ORDER2_STATUS"

# 13. Category Management (Admin)
echo "--> Testing POST /api/v1/admin/categories"
NEW_CAT_RES=$(curl -s -X POST "$BASE_URL/api/v1/admin/categories" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Test Category\",\"emoji\":\"🧪\",\"sort_order\":99}")
echo "$NEW_CAT_RES" | jq .
NEW_CAT_ID=$(echo "$NEW_CAT_RES" | jq -r '.data.id')

echo "--> Testing PUT /api/v1/admin/categories/$NEW_CAT_ID"
UPDATE_CAT_RES=$(curl -s -X PUT "$BASE_URL/api/v1/admin/categories/$NEW_CAT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Updated Test Category\",\"emoji\":\"🧪\",\"sort_order\":99}")
echo "$UPDATE_CAT_RES" | jq .

echo "--> Testing DELETE /api/v1/admin/categories/$NEW_CAT_ID"
DEL_CAT_RES=$(curl -s -X DELETE "$BASE_URL/api/v1/admin/categories/$NEW_CAT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$DEL_CAT_RES" | jq .

# 14. Menu Item Management (Admin)
echo "--> Testing POST /api/v1/admin/menu"
NEW_MENU_RES=$(curl -s -X POST "$BASE_URL/api/v1/admin/menu" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"branch_id\":\"$BRANCH_ID\",\"category_id\":\"$CAT_ID\",\"name\":\"Test Item\",\"description\":\"Test Desc\",\"price\":15000,\"is_available\":true}")
echo "$NEW_MENU_RES" | jq .
NEW_MENU_ID=$(echo "$NEW_MENU_RES" | jq -r '.data.id')

echo "--> Testing PUT /api/v1/admin/menu/$NEW_MENU_ID/availability"
TOGGLE_MENU_RES=$(curl -s -X PUT "$BASE_URL/api/v1/admin/menu/$NEW_MENU_ID/availability" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"is_available\":false}")
echo "$TOGGLE_MENU_RES" | jq .

echo "--> Testing DELETE /api/v1/admin/menu/$NEW_MENU_ID"
DEL_MENU_RES=$(curl -s -X DELETE "$BASE_URL/api/v1/admin/menu/$NEW_MENU_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$DEL_MENU_RES" | jq .

# 15. Settings & Profile Management (Admin)
echo "--> Testing GET /api/v1/admin/settings"
GET_SETTING_RES=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/v1/admin/settings")
echo "$GET_SETTING_RES" | jq .

echo "--> Testing PUT /api/v1/admin/settings"
SETTING_REQ=$(cat <<EOF
{
  "max_delivery_radius_km": 10,
  "base_delivery_fee_near": 0,
  "base_delivery_fee_mid": 8000,
  "base_delivery_fee_far": 12000,
  "near_threshold_km": 1,
  "mid_threshold_km": 5,
  "service_fee": 2000,
  "min_order_amount": 20000,
  "free_delivery_threshold": 0,
  "whatsapp_number": "081234567890",
  "description": "Updated Description",
  "halal_certificate_id": "HALAL-123456"
}
EOF
)
UPDATE_SETTING_RES=$(curl -s -X PUT "$BASE_URL/api/v1/admin/settings" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$SETTING_REQ")
echo "$UPDATE_SETTING_RES" | jq .

echo "--> Testing PUT /api/v1/admin/branches/$BRANCH_ID/status"
BRANCH_STATUS_RES=$(curl -s -X PUT "$BASE_URL/api/v1/admin/branches/$BRANCH_ID/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"is_open\":true}")
echo "$BRANCH_STATUS_RES" | jq .

echo "--> Testing PUT /api/v1/admin/branches/$BRANCH_ID/profile"
BRANCH_PROF_RES=$(curl -s -X PUT "$BASE_URL/api/v1/admin/branches/$BRANCH_ID/profile" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Mareme Kerten Updated\",\"phone\":\"0271-712301\",\"address\":\"Jl. Samratulangi No.65\",\"latitude\":-7.5597,\"longitude\":110.7942}")
echo "$BRANCH_PROF_RES" | jq .

# 16. Owner HQ Endpoints
echo "--> Testing GET /api/v1/owner/dashboard"
OWNER_DASH=$(curl -s -H "Authorization: Bearer $OWNER_TOKEN" "$BASE_URL/api/v1/owner/dashboard")
echo "$OWNER_DASH" | jq .

echo "--> Testing GET /api/v1/owner/orders"
OWNER_ORDERS=$(curl -s -H "Authorization: Bearer $OWNER_TOKEN" "$BASE_URL/api/v1/owner/orders")
echo "$OWNER_ORDERS" | jq .

echo "=== E2E TESTING COMPLETE ==="
