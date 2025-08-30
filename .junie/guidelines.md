
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

### Mudlet - kod źródłowy klienta
W katalogu `docs/sources/Mudlet/` znajdują się kluczowe pliki z kodu źródłowego Mudleta (C++):
- **TRoom.cpp/TRoom.h** - implementacja klasy pokoju z metodami serialization/deserialization
- **TArea.cpp/TArea.h** - implementacja klasy obszaru 
- **TRoomDB.cpp/TRoomDB.h** - baza danych pokoi z metodami zapisu/odczytu
- **TMap.h** - główna klasa mapy
- **TMapLabel.cpp/TMapLabel.h** - etykiety na mapie
- **T2DMap.cpp/T2DMap.h** - renderowanie 2D mapy

### qdatastream.go
znajdziesz tu implementację QDataStream w formie pliku `qdatastream.go`. To dobra baza do zrozumienia formatu binarnego Qt.

### Node.js parser - działająca implementacja
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
```go
package mapparser

import "image/color"

type Map struct {
    Header       Header                `json:"header"`
    Rooms        map[int32]*Room       `json:"rooms"`
    Areas        map[int32]*Area       `json:"areas"`
    Environments []Environment         `json:"environments"`
    CustomLines  []CustomLine          `json:"customLines,omitempty"`
    Labels       []Label               `json:"labels,omitempty"`
}

type Header struct {
    Magic   string `json:"magic"`   // "ATADNOOM"
    Version int8   `json:"version"` // 1, 2, lub 3
}

type Room struct {
    ID          int32  `json:"id"`
    X           int32  `json:"x"`
    Y           int32  `json:"y"`
    Z           int32  `json:"z"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Environment int32  `json:"environment"`
    Exits       []Exit `json:"exits"`
}

type Exit struct {
    Direction string `json:"direction"`  // "north", "south", etc.
    TargetID  int32  `json:"targetId"`   // ID docelowego pokoju
    Lock      bool   `json:"lock"`       // zablokowane wyjście (v3+)
    Weight    int32  `json:"weight"`     // waga przejścia (v3+)
}

type Area struct {
    ID   int32  `json:"id"`
    Name string `json:"name"`
}

type Environment struct {
    Name  string `json:"name"`    // "forest", "city", etc.
    Color int32  `json:"color"`   // RGB color as int32
}

type CustomLine struct {
    X1, Y1, Z1 int32 `json:"x1,y1,z1"`
    X2, Y2, Z2 int32 `json:"x2,y2,z2"`
    Color      int32 `json:"color"`
    Width      int8  `json:"width"`
    Style      int8  `json:"style"`
}

type Label struct {
    X, Y, Z        int32  `json:"x,y,z"`
    Text           string `json:"text"`
    Color          int32  `json:"color"`
    Size           int8   `json:"size"`
    ShowBackground bool   `json:"showBackground"`
}
```

#### API parsera
```go
// Główna funkcja parsowania
func ParseMapFile(filename string) (*Map, error)

// Parsowanie z Reader (dla testów)
func ParseMap(reader io.Reader) (*Map, error)

// Walidacja integralności mapy
func ValidateMap(m *Map) []ValidationError

// Eksport do JSON (debugging)
func ExportToJSON(m *Map, filename string) error

// Statystyki mapy
func GetMapStats(m *Map) MapStats

type MapStats struct {
    TotalRooms       int
    TotalAreas       int
    TotalEnvironments int
    BoundingBox      BoundingBox
    ZLevels          []int32
}

type BoundingBox struct {
    MinX, MinY, MinZ int32
    MaxX, MaxY, MaxZ int32
}
```

### 2. Renderer obrazów (pkg/maprender)

#### Konfiguracja renderowania
```go
package maprender

import (
    "image"
    "image/color"
)

