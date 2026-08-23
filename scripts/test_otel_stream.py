#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""
Simulate live runtime application spans streaming into the Garuda daemon.
"""

import requests
import time
import json

GARUDA_ENDPOINT = "http://localhost:8080/api/v1/telemetry/spans"

spans_to_stream = [
    {
        "service_name": "checkout-service",
        "caller_symbol": "ProcessOrder",
        "caller_package": "github.com/myshra777-ai/garuda/internal/api",
        "target_endpoint": "postgres://localhost:5433/garuda_test",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    },
    {
        "service_name": "checkout-service",
        "caller_symbol": "ChargeCard",
        "caller_package": "github.com/myshra777-ai/garuda/internal/pool",
        "target_endpoint": "unapproved.stripe.payment.driver:443",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    }
]

def main():
    print(f"🛰️ Streaming {len(spans_to_stream)} runtime spans to Garuda at {GARUDA_ENDPOINT}...")
    headers = {
        "Content-Type": "application/json",
        "X-Garuda-Tenant-ID": "00000000-0000-0000-0000-000000000001"
    }
    
    try:
        resp = requests.post(GARUDA_ENDPOINT, json={"spans": spans_to_stream}, headers=headers, timeout=5)
        print(f"✓ Ingestion HTTP Status: {resp.status_code}")
        print(f"✓ Server Response: {resp.text}")
    except Exception as err:
        print(f"❌ Failed to stream spans: {err}")

if __name__ == "__main__":
    main()
