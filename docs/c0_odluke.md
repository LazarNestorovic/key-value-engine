# C0 — Odluke o dizajnu

Evidencija arhitektonskih/API odluka donetih tokom razvoja, radi konzistentnosti u narednim celinama (C0, C1, ...).

## `Get` vraća `(vrednost, found bool, error)`

**Odluka:** Signatura `Get(key string) ([]byte, bool, error)`.

**Obrazloženje:**
- Razdvaja "ključ ne postoji" (normalno, očekivano stanje) od stvarne greške (npr. baza zatvorena, nevalidan ključ).
- Sprečava da prazna vrednost (`[]byte{}` ili `nil`) bude pogrešno protumačena kao "ključ ne postoji" — `found` je eksplicitan signal, ne treba nagađati na osnovu vrednosti.

## Sistemski ključevi koriste prefiks `\x00`

**Odluka:** Ključevi koji počinju bajtom `\x00` su rezervisani za interno stanje sistema i nisu dostupni za korisnički unos.

**Obrazloženje:**
- Bajt `\x00` je nedostupan za slučajan korisnički unos (ne može se uneti sa standardne tastature/CLI-a), pa nema kolizije sa "pravim" ključevima korisnika.
- Od **C12** nadalje koristiće se za interno stanje: rate limiter, HLL (HyperLogLog), CMS (Count-Min Sketch).
