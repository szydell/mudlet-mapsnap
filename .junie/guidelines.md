
# Kompletna instrukcja projektu arkadia-mapsnap

## 📋 Opis projektu

### Cel główny
**arkadia-mapsnap** to biblioteka Go oraz narzędzie CLI do generowania wizualnych fragmentów mapy świata gry MUD "Arkadia". Projekt ma na celu ułatwienie nawigacji i planowania tras przez automatyczne tworzenie obrazków przedstawiających fragment mapy wycentrowany na wybranej lokacji.

### Problem rozwiązywany
- Gracze Arkadii używają Mudleta jako klienta do gry
- Mudlet przechowuje mapy w binarnym formacie, trudnym do przetwarzania
- Brak narzędzi do szybkiego generowania wizualnych fragmentów mapy
- Potrzeba łatwego udostępniania fragmentów map innym graczom
- Planowanie tras wymaga wizualizacji otoczenia danej lokacji

### Funkcjonalności docelowe

#### Biblioteka Go (`pkg/mapsnap`)
- **Parsowanie**: Odczyt plików map Mudleta (.dat) w formacie binarnym
- **Wyszukiwanie**: Lokalizacja pokoju po ID
- **Renderowanie**: Generowanie obrazów fragmentów mapy w formacie WEBP
- **Konfiguracja**: Elastyczne ustawienia wizualizacji (kolory, rozmiary, promienie)

#### CLI Tool (`cmd/mapsnap`)
- **Proste użycie**: `mapsnap -map arkadia.map -room 1234 -output fragment.webp`
- **Konfigurowalność**: Wymiary obrazu, promień wyświetlania, style wizualizacji
- **Tryb debug**: Walidacja map, dump do JSON, szczegółowe logi
- **Batch processing**: Możliwość generowania wielu fragmentów naraz

### Przykład użycia
```bash
# Podstawowe użycie
./mapsnap -map arkadia.map -room 1234 -output lokacja_1234.webp

# Z niestandardowymi parametrami
./mapsnap -map arkadia.map -room 1234 \
  -width 1200 -height 800 \
  -radius 15 -roomsize 12 \
  -output duzy_fragment.webp

# Tryb debug

```

## 🏗️ Architektura rozwiązania

### Struktura projektu
```
arkadia-mapsnap/
├── cmd/
│   └── mapsnap/               # CLI application
│       ├── main.go           # Entry point
│       ├── flags.go          # Command line arguments
│       └── commands.go       # Command handlers
├── pkg/
│   ├── mapparser/            # Map file parsing
│   │   ├── parser.go         # Main parser
│   │   ├── types.go          # Map data structures  
│   │   ├── reader.go         # Binary reading helpers
│   │   └── parser_test.go    # Parser tests
│   ├── maprender/            # Image generation
│   │   ├── renderer.go       # Main rendering engine
│   │   ├── config.go         # Render configuration
│   │   ├── coords.go         # Coordinate transformation
│   │   ├── styles.go         # Visual styles and colors
│   │   └── renderer_test.go  # Renderer tests
│   └── maputils/             # Common utilities
│       ├── search.go         # Room searching and filtering
│       ├── validation.go     # Map validation
│       └── export.go         # Export to json
├── docs/
├── .github/
│   └── workflows/
│       └── ci.yml           # CI/CD pipeline
├── go.mod
├── go.sum
├── Makefile                 # Build automation
└── README.md               # Project documentation
```

### Przepływ danych
```
Plik mapy Mudleta (.map)
           ↓
    [Parser binarny]
           ↓
    Struktura mapy w Go
           ↓
    [Wyszukiwanie pokoju]
           ↓
    [Znajdowanie otoczenia]
           ↓
    [Transformacja współrzędnych]
           ↓
    [Renderowanie obrazu]
           ↓
    Obraz WEBP/PNG
```

## 📚 Źródła referencyjne i dokumentacja formatu

### Mudlet — kod źródłowy klienta
W katalogu `docs/sources/Mudlet/` znajdują się kluczowe pliki z kodu źródłowego Mudleta (C++):
- **TRoom.cpp/TRoom.h** - implementacja klasy pokoju z metodami serialization/deserialization
- **TArea.cpp/TArea.h** - implementacja klasy obszaru 
- **TRoomDB.cpp/TRoomDB.h** - baza danych pokoi z metodami zapisu/odczytu
- **TMap.h** - główna klasa mapy
- **TMapLabel.cpp/TMapLabel.h** - etykiety na mapie
- **T2DMap.cpp/T2DMap.h** - renderowanie 2D mapy

