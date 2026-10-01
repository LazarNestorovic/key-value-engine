# C1 — Konfiguracija (config.json / LoadConfig)

## Šta je urađeno

- `internal/config/config.go`: strukture konfiguracije (`MemtableConfig`, `WALConfig`, `CacheConfig`, `LSMConfig`, `RateLimitConfig`, `BloomConfig`, `StorageConfig`, `Config`) sa `json:"..."` tagovima (snake_case).
- `DefaultConfig() *Config` — vraća podrazumevane vrednosti, inicijalizovane sa imenovanim poljima (named fields) radi čitljivosti.
- `LoadConfig(path string) (*Config, error)` — učitava `Config` iz JSON fajla na datoj putanji, sa fallbackom na `DefaultConfig()` ako fajl ne postoji.
- `config.json` u root-u projekta — primer konfiguracije koja se razlikuje od podrazumevane (radi provere da učitavanje stvarno menja vrednosti).
- `cmd/kvcli/main.go` poziva `config.LoadConfig("config.json")` pre pokretanja `Run`.

## Ručno testiranje — tri scenarija

Testirano pokretanjem `kvcli` binarnog fajla iz privremenog radnog direktorijuma (`/tmp/c1test`), uz `fmt.Printf("%+v\n", cfg)` u `main()` da se vizuelno potvrdi koje vrednosti su stvarno učitane.

### 1. `config.json` postoji i validan je

Kopiran je root-ni `config.json` (sa custom vrednostima: `threshold: 200`, `impl_type: hashmap`, `compaction_type: sizetiered`, itd.) u radni direktorijum.

**Rezultat:** `LoadConfig` vraća `*Config` sa vrednostima **iz fajla** (potvrđeno preko ispisa — `Threshold:200 InstanceNum:2 ImplType:hashmap ... CompactionType:sizetiered ...`), bez greške. Program nastavlja normalno.

### 2. `config.json` ne postoji

Fajl je obrisan iz radnog direktorijuma pre pokretanja.

**Rezultat:** `os.ReadFile` vraća grešku za koju `os.IsNotExist(err) == true`, pa `LoadConfig` **ne vraća grešku** već `DefaultConfig()` (potvrđeno preko ispisa — `Threshold:100 InstanceNum:1 ImplType:skiplist ... CompactionType:leveled ...`, tačno podrazumevane vrednosti). Program nastavlja normalno, kao da fajl nikad nije ni naveden.

### 3. `config.json` postoji, ali je pokvaren (nevalidan JSON)

Napravljen je fajl sa namerno nepotpunim/nevalidnim JSON sadržajem (`{ "memtable": { "threshold": 200, `).

**Rezultat:** `json.Unmarshal` vraća grešku (`unexpected end of JSON input`), `LoadConfig` vraća `nil, err`. U `main()` se greška ispisuje na `os.Stderr` i program se prekida sa `return` **pre** poziva `Run` — CLI se uopšte ne pokreće kad je konfiguracija pokvarena.

**Napomena / mogući sledeći korak:** trenutno `main()` u ovom slučaju izlazi sa exit kodom `0` (samo `return` iz `main`, bez `os.Exit(1)`), iako je došlo do greške. Ako je bitno da skriptovano/CI okruženje detektuje neuspeh preko exit koda, trebalo bi dodati `os.Exit(1)` posle ispisa greške. Za sada nije menjano — ostavljeno kao otvoreno pitanje za dalje celine.

## Zaključak

`LoadConfig` se ponaša po očekivanju:
- fajl postoji i validan je → koristi se sadržaj fajla,
- fajl ne postoji → tiho se koristi `DefaultConfig()` (nije greška),
- fajl postoji ali je nevalidan → prava greška, program se ne pokreće.
