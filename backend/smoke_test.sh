#!/usr/bin/env bash
set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
INVOICES="$BASE_URL/api/invoices"
RESTART_CMD="${RESTART_CMD:-docker compose restart backend}"

passed=0
failed=0

red() { printf '\033[31m%s\033[0m\n' "$1"; }
green() { printf '\033[32m%s\033[0m\n' "$1"; }
bold() { printf '\033[1m%s\033[0m\n' "$1"; }

check() {
    local label="$1" expected="$2" actual="$3"
    if [[ "$expected" == "$actual" ]]; then
        green "  PASS  $label"
        passed=$((passed + 1))
    else
        red "  FAIL  $label"
        red "        expected: $expected"
        red "        actual:   $actual"
        failed=$((failed + 1))
    fi
}

# Splits curl output into body and status code. Sets BODY and STATUS.
request() {
    local method="$1" url="$2" data="${3:-}"
    local response
    if [[ -n "$data" ]]; then
        response=$(curl -sS -X "$method" "$url" \
            -H 'Content-Type: application/json' \
            -d "$data" \
            -w $'\n%{http_code}')
    else
        response=$(curl -sS -X "$method" "$url" -w $'\n%{http_code}')
    fi
    STATUS="${response##*$'\n'}"
    BODY="${response%$'\n'*}"
}

field() {
    echo "$BODY" | jq -r "$1"
}

command -v jq >/dev/null || { red "jq is required"; exit 1; }

