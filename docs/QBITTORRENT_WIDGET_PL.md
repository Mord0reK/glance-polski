# qBittorrent Widget

Widget do wyświetlania statusu torrentów z qBittorrent.

![qBittorrent Widget Preview](images/qbittorrent-widget-preview.png)

## Konfiguracja

```yaml
- type: qbittorrent
  url: http://localhost:8080
  username: admin
  password: adminadmin
```

## Właściwości

| Nazwa | Typ | Wymagane | Domyślnie | Opis |
| ----- | --- | -------- | --------- | ---- |
| url | string | tak | - | Adres URL interfejsu webowego qBittorrent |
| username | string | tak | - | Nazwa użytkownika do logowania |
| password | string | tak | - | Hasło do logowania |
| hide-seeding | boolean | nie | false | Ukryj torrenty, które seedują |
| hide-completed | boolean | nie | false | Ukryj ukończone torrenty |
| show-only-active | boolean | nie | false | Pokaż tylko aktywne torrenty (z prędkością > 0) |
| limit | integer | nie | 10 | Maksymalna liczba wyświetlanych torrentów |
| sort-by | string | nie | progress | Sposób sortowania: `name`, `progress`, `speed`, `eta` |

## Przykłady

### Podstawowa konfiguracja

```yaml
- type: qbittorrent
  url: http://192.168.1.100:8080
  username: admin
  password: twoje_haslo
```

### Tylko pobierane torrenty

```yaml
- type: qbittorrent
  url: http://192.168.1.100:8080
  username: admin
  password: twoje_haslo
  hide-seeding: true
  hide-completed: true
  limit: 5
  sort-by: eta
```

### Aktywne torrenty posortowane po prędkości

```yaml
- type: qbittorrent
  url: http://192.168.1.100:8080
  username: admin
  password: twoje_haslo
  show-only-active: true
  sort-by: speed
  limit: 8
```

## Wyświetlane informacje

Widget wyświetla:

### Podsumowanie
- Liczba pobieranych torrentów
- Liczba seedujących torrentów
- Liczba wstrzymanych torrentów
- Całkowita prędkość pobierania
- Całkowita prędkość wysyłania

### Lista torrentów
- Nazwa torrenta
- Pasek postępu
- Procent ukończenia
- Kategoria (jeśli ustawiona)
- Status (Downloading, Seeding, Paused, itp.)
- Prędkość pobierania (dla aktywnych)
- Szacowany czas do zakończenia (ETA)
- Prędkość wysyłania

## Stany torrentów

| Ikona | Stan |
| ----- | ---- |
| 🔵 (niebieski) | Pobieranie |
| 🟢 (zielony) | Seedowanie |
| ⏸️ (szary) | Wstrzymany |
| ⚠️ (szary) | Zablokowany (stalled) |
| 🕐 (szary) | W kolejce |
| 🔄 (szary) | Sprawdzanie |
| 🔴 (czerwony) | Błąd |

## Wymagania

- qBittorrent z włączonym interfejsem webowym
- Włączone Web UI w ustawieniach qBittorrent (Tools → Options → Web UI)
- Prawidłowe dane uwierzytelniające

## Uwagi dotyczące bezpieczeństwa

- Widget automatycznie zarządza sesją i ponownie loguje się gdy sesja wygaśnie
- Po 3 nieudanych próbach logowania, widget blokuje dalsze próby na 30 minut
- Hasło jest przesyłane bezpośrednio do API qBittorrent, więc zalecane jest używanie HTTPS jeśli qBittorrent jest dostępny z zewnątrz
