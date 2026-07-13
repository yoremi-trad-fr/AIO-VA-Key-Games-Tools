<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime.js';
  import {
    GetLuckSystemPath,
    SetLuckSystemPath,
    ScanGameData,
    SelectPakFile,
    SelectFile,
    SelectDirectory,
    SelectSaveFile,
    StopProcess,
    DefaultKFN,
    SiglusLucaBridge,
    BGMOVIEExtract,
    MusicPakExtract,
    VoicePakExtract,
    AudioConvert,
    ScriptDecompile,
    ScriptCompile,
    PakExtract,
    PakReplace,
    PakFontExtract,
    PakFontReplace,
    FontExtract,
    FontEdit,
    ImageExport,
    ImageImport,
    ImageBatchExport,
    ImageBatchImport,
    DialogueDetectFormat,
    DialogueExtractFile,
    DialogueExtractBatch,
    DialogueImportFile,
    DialogueImportBatch,
    VietnameseFontPatch,
    SupportsLucaMenuDLL,
    ScanLucaMenuKit,
    LucaMenuGenerate,
    SelectScriptTxtFile,
    SelectTsvFile,
    SelectSaveTsvFile,
    SelectSaveScriptFile,
    RldevDisassemble,
    RldevExtract,
    RldevArchive,
    RldevList,
    DefaultBabelRoot,
    RldevCompile,
    RldevCompileBatch,
    RldevOrgTextExport,
    RldevOrgTextImport,
    RldevG00ToPng,
    RldevPngToG00,
    RldevGanToXml,
    RldevXmlToGan,
    RldevNwaToAudio,
    RldevDatToJson,
    RldevDatJsonToBinary,
    RldevSaveInfo,
    RldevSaveMap,
    RldevSaveDoctor,
    RldevSaveDiff,
    RldevSaveGet,
    RldevSaveSet,
    RldevSaveDump,
    RldevSaveExport,
    RldevSaveBuild,
    RldevBabelPrepareRuntime,
    RldevBabelWriteHeader,
    DetectRealLiveVersion,
    SiglusSceneExtract,
    SiglusSceneRebuild,
    SiglusSSDump,
    SiglusSSDumpAll,
    SiglusSSInject,
    SiglusSSInjectAll,
    SiglusGameexeExtract,
    SiglusGameexeRebuild,
    SiglusDBSExtract,
    SiglusDBSRebuild,
    SiglusDBSExportXLSX,
    SiglusDBSBuildFromXLSX,
    SiglusMobilePCKExtract,
    SiglusMobilePCKRebuild,
    SiglusOMVCut,
    SiglusOMVPack,
    SiglusOMV2AVI,
    SiglusOMVPNG,
    SiglusPNGVideo,
    SiglusCombinePNG,
    SiglusScriptRepack,
    SiglusG00ToPng,
    SiglusPngToG00
  } from '../wailsjs/go/main/App.js';

  // ===== State =====
  let activeView = 'hub'; // 'hub' | 'lucksystem' | 'siglus' | 'rldev' | 'outils' | 'about_global'
  let selectedOp = 'decompile';
  let running = false;
  let consoleLines = [];
  let consoleEl;
  let consoleHeight = 160;
  let consoleResizing = false;
  let consoleResizeStartY = 0;
  let consoleResizeStartHeight = 160;
  let consoleMenuVisible = false;
  let consoleMenuX = 0;
  let consoleMenuY = 0;
  let lsPath = '';
  let lucaMenuDllAvailable = false;
  let uiLanguage = 'fr';

  function t(fr, en, language = uiLanguage) {
    return language === 'en' ? en : fr;
  }

  function setUiLanguage(language) {
    uiLanguage = language === 'en' ? 'en' : 'fr';
    try { localStorage.setItem('lucksystem-ui-language', uiLanguage); } catch (_) {}
  }

  // --- Script fields ---
  let pakFile = '';
  let opcodeFile = '';
  let pluginFile = '';
  let charsetVal = 'UTF-8';
  let gameName = '';
  let gamePresets = [];
  let selectedPreset = '';
  let outputDir = '';
  let importDir = '';
  let outputPak = '';

  // --- Siglus -> Luca bridge fields ---
  let siglusLucaLucaDir = '';
  let siglusLucaSiglusDir = '';
  let siglusLucaOutput = '';
  let siglusLucaTargetCol = 2;

  // --- PAK fields ---
  let pakExtSource = '';
  let pakExtOutput = '';
  let pakRepSource = '';
  let pakRepListFile = '';
  let pakRepInput = '';
  let pakRepOutput = '';
  let pakRepUseList = true; // mode par défaut : fichier liste

  // --- BGMOVIE / Video fields ---
  let bgMoviePak = '';
  let bgMovieOutput = '';

  // --- PAK Audio fields ---
  let musicPak = '';
  let musicOutput = '';
  let musicToMp3 = false;
  let voicePak = '';
  let voiceOutput = '';
  let voiceToMp3 = false;
  let audioConvInput = '';
  let audioConvOutput = '';
  let audioConvDirection = 'mp3'; // 'mp3' | 'native'

  // --- PAK Font fields ---
  let pakFontExtSource = '';
  let pakFontExtCharset = 'UTF-8';
  let pakFontExtOutput = '';
  let pakFontRepSource = '';
  let pakFontRepCharset = 'UTF-8';
  let pakFontRepListFile = '';
  let pakFontRepInput = '';
  let pakFontRepSingleFile = '';
  let pakFontRepSingleName = '';
  let pakFontRepAliasFrom = 'info30';
  let pakFontRepAliasTo = 'info32';
  let pakFontRepOutput = '';
  let pakFontRepMode = 'list'; // 'list' | 'dir' | 'single' | 'alias'

  // --- Font Extract ---
  let fontExtCz = '';
  let fontExtInfo = '';
  let fontExtPng = '';
  let fontExtCharset = '';

  // --- Font Edit ---
  let fontEditCz = '';
  let fontEditInfo = '';
  let fontEditTtf = '';
  let fontEditOutCz = '';
  let fontEditOutInfo = '';
  let fontEditCharsetFile = '';
  let fontEditMode = 'append'; // 'redraw' | 'append' | 'insert'
  let fontEditIndex = 0;
  let fontEditArabicMetrics = false;
  let fontEditMetricSetYEnabled = false;
  let fontEditMetricSetY = 0;
  let fontEditMetricYOffset = 0;
  let fontEditMetricXOffset = 0;
  let fontEditMetricWOffset = 0;
  let fontEditArabicConnectorBleed = 0;

  // --- Vietnamese Font Patch ---
  let vietFontRoot = '';
  let vietCharsetFile = '';
  let vietTtfFile = '';
  let vietOutputDir = '';
  let vietSlot = 'en';
  let vietFamily = 'GOTHIC1';
  let vietYMinus2 = false;
  let vietYMinus1 = false;
  let vietY0 = false;
  let vietY1 = false;
  let vietY2 = true;
  let vietY3 = false;
  let vietRedrawLatin = false;

  // --- Luca Menu DLL ---
  let lucaInventory = null;
  let lucaGame = '';
  let lucaSlot = 'en';
  let lucaPatchName = '';
  let lucaPatchVersion = '0.1-gui';
  let lucaExe = '';
  let lucaOutputDir = '';
  let lucaBuildDll = true;
  let lucaProxyChoice = 'version';
  let lucaCustomPatch = '';
  let lucaFillMode = 'fr-safe';
  let lucaSearch = '';
  let lucaEntries = [];

  // --- Image Export ---
  let imgExpBatch = false;
  let imgExpInput = '';
  let imgExpOutput = '';

  // --- Image Import ---
  let imgImpBatch = false;
  let imgImpSource = '';
  let imgImpInput = '';
  let imgImpOutput = '';
  let imgImpFill = false;

  // --- Dialogue Extract ---
  let dlgExtBatch = false;
  let dlgExtInput = '';
  let dlgExtOutput = '';
  let dlgExtLang1 = false;
  let dlgExtLang2 = true;   // default: Lang 2 (typically ENG in AIR)
  let dlgExtLang3 = false;
  let dlgExtLang4 = false;
  let dlgExtDetectedFmt = '';
  let dlgExtMaxCols = 0;

  // --- Dialogue Import ---
  let dlgImpBatch = false;
  let dlgImpScript = '';
  let dlgImpTsv = '';
  let dlgImpOutput = '';
  let dlgImpTargetCol = 2;  // default: Lang 2

  // --- RLdev fields ---
  let rldevSelectedOp = 'kprl_disasm';
  let rlSeenFile = '';
  let rlTemplateSeenFile = '';
  let rlOrgFile = '';
  let rlOrgDir = '';                 // batch input folder for compile
  let rlCompileBatch = false;        // batch mode toggle for compile
  let rlOrgTextMode = 'export';
  let rlOrgTextBatch = false;
  let rlOrgTextFile = '';
  let rlOrgTextDir = '';
  let rlOrgTextUtfFile = '';
  let rlOrgTextUtfDir = '';
  let rlKfnFile = '';
  let rlGameexe = '';
  let rlInterpreter = '';           // path to RealLive.exe (PE version)
  let rlTargetVersion = '';
  let rlTargetVersionAuto = false;
  let rlOutputDir = '';
  let rlEncoding = 'UTF-8';
  let rlOutputTransform = 'NONE';
  let rlForceTransform = false;
  let rlGameId = '';
  let rlDebugInfo = false;
  let rlG00File = '';
  let rlG00Dir = '';
  let rlG00Batch = false;
  let rlG00XmlPath = '';
  let rlPngFile = '';
  let rlPngDir = '';
  let rlPngBatch = false;
  let rlPngXmlPath = '';
  let rlG00Format = 'auto';
  let rlGanFile = '';
  let rlNwaFile = '';
  let rlNwaDir = '';
  let rlNwaBatch = false;
  let rlAudioFormat = 'mp3';
  let rlDatFile = '';
  let rlDatDir = '';
  let rlDatBatch = false;
  let rlDatJsonFile = '';
  let rlDatJsonDir = '';
  let rlDatJsonBatch = false;
  let rlSaveMapPath = '';
  let rlSaveFile = '';
  let rlSaveCompareFile = '';
  let rlSaveProfile = 'read_progress';
  let rlSaveRefs = 'seen[1] seen[100] dword[1]';
  let rlSaveAssignments = 'seen[1]=0';
  let rlSaveTextFile = '';
  let rlSaveBuildOutput = '';
  let rlSaveBackup = true;
  let rlSaveLossless = true;
  let rlSaveMapJson = false;
  let rlSaveDumpAll = false;
  let rlSaveDumpJson = false;
  let rlSaveDoctorJson = false;
  let rlSaveDiffJson = false;
  let rlBabelRoot = '';
  let rlBabelGameDir = '';
  let rlBabelVersion = '1.2.3.5';
  let rlBabelDllMode = 'auto';
  let rlBabelNameEnc = 'western';
  let rlBabelUpdateGameexe = true;
  let rlBabelGlosses = false;

  // --- Siglus fields ---
  let siglusSelectedOp = 'scene_extract';
  let siglusGameName = 'Harmonia - Édition physique';
  const siglusGameKeyOptions = [
    "Kamimachi Sana-chan - Version numérique",
    "Kamimachi Sana-chan - Édition physique",
    "Enkou Shoujo: Rikujo-bu Yukki no Baai",
    "Enkou Shoujo 2: JK Idol Marin no Baai",
    "Seiso de Majime na Kanojo ga, Saikyou Yarisaa ni Kanyu Sareta...?",
    "Hatsukoi 1/1",
    "Hatsukoi 1/1 - Memorical Collection",
    "Hoshi Ori Yume Mirai",
    "Hoshi Ori Yume Mirai - Memorical Collection",
    "Hoshi Ori Yume Mirai: Rikka to Anata no 1 Shuunen Kinen, Icha Love Birthday",
    "Hoshi Ori Yume Mirai - Perfect Edition",
    "Gin'iro, Haruka",
    "Gin'iro, Haruka - Memorical Collection",
    "Tsuki no Kanata de Aimashou - Démo",
    "Tsuki no Kanata de Aimashou",
    "Tsuki no Kanata de Aimashou SSR - Démo",
    "Tsuki no Kanata de Aimashou SSR - Version numérique",
    "Tsuki no Kanata de Aimashou SSR - Édition physique",
    "AIR - Version Android",
    "Kanon - Version Android",
    "Rewrite",
    "Rewrite Harvest festa!",
    "Rewrite: Chroniques du club de recherche occulte - Partie 1",
    "Rewrite: Chroniques du club de recherche occulte - Partie 2",
    "Rewrite+",
    "Rewrite+ - Steam anglais",
    "Harmonia - Édition physique",
    "Harmonia - Édition Steam",
    "Angel Beats! -1st beat-",
    "Summer Pockets",
    "Summer Pockets - Version mobile",
    "Summer Pockets - Steam anglais",
    "Summer Pockets REFLECTION BLUE - Version numérique",
    "Summer Pockets REFLECTION BLUE - Édition physique",
    "CLANNAD - Steam chinois",
    "LOOPERS - Démo",
    "LOOPERS - Version numérique",
    "LOOPERS - Version mobile",
    "LUNARiA - Démo",
    "LUNARiA - Version numérique",
    "LUNARiA - Édition physique",
    "LUNARiA - Version mobile",
    "Tsui no Stella - Démo",
    "Tsui no Stella - Version numérique",
    "LAMUNATION!",
    "Yumeiro Alouette!",
    "Mashiro Summer",
    "Noble Riege!",
    "Kisaragi GOLD STAR",
    "Hatsuyuki Sakura",
    "Karumaruka Circle",
    "Hanasaki Work Spring!",
    "Floral Flowlove",
    "Kinkoi: Golden Loveriche",
    "Yuganda Uso no Koi to Label",
    "Hoshiai no Prism Gear",
    "Jimikko Muchimuchi Iinchou to Dosukebe Choukyou Seikatsu",
    "Yuusha to Odore! - Démo",
    "Yuusha to Odore!",
    "LOVE Destination - Démo",
    "LOVE Destination",
    "Avec amour, d'Einstein - Démo",
    "Avec amour, d'Einstein",
    "Avec amour, d'Einstein - APOLLOCRISIS",
    "Imouto Support",
    "Chikan Senyou Sharyou: Mihattatsu na Karada no Tennis Shoujo Hotaru - Version Android",
    "Seishoujo: Kedakaki Goreijou Shion - Version Android",
    "Seishoujo: Geneki Idol Iroha - Version Android",
    "Seishoujo: Yasashiki Joshikousei Yuuri - Version Android",
    "Seishoujo: Volley-bu no Ace Tsubasa - Version Android",
    "Seishoujo: Hitozuma Onna Kyoushi Ryouka - Version Android",
    "KimiBeta: Kimi wo Betabeta ni Sasete Ageru - Version Android",
    "Shuki Shuki Daisuki!! - Version Android",
    "Saya no Uta - Version Android",
    "MareMareMare SP - Version Android",
    "Kanojotachi no Ryuugi - Version Android",
    "planetarian HD - Steam"
  ];
  let siglusCompressionLevel = 17;
  let siglusFakeCompression = false;
  let siglusScenePck = '';
  let siglusSceneOutputDir = '';
  let siglusSceneInputDir = '';
  let siglusSceneWtf = 'auto';
  let siglusSceneOutputPck = '';
  let siglusSSBatch = false;
  let siglusSSInput = '';
  let siglusSSTsv = '';
  let siglusSSOutput = '';
  let siglusSSCopyText = true;
  let siglusSSSingleLine = false;
  let siglusSSFilterMode = 'smart';
  let siglusSSFormat = 'txt';
  let siglusSSSingleXlsx = false;
  let siglusGameexeDat = '';
  let siglusGameexeIni = '';
  let siglusGameexeOutput = '';
  let siglusGameexeDoubleEncryption = false;
  let siglusDBSFile = '';
  let siglusDBSRaw = '';
  let siglusDBSTxt = '';
  let siglusDBSOutput = '';
  let siglusDBSXlsx = '';
  let siglusDBSXlsxDir = '';
  let siglusDBSOutputDir = '';
  let siglusDBSEncoding = 'shift-jis';
  let siglusDBSDumpAll = false;
  let siglusDBSUnicodeOutput = true;
  let siglusMobilePck = '';
  let siglusMobileDir = '';
  let siglusMobileOutput = '';
  let siglusOMVInput = '';
  let siglusOGVOutput = '';
  let siglusOGVInput = '';
  let siglusOMVOutput = '';
  let siglusOMV2AVIInput = '';
  let siglusOMV2AVIOutput = '';
  let siglusOMVPNGInput = '';
  let siglusOMVPNGOutputDir = '';
  let siglusPNGVideoDir = '';
  let siglusPNGVideoOutput = '';
  let siglusPNGVideoAlpha = false;
  let siglusPNGVideoFPS = '30';
  let siglusCombinePNGDir = '';
  let siglusCombinePNGOutput = '';
  let siglusScriptRepackScript = '';
  let siglusScriptRepackText = '';
  let siglusScriptRepackOutput = '';
  let siglusG00File = '';
  let siglusG00Dir = '';
  let siglusG00Batch = false;
  let siglusG00OutputDir = '';
  let siglusG00XmlPath = '';
  let siglusPngFile = '';
  let siglusPngDir = '';
  let siglusPngBatch = false;
  let siglusPngOutputDir = '';
  let siglusPngXmlPath = '';
  let siglusG00Format = 'auto';

  // --- RLdev operations ---
  // Numbered to reflect the natural translation workflow:
  //   1. List the archive to see what's inside
  //   2. Extract scripts to .org / .utf for translation
  //   3. Compile each translated .org back to .TXT
  //   4. Pack the .TXT files back into a SEEN.txt archive
  const rldevOperations = [
    { id: '_rs1', label: 'KPRL / RLC', section: true },
    { id: 'kprl_list',    label: '1 — List SEEN.txt archive' },
    { id: 'kprl_disasm',  label: '2 — Extract SEEN.txt' },
    { id: 'rlc_compile',  label: '3 — Compile .org / .ke / .avg' },
    { id: 'rlc_org_text', label: 'Extract text ORG' },
    { id: 'kprl_archive', label: '4 — Rebuild SEEN.txt' },
    { id: 'kprl_extract', label: 'Advanced: extract bytecode' },
    { id: '_rs3', label: 'IMAGE (G00)', section: true },
    { id: 'g00_extract', label: 'G00 → PNG' },
    { id: 'g00_import', label: 'PNG → G00' },
    { id: '_rs4', label: 'ANIMATION (GAN)', section: true },
    { id: 'gan_to_xml', label: 'GAN → XML' },
    { id: 'gan_from_xml', label: 'XML → GAN' },
    { id: '_rs5', label: 'AUDIO (BGM)', section: true },
    { id: 'nwa_audio', label: 'NWA → MP3/WAV' },
    { id: '_rs6', label: 'DAT (CG/TCC)', section: true },
    { id: 'dat_to_json', label: 'CGM/TCC → JSON' },
    { id: 'dat_from_json', label: 'JSON → CGM/TCC' },
    { id: '_rs_save', label: 'SAVE', section: true },
    { id: 'save_editor', label: 'RealLive save editor' },
    { id: '_rs7', label: 'BABEL', section: true },
    { id: 'babel_runtime', label: 'Runtime setup' },
    { id: 'babel_header', label: 'global.kh helper' },
  ];

  const siglusOperations = [
    { id: '_sg1', label: 'SCENE.PCK', section: true },
    { id: 'scene_extract', label: 'Extract Scene.pck' },
    { id: 'scene_rebuild', label: 'Rebuild Scene.pck' },
    { id: '_sg2', label: 'TEXT (.SS)', section: true },
    { id: 'ss_dump', label: 'Dump SS text' },
    { id: 'ss_inject', label: 'Inject SS text' },
    { id: '_sg3', label: 'GAMEEXE.DAT', section: true },
    { id: 'gameexe_extract', label: 'Decrypt Gameexe' },
    { id: 'gameexe_rebuild', label: 'Rebuild Gameexe' },
    { id: '_sg4', label: 'DBS', section: true },
    { id: 'dbs_extract', label: 'Dump DBS' },
    { id: 'dbs_rebuild', label: 'Rebuild DBS' },
    { id: 'dbs_xlsx', label: 'Dump DBS XLSX' },
    { id: 'dbs_build', label: 'Build DBS XLSX' },
    { id: '_sg5', label: 'MOBILE / OMV', section: true },
    { id: 'mobile_pck_extract', label: 'Extract mobile PCK' },
    { id: 'mobile_pck_rebuild', label: 'Rebuild mobile PCK' },
    { id: 'omv_cut', label: 'Cut OMV header' },
    { id: 'omv_pack', label: 'Pack OMV' },
    { id: 'omv2avi', label: 'Omv2Avi' },
    { id: 'omv_png', label: 'OMV → PNG sequence' },
    { id: 'png_video', label: 'PNG sequence → OMV/OGV' },
    { id: '_sg6', label: 'ANNEXES', section: true },
    { id: 'g00_extract', label: 'G00 → PNG (vaconv)' },
    { id: 'g00_import', label: 'PNG → G00 (vaconv)' },
    { id: 'combine_png', label: 'Combine PNG' },
    { id: 'script_repack', label: 'Script Repacker' },
  ];

  const gameIdOptions = [
    { id: 'CFV', title: 'Clannad Full Voice' },
    { id: 'LB', title: 'Little Busters!' },
    { id: 'LBEX', title: 'Little Busters! EX' },
    { id: 'LBME', title: 'Little Busters! Memorial Edition' },
    { id: 'LBPE', title: 'Little Busters! PE' },
    { id: 'FIVE', title: '5 -Faibu-' },
    { id: 'SNOW', title: 'Snow Standard Edition' },
    { id: 'KUDO', title: 'Kud Wafter 18+' },
    { id: 'KUDA', title: 'Kud Wafter all-ages' },
    { id: 'PLHD', title: 'Planetarian HD' },
    { id: 'TMPE', title: 'Tomoyo After PE / Memorial Edition' },
    { id: 'ONIU', title: 'Oni Uta' },
    { id: 'ONIUTA', title: 'Oni Uta' },
    { id: 'PING', title: '3P LOVERS' },
    { id: 'KOYO', title: 'Nizuma Koyomi' },
    { id: 'SHINO', title: 'Nizuma Shino' },
    { id: 'TAMA', title: 'Nizuma Tamaki' },
    { id: 'PRIP', title: 'Princess Heart Link package edition' },
    { id: 'PRID', title: 'Princess Heart Link DL edition' },
    { id: 'HINA', title: 'Hinasawa Tomoka no Zettai Joousei' },
    { id: 'LUV', title: 'Lovedori Halation' }
  ];

  const saveProfiles = [
    {
      id: 'read_progress',
      label: 'read.sav progression',
      refs: 'seen[1] seen[100] seen[1000]',
      assignments: 'seen[1]=0'
    },
    {
      id: 'global_flags',
      label: 'save999 intG flags',
      refs: 'intG[0] intG[1] intG[30] intG[31]',
      assignments: 'intG[1]=0'
    },
    {
      id: 'raw_dwords',
      label: 'low-level dwords',
      refs: 'dword[0] dword[1] dword[2]',
      assignments: 'dword[1]=0'
    }
  ];

  // ===== Operations list =====
  const operations = [
    { id: '_s1', label: 'SCRIPT', section: true },
    { id: 'decompile', label: 'Script Decompile' },
    { id: 'compile', label: 'Script Compile' },
    { id: 'siglus_luca', label: 'Siglus -> Luca' },
    { id: '_s2', label: 'PAK (CG)', section: true },
    { id: 'pak_cg_extract', label: 'CG Extract' },
    { id: 'pak_cg_replace', label: 'CG Replace' },
    { id: '_s2v', label: 'PAK (Video)', section: true },
    { id: 'bgmovie_extract', label: 'BGMOVIE Extract' },
    { id: '_s2a', label: 'PAK (Audio)', section: true },
    { id: 'music_extract', label: 'Music Extract' },
    { id: 'voice_extract', label: 'Voice Extract' },
    { id: 'audio_convert', label: 'Ogg / MP3 Convert' },
    { id: '_s2b', label: 'PAK (Font)', section: true },
    { id: 'pak_font_extract', label: 'Font Extract' },
    { id: 'pak_font_replace', label: 'Font Replace' },
    { id: '_s3', label: 'FONT', section: true },
    { id: 'font_extract', label: 'Font Extract' },
    { id: 'font_edit', label: 'Font Edit' },
    { id: '_s3b', label: 'VIET FONT', section: true },
    { id: 'viet_font_patch', label: 'AIR / SG Patch' },
    { id: '_s3c', label: 'DLL HOOK', section: true },
    { id: 'luca_menu_dll', label: 'Luca Menu DLL' },
    { id: '_s4', label: 'IMAGE', section: true },
    { id: 'image_export', label: 'Image Export' },
    { id: 'image_import', label: 'Image Import' },
    { id: '_s5', label: '', section: true },
    { id: 'about', label: 'À propos' },
  ];

  // --- Outils divers operations ---
  const outilsOperations = [
    { id: '_os1', label: 'DIALOGUE (LuckEngine)', section: true },
    { id: 'dlg_extract', label: 'Extract Dialogues' },
    { id: 'dlg_import', label: 'Import Dialogues' },
    { id: '_os2', label: 'IMAGE (Vaconv)', section: true },
    { id: 'vaconv_g00', label: 'G00 → PNG (batch)' },
    { id: 'vaconv_png', label: 'PNG → G00 (batch)' },
  ];

  // ===== Console =====
  // Batched console updates for performance (flush every 80ms instead of per-line)
  let pendingLines = [];
  let flushTimer = null;

  function addLine(text) {
    let cls = '';
    if (text.includes('[OK]')) cls = 'line-ok';
    else if (text.includes('[ERROR]') || text.includes('Panic') || text.includes('Error')) cls = 'line-err';
    else if (text.includes('Warning')) cls = 'line-warn';
    else if (text.startsWith('═') || text.startsWith('─')) cls = 'line-sep';
    else if (text.startsWith('>')) cls = 'line-cmd';
    pendingLines.push({ text, cls });
    if (!flushTimer) {
      flushTimer = setTimeout(flushConsole, 80);
    }
  }
  function flushConsole() {
    if (pendingLines.length > 0) {
      consoleLines = [...consoleLines, ...pendingLines];
      pendingLines = [];
      if (consoleLines.length > 12000) consoleLines = consoleLines.slice(-10000);
      tick().then(() => { if (consoleEl) consoleEl.scrollTop = consoleEl.scrollHeight; });
    }
    flushTimer = null;
  }
  function clearConsole() { consoleLines = []; pendingLines = []; }

  function selectedConsoleText() {
    const selection = window.getSelection();
    const text = selection ? selection.toString() : '';
    return text.trim() ? text : consoleLines.map(line => line.text).join('\n');
  }

  async function writeClipboardText(text) {
    if (!text) return;
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
        return;
      }
    } catch (_) {
      // Fallback below for WebView clipboard restrictions.
    }
    const area = document.createElement('textarea');
    area.value = text;
    area.style.position = 'fixed';
    area.style.left = '-9999px';
    document.body.appendChild(area);
    area.select();
    document.execCommand('copy');
    document.body.removeChild(area);
  }

  async function readClipboardText() {
    try {
      if (navigator.clipboard?.readText) {
        return await navigator.clipboard.readText();
      }
    } catch (_) {
      return '';
    }
    return '';
  }

  function openConsoleMenu(event) {
    event.preventDefault();
    consoleMenuX = event.clientX;
    consoleMenuY = event.clientY;
    consoleMenuVisible = true;
  }

  function closeConsoleMenu() {
    consoleMenuVisible = false;
  }

  async function copyConsoleSelection() {
    await writeClipboardText(selectedConsoleText());
    closeConsoleMenu();
  }

  async function copyConsoleAll() {
    await writeClipboardText(consoleLines.map(line => line.text).join('\n'));
    closeConsoleMenu();
  }

  async function pasteConsoleClipboard() {
    const text = await readClipboardText();
    if (text.trim()) {
      text.replace(/\r\n/g, '\n').replace(/\r/g, '\n').split('\n').forEach(line => addLine(line));
    }
    closeConsoleMenu();
  }

  onMount(async () => {
    try { setUiLanguage(localStorage.getItem('lucksystem-ui-language') || 'fr'); } catch (_) {}
    window.addEventListener('click', closeConsoleMenu);
    window.addEventListener('keydown', closeConsoleMenu);
    EventsOn('log', (msg) => addLine(msg));
    lsPath = await GetLuckSystemPath();
    lucaMenuDllAvailable = await SupportsLucaMenuDLL();
    if (lsPath) {
      addLine('LuckSystem 2.3.2 - Yoremi fork v3.30 GUI');
      addLine('Executable: ' + lsPath);
      // Scan data/ folder for game presets
      gamePresets = (await ScanGameData()) || [];
      if (gamePresets.length > 0) {
        addLine('Found ' + gamePresets.length + ' game preset(s): ' + gamePresets.map(p => p.name).join(', '));
      }
    } else {
      addLine('[ERROR] lucksystem.exe not found!');
      addLine('Place lucksystem.exe next to the GUI, or click "Locate" below.');
    }
    addLine('RLdev 2026 - Go édition v1.3.5');
    addLine('Ready.');
    const kfn = await DefaultKFN();
    if (kfn && !rlKfnFile) {
      rlKfnFile = kfn;
      addLine('KFN détecté : ' + kfn);
    }
    const babel = await DefaultBabelRoot();
    if (babel && !rlBabelRoot) {
      rlBabelRoot = babel;
      addLine('BABEL détecté : ' + babel);
    }
  });
  onDestroy(() => {
    window.removeEventListener('click', closeConsoleMenu);
    window.removeEventListener('keydown', closeConsoleMenu);
    EventsOff('log');
    stopConsoleResize();
  });

  function startConsoleResize(event) {
    event.preventDefault();
    consoleResizing = true;
    consoleResizeStartY = event.clientY;
    consoleResizeStartHeight = consoleHeight;
    window.addEventListener('mousemove', resizeConsole);
    window.addEventListener('mouseup', stopConsoleResize);
  }
  function resizeConsole(event) {
    if (!consoleResizing) return;
    const maxHeight = Math.max(180, window.innerHeight - 130);
    const nextHeight = consoleResizeStartHeight + (consoleResizeStartY - event.clientY);
    consoleHeight = Math.min(maxHeight, Math.max(96, nextHeight));
  }
  function stopConsoleResize() {
    if (!consoleResizing) return;
    consoleResizing = false;
    window.removeEventListener('mousemove', resizeConsole);
    window.removeEventListener('mouseup', stopConsoleResize);
  }

  // ===== Browse helpers =====
  async function browsePak() { const f = await SelectPakFile(); if (f) pakFile = f; }
  async function browseOpcode() { const f = await SelectFile('Select Opcode (.txt)', '*.txt', 'Opcode files'); if (f) { opcodeFile = f; selectedPreset = ''; } }
  async function browsePlugin() { const f = await SelectFile('Select Plugin (.py)', '*.py', 'Python plugins'); if (f) { pluginFile = f; selectedPreset = ''; } }
  async function browseOutputDir() { const d = await SelectDirectory('Select output directory'); if (d) outputDir = d; }
  async function browseImportDir() { const d = await SelectDirectory('Select translated scripts directory'); if (d) importDir = d; }
  async function browseOutputPak() { const f = await SelectSaveFile('Save output PAK', 'SCRIPT_FR.PAK', '*.PAK;*.pak', 'PAK files'); if (f) outputPak = f; }

  async function browseSiglusLucaLucaDir() { const d = await SelectDirectory('Select Luca decompiled scripts folder'); if (d) siglusLucaLucaDir = d; }
  async function browseSiglusLucaSiglusDir() { const d = await SelectDirectory('Select Siglus Full folder'); if (d) siglusLucaSiglusDir = d; }
  async function browseSiglusLucaOutput() { const d = await SelectDirectory('Select patched Luca output folder'); if (d) siglusLucaOutput = d; }

  function applyPreset(presetName) {
    selectedPreset = presetName;
    if (!presetName) { opcodeFile = ''; pluginFile = ''; gameName = ''; return; }
    const p = gamePresets.find(g => g.name === presetName);
    if (p) { opcodeFile = p.opcodeFile; pluginFile = p.pluginFile || ''; gameName = p.gameFlag || ''; }
  }

  async function browsePakExtSource() { const f = await SelectPakFile(); if (f) pakExtSource = f; }
  async function browsePakExtOutput() { const d = await SelectDirectory('Select extraction output'); if (d) pakExtOutput = d; }
  async function browsePakRepSource() { const f = await SelectPakFile(); if (f) pakRepSource = f; }
  async function browsePakRepListFile() { const f = await SelectFile('Sélectionner le fichier liste (_list.txt)', '*.txt', 'Fichiers liste'); if (f) pakRepListFile = f; }
  async function browsePakRepInput() { const d = await SelectDirectory('Select folder with modified files'); if (d) pakRepInput = d; }
  async function browsePakRepOutput() { const f = await SelectSaveFile('Save output PAK', 'FONT.out.PAK', '*.PAK;*.pak', 'PAK files'); if (f) pakRepOutput = f; }

  async function browseBgMoviePak() { const f = await SelectPakFile(); if (f) bgMoviePak = f; }
  async function browseBgMovieOutput() { const d = await SelectDirectory('Select BGMOVIE output folder'); if (d) bgMovieOutput = d; }

  async function browseMusicPak() { const f = await SelectPakFile(); if (f) musicPak = f; }
  async function browseMusicOutput() { const d = await SelectDirectory('Select MUSIC output folder'); if (d) musicOutput = d; }
  async function browseVoicePak() { const f = await SelectPakFile(); if (f) voicePak = f; }
  async function browseVoiceOutput() { const d = await SelectDirectory('Select VOICE output folder'); if (d) voiceOutput = d; }
  async function browseAudioConvInput() {
    const d = await SelectDirectory(audioConvDirection === 'mp3' ? 'Select native Ogg folder' : 'Select MP3 folder');
    if (d) audioConvInput = d;
  }
  async function browseAudioConvOutput() { const d = await SelectDirectory('Select converted audio output folder'); if (d) audioConvOutput = d; }

  async function browsePakFontExtSource() { const f = await SelectPakFile(); if (f) pakFontExtSource = f; }
  async function browsePakFontExtOutput() { const d = await SelectDirectory('Dossier d\'extraction'); if (d) pakFontExtOutput = d; }
  async function browsePakFontRepSource() { const f = await SelectPakFile(); if (f) pakFontRepSource = f; }
  async function browsePakFontRepListFile() { const f = await SelectFile('Sélectionner le fichier liste (_list.txt)', '*.txt', 'Fichiers liste'); if (f) pakFontRepListFile = f; }
  async function browsePakFontRepInput() { const d = await SelectDirectory('Dossier des fichiers modifiés'); if (d) pakFontRepInput = d; }
  async function browsePakFontRepSingleFile() {
    const f = await SelectFile('Sélectionner le fichier à remplacer', '*.*', 'Tous les fichiers');
    if (f) {
      pakFontRepSingleFile = f;
      if (!pakFontRepSingleName) pakFontRepSingleName = f.split(/[\\/]/).pop();
    }
  }
  async function browsePakFontRepOutput() { const f = await SelectSaveFile('Save output PAK', 'FONT.out.PAK', '*.PAK;*.pak', 'PAK files'); if (f) pakFontRepOutput = f; }

  async function browseFontExtCz() { const f = await SelectFile('Select font CZ file', '*.*', 'Font CZ files'); if (f) fontExtCz = f; }
  async function browseFontExtInfo() { const f = await SelectFile('Select info file', '*.*', 'Info files'); if (f) fontExtInfo = f; }
  async function browseFontExtPng() { const f = await SelectSaveFile('Save font PNG', 'font.png', '*.png', 'PNG Images'); if (f) fontExtPng = f; }
  async function browseFontExtCharset() { const f = await SelectSaveFile('Save charset TXT', 'charset.txt', '*.txt', 'Text files'); if (f) fontExtCharset = f; }
  async function browseFontEditCz() { const f = await SelectFile('Select source CZ', '*.*', 'Font CZ files'); if (f) fontEditCz = f; }
  async function browseFontEditInfo() { const f = await SelectFile('Select source info', '*.*', 'Info files'); if (f) fontEditInfo = f; }
  async function browseFontEditTtf() { const f = await SelectFile('Select TTF font', '*.ttf;*.otf', 'Font files'); if (f) fontEditTtf = f; }
  async function browseFontEditOutCz() { const d = await SelectDirectory('Dossier de sortie pour le CZ modifié'); if (d) fontEditOutCz = d + '\\'; }
  async function browseFontEditOutInfo() { const d = await SelectDirectory('Dossier de sortie pour le fichier info'); if (d) fontEditOutInfo = d + '\\'; }
  async function browseFontEditCharset() { const f = await SelectFile('Select charset file', '*.txt', 'Text files'); if (f) fontEditCharsetFile = f; }

  async function browseVietFontRoot() { const d = await SelectDirectory('Select AIR / Planetarian SG files folder'); if (d) vietFontRoot = d; }
  async function browseVietCharset() { const f = await SelectFile('Select full Vietnamese charset (134 chars)', '*.txt', 'Text files'); if (f) vietCharsetFile = f; }
  async function browseVietTtf() { const f = await SelectFile('Select Vietnamese-capable TTF/OTF', '*.ttf;*.otf', 'Font files'); if (f) vietTtfFile = f; }
  async function browseVietOutput() { const d = await SelectDirectory('Select output folder'); if (d) vietOutputDir = d; }

  async function browseImgExpInput() {
    if (imgExpBatch) { const d = await SelectDirectory('Select CZ folder'); if (d) imgExpInput = d; }
    else { const f = await SelectFile('Select CZ file', '*.*', 'CZ image files'); if (f) imgExpInput = f; }
  }
  async function browseImgExpOutput() {
    if (imgExpBatch) { const d = await SelectDirectory('Select PNG output folder'); if (d) imgExpOutput = d; }
    else { const f = await SelectSaveFile('Save PNG', 'output.png', '*.png', 'PNG Images'); if (f) imgExpOutput = f; }
  }
  async function browseImgImpSource() {
    if (imgImpBatch) { const d = await SelectDirectory('Select original CZ folder'); if (d) imgImpSource = d; }
    else { const f = await SelectFile('Select original CZ file', '*.*', 'CZ files'); if (f) imgImpSource = f; }
  }
  async function browseImgImpInput() {
    if (imgImpBatch) { const d = await SelectDirectory('Select PNG folder'); if (d) imgImpInput = d; }
    else { const f = await SelectFile('Select PNG file', '*.png', 'PNG Images'); if (f) imgImpInput = f; }
  }
  async function browseImgImpOutput() {
    if (imgImpBatch) { const d = await SelectDirectory('Select output CZ folder'); if (d) imgImpOutput = d; }
    else { const f = await SelectSaveFile('Save CZ', 'output.cz', '*.*', 'All files'); if (f) imgImpOutput = f; }
  }

  async function locateLuckSystem() {
    lsPath = await SetLuckSystemPath();
    if (lsPath) {
      addLine('Executable set: ' + lsPath);
      gamePresets = (await ScanGameData()) || [];
      if (gamePresets.length > 0) addLine('Found ' + gamePresets.length + ' game preset(s): ' + gamePresets.map(p => p.name).join(', '));
    }
  }

  async function stopProcess() {
    await StopProcess();
  }

  // ===== Actions =====
  async function run(fn) {
    if (running) return;
    running = true;
    try { await fn(); } catch (e) { addLine('[ERROR] ' + e); }
    running = false;
  }

  function startDecompile() { run(() => ScriptDecompile(pakFile, opcodeFile, pluginFile, charsetVal, outputDir, gameName)); }
  function startCompile() { run(() => ScriptCompile(pakFile, opcodeFile, pluginFile, charsetVal, importDir, outputPak, gameName)); }
  function startSiglusLucaBridge() { run(() => SiglusLucaBridge(siglusLucaLucaDir, siglusLucaSiglusDir, siglusLucaOutput, siglusLucaTargetCol)); }
  function startPakExtract() { run(() => PakExtract(pakExtSource, pakExtOutput)); }
  function startBgMovieExtract() { run(() => BGMOVIEExtract(bgMoviePak, bgMovieOutput)); }
  function startMusicExtract() { run(() => MusicPakExtract(musicPak, musicOutput, musicToMp3)); }
  function startVoiceExtract() { run(() => VoicePakExtract(voicePak, voiceOutput, voiceToMp3)); }
  function startAudioConvert() { run(() => AudioConvert(audioConvInput, audioConvOutput, audioConvDirection)); }
  function startPakReplace() {
    const listArg = pakRepUseList ? pakRepListFile : '';
    const dirArg  = pakRepUseList ? '' : pakRepInput;
    run(() => PakReplace(pakRepSource, dirArg, listArg, pakRepOutput));
  }
  function startPakFontExtract() { run(() => PakFontExtract(pakFontExtSource, pakFontExtCharset, pakFontExtOutput)); }
  function startPakFontReplace() {
    const listArg = pakFontRepMode === 'list' ? pakFontRepListFile : '';
    const dirArg  = pakFontRepMode === 'dir' ? pakFontRepInput : '';
    const fileArg = pakFontRepMode === 'single' ? pakFontRepSingleFile : '';
    const nameArg = pakFontRepMode === 'single' ? pakFontRepSingleName : '';
    const aliasFromArg = pakFontRepMode === 'alias' ? pakFontRepAliasFrom : '';
    const aliasToArg = pakFontRepMode === 'alias' ? pakFontRepAliasTo : '';
    run(() => PakFontReplace(pakFontRepSource, pakFontRepCharset, dirArg, listArg, fileArg, nameArg, aliasFromArg, aliasToArg, pakFontRepOutput));
  }
  function startFontExtract() { run(() => FontExtract(fontExtCz, fontExtInfo, fontExtPng, fontExtCharset)); }
  function setFontEditArabicPreset(checked) {
    fontEditArabicMetrics = checked;
  }
  function startFontEdit() {
    const redraw  = fontEditMode === 'redraw';
    const append  = fontEditMode === 'append';
    const index   = (fontEditMode === 'insert') ? fontEditIndex : 0;
    run(() => FontEdit(
      fontEditCz, fontEditInfo, fontEditTtf, fontEditOutCz, fontEditOutInfo, fontEditCharsetFile,
      redraw, append, index,
      fontEditArabicMetrics, fontEditMetricSetYEnabled, fontEditMetricSetY,
      fontEditMetricYOffset, fontEditMetricXOffset, fontEditMetricWOffset,
      fontEditArabicConnectorBleed
    ));
  }

  function getVietYOffsets() {
    const values = [];
    if (vietYMinus2) values.push(-2);
    if (vietYMinus1) values.push(-1);
    if (vietY0) values.push(0);
    if (vietY1) values.push(1);
    if (vietY2) values.push(2);
    if (vietY3) values.push(3);
    return values;
  }
  function startVietnameseFontPatch() {
    run(() => VietnameseFontPatch(vietFontRoot, vietCharsetFile, vietTtfFile, vietOutputDir, vietSlot, vietFamily, getVietYOffsets(), vietRedrawLatin));
  }

  // --- Luca Menu DLL helpers ---
  function lucaProfiles() { return lucaInventory?.profiles || []; }
  function currentLucaProfile() { return lucaProfiles().find(p => p.id === lucaGame) || null; }
  function lucaProxyName() { return lucaProxyChoice; }
  function lucaGameProfiles() { return lucaProfiles().filter(p => !p.id.includes('/')); }

  function lucaFamilyProfiles() {
    const profile = currentLucaProfile();
    if (!profile) return [];
    const root = (profile.id || '').split('/')[0];
    return lucaProfiles().filter(p => p.id === root || p.id.startsWith(root + '/'));
  }

  function lucaSlotEntryCount(profile, slot = lucaSlot) {
    return (profile?.entries || []).filter(e => e.slot === slot).length;
  }

  function lucaSlotSourceProfile(slot = lucaSlot) {
    const current = currentLucaProfile();
    return lucaFamilyProfiles()
      .filter(p => lucaSlotEntryCount(p, slot) > 0)
      .sort((a, b) => {
        const expected = slot === 'jp' ? '/slots-jp' : slot === 'cn' ? '/slots-cn' : '';
        const aPreferred = expected && a.id.toLowerCase().endsWith(expected) ? 1 : 0;
        const bPreferred = expected && b.id.toLowerCase().endsWith(expected) ? 1 : 0;
        if (aPreferred !== bPreferred) return bPreferred - aPreferred;
        const countDiff = lucaSlotEntryCount(b, slot) - lucaSlotEntryCount(a, slot);
        if (countDiff) return countDiff;
        if (a.id === current?.id) return -1;
        if (b.id === current?.id) return 1;
        return a.id.length - b.id.length || a.id.localeCompare(b.id);
      })[0] || null;
  }

  function lucaAvailableSlotCount(slot) { return lucaSlotEntryCount(lucaSlotSourceProfile(slot), slot); }
  function lucaSafeFrenchCount(slot = lucaSlot) {
    const source = lucaSlotSourceProfile(slot);
    return (source?.entries || []).filter(e => e.slot === slot && e.safeAuto && e.suggestedFr).length;
  }
  function slotLabel(slot) {
    if (slot === 'jp') return t('Japonais', 'Japanese');
    if (slot === 'cn') return t('Chinois', 'Chinese');
    return t('Anglais', 'English');
  }
  function byteLen(value) { return new TextEncoder().encode(value || '').length; }
  function lucaEncodedLen(entry, value = entry?.target || '') {
    return entry?.encoding === 'utf-16-le' ? String(value).length * 2 : byteLen(value);
  }
  function entryTooLong(entry) { return entry.budget >= 0 && lucaEncodedLen(entry) > entry.budget; }

  async function loadLucaInventory() {
    lucaInventory = await ScanLucaMenuKit();
    const profiles = lucaProfiles();
    if (!profiles.length) {
      addLine('[ERROR] No Luca DLL patch profiles found.');
      return;
    }
    if (!lucaGame || !profiles.find(p => p.id === lucaGame)) {
      const preferred = profiles.find(p => p.id === 'Kanon') || profiles.find(p => p.id === 'AIR') || profiles[0];
      lucaGame = preferred.id;
    }
    syncLucaProfileDefaults();
    refreshLucaEntries();
    addLine('Luca kit: ' + lucaInventory.kitDir);
    addLine('Luca games: ' + lucaGameProfiles().map(p => p.id).join(', '));
  }

  function syncLucaProfileDefaults() {
    const profile = currentLucaProfile();
    if (!profile) return;
    lucaPatchName = profile.patchGameName || profile.name || lucaGame;
    lucaPatchVersion = profile.patchVersion || '0.1-gui';
    const identity = [profile.id, profile.name, profile.patchGameName, profile.gameExe, lucaGame]
      .filter(Boolean).join(' ').toUpperCase();
    lucaProxyChoice = identity.includes('LBEE') || identity.includes('LITTLE BUSTERS') || identity.includes('LITBUS_WIN32')
      ? 'winmm'
      : (profile.proxyDll === 'winmm' ? 'winmm' : 'version');
  }

  function selectLucaProxy(choice, checked) {
    if (checked) {
      lucaProxyChoice = choice;
      lucaBuildDll = true;
    } else if (lucaProxyChoice === choice) {
      lucaBuildDll = false;
    }
  }

  function refreshLucaEntries(mode = '') {
    const profile = lucaSlotSourceProfile();
    if (!profile) { lucaEntries = []; return; }
    lucaEntries = profile.entries
      .filter(e => e.slot === lucaSlot)
      .map(e => ({ ...e, target: e.target || '', include: false }));
    applyLucaFillMode(mode || lucaFillMode);
  }

  function setLucaGame(value) {
    lucaGame = value;
    lucaExe = '';
    lucaBuildDll = true;
    lucaCustomPatch = '';
    syncLucaProfileDefaults();
    refreshLucaEntries();
  }
  function setLucaSlot(value) { lucaSlot = value; refreshLucaEntries(); }

  function lucaPresetTarget(entry, mode) {
    if (mode === 'fr' || mode === 'fr-safe') return entry.suggestedFr || '';
    if (mode === 'en' || mode === 'en-safe') return entry.suggestedEn || '';
    if (mode === 'ar') return entry.suggestedAr || '';
    if (mode === 'ru') return entry.suggestedRu || '';
    if (mode === 'jp') return entry.suggestedJp || '';
    if (mode === 'cn') return entry.suggestedCn || '';
    return entry.target || '';
  }

  function applyLucaFillMode(mode = lucaFillMode) {
    lucaFillMode = mode;
    const safeMode = mode === 'fr-safe' || mode === 'en-safe';
    lucaEntries = lucaEntries.map(entry => {
      const target = lucaPresetTarget(entry, mode);
      const fits = entry.budget < 0 || lucaEncodedLen(entry, target) <= entry.budget;
      const safe = entry.commonCount >= 4 && (mode !== 'fr-safe' || entry.safeAuto);
      const include = target !== '' && target !== entry.source && fits && (!safeMode || safe);
      return { ...entry, target, include };
    });
  }
  function setLucaFillMode(value) {
    if (value === 'ru') { lucaProxyChoice = 'winmm'; lucaBuildDll = true; }
    applyLucaFillMode(value);
  }
  function clearLucaTargets() { lucaEntries = lucaEntries.map(e => ({ ...e, target: '', include: false })); }
  function lucaVisibleEntries() {
    const q = lucaSearch.trim().toLowerCase();
    return lucaEntries.filter(e => !q || [e.source, e.target, e.context, e.note, e.category, e.textKind]
      .some(v => (v || '').toLowerCase().includes(q)));
  }
  function lucaSelectedEntries() { return lucaEntries.filter(e => e.include && e.target); }

  async function browseLucaExe() {
    const f = await SelectFile(t("Sélectionner l'EXE du jeu Luca", 'Select Luca game EXE'), '*.exe', t('Fichiers exécutables', 'Executable files'));
    if (f) lucaExe = f;
  }
  async function browseLucaOutput() {
    const d = await SelectDirectory(t('Sélectionner le dossier de sortie DLL Luca', 'Select Luca DLL output folder'));
    if (d) lucaOutputDir = d;
  }
  async function browseLucaCustomPatch() {
    const f = await SelectFile(t('Sélectionner un fichier PATCHES personnalisé', 'Select custom PATCHES file'), '*.py', t('Fichiers Python PATCHES', 'Python PATCHES files'));
    if (f) { lucaCustomPatch = f; lucaProxyChoice = 'winmm'; lucaBuildDll = true; }
  }

  function startLucaGenerate() {
    const entries = lucaSelectedEntries().map(e => ({
      rawOffset: e.rawOffset, source: e.source, target: e.target,
      context: e.context, note: e.note, encoding: e.encoding,
      include: e.include, budget: e.budget
    }));
    if (!entries.length && !lucaCustomPatch) {
      addLine(t('[ERROR] Aucune chaîne remplie et sélectionnée pour le slot ', '[ERROR] No filled and selected string for the ') + slotLabel(lucaSlot) + t('.', ' slot.'));
      return;
    }
    run(() => LucaMenuGenerate({
      profileId: lucaGame, gameExe: lucaExe, outputDir: lucaOutputDir,
      patchGameName: lucaPatchName, patchVersion: lucaPatchVersion,
      slot: lucaSlot, buildDll: lucaBuildDll, proxyDll: lucaProxyChoice,
      preset: lucaFillMode, customPatch: lucaCustomPatch, entries
    }));
  }

  function startImageExport() {
    if (imgExpBatch) run(() => ImageBatchExport(imgExpInput, imgExpOutput));
    else run(() => ImageExport(imgExpInput, imgExpOutput));
  }
  function startImageImport() {
    if (imgImpBatch) run(() => ImageBatchImport(imgImpSource, imgImpInput, imgImpOutput, imgImpFill));
    else run(() => ImageImport(imgImpSource, imgImpInput, imgImpOutput, imgImpFill));
  }

  function selectOp(op) {
    if (op.disabled || op.section) return;
    selectedOp = op.id;
    if (selectedOp === 'luca_menu_dll' && !lucaInventory) loadLucaInventory();
  }

  // Reset fields when switching batch mode
  function toggleExpBatch() { imgExpInput = ''; imgExpOutput = ''; }
  function toggleImpBatch() { imgImpSource = ''; imgImpInput = ''; imgImpOutput = ''; }

  // --- Dialogue helpers ---
  async function browseDlgExtInput() {
    if (dlgExtBatch) { const d = await SelectDirectory('Select scripts folder'); if (d) dlgExtInput = d; }
    else { const f = await SelectScriptTxtFile(); if (f) dlgExtInput = f; }
    if (dlgExtInput) await detectDlgFormat();
  }
  async function browseDlgExtOutput() {
    if (dlgExtBatch) { const d = await SelectDirectory('Select output folder'); if (d) dlgExtOutput = d; }
    else {
      const defName = dlgExtInput ? dlgExtInput.replace(/\.txt$/i, '.ext.txt').split(/[\\/]/).pop() : 'dialogues.ext.txt';
      const f = await SelectSaveTsvFile(defName);
      if (f) dlgExtOutput = f;
    }
  }
  async function detectDlgFormat() {
    if (!dlgExtInput) return;
    const target = dlgExtBatch ? '' : dlgExtInput;
    if (!target) return;
    const info = await DialogueDetectFormat(target);
    dlgExtDetectedFmt = info.format || 'Unknown';
    dlgExtMaxCols = info.maxCols || 0;
  }
  function toggleDlgExtBatch() { dlgExtInput = ''; dlgExtOutput = ''; dlgExtDetectedFmt = ''; dlgExtMaxCols = 0; }

  async function browseDlgImpScript() {
    if (dlgImpBatch) { const d = await SelectDirectory('Select original scripts folder'); if (d) dlgImpScript = d; }
    else { const f = await SelectScriptTxtFile(); if (f) dlgImpScript = f; }
  }
  async function browseDlgImpTsv() {
    if (dlgImpBatch) { const d = await SelectDirectory('Select TSV folder'); if (d) dlgImpTsv = d; }
    else { const f = await SelectTsvFile(); if (f) dlgImpTsv = f; }
  }
  async function browseDlgImpOutput() {
    if (dlgImpBatch) { const d = await SelectDirectory('Select output folder'); if (d) dlgImpOutput = d; }
    else {
      const defName = dlgImpScript ? dlgImpScript.split(/[\\/]/).pop() : 'patched.txt';
      const f = await SelectSaveScriptFile(defName);
      if (f) dlgImpOutput = f;
    }
  }
  function toggleDlgImpBatch() { dlgImpScript = ''; dlgImpTsv = ''; dlgImpOutput = ''; }

  function getDlgExtCols() {
    const cols = [];
    if (dlgExtLang1) cols.push(1);
    if (dlgExtLang2) cols.push(2);
    if (dlgExtLang3) cols.push(3);
    if (dlgExtLang4) cols.push(4);
    return cols;
  }

  function startDlgExtract() {
    const cols = getDlgExtCols();
    if (dlgExtBatch) run(() => DialogueExtractBatch(dlgExtInput, dlgExtOutput, cols));
    else run(() => DialogueExtractFile(dlgExtInput, dlgExtOutput, cols));
  }
  function startDlgImport() {
    if (dlgImpBatch) run(() => DialogueImportBatch(dlgImpScript, dlgImpTsv, dlgImpTargetCol, dlgImpOutput));
    else run(() => DialogueImportFile(dlgImpScript, dlgImpTsv, dlgImpTargetCol, dlgImpOutput));
  }

  // ===== RLdev browse helpers =====
  async function browseRlSeen() { const f = await SelectFile('Select SEEN.txt', '*.txt;*.TXT', 'SEEN archives'); if (f) rlSeenFile = f; }
  async function browseRlSeenSave() { const f = await SelectSaveFile('Save SEEN.txt as', 'SEEN.TXT', '*.txt;*.TXT', 'SEEN archives'); if (f) rlSeenFile = f; }
  async function browseRlTemplateSeen() { const f = await SelectFile('Select original/template SEEN.txt', '*.txt;*.TXT', 'SEEN archives'); if (f) rlTemplateSeenFile = f; }
  async function browseRlOrg() {
    if (rlCompileBatch) {
      const d = await SelectDirectory('Select folder with .org / .ke / .avg files');
      if (d) rlOrgDir = d;
    } else {
      const f = await SelectFile('Select .org / .ke / .avg file', '*.org;*.ke;*.avg', 'RLdev scripts');
      if (f) rlOrgFile = f;
    }
  }
  async function browseRlOrgText() {
    if (rlOrgTextBatch) {
      const d = await SelectDirectory('Select folder with .org / .ke files');
      if (d) rlOrgTextDir = d;
    } else {
      const f = await SelectFile('Select .org / .ke file', '*.org;*.ORG;*.ke;*.KE', 'Kepago scripts');
      if (f) rlOrgTextFile = f;
    }
  }
  async function browseRlOrgTextUtf() {
    if (rlOrgTextBatch) {
      const d = await SelectDirectory('Select folder with .utf files');
      if (d) rlOrgTextUtfDir = d;
    } else {
      const f = await SelectFile('Select .utf file', '*.utf;*.UTF', 'UTF text files');
      if (f) rlOrgTextUtfFile = f;
    }
  }
  async function browseRlKfn() { const f = await SelectFile('Select .kfn file', '*.kfn', 'KFN files'); if (f) rlKfnFile = f; }
  async function browseRlGameexe() {
    const f = await SelectFile('Select GAMEEXE.INI', '*.ini;*.INI', 'INI files');
    if (f) {
      rlGameexe = f;
      await refreshRlTargetVersion();
    }
  }
  async function browseRlInterpreter() {
    const f = await SelectFile('Select RealLive / Steam .exe', '*.exe;*.EXE', 'RealLive-compatible interpreter');
    if (f) {
      rlInterpreter = f;
      await refreshRlTargetVersion();
    }
  }
  async function browseRlOutputDir() { const d = await SelectDirectory('Select output directory'); if (d) rlOutputDir = d; }
  async function browseRlG00() {
    if (rlG00Batch) { const d = await SelectDirectory('Select folder with .g00 files'); if (d) rlG00Dir = d; }
    else { const f = await SelectFile('Select .g00 file', '*.g00;*.G00', 'G00 images'); if (f) rlG00File = f; }
  }
  async function browseRlPng() {
    if (rlPngBatch) { const d = await SelectDirectory('Select folder with .png files'); if (d) rlPngDir = d; }
    else { const f = await SelectFile('Select .png file', '*.png;*.PNG', 'PNG images'); if (f) rlPngFile = f; }
  }
  async function browseRlG00Xml() {
    if (rlG00Batch) { const d = await SelectDirectory('Select XML output folder'); if (d) rlG00XmlPath = d; }
    else { const f = await SelectSaveFile('Save metadata XML as', 'image.xml', '*.xml;*.XML', 'G00 metadata XML'); if (f) rlG00XmlPath = f; }
  }
  async function browseRlPngXml() {
    if (rlPngBatch) { const d = await SelectDirectory('Select folder with .xml metadata files'); if (d) rlPngXmlPath = d; }
    else { const f = await SelectFile('Select .xml metadata file', '*.xml;*.XML', 'G00 metadata XML'); if (f) rlPngXmlPath = f; }
  }
  async function browseRlGan() { const f = await SelectFile('Select .gan/.ganxml', '*.gan;*.ganxml', 'GAN files'); if (f) rlGanFile = f; }
  async function browseRlNwa() {
    if (rlNwaBatch) { const d = await SelectDirectory('Select folder with .nwa files'); if (d) rlNwaDir = d; }
    else { const f = await SelectFile('Select .nwa file', '*.nwa;*.NWA', 'NWA audio'); if (f) rlNwaFile = f; }
  }
  async function browseRlDat() {
    if (rlDatBatch) { const d = await SelectDirectory('Select folder with .cgm / .tcc files'); if (d) rlDatDir = d; }
    else { const f = await SelectFile('Select .cgm / .tcc file', '*.cgm;*.CGM;*.tcc;*.TCC', 'RealLive DAT assets'); if (f) rlDatFile = f; }
  }
  async function browseRlDatJson() {
    if (rlDatJsonBatch) { const d = await SelectDirectory('Select folder with DAT JSON files'); if (d) rlDatJsonDir = d; }
    else { const f = await SelectFile('Select DAT JSON file', '*.json;*.JSON', 'DAT JSON files'); if (f) rlDatJsonFile = f; }
  }
  async function browseRlSave() {
    const f = await SelectFile('Select RealLive save', '*.sav;*.SAV', 'RealLive saves');
    if (f) rlSaveFile = f;
  }
  async function browseRlSaveCompare() {
    const f = await SelectFile('Select RealLive save to compare', '*.sav;*.SAV', 'RealLive saves');
    if (f) rlSaveCompareFile = f;
  }
  async function browseRlSaveMapFile() {
    const f = await SelectFile('Select RealLive save', '*.sav;*.SAV', 'RealLive saves');
    if (f) rlSaveMapPath = f;
  }
  async function browseRlSaveMapDir() {
    const d = await SelectDirectory('Select folder with RealLive saves');
    if (d) rlSaveMapPath = d;
  }
  async function browseRlSaveTextInput() {
    const f = await SelectFile('Select rlsave text export', '*.txt;*.TXT;*.rlsavetxt', 'rlsave text exports');
    if (f) rlSaveTextFile = f;
  }
  async function browseRlSaveTextOutput() {
    const f = await SelectSaveFile('Save rlsave text export as', 'save.txt', '*.txt;*.TXT;*.rlsavetxt', 'rlsave text exports');
    if (f) rlSaveTextFile = f;
  }
  async function browseRlSaveBuildOutput() {
    const f = await SelectSaveFile('Save rebuilt RealLive save as', 'rebuilt.sav', '*.sav;*.SAV', 'RealLive saves');
    if (f) rlSaveBuildOutput = f;
  }
  async function browseRlBabelRoot() { const d = await SelectDirectory('Select BABEL folder'); if (d) rlBabelRoot = d; }
  async function browseRlBabelGameDir() { const d = await SelectDirectory('Select game folder'); if (d) rlBabelGameDir = d; }

  // ===== RLdev actions (call backend) =====
  function normalizeRlGameId() { rlGameId = rlGameId.trim().toUpperCase(); }
  function startRlDisasm() { normalizeRlGameId(); run(() => RldevDisassemble(rlSeenFile, rlKfnFile, rlEncoding, rlGameId, rlDebugInfo, rlOutputDir)); }
  function startRlExtract() { run(() => RldevExtract(rlSeenFile, rlOutputDir)); }
  function startRlArchive() { run(() => RldevArchive(rlSeenFile, rlOutputDir, rlTemplateSeenFile)); }
  function startRlList() { run(() => RldevList(rlSeenFile)); }
  function startRlCompile() {
    if (rlCompileBatch) {
      run(() => RldevCompileBatch(rlOrgDir, rlKfnFile, rlGameexe, rlInterpreter, rlTargetVersion, rlEncoding, rlOutputTransform, rlForceTransform, rlOutputDir));
    } else {
      run(() => RldevCompile(rlOrgFile, rlKfnFile, rlGameexe, rlInterpreter, rlTargetVersion, rlEncoding, rlOutputTransform, rlForceTransform, rlOutputDir));
    }
  }
  function toggleCompileBatch() { rlOrgFile = ''; rlOrgDir = ''; }
  function toggleOrgTextBatch() {
    rlOrgTextFile = '';
    rlOrgTextDir = '';
    rlOrgTextUtfFile = '';
    rlOrgTextUtfDir = '';
  }
  function toggleG00Batch() { rlG00File = ''; rlG00Dir = ''; rlG00XmlPath = ''; }
  function togglePngBatch() { rlPngFile = ''; rlPngDir = ''; rlPngXmlPath = ''; }
  function toggleNwaBatch() { rlNwaFile = ''; rlNwaDir = ''; }
  function toggleDatBatch() { rlDatFile = ''; rlDatDir = ''; }
  function toggleDatJsonBatch() { rlDatJsonFile = ''; rlDatJsonDir = ''; }
  function startG00Extract() { run(() => RldevG00ToPng(rlG00Batch ? rlG00Dir : rlG00File, rlOutputDir, rlG00XmlPath, rlG00Batch)); }
  function startOrgText() {
    const orgInput = rlOrgTextBatch ? rlOrgTextDir : rlOrgTextFile;
    if (rlOrgTextMode === 'import') {
      const utfInput = rlOrgTextBatch ? rlOrgTextUtfDir : rlOrgTextUtfFile;
      run(() => RldevOrgTextImport(orgInput, utfInput, rlOutputDir, rlEncoding, rlOrgTextBatch));
      return;
    }
    run(() => RldevOrgTextExport(orgInput, rlOutputDir, rlEncoding, rlOrgTextBatch));
  }
  function startG00Import() { run(() => RldevPngToG00(rlPngBatch ? rlPngDir : rlPngFile, rlOutputDir, rlPngXmlPath, rlG00Format, rlPngBatch)); }
  function startGanToXml() { run(() => RldevGanToXml(rlGanFile, rlOutputDir)); }
  function startGanFromXml() { run(() => RldevXmlToGan(rlGanFile, rlOutputDir)); }
  function startNwaAudio() { run(() => RldevNwaToAudio(rlNwaBatch ? rlNwaDir : rlNwaFile, rlOutputDir, rlAudioFormat, rlNwaBatch)); }
  function startDatToJson() { run(() => RldevDatToJson(rlDatBatch ? rlDatDir : rlDatFile, rlOutputDir, rlDatBatch)); }
  function startDatFromJson() { run(() => RldevDatJsonToBinary(rlDatJsonBatch ? rlDatJsonDir : rlDatJsonFile, rlOutputDir, rlDatJsonBatch)); }
  function startSaveInfo() { run(() => RldevSaveInfo(rlSaveFile)); }
  function startSaveMap() { run(() => RldevSaveMap(rlSaveMapPath || rlSaveFile, rlSaveMapJson)); }
  function startSaveDoctor() { run(() => RldevSaveDoctor(rlSaveMapPath || rlSaveFile, rlSaveDoctorJson)); }
  function startSaveDiff() { run(() => RldevSaveDiff(rlSaveFile, rlSaveCompareFile, rlSaveDiffJson)); }
  function startSaveGet() { run(() => RldevSaveGet(rlSaveFile, rlSaveRefs)); }
  function startSaveSet() { run(() => RldevSaveSet(rlSaveFile, rlSaveAssignments, rlSaveBackup)); }
  function startSaveDump() { run(() => RldevSaveDump(rlSaveFile, rlSaveDumpAll, rlSaveDumpJson)); }
  function startSaveExport() { run(() => RldevSaveExport(rlSaveFile, rlSaveTextFile, rlSaveLossless)); }
  function startSaveBuild() { run(() => RldevSaveBuild(rlSaveTextFile, rlSaveBuildOutput, rlSaveBackup)); }
  function applySaveProfile() {
    const profile = saveProfiles.find((item) => item.id === rlSaveProfile);
    if (!profile) return;
    rlSaveRefs = profile.refs;
    rlSaveAssignments = profile.assignments;
  }
  function startBabelRuntime() { run(() => RldevBabelPrepareRuntime(rlBabelRoot, rlBabelGameDir, rlBabelVersion, rlBabelDllMode, rlBabelNameEnc, rlBabelUpdateGameexe)); }
  function startBabelHeader() { run(() => RldevBabelWriteHeader(rlOutputDir, rlBabelGlosses)); }
  async function refreshRlTargetVersion() {
    const detected = await DetectRealLiveVersion(rlGameexe, rlInterpreter);
    if (detected && (!rlTargetVersion.trim() || rlTargetVersionAuto)) {
      rlTargetVersion = detected;
      rlTargetVersionAuto = true;
    }
  }
  function markRlTargetVersionManual() {
    rlTargetVersionAuto = false;
  }

  // ===== Siglus browse helpers =====
  async function browseSiglusScenePck() { const f = await SelectFile('Select Scene.pck', '*.pck;*.PCK', 'PCK files'); if (f) siglusScenePck = f; }
  async function browseSiglusSceneOutputDir() { const d = await SelectDirectory('Select extraction output'); if (d) siglusSceneOutputDir = d; }
  async function browseSiglusSceneInputDir() { const d = await SelectDirectory('Select folder with .ss files'); if (d) siglusSceneInputDir = d; }
  async function browseSiglusSceneOutputPck() { const f = await SelectSaveFile('Save Scene.pck', 'Scene.pck', '*.pck;*.PCK', 'PCK files'); if (f) siglusSceneOutputPck = f; }

  function toggleSiglusSSBatch() { siglusSSInput = ''; siglusSSTsv = ''; siglusSSOutput = ''; }
  async function browseSiglusSSInput() {
    if (siglusSSBatch) { const d = await SelectDirectory('Select .ss folder'); if (d) siglusSSInput = d; }
    else { const f = await SelectFile('Select .ss file', '*.ss;*.SS', 'SS files'); if (f) siglusSSInput = f; }
  }
  async function browseSiglusSSTsv() {
    if (siglusSelectedOp === 'ss_dump') {
      if (siglusSSFormat === 'xlsx') {
        if (siglusSSBatch && !siglusSSSingleXlsx) { const d = await SelectDirectory('Select xlsx output folder'); if (d) siglusSSTsv = d; }
        else { const f = await SelectSaveFile('Save xlsx dump', 'Scene.xlsx', '*.xlsx;*.XLSX', 'Excel files'); if (f) siglusSSTsv = f; }
      } else if (siglusSSBatch) { const d = await SelectDirectory('Select text output folder'); if (d) siglusSSTsv = d; }
      else { const f = await SelectSaveFile('Save text dump', 'scene.ss.txt', '*.txt;*.TXT', 'Siglus text files'); if (f) siglusSSTsv = f; }
      return;
    }
    if (siglusSSBatch) { const d = await SelectDirectory('Select translated text/xlsx folder'); if (d) siglusSSTsv = d; }
    else { const f = await SelectFile('Select translated text', '*.txt;*.tsv;*.xlsx;*.XLSX', 'Siglus text / Excel files'); if (f) siglusSSTsv = f; }
  }
  async function browseSiglusSSOutput() {
    if (siglusSSBatch) { const d = await SelectDirectory('Select patched .ss output folder'); if (d) siglusSSOutput = d; }
    else { const f = await SelectSaveFile('Save patched .ss', 'patched.ss', '*.ss;*.SS', 'SS files'); if (f) siglusSSOutput = f; }
  }

  async function browseSiglusGameexeDat() { const f = await SelectFile('Select Gameexe.dat', '*.dat;*.DAT', 'Gameexe files'); if (f) siglusGameexeDat = f; }
  async function browseSiglusGameexeIni() { const f = await SelectFile('Select Gameexe.ini', '*.ini;*.INI', 'INI files'); if (f) siglusGameexeIni = f; }
  async function browseSiglusGameexeExtractOutput() { const f = await SelectSaveFile('Save Gameexe.ini', 'Gameexe.ini', '*.ini;*.INI', 'INI files'); if (f) siglusGameexeOutput = f; }
  async function browseSiglusGameexeRebuildOutput() { const f = await SelectSaveFile('Save Gameexe.dat', 'Gameexe.dat', '*.dat;*.DAT', 'Gameexe files'); if (f) siglusGameexeOutput = f; }

  async function browseSiglusDBSFile() { const f = await SelectFile('Select .dbs file', '*.dbs;*.DBS', 'DBS files'); if (f) siglusDBSFile = f; }
  async function browseSiglusDBSRaw() {
    if (siglusSelectedOp === 'dbs_extract') { const f = await SelectSaveFile('Save .dbs.out', 'data.dbs.out', '*.out;*.txt;*.*', 'DBS raw output'); if (f) siglusDBSRaw = f; }
    else { const f = await SelectFile('Select .dbs.out', '*.out;*.*', 'DBS raw output'); if (f) siglusDBSRaw = f; }
  }
  async function browseSiglusDBSTxt() {
    if (siglusSelectedOp === 'dbs_extract') { const f = await SelectSaveFile('Save .dbs.txt', 'data.dbs.txt', '*.txt;*.TXT', 'DBS text'); if (f) siglusDBSTxt = f; }
    else { const f = await SelectFile('Select .dbs.txt', '*.txt;*.TXT', 'DBS text'); if (f) siglusDBSTxt = f; }
  }
  async function browseSiglusDBSOutput() { const f = await SelectSaveFile('Save .dbs', 'data.dbs', '*.dbs;*.DBS', 'DBS files'); if (f) siglusDBSOutput = f; }
  async function browseSiglusDBSXlsx() { const f = await SelectSaveFile('Save DBS xlsx', 'data.dbs.xlsx', '*.xlsx;*.XLSX', 'Excel files'); if (f) siglusDBSXlsx = f; }
  async function browseSiglusDBSXlsxDir() { const d = await SelectDirectory('Select xlsx folder'); if (d) siglusDBSXlsxDir = d; }
  async function browseSiglusDBSOutputDir() { const d = await SelectDirectory('Select DBS output folder'); if (d) siglusDBSOutputDir = d; }

  async function browseSiglusMobilePck() { const f = await SelectFile('Select mobile PCK', '*.pck;*.PCK', 'PCK files'); if (f) siglusMobilePck = f; }
  async function browseSiglusMobileDir() { const d = await SelectDirectory('Select mobile PCK folder'); if (d) siglusMobileDir = d; }
  async function browseSiglusMobileOutput() { const f = await SelectSaveFile('Save mobile PCK', 'data.pck', '*.pck;*.PCK', 'PCK files'); if (f) siglusMobileOutput = f; }

  async function browseSiglusOMVInput() { const f = await SelectFile('Select OMV file', '*.omv;*.OMV', 'OMV files'); if (f) siglusOMVInput = f; }
  async function browseSiglusOGVOutput() { const f = await SelectSaveFile('Save OGV', 'movie.ogv', '*.ogv;*.OGV', 'OGV files'); if (f) siglusOGVOutput = f; }
  async function browseSiglusOGVInput() { const f = await SelectFile('Select OGV file', '*.ogv;*.OGV', 'OGV files'); if (f) siglusOGVInput = f; }
  async function browseSiglusOMVOutput() { const f = await SelectSaveFile('Save OMV', 'movie.omv', '*.omv;*.OMV', 'OMV files'); if (f) siglusOMVOutput = f; }
  async function browseSiglusOMV2AVIInput() { const f = await SelectFile('Select OMV file', '*.omv;*.OMV', 'OMV files'); if (f) siglusOMV2AVIInput = f; }
  async function browseSiglusOMV2AVIOutput() { const f = await SelectSaveFile('Save AVI / OGV', 'movie.avi', '*.avi;*.AVI;*.ogv;*.OGV', 'AVI / OGV files'); if (f) siglusOMV2AVIOutput = f; }
  async function browseSiglusOMVPNGInput() { const f = await SelectFile('Select OMV file', '*.omv;*.OMV', 'OMV files'); if (f) siglusOMVPNGInput = f; }
  async function browseSiglusOMVPNGOutputDir() { const d = await SelectDirectory('Select PNG output folder'); if (d) siglusOMVPNGOutputDir = d; }
  async function browseSiglusPNGVideoDir() { const d = await SelectDirectory('Select PNG sequence folder'); if (d) siglusPNGVideoDir = d; }
  async function browseSiglusPNGVideoOutput() { const f = await SelectSaveFile('Save OMV / OGV', 'movie.omv', '*.omv;*.OMV;*.ogv;*.OGV', 'OMV / OGV files'); if (f) siglusPNGVideoOutput = f; }
  async function browseSiglusCombinePNGDir() { const d = await SelectDirectory('Select PNG folder'); if (d) siglusCombinePNGDir = d; }
  async function browseSiglusCombinePNGOutput() { const f = await SelectSaveFile('Save combined PNG', 'combined.png', '*.png;*.PNG', 'PNG files'); if (f) siglusCombinePNGOutput = f; }
  async function browseSiglusScriptRepackScript() { const f = await SelectFile('Select script file', '*.ss;*.SS;*.dat;*.DAT;*.*', 'Script files'); if (f) siglusScriptRepackScript = f; }
  async function browseSiglusScriptRepackText() { const f = await SelectFile('Select UTF-16 text', '*.txt;*.TXT', 'Text files'); if (f) siglusScriptRepackText = f; }
  async function browseSiglusScriptRepackOutput() {
    const base = siglusScriptRepackScript ? `${siglusScriptRepackScript.split(/[\\/]/).pop()}.out` : 'script.ss.out';
    const f = await SelectSaveFile('Save repacked script', base, '*.ss;*.SS;*.out;*.*', 'Script files');
    if (f) siglusScriptRepackOutput = f;
  }
  async function browseSiglusG00() {
    if (siglusG00Batch) { const d = await SelectDirectory('Select folder with .g00 files'); if (d) siglusG00Dir = d; }
    else { const f = await SelectFile('Select .g00 file', '*.g00;*.G00', 'G00 images'); if (f) siglusG00File = f; }
  }
  async function browseSiglusG00Xml() {
    if (siglusG00Batch) { const d = await SelectDirectory('Select XML output folder'); if (d) siglusG00XmlPath = d; }
    else { const f = await SelectSaveFile('Save metadata XML as', 'image.xml', '*.xml;*.XML', 'G00 metadata XML'); if (f) siglusG00XmlPath = f; }
  }
  async function browseSiglusG00OutputDir() { const d = await SelectDirectory('Select output folder'); if (d) siglusG00OutputDir = d; }
  async function browseSiglusPng() {
    if (siglusPngBatch) { const d = await SelectDirectory('Select folder with .png files'); if (d) siglusPngDir = d; }
    else { const f = await SelectFile('Select .png file', '*.png;*.PNG', 'PNG images'); if (f) siglusPngFile = f; }
  }
  async function browseSiglusPngXml() {
    if (siglusPngBatch) { const d = await SelectDirectory('Select folder with .xml metadata files'); if (d) siglusPngXmlPath = d; }
    else { const f = await SelectFile('Select .xml metadata file', '*.xml;*.XML', 'G00 metadata XML'); if (f) siglusPngXmlPath = f; }
  }
  async function browseSiglusPngOutputDir() { const d = await SelectDirectory('Select output folder'); if (d) siglusPngOutputDir = d; }
  function toggleSiglusG00Batch() { siglusG00File = ''; siglusG00Dir = ''; siglusG00XmlPath = ''; }
  function toggleSiglusPngBatch() { siglusPngFile = ''; siglusPngDir = ''; siglusPngXmlPath = ''; }

  // ===== Siglus actions =====
  function startSiglusSceneExtract() { run(() => SiglusSceneExtract(siglusScenePck, siglusGameName, siglusSceneOutputDir)); }
  function startSiglusSceneRebuild() { run(() => SiglusSceneRebuild(siglusSceneInputDir, siglusGameName, siglusSceneWtf, siglusSceneOutputPck, siglusCompressionLevel, siglusFakeCompression)); }
  function startSiglusSSDump() {
    if (siglusSSBatch) run(() => SiglusSSDumpAll(siglusSSInput, siglusSSTsv, siglusSSCopyText, siglusSSSingleLine, siglusSSFilterMode, siglusSSFormat, siglusSSSingleXlsx));
    else run(() => SiglusSSDump(siglusSSInput, siglusSSTsv, siglusSSCopyText, siglusSSSingleLine, siglusSSFilterMode, siglusSSFormat));
  }
  function startSiglusSSInject() {
    if (siglusSSBatch) run(() => SiglusSSInjectAll(siglusSSInput, siglusSSTsv, siglusSSOutput));
    else run(() => SiglusSSInject(siglusSSInput, siglusSSTsv, siglusSSOutput));
  }
  function startSiglusGameexeExtract() { run(() => SiglusGameexeExtract(siglusGameexeDat, siglusGameName, siglusGameexeOutput)); }
  function startSiglusGameexeRebuild() { run(() => SiglusGameexeRebuild(siglusGameexeIni, siglusGameName, siglusGameexeOutput, siglusGameexeDoubleEncryption, siglusCompressionLevel, siglusFakeCompression)); }
  function startSiglusDBSExtract() { run(() => SiglusDBSExtract(siglusDBSFile, siglusDBSRaw, siglusDBSTxt, siglusDBSEncoding, siglusDBSDumpAll)); }
  function startSiglusDBSRebuild() { run(() => SiglusDBSRebuild(siglusDBSRaw, siglusDBSTxt, siglusDBSOutput, siglusDBSEncoding, siglusCompressionLevel, siglusFakeCompression)); }
  function startSiglusDBSExportXLSX() { run(() => SiglusDBSExportXLSX(siglusDBSFile, siglusDBSXlsx, siglusDBSEncoding)); }
  function startSiglusDBSBuildFromXLSX() { run(() => SiglusDBSBuildFromXLSX(siglusDBSXlsxDir, siglusDBSOutputDir, siglusDBSEncoding, siglusDBSUnicodeOutput, siglusCompressionLevel, siglusFakeCompression)); }
  function startSiglusMobilePCKExtract() { run(() => SiglusMobilePCKExtract(siglusMobilePck, siglusMobileDir)); }
  function startSiglusMobilePCKRebuild() { run(() => SiglusMobilePCKRebuild(siglusMobileDir, siglusMobileOutput)); }
  function startSiglusOMVCut() { run(() => SiglusOMVCut(siglusOMVInput, siglusOGVOutput)); }
  function startSiglusOMVPack() { run(() => SiglusOMVPack(siglusOGVInput, siglusOMVOutput)); }
  function startSiglusOMV2AVI() { run(() => SiglusOMV2AVI(siglusOMV2AVIInput, siglusOMV2AVIOutput)); }
  function startSiglusOMVPNG() { run(() => SiglusOMVPNG(siglusOMVPNGInput, siglusOMVPNGOutputDir)); }
  function startSiglusPNGVideo() { run(() => SiglusPNGVideo(siglusPNGVideoDir, siglusPNGVideoOutput, siglusPNGVideoAlpha, siglusPNGVideoFPS)); }
  function startSiglusCombinePNG() { run(() => SiglusCombinePNG(siglusCombinePNGDir, siglusCombinePNGOutput)); }
  function startSiglusScriptRepack() { run(() => SiglusScriptRepack(siglusScriptRepackScript, siglusScriptRepackText, siglusScriptRepackOutput)); }
  function startSiglusG00Extract() { run(() => SiglusG00ToPng(siglusG00Batch ? siglusG00Dir : siglusG00File, siglusG00OutputDir, siglusG00XmlPath, siglusG00Batch)); }
  function startSiglusG00Import() { run(() => SiglusPngToG00(siglusPngBatch ? siglusPngDir : siglusPngFile, siglusPngOutputDir, siglusPngXmlPath, siglusG00Format, siglusPngBatch)); }
