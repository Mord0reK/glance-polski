# Widget Vikunja - Instrukcja użycia

Widget Vikunja pozwala na wyświetlanie i zarządzanie zadaniami z aplikacji Vikunja bezpośrednio z poziomu dashboard Glance.

## Konfiguracja

Aby skonfigurować widget Vikunja, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: vikunja
  url: https://twoja-instancja-vikunja.pl  # URL do Twojej instancji Vikunja
  token: twoj-token-api                     # Token API z Vikunja
  limit: 10                                  # Maksymalna liczba wyświetlanych zadań (opcjonalnie)
```

### Uzyskiwanie tokenu API

1. Zaloguj się do swojej instancji Vikunja
2. Przejdź do ustawień użytkownika
3. Znajdź sekcję "API Tokens" lub "Tokeny API"
4. Wygeneruj nowy token z odpowiednimi uprawnieniami
5. Skopiuj token do konfiguracji widgetu

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

### 3. Edycja zadania ✏️

Aby edytować zadanie:
1. Kliknij przycisk edycji (ikona ołówka) obok zadania
2. Otworzy się okno modalne z formularzem edycji
3. Możesz zmienić:
   - **Tytuł zadania**: Nowy tytuł zadania
   - **Termin**: Data i godzina wykonania zadania (wybór z kalendarza)
   - **Etykiety**: Zaznacz lub odznacz etykiety z listy dostępnych etykiet
4. Kliknij "Zapisz" aby zapisać zmiany lub "Anuluj" aby anulować

### Uwagi

- Po edycji zadania zaleca się odświeżenie strony aby zobaczyć wszystkie zaktualizowane informacje
- Widget automatycznie odświeża dane co 5 minut
- Tylko zadania niewykonane są wyświetlane w widgecie

## Rozwiązywanie problemów

### Widget nie wyświetla zadań

1. Sprawdź czy URL do instancji Vikunja jest poprawny
2. Sprawdź czy token API jest ważny
3. Sprawdź w konsoli przeglądarki czy nie ma błędów połączenia

### Nie mogę oznaczyć zadania jako wykonane

1. Sprawdź czy token API ma odpowiednie uprawnienia do modyfikacji zadań
2. Sprawdź w konsoli przeglądarki czy operacja nie zwraca błędów

### Etykiety nie są wyświetlane w oknie edycji

1. Sprawdź czy w Twojej instancji Vikunja są utworzone jakiekolwiek etykiety
2. Sprawdź czy token API ma uprawnienia do odczytu etykiet

## Przykładowa konfiguracja

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