### qdatastream.go
Znajdziesz tu implementację QDataStream w formie pliku `qdatastream.go`. To dobra baza do zrozumienia formatu binarnego Qt.

### Node.js parser — działająca implementacja
W katalogu `docs/sources/node-mudlet-map-binary-reader/` znajduje się działający parser Node.js:
- **README.md** - dokumentacja użycia, obsługuje v20 formatu Mudleta
- **index.js** - punkt wejścia z API do read/write/export
- **map-operations.js** - główna logika czytania/zapisywania map
- **models/mudlet-models.js** - definicje struktur MudletMap, MudletRoom, MudletArea, MudletLabel
- **models/qstream-containers.js** - implementacje QMap, QList, QMultiMap dla QDataStream
- **models/qstream-types.js** - podstawowe typy QDataStream (QString, QColor, QPoint, QFont)

**Kluczowe insights z Node.js parsera:**
- Używa biblioteki `qtdatastream` do obsługi binarnego formatu Qt
- Format to QDataStream z zarejestrowanymi QUserType dla struktur Mudleta
- Struktura MudletMap zawiera: version, envColors, areaNames, areas, rooms, labels
- Każdy MudletRoom ma 16 pól standardowych exitów plus special exits w rawSpecialExits
- Obszary (areas) to QMap<QInt, QString> z sortowaniem specjalnym (Default Area = -1 na początku)

### Format binarny - kluczowe informacje
1. **QDataStream format** - Qt's binary serialization format, big-endian. MudletMap rozpoczyna się od qint32 `version` (np. 20). Brak magic stringa w trybie Qt; alternatywnie, w naszym projekcie wspieramy także legacy placeholder z magic "ATADNOOM" + 1‑bajtowa wersja dla testów. 
2. **QString encoding** - UTF-16BE z prefiksem długości jako liczba BAJTÓW (quint32). Wartość 0xFFFFFFFF oznacza pusty (null) string. Po długości następuje dokładnie tyle bajtów (parzysta liczba), które dekodujemy jako UTF‑16BE. 
3. **QMap serialization** - najpierw liczba elementów (qint32), następnie pary klucz→wartość; dla areaNames używany jest specjalny sorter Mudleta (Default Area = -1 pierwsze). 
4. **MudletMap order (początkowa część)** - `version` → `envColors: QMap<int,int>` → `areaNames: QMap<int,QString>` → `mCustomEnvColors` → ... (patrz models w docs/sources/node-.../mudlet-models.js). 
5. **MudletRoom structure** - 16 pól z exitami + environment, weight, name, userData, customLines itp. 
6. **Special exits** - kodowane jako QMultiMap<QUInt, QString> z prefiksami "0"/"1" dla lock status

## 🗂️ Szczegółowa specyfikacja

### 1. Parser map Mudleta (pkg/mapparser)

#### Struktury danych

#### API parsera

### 2. Renderer obrazów (pkg/maprender)

#### Konfiguracja renderowania

#### API renderera

#### Algorytm renderowania

### 3. Narzędzia pomocnicze (pkg/maputils)

#### Wyszukiwanie i filtrowanie

### 4. CLI Tool (cmd/mapsnap)

#### Argumenty wiersza poleceń
```
mapFile := flag.String("map", "", "Path to the Mudlet map file (.map)")
roomID := flag.Int("room", 0, "Room ID to center the map on")
outputFile := flag.String("output", "", "Output file path")
dumpJSON := flag.String("dump-json", "", "Dump map to JSON file")
validate := flag.Bool("validate", false, "Validate map integrity")
showStats := flag.Bool("stats", false, "Show map statistics")
debug := flag.Bool("debug", false, "Enable debug output")
examine := flag.Bool("examine", false, "Examine the binary structure of the map file")
examineQt := flag.Bool("examine-qt", false, "Examine Qt/MudletMap sections and offsets")
timeout := flag.Int("timeout", 30, "Timeout in seconds for parsing operations")
```

#### Przykłady użycia CLI
```bash
# Podstawowe użycie  (wygeneruj obrazek pokazujący fragment mapy wycentrowany na pokoju 1234)
./mapsnap -map arkadia.map -room 1234

# Podstawowe użycie  (wygeneruj obrazek pokazujący fragment mapy wycentrowany na pokoju 1234), zapisz do pliku podziemia.webp
./mapsnap -map arkadia.map -room 1234 -output podziemia.webp

# Walidacja mapy
./mapsnap -map arkadia.map -validate

# Eksport struktury do JSON
./mapsnap -map arkadia.map -dump-json struktura_mapy.json

# Z plikiem konfiguracyjnym
./mapsnap -map arkadia.map -room 1234 -config moja_konfiguracja.yaml
```