</script>

<div id="app">
  <!-- HUB VIEW -->
  {#if activeView === 'hub'}
    <div class="hub-titlebar">
      <span>AIO Key Game Tools</span>
      <span class="hub-subtitle">The ultimate toolbox for VA/Key Games</span>
    </div>
    <div class="hub-content">
      <div class="hub-header">
        <div class="hub-title">🔑 AIO Key Game Tools</div>
        <div class="hub-desc">Select a tool to get started</div>
      </div>
      <div class="hub-grid">
        <button class="hub-card" on:click={() => activeView = 'lucksystem'}>
          <div class="hub-card-title">LuckSystem</div>
          <div class="hub-card-ver">2.3.2 · Yoremi Fork v3.30 GUI</div>
          <div class="hub-card-desc">Scripts, PAK, fonts, images CZ<br>for LuckEngine games</div>
        </button>
        <button class="hub-card" on:click={() => activeView = 'siglus'}>
          <div class="hub-card-title">Siglus Tools</div>
          <div class="hub-card-ver">0.61 · Go port</div>
          <div class="hub-card-desc">SiglusEngine extraction<br>and repackaging</div>
        </button>
        <button class="hub-card" on:click={() => activeView = 'rldev'}>
          <div class="hub-card-title">RLdev 2026</div>
          <div class="hub-card-ver">v1.3 · Go port</div>
          <div class="hub-card-desc">SEEN.txt, Kepago/AVG32, G00, NWA<br>DAT, GAN and Babel for RealLive games</div>
        </button>
      </div>
      <div class="hub-footer">
        <button class="hub-link" on:click={() => activeView = 'outils'}>🛠 Outils divers</button>
        <button class="hub-link" on:click={() => activeView = 'about_global'}>ℹ À propos</button>
      </div>
    </div>

  <!-- ABOUT GLOBAL -->
  {:else if activeView === 'about_global'}
    <div class="titlebar">
      <span>AIO Key Game Tools — À propos</span>
      <button class="titlebar-back" on:click={() => activeView = 'hub'}>← Retour</button>
    </div>
    <div class="form-panel" style="padding:30px 40px">
      <div class="form-title">À propos</div>
      <div class="about-panel">
        <div class="about-logo">AIO Key Game Tools</div>
        <div class="about-subtitle">The ultimate toolbox for Visual Art's / Key Games</div>
        <div class="about-desc">
          Suite d'outils intégrée pour le modding des visual novels Key / Visual Art's.<br><br>
          <strong>LuckSystem v3.30 GUI</strong> — Scripts, PAK, fonts, images CZ, vidéos MVT, extraction audio Ogg/MP3, alias de fontes PAK et générateur Luca Menu DLL<br>
          <strong>RLdev 2026 v1.3.5</strong> — SEEN.txt, Kepago, AVG32, G00, GAN, NWA, DAT, Babel, saves (RealLive)<br>
          <strong>Siglus Tools</strong> — SiglusEngine, Scene.pck, SS, Gameexe, DBS, mobile PCK, OMV<br><br>
          Développé par <strong>Yoremi</strong> · Wails + Svelte
        </div>
        <div class="about-version">v1.0.3 · 2026</div>
      </div>
    </div>

  <!-- SIGLUS TOOLS -->
  {:else if activeView === 'siglus'}
    <div class="titlebar">
      <span>Siglus Tools 0.61 — Go port</span>
      <button class="titlebar-back" on:click={() => activeView = 'hub'}>← Retour</button>
    </div>
    <div class="content">
      <div class="sidebar">
        <div class="sidebar-title">Siglus :</div>
        <div class="sidebar-list">
          {#each siglusOperations as op}
            {#if op.section}
              <div class="sidebar-section">{op.label}</div>
            {:else}
              <div class="sidebar-item" class:active={siglusSelectedOp === op.id} on:click={() => siglusSelectedOp = op.id}>
                {op.label}
              </div>
            {/if}
          {/each}
        </div>
      </div>
      <div class="form-panel">
        <datalist id="siglus-game-options">
          {#each siglusGameKeyOptions as gameKey}
            <option value={gameKey}></option>
          {/each}
        </datalist>

        {#if siglusSelectedOp === 'scene_extract'}
          <div class="form-title">Extract Scene.pck</div>
          <div class="form-group"><label>Scene.pck :</label><div class="form-row"><input type="text" bind:value={siglusScenePck} readonly /><button class="btn" on:click={browseSiglusScenePck}>Select</button></div></div>
          <div class="form-group"><label>Jeu / clé :</label><div class="form-row"><input type="text" bind:value={siglusGameName} list="siglus-game-options" /></div></div>
          <div class="form-group"><label>Dossier de sortie :</label><div class="form-row"><input type="text" bind:value={siglusSceneOutputDir} readonly /><button class="btn" on:click={browseSiglusSceneOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusSceneExtract} disabled={!siglusScenePck || !siglusGameName || !siglusSceneOutputDir}>Extract</button>{/if}</div>

        {:else if siglusSelectedOp === 'scene_rebuild'}
          <div class="form-title">Rebuild Scene.pck</div>
          <div class="form-group"><label>Dossier extrait :</label><div class="form-row"><input type="text" bind:value={siglusSceneInputDir} readonly /><button class="btn" on:click={browseSiglusSceneInputDir}>Select</button></div></div>
          <div class="form-group"><label>Jeu / clé :</label><div class="form-row"><input type="text" bind:value={siglusGameName} list="siglus-game-options" /></div></div>
          <div class="form-group"><label>WTF value :</label><div class="form-row"><input type="text" bind:value={siglusSceneWtf} placeholder="auto" /></div></div>
          <div class="form-group"><label>Compression :</label><div class="form-row"><input type="number" bind:value={siglusCompressionLevel} min="2" max="17" style="width:80px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" /><label class="checkbox-label"><input type="checkbox" bind:checked={siglusFakeCompression} /> Fake compression</label></div></div>
          <div class="form-group"><label>Scene.pck de sortie :</label><div class="form-row"><input type="text" bind:value={siglusSceneOutputPck} readonly /><button class="btn" on:click={browseSiglusSceneOutputPck}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusSceneRebuild} disabled={!siglusSceneInputDir || !siglusGameName || !siglusSceneWtf || !siglusSceneOutputPck}>Rebuild</button>{/if}</div>

        {:else if siglusSelectedOp === 'ss_dump'}
          <div class="form-title">Dump SS text</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSBatch} on:change={toggleSiglusSSBatch} /> Batch mode</label></div></div>
          <div class="form-group"><label>{siglusSSBatch ? 'Dossier .ss :' : 'Fichier .ss :'}</label><div class="form-row"><input type="text" bind:value={siglusSSInput} readonly /><button class="btn" on:click={browseSiglusSSInput}>Select</button></div></div>
          <div class="form-group"><label>Filtre :</label><div class="form-row"><select bind:value={siglusSSFilterMode}><option value="smart">Smart filter</option><option value="all">Export all text</option><option value="full">Full-width only</option></select><label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSCopyText} /> Copy text</label>{#if siglusSSFormat === 'txt'}<label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSSingleLine} /> Une seule ligne</label>{/if}</div>{#if siglusSSFormat === 'txt' && siglusSSSingleLine}<div class="form-hint">Écrit uniquement le slot ●, directement éditable et réinjectable. L'option Copy text est alors implicite.</div>{/if}</div>
          <div class="form-group"><label>Format :</label><div class="form-row"><select bind:value={siglusSSFormat} on:change={() => siglusSSTsv = ''}><option value="txt">TXT Siglus Tools</option><option value="xlsx">XLSX</option></select>{#if siglusSSBatch && siglusSSFormat === 'xlsx'}<label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSSingleXlsx} on:change={() => siglusSSTsv = ''} /> Single workbook</label>{/if}</div></div>
          <div class="form-group"><label>{siglusSSFormat === 'xlsx' ? (siglusSSBatch && !siglusSSSingleXlsx ? 'Dossier Excel :' : 'Classeur Excel :') : (siglusSSBatch ? 'Dossier texte :' : 'Texte de sortie :')}</label><div class="form-row"><input type="text" bind:value={siglusSSTsv} readonly /><button class="btn" on:click={browseSiglusSSTsv}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusSSDump} disabled={!siglusSSInput || !siglusSSTsv}>Dump</button>{/if}</div>

        {:else if siglusSelectedOp === 'ss_inject'}
          <div class="form-title">Inject SS text</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSBatch} on:change={toggleSiglusSSBatch} /> Batch mode</label></div></div>
          <div class="form-group"><label>{siglusSSBatch ? 'Dossier .ss original :' : 'Fichier .ss original :'}</label><div class="form-row"><input type="text" bind:value={siglusSSInput} readonly /><button class="btn" on:click={browseSiglusSSInput}>Select</button></div></div>
          <div class="form-group"><label>{siglusSSBatch ? 'Dossier texte / Excel :' : 'Texte / Excel traduit :'}</label><div class="form-row"><input type="text" bind:value={siglusSSTsv} readonly /><button class="btn" on:click={browseSiglusSSTsv}>Select</button></div></div>
          <div class="form-group"><label>{siglusSSBatch ? 'Dossier .ss patché :' : '.ss patché :'}</label><div class="form-row"><input type="text" bind:value={siglusSSOutput} readonly /><button class="btn" on:click={browseSiglusSSOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusSSInject} disabled={!siglusSSInput || !siglusSSTsv || !siglusSSOutput}>Inject</button>{/if}</div>

        {:else if siglusSelectedOp === 'gameexe_extract'}
          <div class="form-title">Decrypt Gameexe.dat</div>
          <div class="form-group"><label>Gameexe.dat :</label><div class="form-row"><input type="text" bind:value={siglusGameexeDat} readonly /><button class="btn" on:click={browseSiglusGameexeDat}>Select</button></div></div>
          <div class="form-group"><label>Jeu / clé :</label><div class="form-row"><input type="text" bind:value={siglusGameName} list="siglus-game-options" /></div></div>
          <div class="form-group"><label>Gameexe.ini de sortie :</label><div class="form-row"><input type="text" bind:value={siglusGameexeOutput} readonly /><button class="btn" on:click={browseSiglusGameexeExtractOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusGameexeExtract} disabled={!siglusGameexeDat || !siglusGameName || !siglusGameexeOutput}>Decrypt</button>{/if}</div>

        {:else if siglusSelectedOp === 'gameexe_rebuild'}
          <div class="form-title">Rebuild Gameexe.dat</div>
          <div class="form-group"><label>Gameexe.ini :</label><div class="form-row"><input type="text" bind:value={siglusGameexeIni} readonly /><button class="btn" on:click={browseSiglusGameexeIni}>Select</button></div></div>
          <div class="form-group"><label>Jeu / clé :</label><div class="form-row"><input type="text" bind:value={siglusGameName} list="siglus-game-options" /></div></div>
          <div class="form-group"><label>Compression :</label><div class="form-row"><input type="number" bind:value={siglusCompressionLevel} min="2" max="17" style="width:80px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" /><label class="checkbox-label"><input type="checkbox" bind:checked={siglusFakeCompression} /> Fake compression</label><label class="checkbox-label"><input type="checkbox" bind:checked={siglusGameexeDoubleEncryption} /> Double encryption</label></div></div>
          <div class="form-group"><label>Gameexe.dat de sortie :</label><div class="form-row"><input type="text" bind:value={siglusGameexeOutput} readonly /><button class="btn" on:click={browseSiglusGameexeRebuildOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusGameexeRebuild} disabled={!siglusGameexeIni || !siglusGameName || !siglusGameexeOutput}>Rebuild</button>{/if}</div>

        {:else if siglusSelectedOp === 'dbs_extract'}
          <div class="form-title">Dump DBS</div>
          <div class="form-group"><label>Fichier .dbs :</label><div class="form-row"><input type="text" bind:value={siglusDBSFile} readonly /><button class="btn" on:click={browseSiglusDBSFile}>Select</button></div></div>
          <div class="form-group"><label>Encodage ANSI :</label><div class="form-row"><select bind:value={siglusDBSEncoding}><option value="shift-jis">Shift-JIS</option><option value="gbk">GBK</option></select><label class="checkbox-label"><input type="checkbox" bind:checked={siglusDBSDumpAll} /> Export all data</label></div></div>
          <div class="form-group"><label>.dbs.out :</label><div class="form-row"><input type="text" bind:value={siglusDBSRaw} readonly /><button class="btn" on:click={browseSiglusDBSRaw}>Select</button></div></div>
          <div class="form-group"><label>.dbs.txt :</label><div class="form-row"><input type="text" bind:value={siglusDBSTxt} readonly /><button class="btn" on:click={browseSiglusDBSTxt}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusDBSExtract} disabled={!siglusDBSFile || !siglusDBSRaw || !siglusDBSTxt}>Dump</button>{/if}</div>

        {:else if siglusSelectedOp === 'dbs_rebuild'}
          <div class="form-title">Rebuild DBS</div>
          <div class="form-group"><label>.dbs.out :</label><div class="form-row"><input type="text" bind:value={siglusDBSRaw} readonly /><button class="btn" on:click={browseSiglusDBSRaw}>Select</button></div></div>
          <div class="form-group"><label>.dbs.txt :</label><div class="form-row"><input type="text" bind:value={siglusDBSTxt} readonly /><button class="btn" on:click={browseSiglusDBSTxt}>Select</button></div></div>
          <div class="form-group"><label>Encodage ANSI :</label><div class="form-row"><select bind:value={siglusDBSEncoding}><option value="gbk">GBK</option><option value="shift-jis">Shift-JIS</option></select><input type="number" bind:value={siglusCompressionLevel} min="2" max="17" style="width:80px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" /><label class="checkbox-label"><input type="checkbox" bind:checked={siglusFakeCompression} /> Fake compression</label></div></div>
          <div class="form-group"><label>.dbs de sortie :</label><div class="form-row"><input type="text" bind:value={siglusDBSOutput} readonly /><button class="btn" on:click={browseSiglusDBSOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusDBSRebuild} disabled={!siglusDBSRaw || !siglusDBSTxt || !siglusDBSOutput}>Rebuild</button>{/if}</div>

        {:else if siglusSelectedOp === 'dbs_xlsx'}
          <div class="form-title">Dump DBS XLSX</div>
          <div class="form-group"><label>Fichier .dbs :</label><div class="form-row"><input type="text" bind:value={siglusDBSFile} readonly /><button class="btn" on:click={browseSiglusDBSFile}>Select</button></div></div>
          <div class="form-group"><label>Encodage ANSI :</label><div class="form-row"><select bind:value={siglusDBSEncoding}><option value="shift-jis">Shift-JIS</option><option value="gbk">GBK</option></select></div></div>
          <div class="form-group"><label>Classeur XLSX :</label><div class="form-row"><input type="text" bind:value={siglusDBSXlsx} readonly /><button class="btn" on:click={browseSiglusDBSXlsx}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusDBSExportXLSX} disabled={!siglusDBSFile || !siglusDBSXlsx}>Dump XLSX</button>{/if}</div>

        {:else if siglusSelectedOp === 'dbs_build'}
          <div class="form-title">Build DBS from XLSX</div>
          <div class="form-group"><label>Dossier XLSX :</label><div class="form-row"><input type="text" bind:value={siglusDBSXlsxDir} readonly /><button class="btn" on:click={browseSiglusDBSXlsxDir}>Select</button></div></div>
          <div class="form-group"><label>Sortie :</label><div class="form-row"><input type="text" bind:value={siglusDBSOutputDir} readonly /><button class="btn" on:click={browseSiglusDBSOutputDir}>Select</button></div></div>
          <div class="form-group"><label>Format :</label><div class="form-row"><label class="checkbox-label"><input type="checkbox" bind:checked={siglusDBSUnicodeOutput} /> Unicode</label>{#if !siglusDBSUnicodeOutput}<select bind:value={siglusDBSEncoding}><option value="gbk">GBK</option><option value="shift-jis">Shift-JIS</option></select>{/if}<input type="number" bind:value={siglusCompressionLevel} min="2" max="17" style="width:80px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" /><label class="checkbox-label"><input type="checkbox" bind:checked={siglusFakeCompression} /> Fake compression</label></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusDBSBuildFromXLSX} disabled={!siglusDBSXlsxDir || !siglusDBSOutputDir}>Build DBS</button>{/if}</div>

        {:else if siglusSelectedOp === 'mobile_pck_extract'}
          <div class="form-title">Extract mobile PCK</div>
          <div class="form-group"><label>PCK mobile :</label><div class="form-row"><input type="text" bind:value={siglusMobilePck} readonly /><button class="btn" on:click={browseSiglusMobilePck}>Select</button></div></div>
          <div class="form-group"><label>Dossier de sortie :</label><div class="form-row"><input type="text" bind:value={siglusMobileDir} readonly /><button class="btn" on:click={browseSiglusMobileDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusMobilePCKExtract} disabled={!siglusMobilePck || !siglusMobileDir}>Extract</button>{/if}</div>

        {:else if siglusSelectedOp === 'mobile_pck_rebuild'}
          <div class="form-title">Rebuild mobile PCK</div>
          <div class="form-group"><label>Dossier source :</label><div class="form-row"><input type="text" bind:value={siglusMobileDir} readonly /><button class="btn" on:click={browseSiglusMobileDir}>Select</button></div></div>
          <div class="form-group"><label>PCK de sortie :</label><div class="form-row"><input type="text" bind:value={siglusMobileOutput} readonly /><button class="btn" on:click={browseSiglusMobileOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusMobilePCKRebuild} disabled={!siglusMobileDir || !siglusMobileOutput}>Rebuild</button>{/if}</div>

        {:else if siglusSelectedOp === 'omv_cut'}
          <div class="form-title">Cut OMV header</div>
          <div class="form-group"><label>Fichier OMV :</label><div class="form-row"><input type="text" bind:value={siglusOMVInput} readonly /><button class="btn" on:click={browseSiglusOMVInput}>Select</button></div></div>
          <div class="form-group"><label>OGV de sortie :</label><div class="form-row"><input type="text" bind:value={siglusOGVOutput} readonly /><button class="btn" on:click={browseSiglusOGVOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusOMVCut} disabled={!siglusOMVInput || !siglusOGVOutput}>Cut</button>{/if}</div>

        {:else if siglusSelectedOp === 'omv_pack'}
          <div class="form-title">Pack OMV</div>
          <div class="form-group"><label>Fichier OGV :</label><div class="form-row"><input type="text" bind:value={siglusOGVInput} readonly /><button class="btn" on:click={browseSiglusOGVInput}>Select</button></div></div>
          <div class="form-group"><label>OMV de sortie :</label><div class="form-row"><input type="text" bind:value={siglusOMVOutput} readonly /><button class="btn" on:click={browseSiglusOMVOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusOMVPack} disabled={!siglusOGVInput || !siglusOMVOutput}>Pack</button>{/if}</div>

        {:else if siglusSelectedOp === 'omv2avi'}
          <div class="form-title">Omv2Avi</div>
          <div class="form-group"><label>Fichier OMV :</label><div class="form-row"><input type="text" bind:value={siglusOMV2AVIInput} readonly /><button class="btn" on:click={browseSiglusOMV2AVIInput}>Select</button></div></div>
          <div class="form-group"><label>AVI / OGV de sortie :</label><div class="form-row"><input type="text" bind:value={siglusOMV2AVIOutput} readonly /><button class="btn" on:click={browseSiglusOMV2AVIOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusOMV2AVI} disabled={!siglusOMV2AVIInput || !siglusOMV2AVIOutput}>Convert</button>{/if}</div>

        {:else if siglusSelectedOp === 'omv_png'}
          <div class="form-title">OMV → PNG sequence</div>
          <div class="form-group"><label>Fichier OMV :</label><div class="form-row"><input type="text" bind:value={siglusOMVPNGInput} readonly /><button class="btn" on:click={browseSiglusOMVPNGInput}>Select</button></div></div>
          <div class="form-group"><label>Dossier PNG :</label><div class="form-row"><input type="text" bind:value={siglusOMVPNGOutputDir} readonly /><button class="btn" on:click={browseSiglusOMVPNGOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusOMVPNG} disabled={!siglusOMVPNGInput || !siglusOMVPNGOutputDir}>Extract</button>{/if}</div>

        {:else if siglusSelectedOp === 'png_video'}
          <div class="form-title">PNG sequence → OMV/OGV</div>
          <div class="form-group"><label>Dossier PNG :</label><div class="form-row"><input type="text" bind:value={siglusPNGVideoDir} readonly /><button class="btn" on:click={browseSiglusPNGVideoDir}>Select</button></div></div>
          <div class="form-group"><label>Video de sortie :</label><div class="form-row"><input type="text" bind:value={siglusPNGVideoOutput} readonly /><button class="btn" on:click={browseSiglusPNGVideoOutput}>Select</button></div></div>
          <div class="form-group">
            <label>Options :</label>
            <div class="form-row checkbox-row">
              <label class="checkbox-label"><input type="checkbox" bind:checked={siglusPNGVideoAlpha} /> Alpha Siglus</label>
              <span style="min-width:36px;font-size:12px">FPS :</span>
              <input type="text" bind:value={siglusPNGVideoFPS} placeholder="30" style="width:110px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
            </div>
          </div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusPNGVideo} disabled={!siglusPNGVideoDir || !siglusPNGVideoOutput}>Encode</button>{/if}</div>

        {:else if siglusSelectedOp === 'g00_extract'}
          <div class="form-title">G00 → PNG (vaconv)</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={siglusG00Batch} on:change={toggleSiglusG00Batch} /> Batch mode</label></div></div>
          {#if siglusG00Batch}
            <div class="form-group"><label>Dossier G00 :</label><div class="form-row"><input type="text" bind:value={siglusG00Dir} readonly /><button class="btn" on:click={browseSiglusG00}>Select</button></div></div>
            <div class="form-group"><label>Dossier XML :</label><div class="form-row"><input type="text" bind:value={siglusG00XmlPath} readonly placeholder="Auto : dossier de sortie" /><button class="btn" on:click={browseSiglusG00Xml}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>Fichier G00 :</label><div class="form-row"><input type="text" bind:value={siglusG00File} readonly /><button class="btn" on:click={browseSiglusG00}>Select</button></div></div>
            <div class="form-group"><label>Fichier XML :</label><div class="form-row"><input type="text" bind:value={siglusG00XmlPath} readonly placeholder="Auto : même nom que le PNG" /><button class="btn" on:click={browseSiglusG00Xml}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>Dossier de sortie :</label><div class="form-row"><input type="text" bind:value={siglusG00OutputDir} readonly /><button class="btn" on:click={browseSiglusG00OutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusG00Extract} disabled={(siglusG00Batch ? !siglusG00Dir : !siglusG00File) || !siglusG00OutputDir}>Convert</button>{/if}</div>

        {:else if siglusSelectedOp === 'g00_import'}
          <div class="form-title">PNG → G00 (vaconv)</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={siglusPngBatch} on:change={toggleSiglusPngBatch} /> Batch mode</label></div></div>
          {#if siglusPngBatch}
            <div class="form-group"><label>Dossier PNG :</label><div class="form-row"><input type="text" bind:value={siglusPngDir} readonly /><button class="btn" on:click={browseSiglusPng}>Select</button></div></div>
            <div class="form-group"><label>Dossier XML :</label><div class="form-row"><input type="text" bind:value={siglusPngXmlPath} readonly placeholder="Auto : même dossier que les PNG" /><button class="btn" on:click={browseSiglusPngXml}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>Fichier PNG :</label><div class="form-row"><input type="text" bind:value={siglusPngFile} readonly /><button class="btn" on:click={browseSiglusPng}>Select</button></div></div>
            <div class="form-group"><label>Fichier XML :</label><div class="form-row"><input type="text" bind:value={siglusPngXmlPath} readonly placeholder="Auto : même nom que le PNG" /><button class="btn" on:click={browseSiglusPngXml}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>Format G00 :</label><div class="form-row"><select bind:value={siglusG00Format}><option value="auto">Auto</option><option value="0">v0 simple</option><option value="1">v1 compressed</option><option value="2">v2 regions/XML</option></select></div></div>
          <div class="form-group"><label>Dossier de sortie :</label><div class="form-row"><input type="text" bind:value={siglusPngOutputDir} readonly /><button class="btn" on:click={browseSiglusPngOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusG00Import} disabled={(siglusPngBatch ? !siglusPngDir : !siglusPngFile) || !siglusPngOutputDir}>Convert</button>{/if}</div>

        {:else if siglusSelectedOp === 'combine_png'}
          <div class="form-title">Combine PNG</div>
          <div class="form-group"><label>Dossier PNG :</label><div class="form-row"><input type="text" bind:value={siglusCombinePNGDir} readonly /><button class="btn" on:click={browseSiglusCombinePNGDir}>Select</button></div></div>
          <div class="form-group"><label>PNG de sortie :</label><div class="form-row"><input type="text" bind:value={siglusCombinePNGOutput} readonly /><button class="btn" on:click={browseSiglusCombinePNGOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusCombinePNG} disabled={!siglusCombinePNGDir || !siglusCombinePNGOutput}>Combine</button>{/if}</div>

        {:else if siglusSelectedOp === 'script_repack'}
          <div class="form-title">Script Repacker</div>
          <div class="form-group"><label>Script :</label><div class="form-row"><input type="text" bind:value={siglusScriptRepackScript} readonly /><button class="btn" on:click={browseSiglusScriptRepackScript}>Select</button></div></div>
          <div class="form-group"><label>Texte UTF-16 :</label><div class="form-row"><input type="text" bind:value={siglusScriptRepackText} readonly /><button class="btn" on:click={browseSiglusScriptRepackText}>Select</button></div></div>
          <div class="form-group"><label>Script de sortie :</label><div class="form-row"><input type="text" bind:value={siglusScriptRepackOutput} readonly /><button class="btn" on:click={browseSiglusScriptRepackOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusScriptRepack} disabled={!siglusScriptRepackScript || !siglusScriptRepackText || !siglusScriptRepackOutput}>Repack</button>{/if}</div>
        {/if}
      </div>
    </div>

  <!-- OUTILS DIVERS -->
  {:else if activeView === 'outils'}
    <div class="titlebar">
      <span>AIO Key Game Tools — Outils divers</span>
      <button class="titlebar-back" on:click={() => activeView = 'hub'}>← Retour</button>
    </div>
    <div class="content">
      <div class="sidebar">
        <div class="sidebar-title">Outils :</div>
        <div class="sidebar-list">
          {#each outilsOperations as op}
            {#if op.section}
              <div class="sidebar-section">{op.label}</div>
            {:else}
              <div class="sidebar-item" class:active={selectedOp === op.id} on:click={() => selectedOp = op.id}>
                {op.label}
              </div>
            {/if}
          {/each}
        </div>
      </div>
      <div class="form-panel">
        {#if selectedOp === 'dlg_extract'}
          <!-- Dialogue Extract (moved from LuckSystem) -->
          <!-- CONTENT IDENTICAL TO EXISTING dlg_extract BLOCK -->
          <div class="form-title">Extract Dialogues (LuckEngine)</div>
          <div class="form-hint" style="margin-bottom:10px">Extrait les lignes MESSAGE et LOG_BEGIN des scripts LuckEngine (.txt) vers un fichier TSV.</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtBatch} on:change={toggleDlgExtBatch} /> Batch mode</label></div></div>
          <div class="form-group"><label>{dlgExtBatch ? 'Dossier scripts :' : 'Fichier script :'}</label><div class="form-row"><input type="text" bind:value={dlgExtInput} readonly /><button class="btn" on:click={browseDlgExtInput}>Select</button></div></div>
          <div class="form-group"><label>Colonnes :</label><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang1} /> 1</label><label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang2} /> 2</label><label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang3} /> 3</label><label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang4} /> 4</label></div></div>
          <div class="form-group"><label>{dlgExtBatch ? 'Dossier sortie :' : 'Fichier TSV sortie :'}</label><div class="form-row"><input type="text" bind:value={dlgExtOutput} readonly /><button class="btn" on:click={browseDlgExtOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startDlgExtract} disabled={!dlgExtInput || !dlgExtOutput}>Start Extract</button>{/if}</div>
        {:else if selectedOp === 'dlg_import'}
          <div class="form-title">Import Dialogues (LuckEngine)</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={dlgImpBatch} on:change={toggleDlgImpBatch} /> Batch mode</label></div></div>
          <div class="form-group"><label>Colonne cible :</label><div class="form-row"><select bind:value={dlgImpTargetCol}><option value={1}>Lang 1</option><option value={2}>Lang 2</option><option value={3}>Lang 3</option><option value={4}>Lang 4</option></select></div></div>
          <div class="form-group"><label>{dlgImpBatch ? 'Dossier scripts :' : 'Fichier script :'}</label><div class="form-row"><input type="text" bind:value={dlgImpScript} readonly /><button class="btn" on:click={browseDlgImpScript}>Select</button></div></div>
          <div class="form-group"><label>{dlgImpBatch ? 'Dossier TSV :' : 'Fichier TSV :'}</label><div class="form-row"><input type="text" bind:value={dlgImpTsv} readonly /><button class="btn" on:click={browseDlgImpTsv}>Select</button></div></div>
          <div class="form-group"><label>{dlgImpBatch ? 'Dossier sortie :' : 'Fichier sortie :'}</label><div class="form-row"><input type="text" bind:value={dlgImpOutput} readonly /><button class="btn" on:click={browseDlgImpOutput}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startDlgImport} disabled={!dlgImpScript || !dlgImpTsv || !dlgImpOutput}>Start Import</button>{/if}</div>
        {:else}
          <div class="form-title">Outils divers</div>
          <div class="form-hint">Sélectionnez un outil dans le menu de gauche.</div>
        {/if}
      </div>
    </div>

  <!-- RLDEV 2026 -->
  {:else if activeView === 'rldev'}
    <div class="titlebar">
      <span>RLdev 2026 — v1.3 (Go port)</span>
      <button class="titlebar-back" on:click={() => activeView = 'hub'}>← Retour</button>
    </div>
    <div class="content">
      <div class="sidebar">
        <div class="sidebar-title">RLdev 2026 :</div>
        <div class="sidebar-list">
          {#each rldevOperations as op}
            {#if op.section}
              <div class="sidebar-section">{op.label}</div>
            {:else}
              <div class="sidebar-item" class:active={rldevSelectedOp === op.id} on:click={() => rldevSelectedOp = op.id}>
                {op.label}
              </div>
            {/if}
          {/each}
        </div>
      </div>
      <div class="form-panel">
        <!-- KPRL EXTRACT (disassemble) -->
        {#if rldevSelectedOp === 'kprl_disasm'}
          <div class="form-title">2 — Extract SEEN.txt</div>
          <div class="form-hint" style="margin-bottom:10px">Désassemble une archive SEEN.txt en scripts Kepago (.org + .utf/.sjs).</div>
          <div class="form-group"><label>SEEN.txt :</label><div class="form-row"><input type="text" bind:value={rlSeenFile} readonly /><button class="btn" on:click={browseRlSeen}>Select</button></div></div>
          <div class="form-group"><label>KFN file :</label><div class="form-row"><input type="text" bind:value={rlKfnFile} readonly placeholder="Auto : ./KFN/reallive.kfn" /><button class="btn" on:click={browseRlKfn}>Select</button></div></div>
          <div class="form-group"><label>Encodage sortie :</label><div class="form-row"><select bind:value={rlEncoding}><option value="UTF-8">UTF-8</option><option value="CP932">CP932 / Shift-JIS</option><option value="EUC-JP">EUC-JP</option></select></div></div>
          <div class="form-group"><label>Game ID (-G, optionnel) :</label><div class="form-row"><input type="text" bind:value={rlGameId} list="rl-game-id-options" placeholder="ex: KUDO (Kud Wafter 18+)" on:blur={normalizeRlGameId} on:change={normalizeRlGameId} /></div><datalist id="rl-game-id-options">{#each gameIdOptions as game}<option value={game.id} label={game.title}></option>{/each}</datalist></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlDebugInfo} /> Sources debug RealLive (-g / #line)</label></div><div class="form-hint">Pour F3/F5/O uniquement ; garder décoché pour les sources de traduction.</div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startRlDisasm} disabled={!rlSeenFile || !rlKfnFile || !rlOutputDir}>Start Extract</button>{/if}</div>

        {:else if rldevSelectedOp === 'kprl_extract'}
          <div class="form-title">Advanced: extract bytecode</div>
          <div class="form-hint" style="margin-bottom:10px">Décompresse/décrypte les scénarios en fichiers .rl, sans produire de scripts .org.</div>
          <div class="form-group"><label>SEEN.txt :</label><div class="form-row"><input type="text" bind:value={rlSeenFile} readonly /><button class="btn" on:click={browseRlSeen}>Select</button></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startRlExtract} disabled={!rlSeenFile || !rlOutputDir}>Extract</button>{/if}</div>

        <!-- KPRL ARCHIVE (pack into SEEN.txt) -->
        {:else if rldevSelectedOp === 'kprl_archive'}
          <div class="form-title">4 — Rebuild SEEN.txt</div>
          <div class="form-hint" style="margin-bottom:10px">Assemble des fichiers .TXT/.avg compilés dans une archive SEEN.txt.</div>
          <div class="form-group"><label>Input folder (.TXT/.avg) :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-group"><label>Original/template SEEN.txt :</label><div class="form-row"><input type="text" bind:value={rlTemplateSeenFile} readonly placeholder="Optionnel, requis pour Clannad Steam" /><button class="btn" on:click={browseRlTemplateSeen}>Select</button></div></div>
          <div class="form-group"><label>Output SEEN.txt :</label><div class="form-row"><input type="text" bind:value={rlSeenFile} readonly /><button class="btn" on:click={browseRlSeenSave}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startRlArchive} disabled={!rlSeenFile || !rlOutputDir}>Rebuild Archive</button>{/if}</div>

        <!-- KPRL LIST -->
        {:else if rldevSelectedOp === 'kprl_list'}
          <div class="form-title">1 — List SEEN.txt</div>
          <div class="form-group"><label>SEEN.txt :</label><div class="form-row"><input type="text" bind:value={rlSeenFile} readonly /><button class="btn" on:click={browseRlSeen}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startRlList} disabled={!rlSeenFile}>List Contents</button>{/if}</div>

        <!-- RLC COMPILE -->
        {:else if rldevSelectedOp === 'rlc_compile'}
          <div class="form-title">3 — Compile .org / .ke / .avg → .TXT</div>
          <div class="form-hint" style="margin-bottom:10px">Compile les scripts RealLive/Kepago ou AVG32 d'un dossier en mode batch.</div>
          <div class="form-group">
            <div class="form-row checkbox-row">
              <label class="checkbox-label"><input type="checkbox" bind:checked={rlCompileBatch} on:change={toggleCompileBatch} /> Batch mode</label>
            </div>
          </div>
          {#if rlCompileBatch}
            <div class="form-group"><label>Input folder (.org/.ke/.avg) :</label><div class="form-row"><input type="text" bind:value={rlOrgDir} readonly /><button class="btn" on:click={browseRlOrg}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>Script .org / .ke / .avg :</label><div class="form-row"><input type="text" bind:value={rlOrgFile} readonly /><button class="btn" on:click={browseRlOrg}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>KFN file :</label><div class="form-row"><input type="text" bind:value={rlKfnFile} readonly placeholder="Auto : ./KFN/reallive.kfn" /><button class="btn" on:click={browseRlKfn}>Select</button></div></div>
          <div class="form-group"><label>GAMEEXE.INI (optionnel) :</label><div class="form-row"><input type="text" bind:value={rlGameexe} readonly /><button class="btn" on:click={browseRlGameexe}>Select</button></div></div>
          <div class="form-group"><label>Interpréteur RealLive / Steam (optionnel) :</label><div class="form-row"><input type="text" bind:value={rlInterpreter} readonly /><button class="btn" on:click={browseRlInterpreter}>Select</button></div><div class="form-hint">Auto si GAMEEXE.INI pointe vers un dossier contenant RealLive.exe ou SiglusEngine_Steam.exe.</div></div>
          <div class="form-group"><label>Version RealLive (auto si vide) :</label><div class="form-row"><input type="text" bind:value={rlTargetVersion} on:input={markRlTargetVersionManual} list="rl-target-version-options" placeholder="ex: 1.2.3.5 pour CLANNAD 2004" /></div><datalist id="rl-target-version-options"><option value="1.2.3.5"></option><option value="1.2.5.5"></option><option value="1.2.7.0"></option><option value="1.2.9.5"></option><option value="1.3.1.0"></option><option value="1.4.0.5"></option></datalist><div class="form-hint">Rempli automatiquement depuis l'exe RealLive/Steam détecté.</div></div>
          <div class="form-group"><label>Encodage source :</label><div class="form-row"><select bind:value={rlEncoding}><option value="UTF-8">UTF-8</option><option value="CP932">CP932 / Shift-JIS</option><option value="EUC-JP">EUC-JP</option></select></div></div>
          <div class="form-group"><label>Transformation sortie :</label><div class="form-row"><select bind:value={rlOutputTransform}><option value="NONE">NONE / CP932 original</option><option value="WESTERN">WESTERN / CP1252</option><option value="CHINESE">CHINESE</option><option value="KOREAN">KOREAN</option></select></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlForceTransform} /> Force transform</label></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startRlCompile} disabled={(rlCompileBatch ? !rlOrgDir : !rlOrgFile) || !rlOutputDir}>Compile</button>{/if}</div>

        {:else if rldevSelectedOp === 'rlc_org_text'}
          <div class="form-title">Extract text ORG</div>
          <div class="form-group"><label>Mode :</label><div class="form-row"><select bind:value={rlOrgTextMode}><option value="export">Export .utf</option><option value="import">Import .utf</option></select></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlOrgTextBatch} on:change={toggleOrgTextBatch} /> Batch mode</label></div></div>
          {#if rlOrgTextBatch}
            <div class="form-group"><label>ORG/KE folder :</label><div class="form-row"><input type="text" bind:value={rlOrgTextDir} readonly /><button class="btn" on:click={browseRlOrgText}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>Script .org / .ke :</label><div class="form-row"><input type="text" bind:value={rlOrgTextFile} readonly /><button class="btn" on:click={browseRlOrgText}>Select</button></div></div>
          {/if}
          {#if rlOrgTextMode === 'import'}
            {#if rlOrgTextBatch}
              <div class="form-group"><label>UTF folder :</label><div class="form-row"><input type="text" bind:value={rlOrgTextUtfDir} readonly /><button class="btn" on:click={browseRlOrgTextUtf}>Select</button></div></div>
            {:else}
              <div class="form-group"><label>UTF file :</label><div class="form-row"><input type="text" bind:value={rlOrgTextUtfFile} readonly /><button class="btn" on:click={browseRlOrgTextUtf}>Select</button></div></div>
            {/if}
          {/if}
          <div class="form-group"><label>Encodage source :</label><div class="form-row"><select bind:value={rlEncoding}><option value="UTF-8">UTF-8</option><option value="CP932">CP932 / Shift-JIS</option><option value="EUC-JP">EUC-JP</option></select></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startOrgText} disabled={(rlOrgTextBatch ? !rlOrgTextDir : !rlOrgTextFile) || (rlOrgTextMode === 'import' && (rlOrgTextBatch ? !rlOrgTextUtfDir : !rlOrgTextUtfFile)) || !rlOutputDir}>{rlOrgTextMode === 'import' ? 'Import ORG' : 'Export UTF'}</button>{/if}</div>

        <!-- G00 EXTRACT -->
        {:else if rldevSelectedOp === 'g00_extract'}
          <div class="form-title">G00 → PNG</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlG00Batch} on:change={toggleG00Batch} /> Batch mode</label></div></div>
          {#if rlG00Batch}
            <div class="form-group"><label>G00 folder :</label><div class="form-row"><input type="text" bind:value={rlG00Dir} readonly /><button class="btn" on:click={browseRlG00}>Select</button></div></div>
            <div class="form-group"><label>XML folder (optionnel) :</label><div class="form-row"><input type="text" bind:value={rlG00XmlPath} readonly placeholder="Auto : output folder" /><button class="btn" on:click={browseRlG00Xml}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>G00 file :</label><div class="form-row"><input type="text" bind:value={rlG00File} readonly /><button class="btn" on:click={browseRlG00}>Select</button></div></div>
            <div class="form-group"><label>XML file (optionnel) :</label><div class="form-row"><input type="text" bind:value={rlG00XmlPath} readonly placeholder="Auto : same output basename" /><button class="btn" on:click={browseRlG00Xml}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startG00Extract} disabled={(rlG00Batch ? !rlG00Dir : !rlG00File) || !rlOutputDir}>Convert</button>{/if}</div>

        <!-- G00 IMPORT -->
        {:else if rldevSelectedOp === 'g00_import'}
          <div class="form-title">PNG → G00</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlPngBatch} on:change={togglePngBatch} /> Batch mode</label></div></div>
          {#if rlPngBatch}
            <div class="form-group"><label>PNG folder :</label><div class="form-row"><input type="text" bind:value={rlPngDir} readonly /><button class="btn" on:click={browseRlPng}>Select</button></div></div>
            <div class="form-group"><label>XML folder (optionnel) :</label><div class="form-row"><input type="text" bind:value={rlPngXmlPath} readonly placeholder="Auto : same PNG folder" /><button class="btn" on:click={browseRlPngXml}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>PNG file :</label><div class="form-row"><input type="text" bind:value={rlPngFile} readonly /><button class="btn" on:click={browseRlPng}>Select</button></div></div>
            <div class="form-group"><label>XML file (optionnel) :</label><div class="form-row"><input type="text" bind:value={rlPngXmlPath} readonly placeholder="Auto : same PNG basename" /><button class="btn" on:click={browseRlPngXml}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>G00 format :</label><div class="form-row"><select bind:value={rlG00Format}><option value="auto">Auto</option><option value="0">v0 simple</option><option value="1">v1 compressed</option><option value="2">v2 regions/XML</option></select></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startG00Import} disabled={(rlPngBatch ? !rlPngDir : !rlPngFile) || !rlOutputDir}>Convert</button>{/if}</div>

        <!-- GAN TO XML -->
        {:else if rldevSelectedOp === 'gan_to_xml'}
          <div class="form-title">GAN → XML</div>
          <div class="form-group"><label>GAN file :</label><div class="form-row"><input type="text" bind:value={rlGanFile} readonly /><button class="btn" on:click={browseRlGan}>Select</button></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startGanToXml} disabled={!rlGanFile || !rlOutputDir}>Convert</button>{/if}</div>

        <!-- GAN FROM XML -->
        {:else if rldevSelectedOp === 'gan_from_xml'}
          <div class="form-title">XML → GAN</div>
          <div class="form-group"><label>GANXML file :</label><div class="form-row"><input type="text" bind:value={rlGanFile} readonly /><button class="btn" on:click={browseRlGan}>Select</button></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startGanFromXml} disabled={!rlGanFile || !rlOutputDir}>Convert</button>{/if}</div>

        {:else if rldevSelectedOp === 'nwa_audio'}
          <div class="form-title">NWA → MP3/WAV</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlNwaBatch} on:change={toggleNwaBatch} /> Batch mode</label></div></div>
          {#if rlNwaBatch}
            <div class="form-group"><label>NWA folder :</label><div class="form-row"><input type="text" bind:value={rlNwaDir} readonly /><button class="btn" on:click={browseRlNwa}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>NWA file :</label><div class="form-row"><input type="text" bind:value={rlNwaFile} readonly /><button class="btn" on:click={browseRlNwa}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>Output format :</label><div class="form-row"><select bind:value={rlAudioFormat}><option value="mp3">MP3</option><option value="wav">WAV</option></select></div></div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startNwaAudio} disabled={(rlNwaBatch ? !rlNwaDir : !rlNwaFile) || !rlOutputDir}>Convert</button>{/if}</div>

        {:else if rldevSelectedOp === 'dat_to_json'}
          <div class="form-title">CGM/TCC → JSON</div>
          <div class="form-hint" style="margin-bottom:10px">Exporte mode.cgm ou tcdata.tcc vers JSON. TCC expose les courbes RGB ; CGM expose les entrées nom + index quand la table est standard.</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlDatBatch} on:change={toggleDatBatch} /> Batch mode</label></div></div>
          {#if rlDatBatch}
            <div class="form-group"><label>CGM/TCC folder :</label><div class="form-row"><input type="text" bind:value={rlDatDir} readonly /><button class="btn" on:click={browseRlDat}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>CGM/TCC file :</label><div class="form-row"><input type="text" bind:value={rlDatFile} readonly /><button class="btn" on:click={browseRlDat}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startDatToJson} disabled={(rlDatBatch ? !rlDatDir : !rlDatFile) || !rlOutputDir}>Export JSON</button>{/if}</div>

        {:else if rldevSelectedOp === 'dat_from_json'}
          <div class="form-title">JSON → CGM/TCC</div>
          <div class="form-hint" style="margin-bottom:10px">Reconstruit le fichier binaire à partir du champ type du JSON.</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlDatJsonBatch} on:change={toggleDatJsonBatch} /> Batch mode</label></div></div>
          {#if rlDatJsonBatch}
            <div class="form-group"><label>JSON folder :</label><div class="form-row"><input type="text" bind:value={rlDatJsonDir} readonly /><button class="btn" on:click={browseRlDatJson}>Select</button></div></div>
          {:else}
            <div class="form-group"><label>DAT JSON file :</label><div class="form-row"><input type="text" bind:value={rlDatJsonFile} readonly /><button class="btn" on:click={browseRlDatJson}>Select</button></div></div>
          {/if}
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startDatFromJson} disabled={(rlDatJsonBatch ? !rlDatJsonDir : !rlDatJsonFile) || !rlOutputDir}>Rebuild DAT</button>{/if}</div>

        {:else if rldevSelectedOp === 'save_editor'}
          <div class="form-title">RealLive save editor</div>
          <div class="form-group"><label>Profile :</label><div class="form-row"><select bind:value={rlSaveProfile}>{#each saveProfiles as profile}<option value={profile.id}>{profile.label}</option>{/each}</select><button class="btn" on:click={applySaveProfile}>Apply</button></div></div>
          <div class="form-group"><label>Map target :</label><div class="form-row"><input type="text" bind:value={rlSaveMapPath} readonly placeholder="File or folder" /><button class="btn" on:click={browseRlSaveMapFile}>File</button><button class="btn" on:click={browseRlSaveMapDir}>Folder</button></div></div>
          <div class="form-group"><label>Save file :</label><div class="form-row"><input type="text" bind:value={rlSaveFile} readonly /><button class="btn" on:click={browseRlSave}>Select</button></div></div>
          <div class="form-group"><label>Compare with :</label><div class="form-row"><input type="text" bind:value={rlSaveCompareFile} readonly /><button class="btn" on:click={browseRlSaveCompare}>Select</button></div></div>
          <div class="form-group"><label>Variables :</label><div class="form-row"><input type="text" bind:value={rlSaveRefs} placeholder="intG[0] seen[100] dword[1]" /></div></div>
          <div class="form-group"><label>Assignations :</label><div class="form-row"><input type="text" bind:value={rlSaveAssignments} placeholder="intG[30]=0 seen[100]=0" /></div></div>
          <div class="form-group"><label>Text export :</label><div class="form-row"><input type="text" bind:value={rlSaveTextFile} readonly /><button class="btn" on:click={browseRlSaveTextOutput}>Export path</button><button class="btn" on:click={browseRlSaveTextInput}>Build input</button></div></div>
          <div class="form-group"><label>Rebuilt save :</label><div class="form-row"><input type="text" bind:value={rlSaveBuildOutput} readonly /><button class="btn" on:click={browseRlSaveBuildOutput}>Select</button></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveBackup} /> Backup before write</label><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveLossless} /> Lossless export</label></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveMapJson} /> Map JSON</label><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveDoctorJson} /> Doctor JSON</label><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveDiffJson} /> Diff JSON</label><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveDumpAll} /> Dump all intG values</label><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveDumpJson} /> Dump JSON</label></div></div>
          <div class="form-actions">
            {#if running}
              <span class="running-indicator"></span> Running...
            {:else}
              <button class="btn" on:click={startSaveMap} disabled={!rlSaveMapPath && !rlSaveFile}>Map</button>
              <button class="btn" on:click={startSaveDoctor} disabled={!rlSaveMapPath && !rlSaveFile}>Doctor</button>
              <button class="btn" on:click={startSaveInfo} disabled={!rlSaveFile}>Info</button>
              <button class="btn" on:click={startSaveDiff} disabled={!rlSaveFile || !rlSaveCompareFile}>Diff</button>
              <button class="btn" on:click={startSaveGet} disabled={!rlSaveFile || !rlSaveRefs.trim()}>Get</button>
              <button class="btn" on:click={startSaveDump} disabled={!rlSaveFile}>Dump</button>
              <button class="btn" on:click={startSaveExport} disabled={!rlSaveFile || !rlSaveTextFile}>Export</button>
              <button class="btn" on:click={startSaveBuild} disabled={!rlSaveTextFile || !rlSaveBuildOutput}>Build</button>
              <button class="btn btn-primary" on:click={startSaveSet} disabled={!rlSaveFile || !rlSaveAssignments.trim()}>Set</button>
            {/if}
          </div>

        {:else if rldevSelectedOp === 'babel_runtime'}
          <div class="form-title">Babel runtime setup</div>
          <div class="form-hint" style="margin-bottom:10px">Copie rlBabel dans le dossier du jeu, ajoute la map de version si elle existe, et peut préparer GAMEEXE.INI.</div>
          <div class="form-group"><label>BABEL folder :</label><div class="form-row"><input type="text" bind:value={rlBabelRoot} readonly placeholder="Auto : ...\ResCODEX\Rldev2026-go\BABEL" /><button class="btn" on:click={browseRlBabelRoot}>Select</button></div></div>
          <div class="form-group"><label>Game folder :</label><div class="form-row"><input type="text" bind:value={rlBabelGameDir} readonly /><button class="btn" on:click={browseRlBabelGameDir}>Select</button></div></div>
          <div class="form-group"><label>RealLive version :</label><div class="form-row"><input type="text" bind:value={rlBabelVersion} list="babel-version-options" placeholder="ex: 1.2.3.5" /></div><datalist id="babel-version-options"><option value="1.2.3.5"></option><option value="1.2.5.5"></option><option value="1.2.6.4"></option><option value="1.2.7.0"></option><option value="1.2.9.5"></option><option value="1.3.1.0"></option><option value="1.3.2.0"></option><option value="1.4.0.5"></option></datalist></div>
          <div class="form-group"><label>DLL :</label><div class="form-row"><select bind:value={rlBabelDllMode}><option value="auto">Auto by version</option><option value="old">rlBabelF.dll / RealLive 1.2.x</option><option value="new">rlBabel.dll / RealLive 1.2.5+</option></select></div></div>
          <div class="form-group"><label>#NAME_ENC :</label><div class="form-row"><select bind:value={rlBabelNameEnc}><option value="western">Western</option><option value="chinese">Chinese</option><option value="korean">Korean</option><option value="none">No change</option></select></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlBabelUpdateGameexe} /> Update GAMEEXE.INI</label></div><div class="form-hint">Une sauvegarde .bak est créée avant modification.</div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startBabelRuntime} disabled={!rlBabelRoot || !rlBabelGameDir}>Prepare Runtime</button>{/if}</div>

        {:else if rldevSelectedOp === 'babel_header'}
          <div class="form-title">Babel global.kh helper</div>
          <div class="form-hint" style="margin-bottom:10px">Crée un global.kh minimal pour activer la lineation dynamique et charger le module rlBabel.</div>
          <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={rlOutputDir} readonly /><button class="btn" on:click={browseRlOutputDir}>Select</button></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlBabelGlosses} /> Enable glosses</label></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startBabelHeader} disabled={!rlOutputDir}>Create global.kh</button>{/if}</div>
        {/if}
      </div>
    </div>

  <!-- LUCKSYSTEM -->
  {:else if activeView === 'lucksystem'}
  <div class="titlebar">
    <span>LuckSystem 2.3.2 - Yoremi fork v3.30 GUI</span>
    <div style="display:flex;align-items:center;gap:10px">
      <label class="ui-language">
        <span>{t('Interface', 'Interface')}</span>
        <select value={uiLanguage} on:change={(e) => setUiLanguage(e.target.value)}>
          <option value="fr">Français</option>
          <option value="en">English</option>
        </select>
      </label>
      <span class="titlebar-path" on:click={locateLuckSystem} title={t('Cliquer pour modifier', 'Click to change')}>
        {#if lsPath}📁 {lsPath}{:else}⚠ {t('lucksystem.exe introuvable — cliquer pour le localiser', 'lucksystem.exe not found — click to locate')}{/if}
      </span>
      <button class="titlebar-back" on:click={() => activeView = 'hub'}>← {t('Retour', 'Back')}</button>
    </div>
  </div>

  <div class="content">
    <!-- LEFT SIDEBAR -->
    <div class="sidebar">
      <div class="sidebar-title">Select option:</div>
      <div class="sidebar-list">
        {#each operations as op}
          {#if lucaMenuDllAvailable || (op.id !== '_s3c' && op.id !== 'luca_menu_dll')}
            {#if op.section}
              <div class="sidebar-section">{op.label}</div>
            {:else}
              <div class="sidebar-item" class:active={selectedOp === op.id} class:disabled={op.disabled} on:click={() => selectOp(op)}>
                {op.label}
              </div>
            {/if}
          {/if}
        {/each}
      </div>
    </div>

    <!-- RIGHT FORM PANEL -->
    <div class="form-panel">

      <!-- SCRIPT DECOMPILE -->
      {#if selectedOp === 'decompile'}
        <div class="form-title">Script Decompile</div>
        <div class="form-group"><label>SCRIPT.PAK file:</label><div class="form-row"><input type="text" bind:value={pakFile} readonly /><button class="btn" on:click={browsePak}>Select</button></div></div>
        {#if gamePresets.length > 0}
        <div class="form-group"><label>Game preset:</label><div class="form-row"><select value={selectedPreset} on:change={(e) => applyPreset(e.target.value)}><option value="">— Manual —</option>{#each gamePresets as p}<option value={p.name}>{p.name}{p.pluginFile ? ' (plugin)' : ''}</option>{/each}</select></div><div class="form-hint">Auto-fills Opcode, Plugin and Game from data/ folder</div></div>
        {/if}
        <div class="form-group"><label>Opcode file (.txt):</label><div class="form-row"><input type="text" bind:value={opcodeFile} placeholder="e.g. data/AIR.txt" readonly /><button class="btn" on:click={browseOpcode}>Select</button></div></div>
        <div class="form-group"><label>Plugin file (.py):</label><div class="form-row"><input type="text" bind:value={pluginFile} placeholder="e.g. data/AIR.py" readonly /><button class="btn" on:click={browsePlugin}>Select</button></div></div>
        <div class="form-group"><label>Charset:</label><div class="form-row"><select bind:value={charsetVal}><option value="UTF-8">UTF-8</option><option value="ShiftJIS">Shift-JIS</option><option value="GBK">GBK</option></select></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={outputDir} readonly /><button class="btn" on:click={browseOutputDir}>Select</button></div><div class="form-hint">A SCRIPT.PAK subfolder will be created automatically inside</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startDecompile} disabled={!pakFile || !outputDir}>Start Decompile</button>{/if}</div>

      <!-- SCRIPT COMPILE -->
      {:else if selectedOp === 'compile'}
        <div class="form-title">Script Compile</div>
        <div class="form-group"><label>Original SCRIPT.PAK:</label><div class="form-row"><input type="text" bind:value={pakFile} readonly /><button class="btn" on:click={browsePak}>Select</button></div><div class="form-hint">The original unmodified SCRIPT.PAK</div></div>
        {#if gamePresets.length > 0}
        <div class="form-group"><label>Game preset:</label><div class="form-row"><select value={selectedPreset} on:change={(e) => applyPreset(e.target.value)}><option value="">— Manual —</option>{#each gamePresets as p}<option value={p.name}>{p.name}{p.pluginFile ? ' (plugin)' : ''}</option>{/each}</select></div><div class="form-hint">Auto-fills Opcode, Plugin and Game from data/ folder</div></div>
        {/if}
        <div class="form-group"><label>Opcode file (.txt):</label><div class="form-row"><input type="text" bind:value={opcodeFile} readonly /><button class="btn" on:click={browseOpcode}>Select</button></div></div>
        <div class="form-group"><label>Plugin file (.py):</label><div class="form-row"><input type="text" bind:value={pluginFile} readonly /><button class="btn" on:click={browsePlugin}>Select</button></div></div>
        <div class="form-group"><label>Charset:</label><div class="form-row"><select bind:value={charsetVal}><option value="UTF-8">UTF-8</option><option value="ShiftJIS">Shift-JIS</option><option value="GBK">GBK</option></select></div></div>
        <div class="form-group"><label>Translated scripts folder:</label><div class="form-row"><input type="text" bind:value={importDir} readonly /><button class="btn" on:click={browseImportDir}>Select</button></div><div class="form-hint">Select the parent folder containing SCRIPT.PAK (e.g. TRAD), not SCRIPT.PAK itself</div></div>
        <div class="form-group"><label>Output PAK file:</label><div class="form-row"><input type="text" bind:value={outputPak} readonly /><button class="btn" on:click={browseOutputPak}>Select</button></div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startCompile} disabled={!pakFile || !importDir || !outputPak}>Start Compile</button>{/if}</div>

      <!-- SIGLUS -> LUCA BRIDGE -->
      {:else if selectedOp === 'siglus_luca'}
        <div class="form-title">Siglus -> Luca Script Bridge</div>
        <div class="form-hint" style="margin-bottom:10px">
          Importe des lignes traduites depuis des exports Siglus dans des scripts Luca décompilés. Les lignes Luca-only et les découpages à vérifier sont exportés en TSV dans le dossier de sortie.
        </div>
        <div class="form-group"><label>Luca scripts folder:</label><div class="form-row"><input type="text" bind:value={siglusLucaLucaDir} readonly placeholder="SCRIPT.PAK decompiled folder" /><button class="btn" on:click={browseSiglusLucaLucaDir}>Select</button></div><div class="form-hint">Dossier contenant les scripts Luca .txt à patcher.</div></div>
        <div class="form-group"><label>Siglus Full folder:</label><div class="form-row"><input type="text" bind:value={siglusLucaSiglusDir} readonly placeholder="TRAD-silgus\Full" /><button class="btn" on:click={browseSiglusLucaSiglusDir}>Select</button></div><div class="form-hint">Dossier contenant les exports Siglus .ss.txt avec source et traduction.</div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={siglusLucaOutput} readonly placeholder="Luca_from_Siglus_FR" /><button class="btn" on:click={browseSiglusLucaOutput}>Select</button></div><div class="form-hint">Les scripts patchés, <code>hd_candidates.tsv</code> et <code>review.tsv</code> seront écrits ici.</div></div>
        <div class="form-group">
          <label>Target language column:</label>
          <div class="form-row">
            <select bind:value={siglusLucaTargetCol}>
              <option value={1}>Lang 1</option>
              <option value={2}>Lang 2</option>
              <option value={3}>Lang 3</option>
              <option value={4}>Lang 4</option>
            </select>
          </div>
          <div class="form-hint">Colonne Luca à remplacer. Par défaut : Lang 2.</div>
        </div>
        <div class="form-actions">
          {#if running}<span class="running-indicator"></span> Running...
          {:else}<button class="btn btn-primary" on:click={startSiglusLucaBridge}
            disabled={!siglusLucaLucaDir || !siglusLucaSiglusDir || !siglusLucaOutput}>
            Start Bridge
          </button>{/if}
        </div>

      <!-- PAK CG EXTRACT -->
      {:else if selectedOp === 'pak_cg_extract'}
        <div class="form-title">PAK (CG) — Extract</div>
        <div class="form-group"><label>PAK file (CG) :</label><div class="form-row"><input type="text" bind:value={pakExtSource} readonly /><button class="btn" on:click={browsePakExtSource}>Select</button></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={pakExtOutput} readonly /><button class="btn" on:click={browsePakExtOutput}>Select</button></div><div class="form-hint">Le fichier liste <code>&lt;NOM&gt;_list.txt</code> sera généré automatiquement dans ce dossier</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startPakExtract} disabled={!pakExtSource || !pakExtOutput}>Start Extract</button>{/if}</div>

      <!-- BGMOVIE EXTRACT -->
      {:else if selectedOp === 'bgmovie_extract'}
        <div class="form-title">BGMOVIE.PAK — Video Extract</div>
        <div class="form-group"><label>BGMOVIE.PAK file:</label><div class="form-row"><input type="text" bind:value={bgMoviePak} readonly /><button class="btn" on:click={browseBgMoviePak}>Select</button></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={bgMovieOutput} readonly /><button class="btn" on:click={browseBgMovieOutput}>Select</button></div><div class="form-hint">Creates a folder named after the PAK with raw MVT files and a <code>webm</code> subfolder.</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startBgMovieExtract} disabled={!bgMoviePak || !bgMovieOutput}>Extract Videos</button>{/if}</div>

      <!-- MUSIC PAK EXTRACT -->
      {:else if selectedOp === 'music_extract'}
        <div class="form-title">MUSIC.PAK — Audio Extract</div>
        <div class="form-group"><label>MUSIC.PAK file:</label><div class="form-row"><input type="text" bind:value={musicPak} readonly /><button class="btn" on:click={browseMusicPak}>Select</button></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={musicOutput} readonly /><button class="btn" on:click={browseMusicOutput}>Select</button></div><div class="form-hint">Creates native <code>ogg</code> files and a list file usable by PAK replace.</div></div>
        <div class="form-group"><label>Conversion:</label><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={musicToMp3} /> Also create MP3 copies</label></div><div class="form-hint">Uses FFmpeg from PATH when MP3 conversion is enabled.</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startMusicExtract} disabled={!musicPak || !musicOutput}>Extract Music</button>{/if}</div>

      <!-- VOICE PAK EXTRACT -->
      {:else if selectedOp === 'voice_extract'}
        <div class="form-title">VOICE / SYSVOICE.PAK — Audio Extract</div>
        <div class="form-group"><label>VOICE/SYSVOICE PAK file:</label><div class="form-row"><input type="text" bind:value={voicePak} readonly /><button class="btn" on:click={browseVoicePak}>Select</button></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={voiceOutput} readonly /><button class="btn" on:click={browseVoiceOutput}>Select</button></div><div class="form-hint">Creates native <code>ogg</code> files and a list file usable by PAK replace.</div></div>
        <div class="form-group"><label>Conversion:</label><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={voiceToMp3} /> Also create MP3 copies</label></div><div class="form-hint">Uses FFmpeg from PATH when MP3 conversion is enabled.</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startVoiceExtract} disabled={!voicePak || !voiceOutput}>Extract Voices</button>{/if}</div>

      <!-- AUDIO CONVERT -->
      {:else if selectedOp === 'audio_convert'}
        <div class="form-title">Ogg / MP3 — Convert Folder</div>
        <div class="form-group"><label>Mode:</label><div class="form-row"><select bind:value={audioConvDirection}><option value="mp3">Native Ogg -> MP3</option><option value="native">MP3 -> Native Ogg</option></select></div></div>
        <div class="form-group"><label>Input folder:</label><div class="form-row"><input type="text" bind:value={audioConvInput} readonly /><button class="btn" on:click={browseAudioConvInput}>Select</button></div><div class="form-hint">Select the folder containing <code>.ogg</code> or <code>.mp3</code> files.</div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={audioConvOutput} readonly /><button class="btn" on:click={browseAudioConvOutput}>Select</button></div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startAudioConvert} disabled={!audioConvInput || !audioConvOutput}>Convert Audio</button>{/if}</div>

      <!-- PAK CG REPLACE -->
      {:else if selectedOp === 'pak_cg_replace'}
        <div class="form-title">PAK (CG) — Replace</div>
        <div class="form-group"><label>Original PAK file:</label><div class="form-row"><input type="text" bind:value={pakRepSource} readonly /><button class="btn" on:click={browsePakRepSource}>Select</button></div></div>
        <div class="form-group">
          <label>Mode d'entrée :</label>
          <div class="form-row checkbox-row" style="margin-bottom:6px">
            <label class="checkbox-label"><input type="radio" bind:group={pakRepUseList} value={true} /> Fichier liste (<code>*_list.txt</code>)</label>
            <label class="checkbox-label"><input type="radio" bind:group={pakRepUseList} value={false} /> Dossier de fichiers</label>
          </div>
          {#if pakRepUseList}
            <div class="form-row"><input type="text" bind:value={pakRepListFile} placeholder="SYSCG_list.txt" readonly /><button class="btn" on:click={browsePakRepListFile}>Select</button></div>
            <div class="form-hint">Fichier liste généré lors de l'extraction (ex : SYSCG_list.txt)</div>
          {:else}
            <div class="form-row"><input type="text" bind:value={pakRepInput} readonly /><button class="btn" on:click={browsePakRepInput}>Select</button></div>
            <div class="form-hint">Dossier contenant les fichiers modifiés à réinjecter</div>
          {/if}
        </div>
        <div class="form-group"><label>Output PAK:</label><div class="form-row"><input type="text" bind:value={pakRepOutput} readonly /><button class="btn" on:click={browsePakRepOutput}>Select</button></div></div>
        <div class="form-actions">
          {#if running}
            <span class="running-indicator"></span> Running...
          {:else}
            <button class="btn btn-primary" on:click={startPakReplace}
              disabled={!pakRepSource || !pakRepOutput || (pakRepUseList ? !pakRepListFile : !pakRepInput)}>
              Start Replace
            </button>
          {/if}
        </div>

      <!-- PAK FONT EXTRACT -->
      {:else if selectedOp === 'pak_font_extract'}
        <div class="form-title">PAK (Font) — Extract</div>
        <div class="form-group"><label>PAK file (Font) :</label><div class="form-row"><input type="text" bind:value={pakFontExtSource} readonly /><button class="btn" on:click={browsePakFontExtSource}>Select</button></div></div>
        <div class="form-group"><label>Charset :</label><div class="form-row"><select bind:value={pakFontExtCharset}><option value="UTF-8">UTF-8</option><option value="ShiftJIS">Shift-JIS</option><option value="GBK">GBK</option></select></div></div>
        <div class="form-group"><label>Output folder :</label><div class="form-row"><input type="text" bind:value={pakFontExtOutput} readonly /><button class="btn" on:click={browsePakFontExtOutput}>Select</button></div><div class="form-hint">Tous les fichiers du PAK seront extraits ici</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startPakFontExtract} disabled={!pakFontExtSource || !pakFontExtOutput}>Start Extract</button>{/if}</div>

      <!-- PAK FONT REPLACE -->
      {:else if selectedOp === 'pak_font_replace'}
        <div class="form-title">PAK (Font) — Replace</div>
        <div class="form-group"><label>Original PAK file (Font) :</label><div class="form-row"><input type="text" bind:value={pakFontRepSource} readonly /><button class="btn" on:click={browsePakFontRepSource}>Select</button></div></div>
        <div class="form-group"><label>Charset :</label><div class="form-row"><select bind:value={pakFontRepCharset}><option value="UTF-8">UTF-8</option><option value="ShiftJIS">Shift-JIS</option><option value="GBK">GBK</option></select></div></div>
        <div class="form-group">
          <label>Mode d'entrée :</label>
          <div class="form-row checkbox-row" style="margin-bottom:6px">
            <label class="checkbox-label"><input type="radio" bind:group={pakFontRepMode} value="list" /> Fichier liste (<code>*_list.txt</code>)</label>
            <label class="checkbox-label"><input type="radio" bind:group={pakFontRepMode} value="dir" /> Dossier de fichiers</label>
            <label class="checkbox-label"><input type="radio" bind:group={pakFontRepMode} value="single" /> Fichier unique par nom</label>
            <label class="checkbox-label"><input type="radio" bind:group={pakFontRepMode} value="alias" /> Alias de taille compatible</label>
          </div>
          {#if pakFontRepMode === 'list'}
            <div class="form-row"><input type="text" bind:value={pakFontRepListFile} placeholder="FONT__INFO_list.txt" readonly /><button class="btn" on:click={browsePakFontRepListFile}>Select</button></div>
            <div class="form-hint">Fichier liste généré lors de l'extraction (ex : FONT__INFO_list.txt)</div>
          {:else if pakFontRepMode === 'dir'}
            <div class="form-row"><input type="text" bind:value={pakFontRepInput} readonly /><button class="btn" on:click={browsePakFontRepInput}>Select</button></div>
            <div class="form-hint">Remplace uniquement les fichiers du dossier dont le nom existe dans le PAK.</div>
          {:else if pakFontRepMode === 'single'}
            <div class="form-row"><input type="text" bind:value={pakFontRepSingleFile} readonly placeholder="ex : C:\dossier\info30" /><button class="btn" on:click={browsePakFontRepSingleFile}>Select</button></div>
            <div class="form-row" style="margin-top:6px"><input type="text" bind:value={pakFontRepSingleName} placeholder="Nom interne exact : info30 ou 明朝30" /></div>
            <div class="form-hint">Recommandé pour Kanon : faites deux remplacements séparés, <code>info30</code> dans <code>FONT__INFO.PAK</code>, puis <code>明朝30</code> dans <code>FONT_MINCHO.PAK</code>.</div>
          {:else}
            <div class="form-row">
              <span style="min-width:110px;font-size:12px">Copier depuis :</span>
              <input type="text" bind:value={pakFontRepAliasFrom} placeholder="info30 ou 明朝30" />
            </div>
            <div class="form-row" style="margin-top:6px">
              <span style="min-width:110px;font-size:12px">Vers :</span>
              <input type="text" bind:value={pakFontRepAliasTo} placeholder="info32 ou 明朝32" />
            </div>
            <div class="form-hint">Adapte les glyphes de la taille source à la géométrie, la largeur et la longueur d'entrée de la taille cible. Testé sur Kanon : <code>info30 → info32</code>, puis <code>明朝30 → 明朝32</code>.</div>
          {/if}
        </div>
        <div class="form-group"><label>Output PAK :</label><div class="form-row"><input type="text" bind:value={pakFontRepOutput} readonly /><button class="btn" on:click={browsePakFontRepOutput}>Select</button></div></div>
        <div class="form-actions">
          {#if running}
            <span class="running-indicator"></span> Running...
          {:else}
            <button class="btn btn-primary" on:click={startPakFontReplace}
              disabled={!pakFontRepSource || !pakFontRepOutput || (pakFontRepMode === 'list' ? !pakFontRepListFile : pakFontRepMode === 'dir' ? !pakFontRepInput : pakFontRepMode === 'single' ? (!pakFontRepSingleFile || !pakFontRepSingleName) : (!pakFontRepAliasFrom || !pakFontRepAliasTo))}>
              Start Replace
            </button>
          {/if}
        </div>

      <!-- FONT EXTRACT -->
      {:else if selectedOp === 'font_extract'}
        <div class="form-title">Font Extract</div>
        <div class="form-group"><label>Font CZ file (e.g. 明朝32):</label><div class="form-row"><input type="text" bind:value={fontExtCz} readonly /><button class="btn" on:click={browseFontExtCz}>Select</button></div></div>
        <div class="form-group"><label>Info file (e.g. info32):</label><div class="form-row"><input type="text" bind:value={fontExtInfo} readonly /><button class="btn" on:click={browseFontExtInfo}>Select</button></div><div class="form-hint">Must match font size (info32 for 明朝32)</div></div>
        <div class="form-group"><label>Output PNG:</label><div class="form-row"><input type="text" bind:value={fontExtPng} readonly /><button class="btn" on:click={browseFontExtPng}>Select</button></div></div>
        <div class="form-group"><label>Output charset TXT (optional):</label><div class="form-row"><input type="text" bind:value={fontExtCharset} readonly /><button class="btn" on:click={browseFontExtCharset}>Select</button></div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startFontExtract} disabled={!fontExtCz || !fontExtInfo || !fontExtPng}>Start Extract</button>{/if}</div>

      <!-- FONT EDIT -->
      {:else if selectedOp === 'font_edit'}
        <div class="form-title">Font Edit — Modification de glyphes</div>
        <div class="form-hint form-hint-warn">⚠ Font Edit modifie les glyphes d'un fichier CZ avec un TTF. Pour simplement re-packer un PAK de font, utilisez <strong>PAK (Font) → Font Replace</strong>.</div>

        <div class="form-group"><label>Source CZ file:</label><div class="form-row"><input type="text" bind:value={fontEditCz} readonly /><button class="btn" on:click={browseFontEditCz}>Select</button></div></div>
        <div class="form-group"><label>Source info file:</label><div class="form-row"><input type="text" bind:value={fontEditInfo} readonly /><button class="btn" on:click={browseFontEditInfo}>Select</button></div></div>
        <div class="form-group"><label>TTF font file:</label><div class="form-row"><input type="text" bind:value={fontEditTtf} readonly /><button class="btn" on:click={browseFontEditTtf}>Select</button></div></div>

        <div class="form-group">
          <label>Mode :</label>
          <div class="form-row checkbox-row" style="margin-bottom:6px">
            <label class="checkbox-label"><input type="radio" bind:group={fontEditMode} value="redraw" /> Redraw all</label>
            <label class="checkbox-label"><input type="radio" bind:group={fontEditMode} value="append" /> Append to end</label>
            <label class="checkbox-label"><input type="radio" bind:group={fontEditMode} value="insert" /> Insert at index</label>
          </div>
          {#if fontEditMode === 'redraw'}
            <div class="form-hint">Redessine TOUS les glyphes existants avec le TTF. Aucun charset requis.</div>
          {:else if fontEditMode === 'append'}
            <div class="form-hint">Ajoute les caractères du charset à la fin de la police.</div>
          {:else if fontEditMode === 'insert'}
            <div class="form-row" style="margin-top:4px">
              <span style="min-width:90px;font-size:12px">Start index :</span>
              <input type="number" bind:value={fontEditIndex} min="0" style="width:80px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
            </div>
            <div class="form-hint">Insère/remplace à partir de cette position (0-indexé).</div>
          {/if}
        </div>

        {#if fontEditMode !== 'redraw'}
          <div class="form-group"><label>Charset file <span class="required">*</span> :</label><div class="form-row"><input type="text" bind:value={fontEditCharsetFile} readonly /><button class="btn" on:click={browseFontEditCharset}>Select</button></div><div class="form-hint">Fichier texte listant les caractères à ajouter/insérer (ex : accents_fr.txt)</div></div>
        {/if}

        <div class="form-group">
          <label>Metrics adjustment :</label>
          <div class="form-row checkbox-row" style="margin-bottom:6px">
            <label class="checkbox-label"><input type="checkbox" checked={fontEditArabicMetrics} on:change={(e) => setFontEditArabicPreset(e.target.checked)} /> Arabic preset</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={fontEditMetricSetYEnabled} /> Set Y</label>
            {#if fontEditMetricSetYEnabled}
              <input type="number" bind:value={fontEditMetricSetY} style="width:70px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
            {/if}
          </div>
          <div class="form-row" style="gap:8px;flex-wrap:wrap">
            <span style="font-size:12px">Y offset</span>
            <input type="number" bind:value={fontEditMetricYOffset} style="width:70px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
            <span style="font-size:12px">X offset</span>
            <input type="number" bind:value={fontEditMetricXOffset} style="width:70px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
            <span style="font-size:12px">Advance offset</span>
            <input type="number" bind:value={fontEditMetricWOffset} style="width:70px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
            <span style="font-size:12px">Connector bleed</span>
            <input type="number" min="0" max="8" bind:value={fontEditArabicConnectorBleed} style="width:70px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" />
          </div>
          <div class="form-hint">Arabic preset shifts Arabic glyphs toward the Latin baseline. Connector bleed is experimental and stays manual.</div>
        </div>

        <div class="form-group"><label>Output CZ <span class="required">*</span> :</label><div class="form-row"><input type="text" bind:value={fontEditOutCz} placeholder="ex: C:\dossier\ゴシック26" /><button class="btn" on:click={browseFontEditOutCz}>📁</button></div><div class="form-hint">Tapez le chemin complet sans extension — le bouton sélectionne le dossier</div></div>
        <div class="form-group"><label>Output info <span class="required">*</span> :</label><div class="form-row"><input type="text" bind:value={fontEditOutInfo} placeholder="ex: C:\dossier\info26" /><button class="btn" on:click={browseFontEditOutInfo}>📁</button></div><div class="form-hint">Tapez le chemin complet sans extension — requis pour mettre à jour le compte de caractères</div></div>

        <div class="form-actions">
          {#if running}
            <span class="running-indicator"></span> Running...
          {:else}
            <button class="btn btn-primary" on:click={startFontEdit}
              disabled={!fontEditCz || !fontEditInfo || !fontEditTtf || !fontEditOutCz || !fontEditOutInfo
                || (fontEditMode !== 'redraw' && !fontEditCharsetFile)}>
              Start Edit
            </button>
          {/if}
        </div>

      <!-- VIETNAMESE FONT PATCH -->
      {:else if selectedOp === 'viet_font_patch'}
        <div class="form-title">AIR / Planetarian SG — Vietnamese Font Patch</div>
        <div class="form-hint form-hint-warn">This dedicated workflow generates safe Vietnamese font PAKs directly from the original game font folder. Use the full 134-character Vietnamese charset file.</div>

        <div class="form-group"><label>Game files folder:</label><div class="form-row"><input type="text" bind:value={vietFontRoot} readonly placeholder="ex: C:\Games\AIR\files or C:\Games\Planetarian SG\files" /><button class="btn" on:click={browseVietFontRoot}>Select</button></div><div class="form-hint">Select the folder that contains <code>font_win32_1280</code>.</div></div>
        <div class="form-group"><label>Full Vietnamese charset file:</label><div class="form-row"><input type="text" bind:value={vietCharsetFile} readonly placeholder="examples\AIR_vietnamese_full_134.txt" /><button class="btn" on:click={browseVietCharset}>Select</button></div><div class="form-hint">Use the full 134-character charset. The tool keeps the 32 existing characters and injects only the missing 102.</div></div>
        <div class="form-group"><label>TTF / OTF font file:</label><div class="form-row"><input type="text" bind:value={vietTtfFile} readonly /><button class="btn" on:click={browseVietTtf}>Select</button></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={vietOutputDir} readonly /><button class="btn" on:click={browseVietOutput}>Select</button></div><div class="form-hint">Each selected Y value creates a separate subfolder containing ready-to-test PAKs.</div></div>

        <div class="form-group">
          <label>Patch mode:</label>
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietRedrawLatin} /> Experimental: redraw Latin alphabet from TTF</label>
          </div>
          <div class="form-hint">Default mode only injects missing Vietnamese glyphs. Experimental mode redraws A-Z/a-z and already-present Vietnamese glyphs at their original indexes, then injects the missing Vietnamese glyphs.</div>
        </div>

        <div class="form-group">
          <label>Target:</label>
          <div class="form-row">
            <select bind:value={vietSlot}>
              <option value="en">English slot (recommended)</option>
              <option value="zc">Chinese ZC slot</option>
              <option value="all">Both slots</option>
            </select>
            <select bind:value={vietFamily}>
              <option value="GOTHIC1">GOTHIC1 quick test</option>
              <option value="GOTHIC2">GOTHIC2</option>
              <option value="GOTHIC3">GOTHIC3</option>
              <option value="MINCHO">MINCHO</option>
              <option value="MODERN">MODERN</option>
              <option value="all">All English families</option>
            </select>
          </div>
          <div class="form-hint">For first tests, keep English slot + GOTHIC1. Generate all families only after a good Y value is found.</div>
        </div>

        <div class="form-group">
          <label>Y alignment values:</label>
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietYMinus2} /> Y-2</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietYMinus1} /> Y-1</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietY0} /> Y+0</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietY1} /> Y+1</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietY2} /> Y+2</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={vietY3} /> Y+3</label>
          </div>
          <div class="form-hint">Y+2 is the validated AIR value. Select several values to create several test folders at once.</div>
        </div>

        <div class="form-actions">
          {#if running}
            <span class="running-indicator"></span> Running...
          {:else}
            <button class="btn btn-primary" on:click={startVietnameseFontPatch}
              disabled={!vietFontRoot || !vietCharsetFile || !vietTtfFile || !vietOutputDir || getVietYOffsets().length === 0}>
              Generate Vietnamese Font PAKs
            </button>
          {/if}
        </div>

      <!-- LUCA MENU DLL -->
      {:else if selectedOp === 'luca_menu_dll'}
        <div class="form-title">Luca Menu DLL</div>

        {#if !lucaInventory}
          <div class="form-actions" style="justify-content:flex-start">
            <button class="btn btn-primary" on:click={loadLucaInventory}>{t("Charger l'inventaire Luca", 'Load Luca inventory', uiLanguage)}</button>
          </div>
        {:else if !currentLucaProfile()}
          <div class="form-hint form-hint-warn">{t('Aucun profil Luca disponible dans le kit.', 'No Luca profile is available in the kit.', uiLanguage)}</div>
        {:else}
          <div class="form-group">
            <label>{t('Profil et slot :', 'Profile and slot:', uiLanguage)}</label>
            <div class="form-row">
              <select bind:value={lucaGame} on:change={() => setLucaGame(lucaGame)}>
                {#each lucaGameProfiles() as profile}
                  <option value={profile.id}>{profile.name} ({profile.id})</option>
                {/each}
              </select>
              <select value={lucaSlot} on:change={(e) => setLucaSlot(e.target.value)}>
                <option value="en">{t('Slot anglais', 'English slot', uiLanguage)}</option>
                <option value="jp">{t('Slot japonais', 'Japanese slot', uiLanguage)}</option>
                <option value="cn">{t('Slot chinois', 'Chinese slot', uiLanguage)}</option>
              </select>
              <button class="btn" on:click={loadLucaInventory}>Rescan</button>
            </div>
            <div class="form-hint">
              EN {lucaAvailableSlotCount('en')} · JP {lucaAvailableSlotCount('jp')} · CN {lucaAvailableSlotCount('cn')} · {t('FR sûr sur ce slot', 'safe FR strings in this slot', uiLanguage)} {lucaSafeFrenchCount()}
            </div>
            {#if lucaSlotSourceProfile() && lucaSlotSourceProfile().id !== currentLucaProfile().id}
              <div class="form-hint">{t('Inventaire du slot chargé depuis', 'Slot inventory loaded from', uiLanguage)} {lucaSlotSourceProfile().id}.</div>
            {:else if !lucaSlotSourceProfile()}
              <div class="form-hint form-hint-warn">{t('Aucune chaîne du slot', 'No strings from the', uiLanguage)} {slotLabel(lucaSlot)} {t("n'est encore inventoriée pour ce jeu.", 'slot have been inventoried for this game yet.', uiLanguage)}</div>
            {/if}
          </div>

          <div class="form-group">
            <label>{t('EXE du jeu', 'Game EXE', uiLanguage)} <span class="required">*</span> :</label>
            <div class="form-row"><input type="text" bind:value={lucaExe} placeholder={currentLucaProfile().gameExe || t('Sélectionnez le véritable EXE du jeu', 'Select the actual game EXE', uiLanguage)} /><button class="btn" on:click={browseLucaExe}>Select</button></div>
            <div class="form-hint">{t("Sélectionnez l'EXE présent dans le dossier du jeu afin de vérifier les offsets et la taille des chaînes.", 'Select the EXE located in the game folder so offsets and string sizes can be verified.', uiLanguage)}</div>
          </div>

          <div class="form-group">
            <label>{t('Dossier de sortie', 'Output folder', uiLanguage)} <span class="required">*</span> :</label>
            <div class="form-row"><input type="text" bind:value={lucaOutputDir} readonly /><button class="btn" on:click={browseLucaOutput}>Select</button></div>
            <div class="form-hint">{t('Le dossier recevra', 'The folder will receive', uiLanguage)} {lucaCustomPatch ? 'mixed_patches.py, custom_patches.py' : lucaFillMode === 'ru' ? 'mixed_patches.py, russian_preset.py' : 'patches.py'}, patches.h, patches.csv, version.c, {lucaProxyName()}.def {t('et', 'and', uiLanguage)} {lucaProxyName()}.dll {t('si la compilation réussit.', 'if compilation succeeds.', uiLanguage)}</div>
          </div>

          {#if (lucaGame || '').split('/')[0].toUpperCase() === 'LBEE'}
            <div class="form-group">
              <label>{t('Fichier PATCHES personnalisé (facultatif) :', 'Custom PATCHES file (optional):', uiLanguage)}</label>
              <div class="form-row">
                <input type="text" bind:value={lucaCustomPatch} readonly placeholder={t('russian_preset.py ou autre fichier contenant PATCHES', 'russian_preset.py or another file containing PATCHES', uiLanguage)} />
                <button class="btn" on:click={browseLucaCustomPatch}>Select</button>
                {#if lucaCustomPatch}<button class="btn" on:click={() => lucaCustomPatch = ''}>{t('Vider', 'Clear', uiLanguage)}</button>{/if}
              </div>
              <div class="form-hint">{t('Si renseigné, ce fichier remplace le preset russe interne. Le dossier de sortie reste uniquement la destination des fichiers générés.', 'When selected, this file replaces the built-in Russian preset. The output folder remains only the destination for generated files.', uiLanguage)}</div>
            </div>
          {/if}

          <div class="form-group">
            <label>{t('Identité du patch :', 'Patch identity:', uiLanguage)}</label>
            <div class="form-row">
              <input type="text" bind:value={lucaPatchName} placeholder={t('Nom affiché dans luckproxy.log', 'Name shown in luckproxy.log', uiLanguage)} />
              <input type="text" bind:value={lucaPatchVersion} placeholder="Version" style="max-width:140px" />
            </div>
          </div>

          <div class="form-group luca-proxy-group">
            <label>{t('DLL proxy à compiler :', 'Proxy DLL to compile:', uiLanguage)}</label>
            <div class="form-row checkbox-row luca-proxy-row">
              <label class:luca-proxy-selected={lucaBuildDll && lucaProxyChoice === 'version'} class="checkbox-label luca-proxy-option">
                <input type="checkbox" checked={lucaBuildDll && lucaProxyChoice === 'version'} on:change={(e) => selectLucaProxy('version', e.target.checked)} />
                <span><strong>version.dll</strong><small>Proxy Luca standard · x64</small></span>
              </label>
              <label class:luca-proxy-selected={lucaBuildDll && lucaProxyChoice === 'winmm'} class="checkbox-label luca-proxy-option">
                <input type="checkbox" checked={lucaBuildDll && lucaProxyChoice === 'winmm'} on:change={(e) => selectLucaProxy('winmm', e.target.checked)} />
                <span><strong>winmm.dll</strong><small>Little Busters! · PE32/x86</small></span>
              </label>
            </div>
            <div class="form-hint">{t("Choix exclusif : cocher une DLL décoche automatiquement l'autre. Décochez la sélection active pour générer le kit sans compiler.", 'Exclusive choice: selecting one DLL automatically clears the other. Clear the active selection to generate the kit without compiling.', uiLanguage)}</div>
            {#if lucaProxyChoice === 'winmm'}
              <div class="form-hint form-hint-warn"><strong>{t('LBEE sélectionné :', 'LBEE selected:', uiLanguage)}</strong> {t('la GUI compilera', 'the GUI will compile', uiLanguage)} <code>winmm.dll</code> {t('en 32 bits. Installez uniquement cette DLL à côté de', 'as 32-bit. Install only this DLL next to', uiLanguage)} <code>LITBUS_WIN32.exe</code> ; {t("n'utilisez pas", 'do not use', uiLanguage)} <code>version.dll</code>.</div>
            {/if}
          </div>

          <div class="form-group">
            <label>{t('Langue à injecter :', 'Language to inject:', uiLanguage)}</label>
            <div class="form-row checkbox-row luca-toolbar">
              <select value={lucaFillMode} on:change={(e) => setLucaFillMode(e.target.value)}>
                <option value="fr">FR</option>
                <option value="fr-safe">{t('FR (sûr)', 'FR (safe)', uiLanguage)}</option>
                <option value="en">ENG</option>
                <option value="en-safe">{t('ENG (sûr)', 'ENG (safe)', uiLanguage)}</option>
                <option value="ar">{t('Arabe', 'Arabic', uiLanguage)}</option>
                <option value="ru">{t('Russe (LBEE)', 'Russian (LBEE)', uiLanguage)}</option>
                <option value="jp">{t('Japonais', 'Japanese', uiLanguage)}</option>
                <option value="cn">{t('Chinois', 'Chinese', uiLanguage)}</option>
              </select>
              <button class="btn" on:click={clearLucaTargets}>{t('Vider', 'Clear', uiLanguage)}</button>
            </div>
            <div class="form-hint">{t('Les modes sûrs limitent la sélection aux chaînes communes aux quatre jeux et compatibles avec le budget du slot.', 'Safe modes limit selection to strings shared by all four games and compatible with the slot budget.', uiLanguage)}</div>
          </div>

          <div class="form-group">
            <label>{t('Filtre :', 'Filter:', uiLanguage)}</label>
            <div class="form-row">
              <input type="text" bind:value={lucaSearch} placeholder={t('source, cible, contexte...', 'source, target, context...', uiLanguage)} />
              <span class="luca-count">{lucaSelectedEntries().length} {t('sélectionnée(s)', 'selected', uiLanguage)} · {lucaVisibleEntries().length} {t('visible(s)', 'visible', uiLanguage)} · slot {slotLabel(lucaSlot)}</span>
            </div>
          </div>

          <div class="luca-table">
            <div class="luca-row luca-head">
              <div></div>
              <div>{t('Contexte', 'Context', uiLanguage)}</div>
              <div>Source</div>
              <div>{t('Cible', 'Target', uiLanguage)}</div>
              <div>Budget</div>
            </div>
            {#each lucaVisibleEntries() as entry (entry.rawOffset + entry.source)}
              <div class="luca-row" class:entry-warn={entryTooLong(entry)}>
                <div class="luca-check"><input type="checkbox" bind:checked={entry.include} /></div>
                <div>
                  <div class="luca-context">{entry.context}</div>
                  <div class="luca-meta">{entry.rawOffset} · {entry.encoding || 'utf-8'} · {entry.textKind}{entry.commonCount ? ` · ${entry.commonCount}/4` : ''}{entry.risk ? ` · ${entry.risk}` : ''}</div>
                </div>
                <div class="luca-source">{entry.source}</div>
                <div><input type="text" bind:value={entry.target} placeholder={entry.suggestedFr || t('Traduction', 'Translation', uiLanguage)} /></div>
                <div class="luca-budget">{lucaEncodedLen(entry)} / {entry.budget >= 0 ? entry.budget : '?'}</div>
              </div>
            {/each}
          </div>

          <div class="form-actions">
            {#if running}
              <span class="running-indicator"></span> Running...
            {:else}
              <button class="btn btn-primary" on:click={startLucaGenerate}
                disabled={!lucaExe || !lucaOutputDir || (lucaSelectedEntries().length === 0 && !lucaCustomPatch)}>
                {t('Générer le kit DLL', 'Generate DLL kit', uiLanguage)}
              </button>
            {/if}
          </div>
        {/if}

      <!-- IMAGE EXPORT -->
      {:else if selectedOp === 'image_export'}
        <div class="form-title">Image Export (CZ → PNG)</div>
        <div class="form-group">
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={imgExpBatch} on:change={toggleExpBatch} /> Batch mode (entire folder)</label>
          </div>
        </div>
        <div class="form-group"><label>{imgExpBatch ? 'Input CZ folder:' : 'Input CZ file:'}</label><div class="form-row"><input type="text" bind:value={imgExpInput} readonly /><button class="btn" on:click={browseImgExpInput}>Select</button></div></div>
        <div class="form-group"><label>{imgExpBatch ? 'Output PNG folder:' : 'Output PNG file:'}</label><div class="form-row"><input type="text" bind:value={imgExpOutput} readonly /><button class="btn" on:click={browseImgExpOutput}>Select</button></div>
          {#if imgExpBatch}<div class="form-hint">CZ files are converted to PNG. MVT movies are exported as WebM.</div>{/if}
        </div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startImageExport} disabled={!imgExpInput || !imgExpOutput}>Start Export</button>{/if}</div>

      <!-- IMAGE IMPORT -->
      {:else if selectedOp === 'image_import'}
        <div class="form-title">Image Import (PNG → CZ)</div>
        <div class="form-group">
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={imgImpBatch} on:change={toggleImpBatch} /> Batch mode (entire folder)</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={imgImpFill} /> Fill to original size (CZ1 only)</label>
          </div>
        </div>
        <div class="form-group"><label>{imgImpBatch ? 'Original CZ folder:' : 'Original CZ file:'}</label><div class="form-row"><input type="text" bind:value={imgImpSource} readonly /><button class="btn" on:click={browseImgImpSource}>Select</button></div><div class="form-hint">Original CZ file(s) for format reference</div></div>
        <div class="form-group"><label>{imgImpBatch ? 'Input PNG folder:' : 'Input PNG file:'}</label><div class="form-row"><input type="text" bind:value={imgImpInput} readonly /><button class="btn" on:click={browseImgImpInput}>Select</button></div></div>
        <div class="form-group"><label>{imgImpBatch ? 'Output CZ folder:' : 'Output CZ file:'}</label><div class="form-row"><input type="text" bind:value={imgImpOutput} readonly /><button class="btn" on:click={browseImgImpOutput}>Select</button></div>
          {#if imgImpBatch}<div class="form-hint">PNG files matching a CZ source will be converted</div>{/if}
        </div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startImageImport} disabled={!imgImpSource || !imgImpInput || !imgImpOutput}>Start Import</button>{/if}</div>

      <!-- DIALOGUE EXTRACT -->
      {:else if selectedOp === 'dlg_extract'}
        <div class="form-title">Extract Dialogues</div>
        <div class="form-hint" style="margin-bottom:10px">
          Extrait les lignes <strong>MESSAGE</strong> et <strong>LOG_BEGIN</strong> des scripts décompilés (.txt) vers un fichier TSV éditable.<br>
          Les colonnes correspondent aux chaînes entre guillemets dans l'ordre d'apparition. L'attribution des langues varie selon le jeu — vérifiez manuellement.
        </div>
        <div class="form-group">
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtBatch} on:change={toggleDlgExtBatch} /> Batch mode (dossier entier)</label>
          </div>
        </div>
        <div class="form-group">
          <label>Colonnes à extraire :</label>
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang1} /> Lang 1</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang2} /> Lang 2</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang3} /> Lang 3</label>
            <label class="checkbox-label"><input type="checkbox" bind:checked={dlgExtLang4} /> Lang 4</label>
          </div>
          <div class="form-hint">Chaque numéro correspond à la Nième chaîne entre guillemets dans le script. Ex: pour AIR, Lang 1 = JAP, Lang 2 = ENG, Lang 3 = CN.</div>
        </div>
        <div class="form-group"><label>{dlgExtBatch ? 'Dossier scripts (.txt) :' : 'Fichier script (.txt) :'}</label><div class="form-row"><input type="text" bind:value={dlgExtInput} readonly /><button class="btn" on:click={browseDlgExtInput}>Select</button></div>
          {#if dlgExtDetectedFmt}<div class="form-hint">Format détecté : <strong>{dlgExtDetectedFmt}</strong></div>{/if}
        </div>
        <div class="form-group"><label>{dlgExtBatch ? 'Dossier de sortie :' : 'Fichier TSV de sortie :'}</label><div class="form-row"><input type="text" bind:value={dlgExtOutput} readonly /><button class="btn" on:click={browseDlgExtOutput}>Select</button></div>
          {#if dlgExtBatch}<div class="form-hint">Un fichier <code>*.ext.txt</code> sera créé par script contenant des MESSAGE</div>{/if}
        </div>
        <div class="form-actions">
          {#if running}<span class="running-indicator"></span> Running...
          {:else}<button class="btn btn-primary" on:click={startDlgExtract}
            disabled={!dlgExtInput || !dlgExtOutput || (!dlgExtLang1 && !dlgExtLang2 && !dlgExtLang3 && !dlgExtLang4)}>
            Start Extract
          </button>{/if}
        </div>

      <!-- DIALOGUE IMPORT -->
      {:else if selectedOp === 'dlg_import'}
        <div class="form-title">Import Dialogues</div>
        <div class="form-hint" style="margin-bottom:10px">
          Réinjecte les dialogues traduits (TSV) dans les fichiers scripts (.txt).<br>
          Le TSV doit avoir été généré par l'extraction ci-dessus. Supporte MESSAGE et LOG_BEGIN.
        </div>
        <div class="form-group">
          <div class="form-row checkbox-row">
            <label class="checkbox-label"><input type="checkbox" bind:checked={dlgImpBatch} on:change={toggleDlgImpBatch} /> Batch mode (dossier entier)</label>
          </div>
        </div>
        <div class="form-group">
          <label>Colonne cible à réinjecter :</label>
          <div class="form-row">
            <select bind:value={dlgImpTargetCol}>
              <option value={1}>Lang 1 (1ère chaîne)</option>
              <option value={2}>Lang 2 (2ème chaîne)</option>
              <option value={3}>Lang 3 (3ème chaîne)</option>
              <option value={4}>Lang 4 (4ème chaîne)</option>
            </select>
          </div>
          <div class="form-hint">La colonne sélectionnée sera lue dans le TSV et réinjectée dans la Nième chaîne entre guillemets du script.</div>
        </div>
        <div class="form-group"><label>{dlgImpBatch ? 'Dossier scripts originaux :' : 'Fichier script original :'}</label><div class="form-row"><input type="text" bind:value={dlgImpScript} readonly /><button class="btn" on:click={browseDlgImpScript}>Select</button></div>
          <div class="form-hint">Les fichiers .txt décompilés (originaux ou déjà traduits)</div>
        </div>
        <div class="form-group"><label>{dlgImpBatch ? 'Dossier TSV traduits :' : 'Fichier TSV traduit :'}</label><div class="form-row"><input type="text" bind:value={dlgImpTsv} readonly /><button class="btn" on:click={browseDlgImpTsv}>Select</button></div>
          {#if dlgImpBatch}<div class="form-hint">Fichiers <code>*.ext.txt</code> — chaque TSV sera associé au script correspondant</div>{/if}
        </div>
        <div class="form-group"><label>{dlgImpBatch ? 'Dossier de sortie :' : 'Fichier de sortie :'}</label><div class="form-row"><input type="text" bind:value={dlgImpOutput} readonly /><button class="btn" on:click={browseDlgImpOutput}>Select</button></div></div>
        <div class="form-actions">
          {#if running}<span class="running-indicator"></span> Running...
          {:else}<button class="btn btn-primary" on:click={startDlgImport}
            disabled={!dlgImpScript || !dlgImpTsv || !dlgImpOutput}>
            Start Import
          </button>{/if}
        </div>

      <!-- ABOUT -->
      {:else if selectedOp === 'about'}
        <div class="form-title">À propos</div>
        <div class="about-panel">
          <div class="about-logo">LuckSystem</div>
          <div class="about-subtitle">Fork · Yoremi-v3.30 GUI</div>
          <div class="about-desc">
            Interface graphique pour LuckSystem, l'outil de traduction de visual novels Visual Art's / Key.<br>
            Inclut des correctifs CZ (CZ1, CZ4), script, PAK, audio Ogg/MP3, et une interface subprocess.
          </div>
          <div class="about-links">
            <div class="about-link-row">
              <span class="about-link-label">Projet source :</span>
              <span class="about-link-url">https://github.com/wetor/LuckSystem</span>
            </div>
            <div class="about-link-row">
              <span class="about-link-label">Fork Yoremi :</span>
              <span class="about-link-url">https://github.com/yoremi-trad-fr/LuckSystem-2.3.2-Yoremi-Update</span>
            </div>
          </div>
          <div class="about-version">v3.30 GUI · Wails + Svelte</div>
        </div>
      {/if}
    </div>
  </div>

  <!-- CONSOLE (shared by all views) -->
  {/if}
  {#if activeView !== 'hub' && activeView !== 'about_global'}
  <div class="console-wrapper" class:resizing={consoleResizing}>
    <div class="console-resizer" class:resizing={consoleResizing} on:mousedown={startConsoleResize}></div>
    <div class="console-header">
      <span>Console Output</span>
      <div style="display:flex;gap:6px;align-items:center">
        {#if running}
          <button class="console-stop" on:click={stopProcess}>■ Stop</button>
        {/if}
        <button class="console-clear" on:click={clearConsole}>Clear</button>
      </div>
    </div>
    <div class="console" bind:this={consoleEl} style:height={consoleHeight + 'px'} on:contextmenu={openConsoleMenu}>
      {#each consoleLines as line}<div class={line.cls}>{line.text}</div>{/each}
    </div>
    {#if consoleMenuVisible}
      <div
        class="console-context-menu"
        style={`left:${consoleMenuX}px;top:${consoleMenuY}px`}
        on:click|stopPropagation
        on:contextmenu|preventDefault
      >
        <button type="button" on:click={copyConsoleSelection}>Copier</button>
        <button type="button" on:click={copyConsoleAll}>Copier tout</button>
        <button type="button" on:click={pasteConsoleClipboard}>Coller</button>
      </div>
    {/if}
  </div>
  {/if}
</div>