type RenderConfig struct {
    // Wymiary obrazu
    Width  int `json:"width"`
    Height int `json:"height"`
    
    // Obszar renderowania
    Radius int   `json:"radius"`        // promień w jednostkach mapy
    ZLevel int32 `json:"zLevel"`        // poziom Z do renderowania
    
    // Rozmiary elementów
    RoomSize     int `json:"roomSize"`     // rozmiar pokoju w pikselach
    LineWidth    int `json:"lineWidth"`    // szerokość linii połączeń
    FontSize     int `json:"fontSize"`     // rozmiar czcionki (opcjonalnie)
    
    // Kolory
    CenterRoomColor color.RGBA `json:"centerRoomColor"`  // kolor centralnego pokoju
    DefaultRoomColor color.RGBA `json:"defaultRoomColor"` // domyślny kolor pokoju
    LineColor       color.RGBA `json:"lineColor"`        // kolor linii połączeń
    BackgroundColor color.RGBA `json:"backgroundColor"`  // kolor tła
    
    // Opcje wizualne
    ShowRoomNames bool `json:"showRoomNames"` // pokazuj nazwy pokoi
    ShowRoomIDs   bool `json:"showRoomIds"`   // pokazuj ID pokoi
    ShowExitLabels bool `json:"showExitLabels"` // pokazuj etykiety wyjść
    AntiAlias     bool `json:"antiAlias"`     // wygładzanie
    
    // Mapowanie kolorów środowisk na kolory
    EnvironmentColors map[string]color.RGBA `json:"environmentColors"`
}

// Domyślna konfiguracja
func DefaultConfig() RenderConfig {
    return RenderConfig{
        Width:  800,
        Height: 600,
        Radius: 10,
        ZLevel: 0,
        RoomSize: 8,
        LineWidth: 2,
        FontSize: 10,
        CenterRoomColor: color.RGBA{R: 255, G: 0, B: 0, A: 255}, // czerwony
        DefaultRoomColor: color.RGBA{R: 100, G: 100, B: 100, A: 255}, // szary
        LineColor: color.RGBA{R: 200, G: 200, B: 200, A: 255}, // jasny szary
        BackgroundColor: color.RGBA{R: 0, G: 0, B: 0, A: 255}, // czarny
        ShowRoomNames: false,
        ShowRoomIDs: false,
        ShowExitLabels: false,
        AntiAlias: true,
        EnvironmentColors: map[string]color.RGBA{
            "city":     {R: 150, G: 150, B: 150, A: 255},
            "forest":   {R: 0, G: 150, B: 0, A: 255},
            "mountain": {R: 139, G: 69, B: 19, A: 255},
            "water":    {R: 0, G: 100, B: 200, A: 255},
            "desert":   {R: 238, G: 203, B: 173, A: 255},
        },
    }
}
```

#### API renderera
```go
// Główna funkcja renderowania
func RenderMapFragment(m *mapparser.Map, centerRoomID int32, config RenderConfig) (image.Image, error)

// Znajdowanie pokoi w promieniu
func FindRoomsInRadius(m *mapparser.Map, centerRoomID int32, radius int, zLevel int32) ([]*mapparser.Room, error)

// Transformacja współrzędnych mapa -> obraz
func CalculateCoordTransform(rooms []*mapparser.Room, config RenderConfig) CoordTransform

type CoordTransform struct {
    ScaleX, ScaleY float64
    OffsetX, OffsetY float64
    CenterX, CenterY int32  // współrzędne centralnego pokoju na mapie
}

func (ct CoordTransform) MapToImage(mapX, mapY int32) (imgX, imgY int)

// Zapis do różnych formatów
func SaveAsWebP(img image.Image, filename string) error
func SaveAsPNG(img image.Image, filename string) error  

