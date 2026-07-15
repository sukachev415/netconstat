#!/bin/bash
# Download free ASN database (O-X-L, no registration required)
set -e

ASN_DIR="$(dirname "$0")/../data/asn"
mkdir -p "$ASN_DIR"

echo "Downloading ASN database from O-X-L..."
curl -L "https://geoip.oxl.app/file/asn_ipv4_small.mmdb.zip" -o /tmp/asn.zip
unzip -o /tmp/asn.zip -d "$ASN_DIR/"
rm -f /tmp/asn.zip

echo "ASN database saved to $ASN_DIR/"
ls -lh "$ASN_DIR/"*.mmdb
