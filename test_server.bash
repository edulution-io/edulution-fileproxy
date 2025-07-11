#!/bin/bash

declare -A USERS
USERS["netzint1"]="Muster!"
USERS["netzint2"]="Muster!"
USERS["netzint3"]="Muster!"

check_status() {
    local status="$1"
    local step="$2"
    if [[ "$status" =~ ^2 ]]; then
        echo "✅ $step erfolgreich (Status $status)"
    else
        echo "❌ $step fehlgeschlagen (Status $status)"
        exit 1
    fi
}

timed_step() {
    local label="$1"
    shift
    local start end duration status

    start=$(date +%s%3N)  # Zeit in Millisekunden

    # führe den Befehl aus, gib nur Statuscode zurück
    status=$("$@" 2>/dev/null)
    
    end=$(date +%s%3N)
    duration=$((end - start))

    check_status "$status" "$label"
    echo "⏱️ Dauer für $label: ${duration} ms"
    echo
}


run_webdav_test() {
    USER="$1"
    PASS="$2"
    DAV_URL="$3"
    TMPDIR=$(mktemp -d)
    LOCALFILE="$TMPDIR/testfile.txt"
    LOCALCOPY="$TMPDIR/testfile_copy.txt"
    DAV_FILE="testfile.txt"
    RENAMED_FILE="renamed.txt"
    SUBDIR="subfolder"

    echo "🔧 Starte WebDAV-Test für $USER"

    echo "📁 Erstelle Ordner $SUBDIR"
    timed_step "Ordnererstellung ($USER)" \
        curl -k -u "$USER:$PASS" -X MKCOL -s -o /dev/null -w "%{http_code}" "$DAV_URL/$SUBDIR"

    echo "📤 Lade Datei hoch"
    echo "Inhalt $(date)" > "$LOCALFILE"
    timed_step "Datei-Upload ($USER)" \
        curl -k -u "$USER:$PASS" -T "$LOCALFILE" -s -o /dev/null -w "%{http_code}" "$DAV_URL/$DAV_FILE"

    echo "📥 Lade Datei herunter"
    start=$(date +%s%3N)
    status=$(curl -k -u "$USER:$PASS" -o "$LOCALCOPY" -s -w "%{http_code}" "$DAV_URL/$DAV_FILE")
    end=$(date +%s%3N)
    duration=$((end - start))
    check_status "$status" "Datei-Download ($USER)"
    echo "⏱️ Dauer für Datei-Download ($USER): ${duration} ms"
    diff "$LOCALFILE" "$LOCALCOPY" && echo "✅ Dateiinhalt stimmt überein ($USER)"
    echo

    echo "✏️ Benenne Datei um"
    timed_step "Datei-Umbenennung ($USER)" \
        curl -k -u "$USER:$PASS" -X MOVE -H "Destination: $DAV_URL/$RENAMED_FILE" -s -o /dev/null -w "%{http_code}" "$DAV_URL/$DAV_FILE"

    echo "🗑️ Lösche Datei"
    timed_step "Datei-Löschung ($USER)" \
        curl -k -u "$USER:$PASS" -X DELETE -s -o /dev/null -w "%{http_code}" "$DAV_URL/$RENAMED_FILE"

    echo "🗑️ Lösche Ordner"
    timed_step "Ordner-Löschung ($USER)" \
        curl -k -u "$USER:$PASS" -X DELETE -s -o /dev/null -w "%{http_code}" "$DAV_URL/$SUBDIR"

    rm -rf "$TMPDIR"
    echo "✅ WebDAV-Test abgeschlossen für $USER"
    echo "========================================"
    echo
}

for USERNAME in "${!USERS[@]}"; do
    PASSWORD="${USERS[$USERNAME]}"
    echo "🔐 Teste Benutzer: $USERNAME"

    # Setze Variablen dynamisch
    DAV_URL="https://10.0.0.3:8443/webdav/default-school/students/niclass/$USERNAME"
    USER="$USERNAME"
    PASS="$PASSWORD"

    # Dann ruf das Testszenario hier auf – z. B. als Funktion
    run_webdav_test "$USER" "$PASS" "$DAV_URL"
done