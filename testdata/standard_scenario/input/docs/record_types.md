# Record Types

- [Record Types](#record-types)
  - [Mutable Records](#mutable-records)
  - [Inmutable Records](#inmutable-records)
  - [Plain Records](#plain-records)
  - [Affixed Records](#affixed-records)
  - [Local Records](#local-records)
  - [Central Records](#central-records)

The **Blind Records** are classified in types bases on three different criteria, one based on their mutability, the other based on how their content is encoded, and lastly based on the signature of the record.

- Depending on the mutability we have two types of records: **Mutable** and **Inmutable**.
  This classification is defined at the application level.
  For example a `Patient Record` is `Mutable`, and a `Study Record` is `Inmutable`.

- Depending on the encoding we also have two types of records: **Plain** and **Affixed**.
  This also is defined at the application level, but it must respect the size limitations.
  For example a `Patient Record` is `plain`, and a `Study Record` is `affixed`.

- Depending on the signature we have two types of records: **Local** and **Central**.
  This classification is by design of the **Blind Records**, and the application could not change that.
  Any record created in a **Device** and used there is `local`, and any record downloaded from the **Central Platform** is `central`.

The three criteria gives the following types of records:

| Criteria   | Types               |
|------------|---------------------|
| Mutability | Mutable / Inmutable |
| Encoding   | Plain / Affixed     |
| Signature  | Local / Central     |

## Mutable Records

**Mutable Records** are records that can be updated along their lifetime, and the filename is composed of the following parts:

`<TS>-<ID>.<ext>`

- `TS`: lowercase no-dash `UUIDv7` that represents the timestamp of the creation/update of the record.

- `ID`: lowercase no-dash `UUIDv4` used as unique identifier of the record.

- `ext`: lowercase 3 characters extension that represents the type and version of record, in this case `rm1` (Record Mutable Version 1).

On record creation the `TS` is the creation timestamp, and on record update the `TS` is the update timestamp.

The dash (`-`) allows for a human to quickly select any part with a double click.

For example: `018762219560790cae18ff9aaf665f7e-4ed4a83f8a0647eeb013c733c1ad2f61.rm1`

## Inmutable Records

**Inmutable Records** are records that can not be updated along their lifetime, they represent a unique event, and the filename is composed of the following parts:

`<TS>.<ext>`

- `TS`: lowercase no-dash `UUIDv7` that represents the timestamp of the creation of the record.

- `ext`: lowercase 3 characters extension that represents the type and version of record, in this case `ri1` (Record Inmutable Version 1).
  For example: `018762219560790d969d556626a38746.ri1`

## Plain Records

**Plain Records** are records that are created encoding all the information payload inside the record itself, and remain in that way for the lifetime of the record.

> [!IMPORTANT]
> This kind of records should be limited in size so all the encoded record has less than 1 MB.

## Affixed Records

**Affixed Records** are records that are created encoding one part of the information inside the record payload itself, and the other part of the information, called `Affix`, is stored in a separate encrypted file.

> [!NOTE]
> This kind of records should be used for large amount of information, because the `Affix` is not limited in size.
> Only the first encoded part should remain less than 1 MB.

## Local Records

**Local Records** are records that are created locally in the **Device** and used in the same **Device** that signed them.

## Central Records

**Central Records** are records that were uploaded to the **Central Platform**, were re-signed centrally by the **Platform** and distributed to other **Devices**.
If the record is downloaded by the **Device** that originally created the record, it's going to be a **Central Record** re-signed by the **Platform**, even it's located in the same **Device** that created it.
