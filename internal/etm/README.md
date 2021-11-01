# etm

This package contains code retrieved from https://github.com/codahale/etm on 2021-10-28.  It implements the crypto.AEAD interface using AES-CBC encryption and sha hashing algorithms.  It was stripped of all aead constructions other than `AEAD_AES_256_CBC_HMAC_SHA_512` to fit our usage.