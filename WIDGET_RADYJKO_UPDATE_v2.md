# Widget Radyjko - Aktualizacja v2 - Ulepszenia UX

## Zmiany wprowadzone

### ✅ 1. Dodane ikony stacji radiowych
- Skopiowano wszystkie ikony PNG ze katalogu `Radyjko/ikony/` do `/internal/glance/static/images/radyjko/`
- Dodano metodę `GetIconURL()` w strukturze `station` do zwracania pełnego URL ikony
- Ikony są teraz wyświetlane w głównym headerze playera
- Fallback na gradient, gdy ikona się nie załaduje

### ✅ 2. Przesunięcie listy stacji do popover menu
- Usunięto dużą listę stacji ze głównego widgetu
- Lista stacji jest teraz dostępna w oddzielnym popover menu (kliknij ikonę "☰")
- Zmniejszenie wysokości widgetu z ~500px do ~280px
- Cleaner, bardziej skupiony interfejs
- Popover zawiera miniaturki ikon stacji

### ✅ 3. Dodane przyciski nawigacji
- **Przycisk "Poprzednia"** (|<) - przechodzi do poprzedniej stacji
- **Przycisk "Następna"** (>|) - przechodzi do następnej stacji
- Przyciski są dostępne zawsze, umożliwiają szybką zmianę stacji bez otwierania menu
- Przyciski są wyczarowane (mniej kontrastowe) niż play/pause

### ✅ 4. Zmienione ikony kontrolek
- Zmieniona ikona **pauzy** - teraz bardziej minimalistyczna (dwa prostokąty zamiast bardziej ozdobnej)
- Zmieniona ikona **głośności** - teraz ikona uśmiechniętej buźki (emoji 😊) zastąpiona na ikę głośnika
- Ikony lepiej pasują do minimalistycznego stylu Glance'a
- Ikonę "menu stacji" na bardziej rozpoznawalną (3 poziome linie ☰)

## Struktura plików

```
internal/glance/
├── widget-radyjko.go (UPDATED - dodana metoda GetIconURL)
├── templates/
│   └── radyjko.html (UPDATED - nowy layout z popoverem)
├── static/
│   ├── css/
│   │   └── widget-radyjko.css (UPDATED - nowe style dla przycisków i popovers)
│   ├── js/
│   │   └── radyjko.js (UPDATED - obsługa nowych przycisków i popovers)
│   └── images/
│       └── radyjko/
│           ├── eska-siedlce.png
│           ├── krzakfm.png
│           ├── meloradio.png
│           ├── murzynfm.png
│           ├── openfm-500partyhits.png
│           ├── openfm-dance.png
│           ├── openfm-vixa.png
│           ├── radio-freee.png
│           ├── radio-kierowcow.png
│           ├── radio-zet.png
│           ├── radiozet-dance.png
│           ├── rmf-fm.png
│           ├── rmf-hard-and-heavy.png
│           ├── rmf-maxx.png
│           ├── rp-djmixes.png
│           ├── rp-kanalglowny.png
│           ├── voxfm-bestlista.png
│           ├── voxfm-djmix.png
│           └── voxfm.png
```

## Nowy layout playerera

```
┌─────────────────────────────────────┐
│  [Ikona]  Teraz odtwarzam           │  <- Header z ikoną i nazwą stacji
│           Nazwa stacji              │
├─────────────────────────────────────┤
│  [◀]  [▶]  [▶|]  [🔊] ━━━  [☰]    │  <- Kontrolki
│  Poprz Play  Następ Głośn      Menu │
└─────────────────────────────────────┘
```

Popover (kliknięcie ☰):
```
┌──────────────────────┐
│ STACJE RADIOWE       │
├──────────────────────┤
│ [🎵] Stacja 1        │
│ [🎵] Stacja 2        │
│ [🎵] Stacja 3        │
│ ...                  │
└──────────────────────┘
```

## Ulepszenia funkcjonalne

1. **Szybsza nawigacja** - przyciski poprzednia/następna zamiast otwierania menu
2. **Bardziej zwarte** - mniejsze wysokości widgetu, lepiej pasuje na dashboards
3. **Wizualnie atrakcyjne** - ikony stacji robią widżet bardziej rozpoznawalnym
4. **Przystępne dla mobilnych** - responsywny popover menu
5. **Graceful degradation** - ikony które się nie załadują mają fallback na gradient

## Testy

Aby przetestować:

```bash
go build -o glance .
./glance
```

Następnie otwórz `http://localhost:8080` i dodaj do konfiguracji:

```yaml
- type: radyjko
```

## Notatki techniczne

- Widget używa `GetIconURL()` do zwracania URL ścieżek do ikon
- Popover menu korzysta z systemowego popovers Glance'a
- Przyciski poprzednia/następna mają dedykowany styl (mniej kontrastowy background)
- Obsługa błędów załadowania obrazów przy pomocy `onerror` eventów
- CSS używa fallback gradientu dla ikonek, które się nie załadują
