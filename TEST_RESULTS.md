# TEST_RESULTS

## Формат ведения

Каждый прогон добавляется **сверху**, сразу под этим разделом.
Формат заголовка: `## Прогон: YYYY-MM-DD HH:MM TZ (±HHMM)`.
Результаты — только таблицей. После таблицы — блоки `### Сводка` и `### Падения`.

---

## Прогон: 2026-09-25 07:08 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestParseSpecialty | pdf2table | PASS | 0.00s |
| 6 | TestMetaPairs | pdf2table | PASS | 0.00s |
| 7 | TestLargestTable | pdf2table | PASS | 0.00s |
| 8 | TestSampleMetadata | pdf2table | PASS | 15.73s |
| 9 | TestSampleTextIsNotLost | pdf2table | PASS | 0.00s |
| 10 | TestSampleTablesAndNormalize | pdf2table | PASS | 0.00s |
| 11 | TestSampleGoldenPlan | pdf2table | PASS | 0.00s |
| 12 | TestSampleSplit | pdf2table | PASS | 14.96s |
| 13 | TestSampleWriters | pdf2table | PASS | 0.99s |

### Сводка

Всего: 13
PASS: 13
FAIL: 0
SKIP: 0
Время: 31.687s
Примечание: с учётом подтестов (по одному на каждый из 13 PDF) — 78 PASS, 0 FAIL, 0 SKIP.

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: main
Сэмплы: 13 golden PDF из `../energydocs` (переопределяется `PDF2TABLE_SAMPLES`).
Проверено: метаданные код/название для каждого PDF; отсутствие потери текста внутри сетки; прямоугольность main-таблицы; непустые колонки после normalize; golden-строки учебного плана 09.02.12; split-режим; все форматы вывода (csv/md/html/json/xlsx) с метаданными и детерминизмом.

---

## Прогон: 2026-09-25 07:02 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestParseSpecialty | pdf2table | PASS | 0.00s |
| 6 | TestMetaPairs | pdf2table | PASS | 0.00s |
| 7 | TestSampleTables | pdf2table | SKIP | 0.00s |
| 8 | TestSplitTables | pdf2table | SKIP | 0.00s |
| 9 | TestLargestTable | pdf2table | PASS | 0.00s |
| 10 | TestNormalize | pdf2table | SKIP | 0.00s |
| 11 | TestWriteXLSX | pdf2table | SKIP | 0.00s |
| 12 | TestSamplePlanValues | pdf2table | SKIP | 0.00s |

### Сводка

Всего: 12
PASS: 7
FAIL: 0
SKIP: 5
Время: 0.003s

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: main
Примечание: SKIP — sample PDF не найден (`PDF2TABLE_SAMPLE`, `testdata/document.pdf`, `/home/fworld/document.pdf`). Функциональность проверена прогоном CLI по всем 13 PDF в `../energydocs` — specialty_code/specialty_name извлечены корректно.

---

## Прогон: 2026-09-24 03:25 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestSampleTables | pdf2table | PASS | 0.96s |
| 6 | TestSplitTables | pdf2table | PASS | 1.89s |
| 7 | TestLargestTable | pdf2table | PASS | 0.00s |
| 8 | TestNormalize | pdf2table | PASS | 0.95s |
| 9 | TestWriteXLSX | pdf2table | PASS | 1.02s |
| 10 | TestSamplePlanValues | pdf2table | PASS | 0.97s |

### Сводка

Всего: 10
PASS: 10
FAIL: 0
SKIP: 0
Время: 5.792s

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: n/a (не git-репозиторий)

---

## Прогон: 2026-09-24 03:15 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestSampleTables | pdf2table | PASS | 0.95s |
| 6 | TestLargestTable | pdf2table | PASS | 0.00s |
| 7 | TestNormalize | pdf2table | PASS | 0.96s |
| 8 | TestWriteXLSX | pdf2table | PASS | 1.01s |
| 9 | TestSamplePlanValues | pdf2table | PASS | 0.95s |

### Сводка

Всего: 9
PASS: 9
FAIL: 0
SKIP: 0
Время: 3.873s

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: n/a (не git-репозиторий)

---

## Прогон: 2026-09-24 03:13 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestSampleTables | pdf2table | PASS | 0.95s |
| 6 | TestLargestTable | pdf2table | PASS | 0.00s |
| 7 | TestNormalize | pdf2table | PASS | 0.94s |
| 8 | TestSamplePlanValues | pdf2table | PASS | 0.96s |

### Сводка

Всего: 8
PASS: 8
FAIL: 0
SKIP: 0
Время: 2.849s

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: n/a (не git-репозиторий)

---

## Прогон: 2026-09-24 03:09 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestSampleTables | pdf2table | PASS | 0.94s |
| 6 | TestLargestTable | pdf2table | PASS | 0.00s |
| 7 | TestSamplePlanValues | pdf2table | PASS | 0.93s |

### Сводка

Всего: 7
PASS: 7
FAIL: 0
SKIP: 0
Время: 1.873s

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: n/a (не git-репозиторий)

---

## Прогон: 2026-09-24 03:07 MSK (+0300)

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestUniqueSorted | pdf2table | PASS | 0.00s |
| 2 | TestBand | pdf2table | PASS | 0.00s |
| 3 | TestCleanText | pdf2table | PASS | 0.00s |
| 4 | TestPageRunsGroupsChars | pdf2table | PASS | 0.00s |
| 5 | TestSampleTables | pdf2table | PASS | 0.95s |
| 6 | TestSamplePlanValues | pdf2table | PASS | 0.97s |

### Сводка

Всего: 6
PASS: 6
FAIL: 0
SKIP: 0
Время: 1.926s

### Падения

| Тест | Причина |
|------|---------|
| нет | — |

Команда: `go test ./... -v -count=1`
Go: go1.26.3 linux/amd64
Ветка: n/a (не git-репозиторий)
