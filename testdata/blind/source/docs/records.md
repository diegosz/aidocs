# Blind Records

- [Blind Records](#blind-records)
  - [Record Filenames](#record-filenames)
  - [Record Folders](#record-folders)

**Blind Records** are a way to store and share information records in a secure and private way, with the added capability of proactively synchronizing them between a **Central Platform** and other devices, even when we are not able the access the encrypted information of the records, we do the synchronization `blindly`.

The **Blind Records** are a mix of digital envelope and digital signature, since they are encrypted and signed by a **User**, and further signed by the **Device**.
The only way the access the record information is by having access to the **User** key, that is only possible when the **User** is logged in.

The **Blind Records** provides `Message Integrity`, `Authentication`, `NonRepudiation` and  `Confidentiality` (except for the information included in the header).

Having the capability to synchronize **Blind Records** when there is no logged **User** is a critical feature because the **Devices** could be offline for long periods of time, and only have connectivity for intermittent short periods that could happen when there is no **User** around.

Another important aspect of the **Blind Records** is that they use **Attribute Based Encryption** so the encrypted information can be accessed by a group of **Users** that have **Credentials** matching the required attributes.

The **Blind Records** are stored in a **Blind Record Repository** that is a set of folders and files organized in a way that allows to quickly find the records needed and to synchronize them with a **Central Platform** minimizing the amount of data required to transmit.

The **Central Platform** is a group of services for the synchronization of the **Devices**, the long term storage of the **Blind Records**, the provisioning of the **Devices** and the management of the **Users**.
It also allows for the sharing of the **Blind Records** between different groups of **Users**, even from different **Organizations**.

Together the **Blind Records**, the **Devices** and the **Central Platform** with it's services conform the **Blind Records System**.

## Record Filenames

The filename of the records provides identification and a sorteable timestamp of (creation), both required for synchronization and storing.

The filename is composed of several parts depending on the type of record, but typically they have the following format:

`<TS>-<ID>.<ext>` or `<TS>.<ext>`

It's composed of a mandatory timestamp of creation (`TS`), that could also be used as identifier, and optional unique `ID`, and an extension.

For the `TS` we use an `UUID Version 7`, that is a 128-bit number generated in a way that is intended to make it practically unique in space and time that also is binary and lexicographical time sortable.

> [!NOTE]
> Starting the filename with a `UUIDv7` used as timestamp allows to do quick searches by a time prefix, for example the 5 first characters represents a little more than a 3 days block, 4 characters represents around 50 days.

For the optional unique `ID`, we use another `UUID` that typically is a Non-time-ordered fully random `UUID Version 4`.
This allows to have a unique identifier for the record that does not change with the time.
Using `UUIDv4` are opaque and do not reveal any information about the time of creation.

For the filenames we use lower-case characters with the hexadecimal representation of the `UUIDs` without dashes, if more than one `UUID` is used, then they are separated by a dash (`-`).

For the extensions we also use lower-case, and include a version number in the last character starting with `1`.
The last character versioning is for eventually supporting multiple versions in the same folder structure, and also to facilitate any version migration if required.

> The extensions allows to do a filtered search.

All the components of the filename, are also included inside of the content of the file protected by signatures.
So any change in the filename will be detected and reported.

> [!IMPORTANT]
> One requirement for the filenames is to be supported by the major cloud blob services.

## Record Folders

We are going to use a folder structure that splits the records in 3 day blocks.

For example:

```sh
recs
├── 0186e
│   ├── 0186e0edfb4b725eb56a145e767174ee-0186e08ac21f70ddaf2bca475a2eacc6.rm1
│   └── 0186e08ac21f70ddaf2bca475a2eadd6.ri1
├── 0187f
│   └── 0187f1ed62be7638b309d7535c872d59.ri1
└── 01880
    └── 018800ed9e6277efaf4893edbcf57aa9-0186e0ed9e6477ef962ca270aacc132a.rm1
```

The 3 day blocks are represented by the first 5 characters of the `TSv7` of the record, it allows to quickly find records by time range, and to process the records folder by folder limiting the amount of resources required per time block and the maximum amount of records per folder.
