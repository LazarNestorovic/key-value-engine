# C2 — Record format

- Redosled bajtova: BigEndian, svuda isti za upis i čitanje.
- Timestamp (16B): 8B Unix sekunde + 8B nanosekunde unutar sekunde. Zbog BigEndian-a, poređenje bajt po bajt daje isti redosled kao poređenje vremena. Gornja 4 bajta nanosekundi su uvek 0. Poznato ograničenje: sistemski sat može da se pomeri unazad.
- Record u memoriji ima samo Timestamp, Tombstone, Key, Value. CRC i dužine postoje samo na disku; dužine se izvode iz ključa i vrednosti.
- Ključ je string (nepromenljiv, može biti ključ Go mape), vrednost []byte.
- CRC32 (IEEE) se računa nad svim bajtovima zapisa osim samog CRC polja, poslednji pri serijalizaciji.
- Tombstone zapis nema vrednost: ValueSize = 0, vrednost se ne upisuje čak ni ako je postavljena u strukturi.
- Deserialize čita jedan zapis i vraća broj pročitanih bajtova; petlju kroz više zapisa (WAL segment) piše pozivalac.
- Greške: ErrShortInput (ulaz prekratak za zapis), ErrDamagedData (CRC se ne poklapa).
- Poznato ograničenje: oštećeno polje dužine u poslednjem zapisu segmenta daje ErrShortInput i ne razlikuje se od zapisa prekinutog padom sistema.
