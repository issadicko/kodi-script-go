# KodiScript Go SDK

Un interpréteur KodiScript v1.2 pour Go, conçu comme module à intégrer dans vos projets.

## Installation

```bash
go get github.com/kodi-script/kodi-go
```

## Utilisation Rapide

```go
package main

import (
    "fmt"
    kodi "github.com/kodi-script/kodi-go"
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

| Fonction | Description |
|----------|-------------|
| `print(...)` | Affiche des valeurs |
| `toString(val)` | Convertit en string |
| `toNumber(val)` | Convertit en nombre |
| `base64Encode(str)` | Encode en Base64 |
| `base64Decode(str)` | Décode du Base64 |
| `urlEncode(str)` | Encode pour URL |
| `urlDecode(str)` | Décode une URL |
| `jsonParse(str)` | Parse du JSON |
| `jsonStringify(val)` | Sérialise en JSON |
| `typeOf(val)` | Retourne le type |
| `isNull(val)` | Vérifie si null |

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

// Point-virgule optionnel
let x = 1
let y = 2;  // Les deux sont valides
```

## Tests

```bash
go test ./... -v
```
