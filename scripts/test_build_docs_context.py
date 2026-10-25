#!/usr/bin/env python3
"""Regression checks for the structured capability metadata loader."""

import unittest

from build_docs_context import flatten_capabilities


class FlattenCapabilitiesTests(unittest.TestCase):
    def test_category_list_schema_preserves_category_and_name(self):
        rows = flatten_capabilities({
            "categories": [{
                "name": "Talon",
                "capabilities": [{
                    "name": "Pre-edit briefing",
                    "status": "implemented",
                }],
            }]
        })
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0]["name"], "Pre-edit briefing")
        self.assertEqual(rows[0]["category"], "Talon")
        self.assertEqual(rows[0]["status"], "implemented")

    def test_legacy_name_to_capability_map_still_loads(self):
        rows = flatten_capabilities({
            "capabilities": {
                "Runtime evidence": {
                    "status": "beta",
                    "supported_semantics": "Captured spans only",
                }
            }
        })
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0]["name"], "Runtime evidence")
        self.assertEqual(rows[0]["status"], "beta")

    def test_empty_current_category_list_is_valid(self):
        self.assertEqual(flatten_capabilities({"categories": []}), [])


if __name__ == "__main__":
    unittest.main(verbosity=2)
