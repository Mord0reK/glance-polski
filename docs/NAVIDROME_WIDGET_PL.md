# Widget Navidrome - Instrukcja użycia

Widget Navidrome umożliwia odtwarzanie muzyki z Twojego serwera Navidrome (lub innego kompatybilnego z API Subsonic) bezpośrednio na dashboardzie Glance.

## Konfiguracja

Aby skonfigurować widget Navidrome, dodaj następującą sekcję do swojego pliku `glance.yml`:

```yaml
- type: navidrome
  url: https://twoj-navidrome.pl  # Adres URL Twojego serwera Navidrome
  user: twoj_login                # Nazwa użytkownika
  token: twoje_haslo              # Hasło użytkownika (lub token)
  salt: losowy_ciag               # Opcjonalnie: własny ciąg "salt" do autoryzacji (domyślnie generowany automatycznie)
```

### Wymagania

Widget korzysta z API Subsonic, więc powinien działać z każdym serwerem kompatybilnym z tym standardem (Navidrome, Gonic, Airsonic itp.), o ile serwer obsługuje wymagane endpointy (`getPlaylists`, `getPlaylist`, `stream`, `getCoverArt`).

## Funkcje widgetu

### 1. Odtwarzacz muzyki

Widget oferuje pełnoprawny odtwarzacz z następującymi funkcjami:
- **Wybór playlisty**: Rozwijana lista wszystkich dostępnych playlist z serwera.
- **Okładka albumu**: Wyświetla okładkę aktualnie odtwarzanego utworu.
- **Informacje o utworze**: Tytuł i wykonawca.
- **Sterowanie**: Przyciski Play/Pause, Poprzedni, Następny.

### 2. Funkcje dodatkowe

- **Losowe odtwarzanie (Shuffle)**: Przycisk pozwalający na losowe mieszanie utworów w wybranej playliście.
- **Pętla (Loop)**: Przycisk włączający powtarzanie playlisty po jej zakończeniu.
- **Regulacja głośności**: Suwak pozwalający dostosować głośność odtwarzania.

### 3. Zapamiętywanie stanu

Widget automatycznie zapamiętuje swoje ustawienia w przeglądarce:
- Ostatnio wybrana playlista
- Ostatnio odtwarzany utwór
- Poziom głośności
- Stan przycisków Shuffle i Loop

Dzięki temu po odświeżeniu strony możesz szybko wrócić do słuchania muzyki w miejscu, w którym skończyłeś.

## Rozwiązywanie problemów

### Brak playlist
Jeśli lista playlist jest pusta, upewnij się, że:
1. Dane logowania (url, user, token) są poprawne.
2. Użytkownik ma uprawnienia do korzystania z API i widzi playlisty w Navidrome.
3. Serwer Navidrome jest dostępny z sieci, w której działa Glance.

### Problemy z odtwarzaniem
Jeśli utwory nie chcą się odtwarzać:
1. Sprawdź w konsoli przeglądarki (F12), czy nie ma błędów związanych z CORS lub Mixed Content (jeśli Glance jest na HTTPS, a Navidrome na HTTP).
2. Upewnij się, że format plików muzycznych jest wspierany przez Twoją przeglądarkę.
