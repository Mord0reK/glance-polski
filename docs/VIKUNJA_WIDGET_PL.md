# Widget Vikunja - Instrukcja użycia

Widget Vikunja pozwala na wyświetlanie i zarządzanie zadaniami z aplikacji Vikunja bezpośrednio z poziomu dashboard Glance.

## Konfiguracja

Aby skonfigurować widget Vikunja, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: vikunja
  url: https://twoja-instancja-vikunja.pl  # URL do Twojej instancji Vikunja
  token: twoj-token-api                     # Token API z Vikunja
  project-id: 1                             # ID projektu do tworzenia nowych zadań (opcjonalnie, domyślnie 1)
  limit: 10                                  # Maksymalna liczba wyświetlanych zadań (opcjonalnie)
  # Integracja z Affine (opcjonalnie)
  affine-url: https://twoja-instancja-affine.pl      # URL do Twojej instancji Affine
  affine-email: twoj-email@example.com                # Email do logowania Affine
  affine-password: twoje-haslo-affine                 # Hasło do logowania Affine
```

### Uzyskiwanie tokenu API

1. Zaloguj się do swojej instancji Vikunja
2. Przejdź do ustawień użytkownika
3. Znajdź sekcję "API Tokens" lub "Tokeny API"
4. Wygeneruj nowy token z odpowiednimi uprawnieniami
5. Skopiuj token do konfiguracji widgetu

### Znajdowanie ID projektu

Aby znaleźć ID projektu w Vikunja:
1. Otwórz projekt w przeglądarce
2. Sprawdź URL - ID projektu znajduje się w adresie (np. `/projects/5` oznacza ID projektu = 5)
3. Użyj tego ID w konfiguracji `project-id`

**Uwaga**: Parametr `project-id` określa, w którym projekcie będą tworzone nowe zadania. Jeśli masz wiele projektów, ustaw ID projektu, w którym chcesz tworzyć zadania. Domyślnie używany jest projekt o ID 1.

## Funkcje widgetu

### 1. Wyświetlanie zadań

Widget automatycznie pobiera i wyświetla zadania z Vikunja:
- **Koniec za**: Czas pozostały do terminu wykonania zadania
- **Treść zadania**: Tytuł zadania
- **Etykiety**: Etykiety przypisane do zadania (z kolorami)

Zadania są sortowane według daty - zadania z najbliższym terminem są wyświetlane jako pierwsze.

### 2. Oznaczanie zadania jako wykonane ✓

Aby oznaczyć zadanie jako wykonane:
1. Kliknij w checkbox (pole wyboru) obok zadania
2. Potwierdź operację w wyświetlonym dialogu
3. Zadanie zostanie automatycznie usunięte z listy po oznaczeniu jako wykonane

### 3. Dodawanie nowego zadania ➕

Aby dodać nowe zadanie:
1. Kliknij przycisk "+" (plus) w prawym górnym rogu widgetu
2. Otworzy się okno modalne z formularzem tworzenia zadania
3. Wprowadź:
   - **Tytuł zadania**: Nazwa nowego zadania (wymagane)
   - **Termin**: Data i godzina wykonania zadania (opcjonalnie)
   - **Etykiety**: Zaznacz etykiety, które chcesz przypisać do zadania (opcjonalnie)
   - **Link do notatki Affine**: URL do powiązanej notatki w Affine (opcjonalnie)
4. Kliknij "Utwórz" aby utworzyć zadanie lub "Anuluj" aby anulować
5. Widget automatycznie odświeży się i wyświetli nowo utworzone zadanie

### 4. Edycja zadania ✏️

Aby edytować zadanie:
1. Kliknij przycisk edycji (ikona ołówka) obok zadania
2. Otworzy się okno modalne z formularzem edycji
3. Możesz zmienić:
   - **Tytuł zadania**: Nowy tytuł zadania
   - **Termin**: Data i godzina wykonania zadania (wybór z kalendarza)
   - **Etykiety**: Zaznacz lub odznacz etykiety z listy dostępnych etykiet
   - **Link do notatki Affine**: URL do powiązanej notatki w Affine (opcjonalnie)
4. Kliknij "Zapisz" aby zapisać zmiany lub "Anuluj" aby anulować

### 5. Integracja z Affine 📝

Widget Vikunja oferuje integrację z Affine - aplikacją do tworzenia notatek. Ta funkcja pozwala powiązać zadania Vikunja z notatkami w Affine.

#### Konfiguracja integracji z Affine

Aby włączyć integrację z Affine, dodaj następujące parametry do konfiguracji widgetu:

```yaml
- type: vikunja
  url: https://twoja-instancja-vikunja.pl
  token: twoj-token-api
  # Parametry Affine
  affine-url: https://twoja-instancja-affine.pl
  affine-email: twoj-email@example.com
  affine-password: twoje-haslo-affine
