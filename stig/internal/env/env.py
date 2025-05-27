"""
Package env provides functions to lookup environment variables.
"""

import os

def lookup(key: str) -> str:
  return os.getenv(key)

def lookupInt(key: str) -> int:
  return int(lookup(key))

def lookupFloat(key: str) -> float:
  return float(lookup(key))

def lookupBool(key: str) -> bool:
  return lookup(key).lower() == "true"