#### Format pliku konfiguracyjnego (YAML)
```yaml
# moja_konfiguracja.yaml
render:
  width: 1200
  height: 800
  radius: 15
  roomSize: 10
  showRoomNames: true
  showRoomIds: false
  
colors:
  centerRoom: "#FF0000"
  background: "#000000"
  defaultRoom: "#808080"
  line: "#C0C0C0"
  
environments:
  city: "#969696"
  forest: "#00AA00"
  mountain: "#8B4513"
  water: "#0066CC"
  desert: "#EECBAD"
  
output:
  format: "webp"
  quality: 85
```

## 🧪 Testowanie

### Struktura testów

### Testy kluczowych funkcji

## 📦 Dependency Management

go 1.24+

## 🚀 Roadmap rozwoju

### Faza 1: MVP (Minimum Viable Product)
- [ ] Parser formatu binarnego Mudleta
- [ ] Podstawowy renderer obrazów WEBP
- [ ] CLI z podstawowymi flagami
- [ ] Testy jednostkowe parsera
- [ ] Dokumentacja API

### Faza 2: Ekosystem
- [ ] Docker images
- [ ] GitHub Actions dla CI/CD
- [ ] Dokumentacja

### Faza 3: Rozszerzenie funkcjonalności
- [ ] HTTP API server
- [ ] Wsparcie dla wielu formatów wyjściowych (PNG, JPEG, SVG)
- [ ] Predefiniowane style wizualne
- [ ] Batch processing
- [ ] Pliki konfiguracyjne
- [ ] Optymalizacja wydajności

## 📚 Dokumentacja

### Struktura dokumentacji
```
README.md              # Główna dokumentacja
CHANGELOG.md           # Historia zmian
docs/
├── INSTALLATION.md        # Instrukcje instalacji
├── QUICK_START.md         # Szybki start
├── API_REFERENCE.md       # Dokumentacja API biblioteki
├── CLI_REFERENCE.md       # Dokumentacja CLI
├── FORMAT_SPECIFICATION.md # Specyfikacja formatu map Mudleta
├── CONFIGURATION.md       # Konfiguracja i personalizacja
├── EXAMPLES.md           # Przykłady użycia
├── TROUBLESHOOTING.md    # Rozwiązawanie problemów
├── CONTRIBUTING.md       # Wytyczne dla kontrybutorów

```

## 🔧 Wskazówki implementacyjne

### 1. Kolejność implementacji
1. **Parser map** - zacznij od solidnych podstaw
2. **Podstawowy renderer** - prosty rendering prostokątów i linii
3. **CLI skeleton** - podstawowa struktura argumentów
4. **Testy** - równolegle z implementacją
5. **Optymalizacja** - po osiągnięciu funkcjonalności
6. **Dokumentacja** - na końcu każdej fazy

### 2. Debugowanie parsera

### 3. Obsługa błędów
```
// Używaj wrapped errors dla lepszego debugowania
if err := parseRoom(reader, version); err != nil {
    return fmt.Errorf("parsing room at offset %d: %w", offset, err)
}
```
Obsługa błędów zamykania plików: defer file.Close()
```
// Zalecenia dla Go 1.20+ (projekt używa Go 1.24+):
// - Stosuj nazwane wartości zwracane (err).
// - Łącz błąd parsowania i błąd zamknięcia przez errors.Join.

func exampleCloseHandling(path string) (err error) {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil {
            if err != nil {
                err = errors.Join(err, fmt.Errorf("close: %w", cerr))
            } else {
                err = fmt.Errorf("close: %w", cerr)
            }
        }
    }()

    // ... praca z plikiem ...
    return nil
}
```


### 4. Wydajność
- Używaj `bufio.Reader` dla dużych plików
- Implementuj lazy loading dla map z tysiącami pokoi
- Cache'uj często używane obliczenia (transformacje współrzędnych)
- Rozważ goroutines dla renderowania równoległego

### 5. Testowanie z prawdziwymi danymi
- Testuj na prawdziwych plikach map Arkadii
- Porównuj wyniki z Node.js parserem (tam gdzie działa poprawnie)
- Używaj golden files dla testów wizualnych
- Implementuj testy regresji dla różnych wersji map

