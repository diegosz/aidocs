---
title: "Blind Records Encoding"
description: "Encoding format for blind records"
category: "reference"
tags: ["encoding", "format", "binary"]
estimatedTokens: 450
---

# Blind Records Encoding

This document describes the encoding format used for blind records.

## Binary Format

The records use a compact binary encoding optimized for storage and transmission.

## Field Layout

Each record contains:
- Header (fixed size)
- Payload (variable size)
- Signature (fixed size)
