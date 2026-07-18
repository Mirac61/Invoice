#!/usr/bin/env bash
#
# Testet alle HTTP-Methoden der Invoice-API mit Mock-Daten.
# Voraussetzung: Server läuft (go run .) und `jq` ist installiert.
#
# Benutzung:
#   ./test-api.sh            # alle Endpunkte der Reihe nach
#   BASE=http://host:8080 ./test-api.sh

set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
API="$BASE/api/invoices"

# Farben nur, wenn wir in ein Terminal schreiben
if [ -t 1 ]; then
	BOLD=$'\e[1m'; DIM=$'\e[2m'; RESET=$'\e[0m'
else
	BOLD=""; DIM=""; RESET=""
fi

step() { echo; echo "${BOLD}==> $*${RESET}"; }

# call METHOD PATH [JSON-BODY]
# Gibt Statuscode + hübsch formatierten Body aus und liefert den Body zurück.
call() {
	local method="$1" path="$2" body="${3:-}"
	local args=(-sS -X "$method" -w $'\n%{http_code}')
	if [ -n "$body" ]; then
		args+=(-H "Content-Type: application/json" -d "$body")
	fi
	echo "${DIM}$method $path${RESET}" >&2

	local raw code out
	raw="$(curl "${args[@]}" "$BASE$path")"
	code="$(tail -n1 <<<"$raw")"
	out="$(sed '$d' <<<"$raw")"

	echo "  ${DIM}HTTP $code${RESET}" >&2
	if [ -n "$out" ]; then
		echo "$out" | jq . >&2 2>/dev/null || echo "  $out" >&2
	fi
	echo "$out"   # nur der Body geht auf stdout (für Weiterverarbeitung)
}

# ---- Mock-Daten -------------------------------------------------------------

new_invoice='{
  "paymentDueAt": "2026-09-01T00:00:00Z",
  "sender": {
    "name": "Muster GmbH", "street": "Hauptstr. 1", "zip": "10115",
    "city": "Berlin", "country": "DE", "email": "billing@muster.de", "taxId": "DE123456789"
  },
  "recipient": {
    "name": "Kunde AG", "street": "Marktplatz 5", "zip": "80331",
    "city": "München", "country": "DE", "email": "einkauf@kunde.ag"
  },
  "vatRate": 0.19,
  "items": [
    { "description": "Beratung", "quantity": 10, "unitPrice": 120.0, "unit": "Std" },
    { "description": "Lizenz",   "quantity": 1,  "unitPrice": 499.0, "unit": "Stück" }
  ],
  "notes": "Zahlbar innerhalb von 14 Tagen."
}'

# PUT ersetzt die ganze Rechnung (gleiche Struktur wie beim Anlegen)
put_invoice="${new_invoice/Beratung/Beratung (überarbeitet)}"

# PATCH ändert nur einzelne Felder
patch_invoice='{
  "vatRate": 0.07,
  "notes": "Reduzierter Steuersatz."
}'

# ---- Ablauf -----------------------------------------------------------------

step "POST   $API  (neue Rechnung anlegen)"
created="$(call POST "/api/invoices" "$new_invoice")"
id="$(jq -r '.id' <<<"$created")"
echo "${BOLD}Angelegte ID: $id${RESET}"

step "GET    $API  (alle Rechnungen)"
call GET "/api/invoices" >/dev/null

step "GET    $API/$id  (einzelne Rechnung)"
call GET "/api/invoices/$id" >/dev/null

step "PUT    $API/$id  (komplett ersetzen)"
call PUT "/api/invoices/$id" "$put_invoice" >/dev/null

step "PATCH  $API/$id  (teilweise ändern)"
call PATCH "/api/invoices/$id" "$patch_invoice" >/dev/null

step "DELETE $API/$id  (löschen)"
call DELETE "/api/invoices/$id" >/dev/null

step "GET    $API/$id  (sollte 404 sein)"
call GET "/api/invoices/$id" >/dev/null || true

echo; echo "${BOLD}Fertig.${RESET}"