Ten kompletny plan powinien pozwolić na stworzenie funkcjonalnego i użytecznego narzędzia dla społeczności graczy Arkadii.



## 📝 Aktualizacje: QDataStream i MudletLabel (2025-08-30)

Poniższa sekcja dokumentuje praktyczne wnioski i pułapki wykryte podczas pracy nad examine-qt oraz parserem dużych map Mudleta (v20). Zostały zweryfikowane na plikach testowych: tests/fixtures/2_rooms_map/2lok.dat oraz tests/fixtures/large_maps/2025-05-27#15-06-15map.dat.

- QString (Qt QDataStream):
  - Długość zapisywana jest jako quint32 reprezentujący LICZBĘ BAJTÓW UTF-16BE, nie liczbę znaków.
  - Wartość 0xFFFFFFFF oznacza null/empty string i powinna zwrócić pusty string (bez czytania kolejnych danych).
  - Długość musi być parzysta (pełne 16-bitowe QChar). Nieprawidłowa długość sugeruje rozjechany strumień wcześniej w pliku.

- MudletLabel (kolejność pól):
  - Zgodnie z referencją Node.js (v20) i źródłami Mudleta etykieta serializuje się w kolejności:
    1) id: int
    2) pos: QVector3D → 3 x double
    3) dummy1: double
    4) dummy2: double
    5) size: QPair<double,double> → 2 x double
    6) text: QString
    7) fgColor: QColor
    8) bgColor: QColor
    9) pixMap: QPixmap (często PNG inline)
    10) noScaling: bool
    11) showOnTop: bool
  - Kluczowe: łącznie 7 odczytów double przed QString (3 + 2 + 2). Dodatkowy odczyt double rozjedzie strumień i spowoduje błędy QString.

- QPixmap/Png w etykietach (krytyczna pułapka):
  - Po polu QPixmap występują często dane PNG zaczynające się od magic 0x89504E47. Przy pomijaniu PNG należy skanować do znacznika 'IEND' (0x49 0x45 0x4E 0x44) i KONIECZNIE skonsumować również 4‑bajtowy CRC po IEND.
  - W praktyce: po znalezieniu 'IEND' trzeba wykonać Skip(8) – 4 bajty IEND + 4 bajty CRC, aby ustawić pozycję dokładnie za obrazem. Samo zjedzenie 'IEND' pozostawia CRC, które rozbija kolejny odczyt (np. QString).

- Examine-qt (diagnozowanie dużych plików):
  - Wypisywanie offsetów @Position() przed/po kluczowych sekcjach bardzo pomaga zlokalizować rozjazdy.
  - Jeżeli analiza etykiet jest problematyczna lub kosztowna, można:
    - Użyć env MAPSNAP_SKIP_LABELS=1 w parserze, który skorzysta z heurystyki przeskoku do sekcji rooms.
    - Ograniczyć diagnostykę (np. wypisywać peek 8 bajtów tylko dla kilku pierwszych etykiet).
  - Nie zwiększaj defaultowego timeoutu > 30s. Mudlet wczytuje duże mapy ~1s; dłuższe czasy wskazują na błąd w parserze (np. nieskończone skanowanie PNG).

- Wydajność i bezpieczeństwo strumienia:
  - Zawsze owijaj io.Reader w bufio.Reader na wejściu parsera.
  - Unikaj wielokrotnego “pełzania” po tych samych danych; przy skanowaniu PNG przesuwaj się o 1 bajt i sprawdzaj okno 4 bajtów.
  - Waliduj sensowne zakresy (np. liczniki QMap/QList < rozsądny próg) zanim wejdziesz w pętle.

- Flagi i zmienne ułatwiające debug:
  - mapsnap -examine-qt -map <plik> – wypisuje strukturę QDataStream z offsetami.
  - MAPSNAP_DEBUG=1 – parser wypisze wybrane etapy z pozycjami w strumieniu.
  - MAPSNAP_SKIP_LABELS=1 – w parserze pominie ciężką sekcję etykiet, używając heurystyki odszukania początku rooms.

Te zasady zostały już odzwierciedlone w kodzie:
- cmd/mapsnap/examine_qt.go: poprawna liczba double w MudletLabel i skip PNG do IEND+CRC.
- pkg/mapparser/parser.go: skipPNG również konsumuje CRC (IEND+CRC).
