"""Shared pytest fixtures for the Debian translator tests.

T-04-08: the preseed emitter refuses to run without DEBATEOS_HASHED_PASSWORD
(no default password may be baked into an image). Tests that emit a profile
tree need a valid crypt hash in the environment; this autouse fixture supplies
a throwaway SHA-512 test hash so the emit path exercises real substitution.
Tests that assert the hard-fail behavior unset the var themselves.
"""

import os

import pytest

# A syntactically valid SHA-512 crypt hash (throwaway, test-only).
_TEST_HASHED_PASSWORD = (
    "$6$rounds=656000$debateostest$"
    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789./"
)


@pytest.fixture(autouse=True)
def _debateos_test_password(monkeypatch):
    monkeypatch.setenv("DEBATEOS_HASHED_PASSWORD", _TEST_HASHED_PASSWORD)
