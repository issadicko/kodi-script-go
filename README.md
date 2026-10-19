# KodiScript Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/issadicko/kodi-script-go.svg)](https://pkg.go.dev/github.com/issadicko/kodi-script-go)
[![Release](https://img.shields.io/github/v/release/issadicko/kodi-script-go)](https://github.com/issadicko/kodi-script-go/releases)

A lightweight, embeddable scripting language interpreter for Go applications.

📖 **Documentation complète** : [docs-kodiscript.dickode.net](https://docs-kodiscript.dickode.net/)

## 🎯 Pourquoi KodiScript ?

Vous avez déjà eu besoin d'exécuter du code dynamiquement dans votre application ? De laisser vos utilisateurs (admins) définir des règles métier sans recompiler tout le projet ? C'est exactement pour ça que KodiScript existe.

**KodiScript** est un langage de script minimaliste, pensé pour être :

- **Simple à apprendre** — Une syntaxe épurée, proche du JavaScript, que n'importe qui peut comprendre en quelques minutes
- **Léger** — Pas de dépendances lourdes, juste l'essentiel pour faire le travail
- **Sécurisé** — Exécution sandboxée, vos utilisateurs peuvent écrire des scripts sans risquer de casser votre système
- **Facile à intégrer** — Quelques lignes de code suffisent pour l'embarquer dans votre projet Go

Imaginez : un admin qui configure des règles de validation, un workflow qui s'adapte selon le contexte, ou des transformations de données à la volée. Tout ça devient possible sans toucher à votre code source.

## Installation

```bash
go get github.com/issadicko/kodi-script-go@v0.3.0
```

## Utilisation Rapide

```go
package main

import (
    "fmt"
    kodi "github.com/issadicko/kodi-script-go"
)

func main() {
    // Exécution simple
    result := kodi.Run(`
        let name = "Kodi"
        let version = 1.2
        print("Hello " + name)
    `, nil)

    for _, line := range result.Output {
        fmt.Println(line)
    }
}
```

## Injection de Variables

```go
vars := map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Alice",
        "role": "admin",
    },
    "config": map[string]interface{}{
        "debug": true,
    },
}

result := kodi.Run(`
    let greeting = "Hello " + user.name
    let status = user?.active ?: "offline"
    print(greeting)
`, vars)
```

## Fonctions Natives

### Chaînes de caractères
| Fonction | Description |
|----------|-------------|
| `print(...)` | Affiche des valeurs |
| `toString(val)` | Convertit en string |
| `toNumber(val)` | Convertit en nombre |
| `length(str)` | Longueur d'une chaîne |
| `substring(str, start, [end])` | Extrait une sous-chaîne |
| `toUpperCase(str)` | Convertit en majuscules |
| `toLowerCase(str)` | Convertit en minuscules |
| `trim(str)` | Supprime les espaces |
| `replace(str, old, new)` | Remplace du texte |
| `split(str, sep)` | Sépare en tableau |
| `join(arr, sep)` | Joint un tableau |
| `contains(str, substr)` | Vérifie si contient |
| `startsWith(str, prefix)` | Vérifie le début |
| `endsWith(str, suffix)` | Vérifie la fin |
| `indexOf(str, substr)` | Position d'une sous-chaîne |

### Math
| Fonction | Description |
|----------|-------------|
| `abs(n)` | Valeur absolue |
| `floor(n)` | Arrondi inférieur |
| `ceil(n)` | Arrondi supérieur |
| `round(n)` | Arrondi |
| `min(a, b, ...)` | Minimum |
| `max(a, b, ...)` | Maximum |
| `pow(base, exp)` | Puissance |
| `sqrt(n)` | Racine carrée |
| `sin(n)`, `cos(n)`, `tan(n)` | Trigonométrie |
| `log(n)`, `log10(n)`, `exp(n)` | Logarithmes |

### Random
| Fonction | Description |
|----------|-------------|
| `random()` | Nombre aléatoire [0, 1) |
| `randomInt(min, max)` | Entier aléatoire |
| `randomUUID()` | UUID v4 aléatoire |

### Crypto
| Fonction | Description |
|----------|-------------|
| `md5(str)` | Hash MD5 |
| `sha1(str)` | Hash SHA-1 |
| `sha256(str)` | Hash SHA-256 |

### JSON / Encodage
| Fonction | Description |
|----------|-------------|
| `jsonParse(str)` | Parse du JSON |
| `jsonStringify(val)` | Sérialise en JSON |
| `base64Encode(str)` | Encode en Base64 |
| `base64Decode(str)` | Décode du Base64 |
| `urlEncode(str)` | Encode pour URL |
| `urlDecode(str)` | Décode une URL |

### Tableaux
| Fonction | Description |
|----------|-------------|
| `sort(arr, [order])` | Trie (asc/desc) |
| `sortBy(arr, field, [order])` | Trie par champ |
| `reverse(arr)` | Inverse l'ordre |
| `size(arr)` | Taille du tableau |
| `first(arr)` | Premier élément |
| `last(arr)` | Dernier élément |
| `slice(arr, start, [end])` | Extrait une portion |
| `range([start], end)`, `sum(arr)`, `avg(arr)` | Intervalle, somme, moyenne |
| `unique(arr)`, `flatten(arr)`, `push(arr, ...)`, `concat(...)` | Dédoublonne, aplatit, ajoute, concatène |
| `chunk(arr, n)` | Découpe en tableaux de `n` éléments |
| `compact(arr)` | Retire les `null` |
| `pluck(arr, chemin)` | La valeur au chemin (`"client.ville"`) pour chaque élément |
| `zip(arr1, arr2, ...)` | `[[a0, b0], [a1, b1], …]`, à la longueur du plus court |
| `fromEntries(paires)` | Objet bâti depuis des paires `[clé, valeur]` |
| `difference(a, b)`, `intersection(a, b)`, `union(...)` | Opérations d'ensemble (type et valeur comparés) |

Un tableau `null` se lit comme un tableau vide.

### Regroupements
La clé est un **chemin** lu sur chaque élément, ou une **fonction** `fn(item, index)`.

| Fonction | Description |
|----------|-------------|
| `groupBy(arr, clé)` | Objet `{clé: [éléments]}` |
| `countBy(arr, clé)` | Objet `{clé: nombre}` |
| `indexBy(arr, clé)` | Objet `{clé: élément}`, le dernier l'emporte |
| `uniqueBy(arr, clé)` | Le premier élément de chaque clé |
| `sumBy(arr, clé)` | Somme des clés (les `null` sont ignorées) |
| `minBy(arr, clé)`, `maxBy(arr, clé)` | L'élément de plus petite / plus grande clé |
| `partition(arr, clé)` | `[[clé vraie], [les autres]]` |

### Types
| Fonction | Description |
|----------|-------------|
| `typeOf(val)` | Retourne le type |
| `isNull(val)` | Vérifie si null |
| `isNumber(val)` | Vérifie si nombre |
| `isString(val)` | Vérifie si chaîne |
| `isBool(val)` | Vérifie si booléen |

### Date/Heure
Une **date** est un nombre de millisecondes depuis 1970 (UTC) ou un texte ISO-8601 : `2026-10-19`,
`2026-10-19T08:30`, `2026-10-19T08:30:00.250Z`, `2026-10-19 08:30:00+01:00`. Sans fuseau, le texte
est en UTC. Une date illisible (`2026-02-29`, `"abc"`, `null`) donne `null`. Tous les calculs de
calendrier se font en **UTC** ; une fonction qui produit une date rend des millisecondes.

| Fonction | Description |
|----------|-------------|
| `now()` | Instant actuel en ms |
| `date()`, `time()`, `datetime()` | Date `YYYY-MM-DD`, heure `HH:mm:ss`, instant ISO (UTC) |
| `timestamp([date])` | La date en ms (maintenant sans argument) |
| `isDate(valeur)` | Vrai si la valeur se lit comme une date |
| `parseDate(texte, format)` | Lit `"19/10/2026"` avec `"DD/MM/YYYY"` ; `null` si le texte ne suit pas le format |
| `toISO(date)` | `YYYY-MM-DDTHH:mm:ss.SSSZ` |
| `formatDate(date, [fmt])` | Jetons `YYYY`, `YY`, `MM`, `DD`, `HH`, `mm`, `ss`, `SSS` |
| `year`, `month`, `day`, `hour`, `minute`, `second` `([date])` | Champs de la date |
| `dayOfWeek([date])` | Jour de semaine (0 = dimanche) |
| `daysInMonth(date)`, `isWeekend(date)` | Jours du mois, samedi ou dimanche |
| `addDays`, `addHours`, `addMinutes`, `addSeconds` `(date, n)` | Date décalée |
| `addMonths(date, n)`, `addYears(date, n)` | Décalage de calendrier ; un jour absent du mois visé devient son dernier jour |
| `diffDays`, `diffHours`, `diffMinutes` `(a, b)` | Unités entières de `a` à `b` (négatif si `b` précède) |
| `diffMonths(a, b)`, `diffYears(a, b)` | Mois / années de calendrier entiers de `a` à `b` |
| `age(naissance, [à])` | Années entières révolues (à maintenant par défaut) |
| `startOfDay`, `endOfDay`, `startOfMonth`, `endOfMonth` `(date)` | Bornes de la période, en ms |

## 🔌 Extensibilité

KodiScript est conçu pour être **extensible**. Vous pouvez enrichir le langage en ajoutant vos propres fonctions natives, permettant aux scripts d'interagir avec votre système.

### Fonctions personnalisées

```go
script := kodi.New(`
    let result = myCustomFunc("hello")
    print(result)
`)

script.RegisterFunction("myCustomFunc", func(args ...interface{}) (interface{}, error) {
    return strings.ToUpper(args[0].(string)), nil
})

result := script.Execute()
```

### Exemple : intégration métier

```go
script := kodi.New(`
    let user = fetchUser(userId)
    let discount = calculateDiscount(user.tier, orderTotal)
    
    if (discount > 0) {
        sendNotification(user.email, "Vous avez " + discount + "% de réduction!")
    }
    
    return discount
`)

script.RegisterFunction("fetchUser", func(args ...interface{}) (interface{}, error) {
    id := args[0].(float64)
    // Appel à votre base de données
    return map[string]interface{}{
        "id": id, "name": "Alice", "tier": "gold", "email": "alice@example.com",
    }, nil
})

script.RegisterFunction("calculateDiscount", func(args ...interface{}) (interface{}, error) {
    tier := args[0].(string)
    switch tier {
    case "gold": return 20, nil
    case "silver": return 10, nil
    default: return 5, nil
    }
})

script.RegisterFunction("sendNotification", func(args ...interface{}) (interface{}, error) {
    // Envoyer une notification
    return true, nil
})

result := script.WithVariable("userId", 123).WithVariable("orderTotal", 100).Execute()
```

Cela permet à vos utilisateurs d'écrire des scripts puissants tout en gardant le contrôle sur les fonctionnalités exposées.

## Syntaxe KodiScript v1.2

```javascript
// Variables
let name = "Kodi"
let version = 1.2

// Null-safety
let status = user?.active ?: "offline"

// Conditions
if (version > 1.0) {
    print("Modern version")
} else {
    print("Legacy version")
}

// Expressions multi-lignes (continuation automatique)
let total = 10 +
            20 +
            30

// Return statement (arrête l'exécution et retourne la valeur)
let x = 10
if (x > 5) {
    return "grand"  // Retour anticipé
}
return "petit"

// Point-virgule optionnel
let a = 1
let b = 2;  // Les deux sont valides
```

## Tests

```bash
go test ./... -v
```

## Other Implementations

| Language | Package |
|----------|---------|  
| **Kotlin** | [Maven Central](https://central.sonatype.com/artifact/io.github.issadicko/kodi-script) |
| **Dart** | [pub.dev](https://pub.dev/packages/kodi_script) |
| **TypeScript** | [npm](https://www.npmjs.com/package/@issadicko/kodi-script) |

## License

MIT
