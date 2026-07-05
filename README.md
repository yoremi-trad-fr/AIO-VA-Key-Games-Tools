# AIO VA / Key Game Tools

Interface graphique open source pour regrouper les principaux outils de modding et de traduction des jeux VisualArt's / Key.

L'application fonctionne comme un wrapper Wails/Go : elle ne remplace pas les outils d'origine, elle les lance depuis le dossier `bin/` avec une interface commune et une console integree.

## Outils integres

- **LuckSystem 2.3.2 - fork Yoremi v3.25** : scripts, PAK, fontes, images CZ, videos BGMOVIE/MVT, extraction audio MUSIC/VOICE en Ogg/MP3, preset arabe Font Edit, bridge Siglus -> Luca et workflows dialogue TSV.
- **RLdev2026-Go 1.3.5** : extraction/rebuild `SEEN.TXT`, compilation `.org/.ke/.avg`, G00, GAN, NWA, DAT, Babel, edition/export/import de saves RealLive avec `rlsave`.
- **Siglus Tools 0.61 - Go port** : `Scene.pck`, scripts `.ss`, `Gameexe`, DBS, mobile PCK, OMV et conversions annexes.

## Nouveautes RLdev integrees

## Nouveautes LuckSystem 3.25 integrees

- Panneau `PAK (Audio)` avec extraction `MUSIC.PAK`, `VOICE.PAK` et `SYSVOICE.PAK`.
- Sortie native Ogg, fichier liste compatible reinjection PAK, et copie MP3 optionnelle via FFmpeg.
- Conversion de dossiers audio `Ogg -> MP3` et `MP3 -> Ogg` depuis l'interface AIO.

## Rattrapage LuckSystem 3.24 integre

- `Font Edit` expose maintenant le preset arabe et les reglages manuels `Set Y`, `Y offset`, `X offset`, `Advance offset` et `Connector bleed`.
- `PAK (Font) -> Font Replace` prend en charge le remplacement d'un fichier unique par nom interne exact, utile pour `info30`, `info32` ou `明朝30`.

## Nouveautes RLdev integrees

- Panneau `RealLive save editor` avec Info, Map, Doctor, Diff, Get, Dump, Set, Export texte et Build `.sav`.
- Profils generiques pour `read.sav`, `save999.sav` et les dwords bas niveau.
- Export `rlsave` avec extension `.txt` par defaut.
- Console partagee redimensionnable, pratique pour les gros logs batch.
- Support des nouveaux batchs paralleles de `rlsave` et `vaconv` cote outils.

## Structure attendue

```text
AIO-VA-Key-Games-Tools/
  GUI-Sources/
    bin/
      lucksystem.exe
      kprl16.exe
      rlc2026.exe
      vaconv.exe
      rlxml.exe
      rlsave.exe
      ...
    frontend/
    app.go
  Siglus go/
  README.md
```

Les binaires peuvent aussi etre places a cote de l'executable final, dans un dossier `bin/`. L'application essaie aussi de retrouver certains outils depuis le `PATH`, mais le dossier `bin/` reste la methode recommandee pour une release.

## Utilisation

1. Ouvrir l'application AIO.
2. Choisir le moteur dans l'ecran d'accueil : LuckSystem, RLdev, Siglus ou outils divers.
3. Selectionner les fichiers/dossiers demandes par le panneau.
4. Lancer l'action et suivre la sortie dans la console.

Les operations qui modifient des saves RealLive creent des sauvegardes `.bak` par defaut quand `rlsave` ecrit dans un fichier existant.

## Build de la GUI

Prerequis :

- Go 1.22 ou plus recent.
- Node.js/npm pour le frontend Svelte.
- Wails v2 installe sur la machine de build.

Depuis `GUI-Sources` :

```powershell
npm install --prefix frontend
wails build
```

Pour tester seulement les parties Go :

```powershell
go test ./...
```

Pour reconstruire uniquement le frontend :

```powershell
npm run build --prefix frontend
```

## Notes

- Le projet AIO reste un wrapper : les bugs de format propres aux moteurs sont generalement corriges dans les depots/outils correspondants.
- Les chemins longs Windows sont mieux geres en privilegiant les modes dossier et les sorties dediees.
- Avant une operation destructive sur un jeu ou une save, garder une copie du dossier original reste recommande.
