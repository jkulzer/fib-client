# Installation

## CI/CD Artefakte

Durch den neuesten GitHub Release (https://github.com/jkulzer/fib-client/releases/latest) lässt sich die neueste APK herunterladen.

## Manuelle Kompilierung

1. Den Paketmanager Nix installieren (https://nixos.org/download)

2. Die Dev-Umgebung starten
```bash
nix develop
```
Diese Entwicklungsumgebung konfiguriert automatisch externe Dependencies der Software, wie Pakete oder die Android NDK.

Die folgenden Schritte werden alle in der Nix Shell ausgeführt

3. Dependencies installieren

```bash
go mod tidy
```

4. Die App kompilieren

```bash
make package
```

Daraufhin wird die App kompiliert. Die erstellte APK hat dann den Namen `fib_client.apk`
