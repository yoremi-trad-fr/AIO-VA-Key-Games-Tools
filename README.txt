New project draft
This project is based on the observation that there are currently three main open source tools for decompiling/recompiling Key/VA game files.
LuckSystem/ Siglus / Rldev

The idea is to combine these three pieces of software into a single GUI, coded in Go and open source.
The versions of these software programmes are: 
-LuckSystem 2.3.2-Fork-Yoremi
-Siglus Tools 0.61
-Rldev- fork 2026

Lucksystem is already coded in Go and has a GUI (wrapper), which will serve as the basis for the project. I have started porting Siglus, and Rldev, which is natively coded in Ocaml, remains. I have started porting Siglus, and Rldev, which is natively coded in Ocaml, remains.

RLDEV stat : 

Ce qui reste à porter en GO📋
1. rlc backend (~8 500 lignes OCaml)
Fichier OCaml	Lignes	Description	Priorité
keTypes.ml	265	Types de base, registre des fonctions KFN	🔴
memory.ml + variables.ml	665	Table de symboles, scoping, allocation	🔴
codegen.ml + bytecodeGen.ml	411	Génération IR et bytecode binaire	🔴
compilerFrame.ml	1 315	Pipeline principal AST → bytecode	🔴
expr.ml	1 383	Normalisation/transformation d'expressions	🔴
function.ml + funcAsm.ml	936	Compilation des appels de fonctions	🟡
intrinsic.ml	410	Builtins (defined?, assert, etc.)	🟡
textout.ml	506	Compilation de texte dynamique	🟡
goto.ml	261	goto_on, goto_case, goto_if	🟡
select.ml	222	Menus de sélection	🟡
rlBabel.ml	260	Encodage multi-langue (latin/CJK)	🟡
ini.ml + iniParser.mly + iniLexer.mll	261	Parseur GAMEEXE.INI	🟢
directive.ml	110	Compilation des directives	🟢
global.ml + meta.ml	140	État global, métadonnées	🟢
app.ml + main.ml	591	CLI du compilateur rlc	🟢

2. common non porté (~7 100 lignes OCaml)
Fichier OCaml	Lignes	Description	Priorité
text.ml	623	Manipulation Unicode/Shift_JIS	🔴
textTransforms.ml	537	Encodage bytecode ↔ texte	🔴
cp932.ml + cp932_in.ml	1 391	Tables Shift_JIS	🟡
cp936.ml + cp936_in.ml	997	Tables GB2312/GBK	🟡
cp949.ml + cp949_in.ml	1 716	Tables EUC-KR	🟡
game.ml + gameParser.mly + gameLexer.mll	1 240	Parseur de données de jeu	🟡
kfnTypes.ml + kfnParser.mly + kfnLexer.mll	411	Parseur reallive.kfn	🔴
cast.ml + castParser.mly + castLexer.mll	269	Parseur de fichiers cast	🟢
optpp.ml	423	Pretty-printer et logging	🟢
iMap.ml + iSet.ml + avlTree.ml	569	Structures de données	🟢 (Go a map/slice)

3. Outils secondaires (~716 lignes OCaml)
Outil	Lignes	Description	Priorité
rlxml	419	Convertisseur GAN ↔ XML	🟡
cvtkpacres	297	Convertisseur de ressources	🟢
docsrc	217	Générateur de doc	🟢
Résumé chiffré
	OCaml	Go porté	Reste estimé
Lignes source	30 377	8 883	~12 000-15 000 Go
Tests	—	2 210	—
Couverture	100%	~45%	~55%