```

#### Dodawanie linku do notatki Affine

Podczas tworzenia lub edycji zadania:
1. W polu "Link do notatki Affine" wklej pełny URL do notatki
2. Format URL: `https://affine-url/workspace/WORKSPACE_ID/PAGE_ID`
3. Glance automatycznie pobierze tytuł notatki z Affine
4. W tabeli zadań pojawi się ikona dokumentu z tytułem notatki

#### Wyświetlanie powiązanych notatek

Jeśli zadanie ma powiązaną notatkę Affine:
- W kolumnie "Notatka" wyświetli się ikona dokumentu z tytułem notatki
- Kliknięcie na link otworzy notatkę w Affine w nowej karcie przeglądarki
- Tytuł notatki jest automatycznie pobierany z Affine przy każdym odświeżeniu widgetu

#### Jak znaleźć URL notatki w Affine

1. Otwórz notatkę w Affine
2. Skopiuj URL z paska adresu przeglądarki
3. URL powinien mieć format: `https://your-affine.com/workspace/xxx.../yyy...`
4. Wklej ten URL do pola "Link do notatki Affine" w formularzu zadania

### Uwagi

- Po edycji zadania zaleca się odświeżenie strony aby zobaczyć wszystkie zaktualizowane informacje
- Widget automatycznie odświeża dane co 5 minut
- Tylko zadania niewykonane są wyświetlane w widgecie

## Rozwiązywanie problemów

### Widget nie wyświetla zadań

1. Sprawdź czy URL do instancji Vikunja jest poprawny
2. Sprawdź czy token API jest ważny
3. Sprawdź w konsoli przeglądarki czy nie ma błędów połączenia

### Nie mogę oznaczyć zadania jako wykonane lub utworzyć nowego zadania

1. Sprawdź czy token API ma odpowiednie uprawnienia do modyfikacji zadań
2. Sprawdź w konsoli przeglądarki czy operacja nie zwraca błędów
3. Upewnij się, że Twoja instancja Vikunja jest dostępna i działa poprawnie

### Etykiety nie są wyświetlane w oknie edycji lub tworzenia

1. Sprawdź czy w Twojej instancji Vikunja są utworzone jakiekolwiek etykiety
2. Sprawdź czy token API ma uprawnienia do odczytu etykiet

## Przykładowa konfiguracja

### Podstawowa konfiguracja

```yaml
pages:
  - name: Moja strona główna
    columns:
      - size: small
        widgets:
          - type: vikunja
            url: https://tasks.example.com
            token: abc123xyz789...
            limit: 15
```

### Konfiguracja z integracją Affine

```yaml
pages:
  - name: Moja strona główna
    columns:
      - size: small
        widgets:
          - type: vikunja
            url: https://tasks.example.com
            token: abc123xyz789...
            limit: 15
            affine-url: https://affine.example.com
            affine-email: user@example.com
            affine-password: secure-password-here
```