// Generowanie różnych stylów map
func RenderTopographicStyle(m *mapparser.Map, centerRoomID int32, config RenderConfig) (image.Image, error)
func RenderMinimalistStyle(m *mapparser.Map, centerRoomID int32, config RenderConfig) (image.Image, error)
func RenderDetailedStyle(m *mapparser.Map, centerRoomID int32, config RenderConfig) (image.Image, error)
```

#### Algorytm renderowania
```go
func RenderMapFragment(m *mapparser.Map, centerRoomID int32, config RenderConfig) (image.Image, error) {
    // 1. Znajdź centralny pokój
    centerRoom, exists := m.Rooms[centerRoomID]
    if !exists {
        return nil, fmt.Errorf("room %d not found", centerRoomID)
    }
    
    // 2. Zbierz pokoje w promieniu na danym poziomie Z
    rooms, err := FindRoomsInRadius(m, centerRoomID, config.Radius, config.ZLevel)
    if err != nil {
        return nil, err
    }
    
    // 3. Oblicz transformację współrzędnych
    transform := CalculateCoordTransform(rooms, config)
    
    // 4. Stwórz canvas i kontekst rysowania
    img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
    gc := setupDrawingContext(img, config)
    
    // 5. Narysuj tło
    drawBackground(gc, config)
    
    // 6. Narysuj połączenia (linie) między pokojami
    drawConnections(gc, rooms, transform, config, m)
    
    // 7. Narysuj pokoje
    drawRooms(gc, rooms, centerRoom, transform, config, m)
    
    // 8. Dodaj etykiety i tekst (opcjonalnie)
    if config.ShowRoomNames || config.ShowRoomIDs {
        drawLabels(gc, rooms, centerRoom, transform, config)
    }
    
    return img, nil
}
```

### 3. Narzędzia pomocnicze (pkg/maputils)

#### Wyszukiwanie i filtrowanie
```go
package maputils

// Wyszukiwanie pokoi
func FindRoomByName(m *mapparser.Map, name string) []*mapparser.Room
func FindRoomsByArea(m *mapparser.Map, areaID int32) []*mapparser.Room
func FindRoomsByEnvironment(m *mapparser.Map, envName string) []*mapparser.Room

// Analiza połączeń
func GetConnectedRooms(m *mapparser.Map, roomID int32) []*mapparser.Room
func FindShortestPath(m *mapparser.Map, fromID, toID int32) ([]int32, error)
func AnalyzeConnectivity(m *mapparser.Map) ConnectivityReport

type ConnectivityReport struct {
    DisconnectedRooms []int32
    DeadEnds         []int32
    Hubs             []int32 // pokoje z więcej niż 4 wyjściami
}

// Statystyki obszarów
func GetAreaStatistics(m *mapparser.Map) map[int32]AreaStats

type AreaStats struct {
    RoomCount    int
    BoundingBox  BoundingBox
    Environments map[string]int
}
```

### 4. CLI Tool (cmd/mapsnap)

#### Argumenty wiersza poleceń
```go
type CLIFlags struct {
    // Podstawowe
    MapFile  string // -map
    RoomID   int32  // -room
    Output   string // -output
    
    // Wymiary i renderowanie
    Width    int // -width
    Height   int // -height
    Radius   int // -radius
    RoomSize int // -roomsize
    ZLevel   int32 // -zlevel
    
    // Tryby pracy
    Validate bool   // -validate
    Debug    bool   // -debug
    DumpJSON string // -dump-json
    
    // Style i kolory
    Style            string // -style (default, topographic, minimal, detailed)
    CenterRoomColor  string // -center-color
    BackgroundColor  string // -bg-color
    ShowRoomNames    bool   // -show-names
    ShowRoomIDs      bool   // -show-ids
    
    // Format wyjściowy
    Format  string // -format (webp, png, jpeg)
    Quality int    // -quality (dla JPEG)
    
    // Batch processing
    BatchFile string // -batch (plik z listą pokoi do wygenerowania)
    
    // Konfiguracja
    ConfigFile string // -config (plik YAML/JSON z konfiguracją)
}
```

#### Przykłady użycia CLI
```bash
# Podstawowe użycie
./mapsnap -map arkadia.map -room 1234

# Niestandardowe wymiary i styl
./mapsnap -map arkadia.map -room 1234 \
  -width 1200 -height 800 \
  -style topographic \
  -show-names

# Generowanie dla konkretnego poziomu Z
./mapsnap -map arkadia.map -room 1234 -zlevel -1 -output podziemia.webp

# Walidacja mapy
./mapsnap -map arkadia.map -validate

# Eksport struktury do JSON
./mapsnap -map arkadia.map -dump-json struktura_mapy.json

# Batch processing
./mapsnap -map arkadia.map -batch lokacje.txt

