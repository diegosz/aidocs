# Blind Keys

- [Blind Keys](#blind-keys)
  - [Platform Keys](#platform-keys)
    - [Platform Record Repository Fallback Key (`PRRFKEY`)](#platform-record-repository-fallback-key-prrfkey)
  - [Tenant Keys](#tenant-keys)
    - [Tenant Record Repository Fallback Key (`TRRFKEY`)](#tenant-record-repository-fallback-key-trrfkey)
    - [Tenant Central Blind Signer Key (`TCBSKEY`)](#tenant-central-blind-signer-key-tcbskey)
    - [Blinder Tenant Key (`BTKEY`)](#blinder-tenant-key-btkey)
    - [Tenant Attribute Based Encryption](#tenant-attribute-based-encryption)
  - [Device Keys](#device-keys)
  - [User Keys](#user-keys)

For the **Blind Records System** to function, we need to have a set of **Keys** that are used to provide the different services and capabilities.
Some of the **Keys** are used by the **Users**, some by the **Devices** and others by the **Central Platform** services at the **Platform** or **Tenant** level.

The **Keys** are encoded using the [JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517) format.
Some `ABE` (Attribute Based Encryption) **Keys** are also encoded using `JSON Compact Serialization` semantics, similar to `JWK`.

> **Edge** means the **Device** side.

## Platform Keys

The **Platform Keys** are the keys used to provide services and capabilities for all the **Tenants** managed by the **Platform**.

| Key       | Description                             | Edge   | Platform |
|-----------|-----------------------------------------|--------|----------|
| `PRRFKEY` | Platform Record Repository Fallback Key | Public | Private  |

### Platform Record Repository Fallback Key (`PRRFKEY`)

The `PRRFKEY` (Platform Record Repository Fallback Key) is an `elliptic.P256` key pair, used as last resort **Platform** fallback key to recover the `RKEY` (Record Key) from the `PFRKEY` (Platform Fallback Record Key) in each record.

The private fallback key `PRRFKEY` is used by the **Platform** to recover the `RKEY` record key from the `PFRKEY` attached to each `Local` record.

The public fallback key `PRRFKEY` is used by the **Edge** to encrypt the `RKEY` inside the `PFRKEY` recovery record key attached to the `Local` records.

> The **Platform** fallback recovery keys could be disabled by the **Platform** for all **Tenants** or each **Tenant** could disable it individually for it self.

## Tenant Keys

The Tenant Keys are the keys used to provide services and capabilities for each **Tenant**.

| Key          | Description                                         | Edge           | Platform |
|--------------|-----------------------------------------------------|----------------|----------|
| `TRRFKEY`    | Tenant Record Repository Fallback Key               | Public         | Private  |
| `TCBSKEY`    | Tenant Central Blind Signer Key                     | Public         | Private  |
| `BTKEY`      | Blinder Tenant Key                                  | Private/Public | -        |
| `TABESCHEME` | Tenant Attribute Based Encryption Scheme            | Present        | Present  |
| `TABEMSK`    | Tenant Attribute Based Encryption Master Secret Key |                |          |
| `TABEMPK`    | Tenant Attribute Based Encryption Master Public Key |                |          |

### Tenant Record Repository Fallback Key (`TRRFKEY`)

The `TRRFKEY` (Tenant Record Repository Fallback Key) is an `elliptic.P256` key pair, used as last resort **Tenant** fallback key to recover the `RKEY` from the `TFRKEY` (Tenant Fallback Record Key) in each record.

The private fallback key `TRRFKEY` is used by the **Tenant** to recover the `RKEY` record key from the `TFRKEY` attached to each `Local` record.

The public fallback key `TRRFKEY` is used by the **Edge** to encrypt the `RKEY` inside the `TFRKEY` recovery record key attached to the `Local` records.

> The **Tenant** fallback recovery keys could be disabled by the **Platform** for all **Tenants** or each **Tenant** could disable it individually for it self.

### Tenant Central Blind Signer Key (`TCBSKEY`)

The `TCBSKEY` (Tenant Central Blind Signer Key) is an `EdDSA elliptic ed25519` key pair, used to sign and verify the `Central` records created in the **Central Platform** at the **Tenant** level.

The `TCBSKEY` public key is distributed to the **Devices** during provisioning, for doing the `Central` records signature verification.

### Blinder Tenant Key (`BTKEY`)

The `BTKEY` (Blinder Tenant Key) is an `elliptic.P256` key pair, used at the **Edge** to encrypt/decrypt the `BRKEY` (Blinded Record Key) in/from the `BTRKEY` (Blinded Tenant Record Key) that is attached to each `Local` record.

The `BTKEY` is deployed on the **Devices** during provisioning, and is only used in the **Devices**.

### Tenant Attribute Based Encryption

Attributed Based Encryption (`ABE`) allows for the generation of a decryption key based on the definition of attributes in a **Policy**.

The `TABESCHEME` (Tenant Attribute Based Encryption Scheme) is a `CP-ABE FAME` (Ciphertext-Policy Attribute-Based Encryption with Forward and Attribute Modification Equivalence) scheme to embed **Policies** into our data, and supporting **Users** providing **Credentials** with various attributes to define the claims they have to access the data.

The `TABEMSK` (Tenant Attribute Based Encryption Master Secret Key) is used by the **Central Platform** in a **Tenant** level to create **User** **Credentials** with attributes that should match the **Policy** to decrypt the payloads.

The `TABEMPK` (Tenant Attribute Based Encryption Master Public Key) is used in to encrypt payloads with an attributes boolean expression **Policy**.

## Device Keys

The **Device Keys** are the keys provisioned to the **Devices**.

| Key      | Description | Edge    | Platform |
|----------|-------------|---------|----------|
| `DEVKEY` | Device Key  | Private | Public   |

The `DEVKEY` (Device Key) is an `EdDSA elliptic ed25519` key pair, used to sign and verify the `Local` records created in the **Device**.

## User Keys

The **User Keys** are the keys generated for each **User**.

| Key      | Description | User    | Platform |
|----------|-------------|---------|----------|
| `USRKEY` | User Key    | Private | Public   |

The `USRKEY` (User Key) is an `EdDSA elliptic ed25519` key pair, used to sign and verify the `Local` records created by each **User**.
