# KodiScript Go SDK

Un interpréteur KodiScript v1.2 pour Go, conçu comme module à intégrer dans vos projets.

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
go get github.com/issadicko/kodi-script-go
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

### Types
| Fonction | Description |
|----------|-------------|
| `typeOf(val)` | Retourne le type |
| `isNull(val)` | Vérifie si null |
| `isNumber(val)` | Vérifie si nombre |
| `isString(val)` | Vérifie si chaîne |
| `isBool(val)` | Vérifie si booléen |

## Fonctions Personnalisées

```go
script := kodi.New(`
    let result = myCustomFunc("hello")
`)

script.RegisterFunction("myCustomFunc", func(args ...interface{}) (interface{}, error) {
    return strings.ToUpper(args[0].(string)), nil
})

result := script.Execute()
```

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
