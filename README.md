# AIO VA / Key Game Tools

Interface graphique open source pour regrouper les principaux outils de modding et de traduction des jeux VisualArt's / Key.

L'application fonctionne comme un wrapper Wails/Go : elle ne remplace pas les outils d'origine, elle les lance depuis le dossier `bin/` avec une interface commune et une console integree.

## Outils integres

- **LuckSystem 2.3.2 - fork Yoremi v3.30** : scripts, PAK, fontes, images CZ, videos BGMOVIE/MVT, extraction audio MUSIC/VOICE en Ogg/MP3, bridge Siglus -> Luca, workflows dialogue TSV et generateur Luca Menu DLL.
- **RLdev2026-Go 1.3.5** : extraction/rebuild `SEEN.TXT`, compilation `.org/.ke/.avg`, G00, GAN, NWA, DAT, Babel, edition/export/import de saves RealLive avec `rlsave`.
- **Siglus Tools 0.61 - Go port** : `Scene.pck`, scripts `.ss`, `Gameexe`, DBS, mobile PCK, OMV et conversions annexes.

## Profils Siglus Harmonia

- Deux profils distincts sont proposes : `Harmonia - Édition physique` et
  `Harmonia - Édition Steam`. Ils partagent la meme cle de chiffrement.
- Les deux PCK d'origine utilisent le WTF `0x166`. Lors d'un rebuild en mode
  `auto`, cette valeur sert de secours si `_siglus_pck.json` est absent.
- Le simple alias `Harmonia` reste accepte et selectionne l'edition physique.
- La liste affiche des titres romanises et des libelles d'edition en francais ;
  les anciens titres japonais restent acceptes dans les commandes.
- Le dump TXT propose un mode `Une seule ligne` : il conserve uniquement le
  slot traduction `●`, afin d'eviter la copie source/traduction en double tout
  en restant directement compatible avec la reinjection Siglus.
- Le dump SS exporte par defaut toutes les chaines non vides, y compris les
  dialogues entierement en anglais. Le mode `Dialogue only` retire uniquement
  les identifiants techniques connus ; `Japanese only (legacy)` conserve
  l'ancien filtre non-ASCII.

## Nouveautes RLdev integrees

## Nouveautes LuckSystem 3.30 integrees

- Generateur `Luca Menu DLL` pour AIR, Kanon, Harmonia HD, Loopers et Little Busters! English Edition.
- Proxy `version.dll` x64 ou `winmm.dll` x86 pour LBEE, avec verification des offsets et du budget des chaines.
- Presets d'injection FR/ENG surs, arabe, russe, japonais et chinois ; fichier `PATCHES` personnalise pour LBEE.
- Selecteur persistant francais/anglais pour le nouveau workflow.
- `PAK (Font) -> Font Replace` sait maintenant creer un alias interne, par exemple `info30` vers `info32`.

## Nouveautes LuckSystem 3.26 conservees

- Correctif LBEE/CZ3 : conservation des 8 octets d'en-tete et placement correct de la table LZW a `HeaderLength`.
- Import batch PNG -> CZ : les PNG nommes `*.cz3.png` peuvent maintenant correspondre aux entrees PAK sans extension `.cz3`.

## Rattrapage LuckSystem 3.25 integre

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
