# Integracja Affine z widgetem Vikunja

Widget Vikunja w Glance oferuje integrację z Affine - aplikacją do tworzenia notatek i dokumentacji. Ta funkcja pozwala powiązać zadania Vikunja z notatkami w Affine, tworząc spójny system zarządzania zadaniami i dokumentacją.

## Czym jest Affine?

Affine to open-source'owa aplikacja do tworzenia notatek, dokumentów i baz wiedzy. Jest alternatywą dla narzędzi takich jak Notion czy Obsidian, oferując możliwość self-hostingu.

Więcej informacji: https://affine.pro/

## Konfiguracja

### Wymagania

1. Działająca instancja Affine (self-hosted lub cloud)
2. Konto użytkownika w Affine z emailem i hasłem
3. Skonfigurowany widget Vikunja w Glance

### Konfiguracja w glance.yml

Dodaj następujące parametry do swojego widgetu Vikunja:

```yaml
pages:
  - name: Strona główna
    columns:
      - size: small
        widgets:
          - type: vikunja
            url: https://vikunja.example.com
            token: your-vikunja-api-token
            # Konfiguracja Affine
            affine-url: https://affine.example.com
            affine-email: user@example.com
            affine-password: your-affine-password
```

### Parametry konfiguracyjne

- `affine-url` - URL Twojej instancji Affine (z protokołem https://)
- `affine-email` - Email używany do logowania w Affine
- `affine-password` - Hasło do konta Affine

**Uwaga bezpieczeństwa**: Hasło jest przechowywane w pliku konfiguracyjnym. Upewnij się, że plik glance.yml jest odpowiednio zabezpieczony.

## Użycie

### Dodawanie linku do notatki podczas tworzenia zadania

1. Kliknij przycisk "+" w widgecie Vikunja
2. Wypełnij szczegóły zadania (tytuł, termin, etykiety)
3. W polu "Link do notatki Affine" wklej URL do notatki
4. Kliknij "Utwórz"

### Dodawanie linku do notatki w istniejącym zadaniu

1. Kliknij ikonę edycji (ołówek) przy zadaniu
2. W polu "Link do notatki Affine" wklej URL do notatki
3. Kliknij "Zapisz"

### Jak uzyskać URL notatki w Affine

1. Otwórz notatkę w Affine
2. Skopiuj URL z paska adresu przeglądarki
3. URL powinien mieć format: `https://affine.example.com/workspace/WORKSPACE_ID/PAGE_ID`

Przykład:
```
https://affine.example.com/workspace/e225f684-c91e-4d09-b4d9-fb9793257808/kJShB0ELuSgZ2VkVXU3dn
```

### Wyświetlanie powiązanych notatek

Po dodaniu linku do notatki:
- W kolumnie "Notatka" w tabeli zadań pojawi się ikona dokumentu
- Obok ikony wyświetli się tytuł notatki z Affine
- Kliknięcie na link otworzy notatkę w nowej karcie

## Jak to działa

1. **Parsowanie URL**: Glance wyciąga workspace ID i page ID z podanego URL
2. **Logowanie**: Glance loguje się do Affine używając podanych danych
3. **Pobieranie danych**: Przez GraphQL API, Glance pobiera tytuł i metadane notatki
4. **Wyświetlanie**: Tytuł notatki jest wyświetlany przy zadaniu z możliwością kliknięcia

## API używane przez integrację

### Endpoint logowania
```
POST https://affine-url/api/auth/sign-in
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
```

### Endpoint GraphQL (pobieranie notatki)
```
POST https://affine-url/graphql
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "query": "query getWorkspacePageById($workspaceId: String!, $pageId: String!) { workspace(id: $workspaceId) { doc(docId: $pageId) { id mode title summary } } }",
  "variables": {
    "workspaceId": "workspace-id",
    "pageId": "page-id"
  },
  "operationName": "getWorkspacePageById"
}
```

## Rozwiązywanie problemów

### Tytuł notatki nie pojawia się przy zadaniu

**Możliwe przyczyny:**
1. Nieprawidłowe dane logowania do Affine
2. Niepoprawny format URL notatki
3. Brak dostępu do notatki

**Rozwiązanie:**
1. Sprawdź czy dane logowania (email i hasło) są poprawne
2. Upewnij się, że URL ma format: `https://affine-url/workspace/WORKSPACE_ID/PAGE_ID`
3. Sprawdź czy użytkownik ma dostęp do notatki w Affine

### Błąd "failed to sign in to Affine"

**Możliwe przyczyny:**
1. Niepoprawny email lub hasło
2. Affine jest niedostępny
3. Niepoprawny URL do Affine

**Rozwiązanie:**
1. Sprawdź dane logowania
2. Sprawdź czy Affine jest dostępny pod podanym URL
3. Upewnij się, że URL zaczyna się od `https://` lub `http://`

### Błąd "could not extract workspace ID and page ID from URL"

**Możliwa przyczyna:** Niepoprawny format URL

**Rozwiązanie:** Upewnij się, że URL ma format: `https://affine-url/workspace/WORKSPACE_ID/PAGE_ID`

## Bezpieczeństwo

### Przechowywanie danych uwierzytelniających

Dane logowania do Affine są przechowywane w pliku `glance.yml`. Zalecenia:

1. Ustaw odpowiednie uprawnienia do pliku:
   ```bash
   chmod 600 glance.yml
   ```

2. Nie commituj pliku glance.yml z hasłami do repozytorium Git

3. Użyj zmiennych środowiskowych (jeśli wspierane):
   ```yaml
   affine-password: ${AFFINE_PASSWORD}
   ```

### Komunikacja z Affine

- Wszystkie połączenia powinny używać HTTPS
- Token autoryzacyjny jest używany tylko podczas sesji
- Token nie jest przechowywany - każde żądanie wymaga nowego logowania

## Przykłady użycia

### Przypadek użycia 1: Dokumentacja projektu

Masz projekt w Vikunja z wieloma zadaniami. Dla każdego zadania tworzysz notatkę w Affine z:
- Szczegółową specyfikacją
- Notatkami ze spotkań
- Linkami do zasobów

Powiązując zadania z notatkami możesz szybko przejść do dokumentacji klikając ikonę przy zadaniu.

### Przypadek użycia 2: Nauka

Używasz Vikunja do śledzenia postępów w nauce różnych tematów. W Affine prowadzisz szczegółowe notatki z każdego tematu. Łącząc zadania z notatkami masz szybki dostęp do materiałów edukacyjnych.

### Przypadek użycia 3: Kartkówki i egzaminy

Tworzysz zadania dla nadchodzących kartkówek i egzaminów. Każde zadanie ma powiązaną notatkę w Affine z:
- Zakresem materiału
- Notatkami z lekcji
- Zadaniami do przećwiczenia

## Często zadawane pytania

### Q: Czy muszę mieć integrację Affine, aby używać widgetu Vikunja?
A: Nie, integracja z Affine jest opcjonalna. Widget Vikunja działa normalnie bez konfiguracji Affine.

### Q: Czy Glance przechowuje moje dane z Affine?
A: Nie, Glance tylko pobiera tytuł notatki do wyświetlenia. Dane nie są przechowywane.

### Q: Czy mogę użyć publicznej instancji Affine (affine.pro)?
A: Tak, możesz używać dowolnej instancji Affine - zarówno self-hosted jak i oficjalnej.

### Q: Co się stanie jeśli zmienię tytuł notatki w Affine?
A: Przy następnym odświeżeniu widgetu Vikunja (domyślnie co 5 minut) tytuł zostanie zaktualizowany.

### Q: Czy mogę powiązać wiele notatek z jednym zadaniem?
A: Obecnie można powiązać tylko jedną notatkę z zadaniem.

## Przydatne linki

- [Dokumentacja Affine](https://docs.affine.pro/)
- [Dokumentacja Vikunja](https://vikunja.io/docs/)
- [Repozytorium Glance](https://github.com/glanceapp/glance)