due_date=$(date -u -v+14d '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null \
    || date -u -d '+14 days' '+%Y-%m-%dT%H:%M:%SZ')

draft_payload=$(cat <<EOF
{
  "paymentDueAt": "$due_date",
  "sender": {
    "name": "Sender GmbH", "street": "Hauptstr. 1", "zip": "70173",
    "city": "Stuttgart", "country": "DE", "email": "billing@sender.de"
  },
  "recipient": {
    "name": "Recipient GmbH", "street": "Nebenstr. 2", "zip": "70174",
    "city": "Stuttgart", "country": "DE"
  },
  "items": [
    { "description": "Beratung", "quantity": 3, "unitPrice": 150, "unit": "h" }
  ],
  "vatRate": 0.19,
  "notes": "initial"
}
EOF
)

bold "== 1. POST /invoices creates a draft =="
request POST "$INVOICES" "$draft_payload"
check "status code" "201" "$STATUS"
check "status is draft" "draft" "$(field '.status')"
check "net total" "450" "$(field '.netTotal')"
check "vat amount" "85.5" "$(field '.vatAmount')"
check "gross total" "535.5" "$(field '.grossTotal')"
check "invoice number empty on draft" "" "$(field '.invoiceNumber')"
check "item position is 1-based" "1" "$(field '.items[0].position')"
check "item total" "450" "$(field '.items[0].total')"
check "item has an id" "true" "$(field '.items[0].id | length > 0')"

invoice_id=$(field '.id')
created_at=$(field '.createdAt')
[[ -n "$invoice_id" && "$invoice_id" != "null" ]] || { red "no invoice id, aborting"; exit 1; }
echo "  invoice id: $invoice_id"

bold "== 2. GET /invoices/:id returns the same data =="
request GET "$INVOICES/$invoice_id"
check "status code" "200" "$STATUS"
check "gross total" "535.5" "$(field '.grossTotal')"
check "items are loaded" "1" "$(field '.items | length')"
check "notes" "initial" "$(field '.notes')"

bold "== 3. GET /invoices includes items (not null) =="
request GET "$INVOICES"
check "status code" "200" "$STATUS"
check "invoice is in the list" "1" "$(field "[.[] | select(.id == \"$invoice_id\")] | length")"
check "items are not null" "1" "$(field "[.[] | select(.id == \"$invoice_id\")][0].items | length")"

bold "== 4. PATCH notes leaves everything else untouched =="
request PATCH "$INVOICES/$invoice_id" '{"notes": "please pay by end of month"}'
check "status code" "200" "$STATUS"
check "notes updated" "please pay by end of month" "$(field '.notes')"
check "gross total unchanged" "535.5" "$(field '.grossTotal')"
check "recipient unchanged" "Recipient GmbH" "$(field '.recipient.name')"
check "created at unchanged" "$created_at" "$(field '.createdAt')"

bold "== 5. PATCH items recalculates totals and renumbers positions =="
request PATCH "$INVOICES/$invoice_id" \
    '{"items": [{"description": "Buch", "quantity": 1, "unitPrice": 20}, {"description": "Versand", "quantity": 2, "unitPrice": 5}], "vatRate": 0.07}'
check "status code" "200" "$STATUS"
check "net total" "30" "$(field '.netTotal')"
check "vat amount" "2.1" "$(field '.vatAmount')"
check "gross total" "32.1" "$(field '.grossTotal')"
check "item count" "2" "$(field '.items | length')"
check "first position" "1" "$(field '.items[0].position')"
check "second position" "2" "$(field '.items[1].position')"
check "new items got ids" "true" "$(field '[.items[].id | length > 0] | all')"

bold "== 6. PUT cannot overwrite server-managed fields =="
request PUT "$INVOICES/$invoice_id" "$(cat <<EOF
{
  "status": "paid",
  "invoiceNumber": "HACKED-001",
  "createdAt": "1999-01-01T00:00:00Z",
  "paymentDueAt": "$due_date",
  "sender": {"name": "Sender GmbH", "street": "Hauptstr. 1", "zip": "70173", "city": "Stuttgart", "country": "DE"},
  "recipient": {"name": "Renamed GmbH", "street": "Nebenstr. 2", "zip": "70174", "city": "Stuttgart", "country": "DE"},
  "items": [{"description": "Beratung", "quantity": 1, "unitPrice": 100}],
  "vatRate": 0.19
}
EOF
)"
check "status code" "200" "$STATUS"
check "status still draft" "draft" "$(field '.status')"
check "invoice number not hijacked" "" "$(field '.invoiceNumber')"
check "created at not hijacked" "$created_at" "$(field '.createdAt')"
check "recipient was updated" "Renamed GmbH" "$(field '.recipient.name')"

bold "== 7. Invalid input is rejected =="
request PATCH "$INVOICES/$invoice_id" '{"vatRate": 1.5}'
check "vat rate above 1 rejected" "400" "$STATUS"
request PATCH "$INVOICES/$invoice_id" '{"items": [{"description": "X", "quantity": -1, "unitPrice": 10}]}'
check "negative quantity rejected" "400" "$STATUS"
request GET "$INVOICES/00000000-0000-0000-0000-000000000000"
check "unknown id returns 404" "404" "$STATUS"

bold "== 8. POST /invoices/:id/issue assigns a number =="
request POST "$INVOICES/$invoice_id/issue"
check "status code" "200" "$STATUS"
check "status is issued" "issued" "$(field '.status')"
check "invoice number format" "true" "$(field '.invoiceNumber | test("^[0-9]{4}-[0-9]{4}$")')"
check "issued at is set" "true" "$(field '.issuedAt != null and (.issuedAt | startswith("0001") | not)')"
issued_number=$(field '.invoiceNumber')
echo "  invoice number: $issued_number"

bold "== 9. Issued invoices are frozen =="
request POST "$INVOICES/$invoice_id/issue"
check "second issue returns 409" "409" "$STATUS"
request PATCH "$INVOICES/$invoice_id" '{"notes": "too late"}'
check "patch after issue returns 409" "409" "$STATUS"
request PUT "$INVOICES/$invoice_id" "$draft_payload"
check "put after issue returns 409" "409" "$STATUS"
request DELETE "$INVOICES/$invoice_id"
check "delete after issue returns 409" "409" "$STATUS"

bold "== 10. Invoice numbers are sequential =="
request POST "$INVOICES" "$draft_payload"
second_id=$(field '.id')
request POST "$INVOICES/$second_id/issue"
second_number=$(field '.invoiceNumber')
expected_next=$(awk -F- -v y="${issued_number%%-*}" '{printf "%s-%04d", y, $2 + 1}' <<<"$issued_number")
check "next number follows previous" "$expected_next" "$second_number"

bold "== 11. Drafts can be deleted =="
request POST "$INVOICES" "$draft_payload"
throwaway_id=$(field '.id')
request DELETE "$INVOICES/$throwaway_id"
check "status code" "204" "$STATUS"
request GET "$INVOICES/$throwaway_id"
check "gone afterwards" "404" "$STATUS"

bold "== 12. Data survives a restart =="
if [[ "${SKIP_RESTART:-0}" == "1" ]]; then
    echo "  skipped (SKIP_RESTART=1)"
else
    echo "  running: $RESTART_CMD"
    if $RESTART_CMD >/dev/null 2>&1; then
        for _ in {1..30}; do
            request GET "$INVOICES" && [[ "$STATUS" == "200" ]] && break
            sleep 1
        done
        request GET "$INVOICES/$invoice_id"
        check "invoice still exists" "200" "$STATUS"
        check "invoice number persisted" "$issued_number" "$(field '.invoiceNumber')"
        check "items persisted" "1" "$(field '.items | length')"
        check "status persisted" "issued" "$(field '.status')"
        request GET "$INVOICES/$throwaway_id"
        check "deleted invoice stays deleted" "404" "$STATUS"
    else
        red "  restart command failed, skipping persistence checks"
        failed=$((failed + 1))
    fi
fi

echo
bold "== Summary =="
green "passed: $passed"
if (( failed > 0 )); then
    red "failed: $failed"
    exit 1
fi
green "failed: 0"
