# Odluke o dizajnu

Evidencija arhitektonskih/API odluka donetih tokom razvoja, radi konzistentnosti u narednim celinama (C0, C1, ...).

## C0 — `Get` vraća `(vrednost, found bool, error)`

**Odluka:** Signatura `Get(key string) ([]byte, bool, error)`.

**Obrazloženje:**
- Razdvaja "ključ ne postoji" (normalno, očekivano stanje) od stvarne greške (npr. baza zatvorena, nevalidan ključ).
- Sprečava da prazna vrednost (`[]byte{}` ili `nil`) bude pogrešno protumačena kao "ključ ne postoji" — `found` je eksplicitan signal, ne treba nagađati na osnovu vrednosti.

## C0 — Sistemski ključevi koriste prefiks `\x00`

**Odluka:** Ključevi koji počinju bajtom `\x00` su rezervisani za interno stanje sistema i nisu dostupni za korisnički unos.

**Obrazloženje:**
- Bajt `\x00` je nedostupan za slučajan korisnički unos (ne može se uneti sa standardne tastature/CLI-a), pa nema kolizije sa "pravim" ključevima korisnika.
- Od **C12** nadalje koristiće se za interno stanje: rate limiter, HLL (HyperLogLog), CMS (Count-Min Sketch).

## C1 — `LoadConfig` ručno testiran za sva tri scenarija

**Odluka/status:** `internal/config.LoadConfig(path string) (*Config, error)` je ručno testiran pokretanjem `kvcli` binarnog fajla za sledeća tri scenarija, potvrđeno preko `fmt.Printf("%+v\n", cfg)` ispisa u `main()`:

1. **Fajl postoji i validan je** → učitane su vrednosti iz fajla (ne default).
2. **Fajl ne postoji** → `os.IsNotExist(err) == true`, `LoadConfig` tiho vraća `DefaultConfig()` bez greške (nije tretirano kao failure).
3. **Fajl postoji ali je nevalidan JSON** → `json.Unmarshal` vraća grešku, `LoadConfig` vraća `nil, err`; `main()` ispisuje grešku na `os.Stderr` i prekida se pre `Run` (CLI se ne pokreće sa pokvarenom konfiguracijom).

Detaljan opis testova i otvoreno pitanje (exit kod pri grešci) u [c1.md](c1.md).