# Z plikiem konfiguracyjnym
./mapsnap -map arkadia.map -room 1234 -config moja_konfiguracja.yaml
```

#### Format pliku batch
```
# Plik lokacje.txt
1234:fragment_1234.webp
5678:fragment_5678.webp
9012:podziemia_9012.webp:zlevel=-1
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
```
tests/
├── unit/
│   ├── parser_test.go       # Testy parsera
│   ├── renderer_test.go     # Testy renderera
│   └── utils_test.go        # Testy narzędzi
├── integration/
│   ├── full_pipeline_test.go # Testy całego pipeline
│   └── cli_test.go          # Testy CLI
├── fixtures/
│   ├── sample_maps/         # Przykładowe mapy do testów
│   ├── expected_outputs/    # Oczekiwane wyniki
│   └── corrupted_maps/      # Uszkodzone mapy do testów błędów
└── benchmark/
    ├── parser_bench_test.go
    └── render_bench_test.go
```

### Testy kluczowych funkcji
```go
func TestParseCompleteMap(t *testing.T) {
    m, err := mapparser.ParseMapFile("fixtures/arkadia_sample.map")
    require.NoError(t, err)
    
    assert.Equal(t, "ATADNOOM", m.Header.Magic)
    assert.True(t, len(m.Rooms) > 0)
    assert.True(t, len(m.Areas) > 0)
    
    // Sprawdź integralność połączeń
    for _, room := range m.Rooms {
        for _, exit := range room.Exits {
            if exit.TargetID > 0 {
                _, exists := m.Rooms[exit.TargetID]
                assert.True(t, exists, "Room %d->%d: target not found", room.ID, exit.TargetID)
            }
        }
    }
}

func TestRenderFragment(t *testing.T) {
    m := loadTestMap(t)
    config := maprender.DefaultConfig()
    
    img, err := maprender.RenderMapFragment(m, 1234, config)
    require.NoError(t, err)
    
    bounds := img.Bounds()
    assert.Equal(t, config.Width, bounds.Dx())
    assert.Equal(t, config.Height, bounds.Dy())
    
    // Sprawdź czy centralny pokój jest wyróżniony
    centerX, centerY := bounds.Dx()/2, bounds.Dy()/2
    centerColor := img.At(centerX, centerY)
    // assert że kolor to czerwony (centralny pokój)
}

func BenchmarkParseMap(b *testing.B) {
    for i := 0; i < b.N; i++ {
        mapparser.ParseMapFile("fixtures/large_map.map")
    }
}
```

## 📦 Dependency Management

### go.mod
```go
module github.com/szydell/arkadia-mapsnap

go 1.21

require (
    github.com/HugoSmits86/nativewebp v0.0.0-20220101000000-abcdef123456
    github.com/golang/freetype v0.0.0-20170609013337-24b699ab12dc
    github.com/spf13/cobra v1.7.0
    github.com/spf13/viper v1.16.0
    gopkg.in/yaml.v3 v3.0.1
)

require (
    // Indirect dependencies...
)
```

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
docs/
├── README.md              # Główna dokumentacja
├── INSTALLATION.md        # Instrukcje instalacji
├── QUICK_START.md         # Szybki start
├── API_REFERENCE.md       # Dokumentacja API biblioteki
├── CLI_REFERENCE.md       # Dokumentacja CLI
├── FORMAT_SPECIFICATION.md # Specyfikacja formatu map Mudleta
├── CONFIGURATION.md       # Konfiguracja i personalizacja
├── EXAMPLES.md           # Przykłady użycia
├── TROUBLESHOOTING.md    # Rozwiązawanie problemów
├── CONTRIBUTING.md       # Wytyczne dla kontrybutorów
└── CHANGELOG.md          # Historia zmian
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
```go
// Dodaj hex dump dla porównania z referencyjną implementacją
func debugHexDump(data []byte, offset int) {
    fmt.Printf("Offset %d:\n", offset)
    for i := 0; i < len(data) && i < 64; i += 16 {
        fmt.Printf("%08x: ", offset+i)
        for j := i; j < i+16 && j < len(data); j++ {
            fmt.Printf("%02x ", data[j])
        }
        fmt.Println()
    }
}
```

### 3. Obsługa błędów
```go
// Używaj wrapped errors dla lepszego debugowania
if err := parseRoom(reader, version); err != nil {
    return fmt.Errorf("parsing room at offset %d: %w", offset, err)
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
