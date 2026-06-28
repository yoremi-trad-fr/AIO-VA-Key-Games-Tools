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
    RldevSaveGet,
    RldevSaveSet,
    RldevSaveDump,
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
  let lsPath = '';

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

  // --- PAK fields ---
  let pakExtSource = '';
  let pakExtOutput = '';
  let pakRepSource = '';
  let pakRepListFile = '';
  let pakRepInput = '';
  let pakRepOutput = '';
  let pakRepUseList = true; // mode par défaut : fichier liste

  // --- PAK Font fields ---
  let pakFontExtSource = '';
  let pakFontExtCharset = 'UTF-8';
  let pakFontExtOutput = '';
  let pakFontRepSource = '';
  let pakFontRepCharset = 'UTF-8';
  let pakFontRepListFile = '';
  let pakFontRepInput = '';
  let pakFontRepOutput = '';
  let pakFontRepUseList = true; // mode par défaut : fichier liste

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
  let rlSaveFile = '';
  let rlSaveRefs = 'intG[0] intG[6] intG[30] intG[31]';
  let rlSaveAssignments = 'intG[6]=0 intG[30]=0';
  let rlSaveBackup = true;
  let rlSaveDumpAll = false;
  let rlSaveDumpJson = false;
  let rlBabelRoot = '';
  let rlBabelGameDir = '';
  let rlBabelVersion = '1.2.3.5';
  let rlBabelDllMode = 'auto';
  let rlBabelNameEnc = 'western';
  let rlBabelUpdateGameexe = true;
  let rlBabelGlosses = false;

  // --- Siglus fields ---
  let siglusSelectedOp = 'scene_extract';
  let siglusGameName = 'Harmonia';
  let siglusCompressionLevel = 17;
  let siglusFakeCompression = false;
  let siglusScenePck = '';
  let siglusSceneOutputDir = '';
  let siglusSceneInputDir = '';
  let siglusSceneWtf = '0x166';
  let siglusSceneOutputPck = '';
  let siglusSSBatch = false;
  let siglusSSInput = '';
  let siglusSSTsv = '';
  let siglusSSOutput = '';
  let siglusSSCopyText = true;
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

  // ===== Operations list =====
  const operations = [
    { id: '_s1', label: 'SCRIPT', section: true },
    { id: 'decompile', label: 'Script Decompile' },
    { id: 'compile', label: 'Script Compile' },
    { id: '_s2', label: 'PAK (CG)', section: true },
    { id: 'pak_cg_extract', label: 'CG Extract' },
    { id: 'pak_cg_replace', label: 'CG Replace' },
    { id: '_s2b', label: 'PAK (Font)', section: true },
    { id: 'pak_font_extract', label: 'Font Extract' },
    { id: 'pak_font_replace', label: 'Font Replace' },
    { id: '_s3', label: 'FONT', section: true },
    { id: 'font_extract', label: 'Font Extract' },
    { id: 'font_edit', label: 'Font Edit' },
    { id: '_s3b', label: 'VIET FONT', section: true },
    { id: 'viet_font_patch', label: 'AIR / SG Patch' },
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

  onMount(async () => {
    EventsOn('log', (msg) => addLine(msg));
    lsPath = await GetLuckSystemPath();
    if (lsPath) {
      addLine('LuckSystem 2.3.2 - Yoremi fork v3.20 GUI');
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
    addLine('RLdev 2026 - Go édition v1.3.4');
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
  onDestroy(() => { EventsOff('log'); });

  // ===== Browse helpers =====
  async function browsePak() { const f = await SelectPakFile(); if (f) pakFile = f; }
  async function browseOpcode() { const f = await SelectFile('Select Opcode (.txt)', '*.txt', 'Opcode files'); if (f) { opcodeFile = f; selectedPreset = ''; } }
  async function browsePlugin() { const f = await SelectFile('Select Plugin (.py)', '*.py', 'Python plugins'); if (f) { pluginFile = f; selectedPreset = ''; } }
  async function browseOutputDir() { const d = await SelectDirectory('Select output directory'); if (d) outputDir = d; }
  async function browseImportDir() { const d = await SelectDirectory('Select translated scripts directory'); if (d) importDir = d; }
  async function browseOutputPak() { const f = await SelectSaveFile('Save output PAK', 'SCRIPT_FR.PAK', '*.PAK;*.pak', 'PAK files'); if (f) outputPak = f; }

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

  async function browsePakFontExtSource() { const f = await SelectPakFile(); if (f) pakFontExtSource = f; }
  async function browsePakFontExtOutput() { const d = await SelectDirectory('Dossier d\'extraction'); if (d) pakFontExtOutput = d; }
  async function browsePakFontRepSource() { const f = await SelectPakFile(); if (f) pakFontRepSource = f; }
  async function browsePakFontRepListFile() { const f = await SelectFile('Sélectionner le fichier liste (_list.txt)', '*.txt', 'Fichiers liste'); if (f) pakFontRepListFile = f; }
  async function browsePakFontRepInput() { const d = await SelectDirectory('Dossier des fichiers modifiés'); if (d) pakFontRepInput = d; }
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
  function startPakExtract() { run(() => PakExtract(pakExtSource, pakExtOutput)); }
  function startPakReplace() {
    const listArg = pakRepUseList ? pakRepListFile : '';
    const dirArg  = pakRepUseList ? '' : pakRepInput;
    run(() => PakReplace(pakRepSource, dirArg, listArg, pakRepOutput));
  }
  function startPakFontExtract() { run(() => PakFontExtract(pakFontExtSource, pakFontExtCharset, pakFontExtOutput)); }
  function startPakFontReplace() {
    const listArg = pakFontRepUseList ? pakFontRepListFile : '';
    const dirArg  = pakFontRepUseList ? '' : pakFontRepInput;
    run(() => PakFontReplace(pakFontRepSource, pakFontRepCharset, dirArg, listArg, pakFontRepOutput));
  }
  function startFontExtract() { run(() => FontExtract(fontExtCz, fontExtInfo, fontExtPng, fontExtCharset)); }
  function startFontEdit() {
    const redraw  = fontEditMode === 'redraw';
    const append  = fontEditMode === 'append';
    const index   = (fontEditMode === 'insert') ? fontEditIndex : 0;
    run(() => FontEdit(fontEditCz, fontEditInfo, fontEditTtf, fontEditOutCz, fontEditOutInfo, fontEditCharsetFile, redraw, append, index));
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

  function startImageExport() {
    if (imgExpBatch) run(() => ImageBatchExport(imgExpInput, imgExpOutput));
    else run(() => ImageExport(imgExpInput, imgExpOutput));
  }
  function startImageImport() {
    if (imgImpBatch) run(() => ImageBatchImport(imgImpSource, imgImpInput, imgImpOutput, imgImpFill));
    else run(() => ImageImport(imgImpSource, imgImpInput, imgImpOutput, imgImpFill));
  }

  function selectOp(op) { if (!op.disabled && !op.section) selectedOp = op.id; }

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
  function startSaveGet() { run(() => RldevSaveGet(rlSaveFile, rlSaveRefs)); }
  function startSaveSet() { run(() => RldevSaveSet(rlSaveFile, rlSaveAssignments, rlSaveBackup)); }
  function startSaveDump() { run(() => RldevSaveDump(rlSaveFile, rlSaveDumpAll, rlSaveDumpJson)); }
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
    if (siglusSSBatch) run(() => SiglusSSDumpAll(siglusSSInput, siglusSSTsv, siglusSSCopyText, siglusSSFilterMode, siglusSSFormat, siglusSSSingleXlsx));
    else run(() => SiglusSSDump(siglusSSInput, siglusSSTsv, siglusSSCopyText, siglusSSFilterMode, siglusSSFormat));
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
          <div class="hub-card-ver">2.3.2 · Yoremi Fork v3.20 GUI</div>
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
          <strong>LuckSystem v3.20 GUI</strong> — Scripts, PAK, fonts, images CZ (LuckEngine)<br>
          <strong>RLdev 2026 v1.3</strong> — SEEN.txt, Kepago, AVG32, G00, GAN, NWA, DAT, Babel (RealLive)<br>
          <strong>Siglus Tools</strong> — SiglusEngine, Scene.pck, SS, Gameexe, DBS, mobile PCK, OMV<br><br>
          Développé par <strong>Yoremi</strong> · Wails + Svelte
        </div>
        <div class="about-version">v1.0 · 2026</div>
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
          <option value="Harmonia"></option>
          <option value="Planetarian HD Steam"></option>
          <option value="Rewrite"></option>
          <option value="Rewrite Harvest festa!"></option>
          <option value="Rewrite+"></option>
          <option value="Rewrite+ Steam English"></option>
          <option value="Angel Beats! -1st beat-"></option>
          <option value="Summer Pockets"></option>
          <option value="Summer Pockets Steam English"></option>
          <option value="Summer Pockets REFLECTION BLUE DL"></option>
          <option value="Summer Pockets REFLECTION BLUE PKG"></option>
          <option value="CLANNAD Steam Chinese"></option>
          <option value="LOOPERS DL"></option>
          <option value="LUNARiA DL"></option>
          <option value="LUNARiA PKG"></option>
          <option value="AIR Android"></option>
          <option value="Kanon Android"></option>
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
          <div class="form-group"><label>WTF value :</label><div class="form-row"><input type="text" bind:value={siglusSceneWtf} placeholder="0x166" /></div></div>
          <div class="form-group"><label>Compression :</label><div class="form-row"><input type="number" bind:value={siglusCompressionLevel} min="2" max="17" style="width:80px;height:26px;padding:0 6px;border:1px solid #c0c0c0;border-radius:2px" /><label class="checkbox-label"><input type="checkbox" bind:checked={siglusFakeCompression} /> Fake compression</label></div></div>
          <div class="form-group"><label>Scene.pck de sortie :</label><div class="form-row"><input type="text" bind:value={siglusSceneOutputPck} readonly /><button class="btn" on:click={browseSiglusSceneOutputPck}>Select</button></div></div>
          <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startSiglusSceneRebuild} disabled={!siglusSceneInputDir || !siglusGameName || !siglusSceneWtf || !siglusSceneOutputPck}>Rebuild</button>{/if}</div>

        {:else if siglusSelectedOp === 'ss_dump'}
          <div class="form-title">Dump SS text</div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSBatch} on:change={toggleSiglusSSBatch} /> Batch mode</label></div></div>
          <div class="form-group"><label>{siglusSSBatch ? 'Dossier .ss :' : 'Fichier .ss :'}</label><div class="form-row"><input type="text" bind:value={siglusSSInput} readonly /><button class="btn" on:click={browseSiglusSSInput}>Select</button></div></div>
          <div class="form-group"><label>Filtre :</label><div class="form-row"><select bind:value={siglusSSFilterMode}><option value="smart">Smart filter</option><option value="all">Export all text</option><option value="full">Full-width only</option></select><label class="checkbox-label"><input type="checkbox" bind:checked={siglusSSCopyText} /> Copy text</label></div></div>
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
          <div class="form-group"><label>Save file :</label><div class="form-row"><input type="text" bind:value={rlSaveFile} readonly /><button class="btn" on:click={browseRlSave}>Select</button></div></div>
          <div class="form-group"><label>Variables :</label><div class="form-row"><input type="text" bind:value={rlSaveRefs} placeholder="intG[0] intG[6] intG[30]" /></div></div>
          <div class="form-group"><label>Assignations :</label><div class="form-row"><input type="text" bind:value={rlSaveAssignments} placeholder="intG[6]=0 intG[30]=0" /></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveBackup} /> Backup before write</label></div></div>
          <div class="form-group"><div class="form-row checkbox-row"><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveDumpAll} /> Dump all intG values</label><label class="checkbox-label"><input type="checkbox" bind:checked={rlSaveDumpJson} /> JSON</label></div></div>
          <div class="form-actions">
            {#if running}
              <span class="running-indicator"></span> Running...
            {:else}
              <button class="btn" on:click={startSaveInfo} disabled={!rlSaveFile}>Info</button>
              <button class="btn" on:click={startSaveGet} disabled={!rlSaveFile || !rlSaveRefs.trim()}>Get</button>
              <button class="btn" on:click={startSaveDump} disabled={!rlSaveFile}>Dump</button>
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
    <span>LuckSystem 2.3.2 - Yoremi fork v3.20 GUI</span>
    <div style="display:flex;align-items:center;gap:10px">
      <span class="titlebar-path" on:click={locateLuckSystem} title="Click to change">
        {#if lsPath}📁 {lsPath}{:else}⚠ lucksystem.exe not found - Click to locate{/if}
      </span>
      <button class="titlebar-back" on:click={() => activeView = 'hub'}>← Retour</button>
    </div>
  </div>

  <div class="content">
    <!-- LEFT SIDEBAR -->
    <div class="sidebar">
      <div class="sidebar-title">Select option:</div>
      <div class="sidebar-list">
        {#each operations as op}
          {#if op.section}
            <div class="sidebar-section">{op.label}</div>
          {:else}
            <div class="sidebar-item" class:active={selectedOp === op.id} class:disabled={op.disabled} on:click={() => selectOp(op)}>
              {op.label}
            </div>
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

      <!-- PAK CG EXTRACT -->
      {:else if selectedOp === 'pak_cg_extract'}
        <div class="form-title">PAK (CG) — Extract</div>
        <div class="form-group"><label>PAK file (CG) :</label><div class="form-row"><input type="text" bind:value={pakExtSource} readonly /><button class="btn" on:click={browsePakExtSource}>Select</button></div></div>
        <div class="form-group"><label>Output folder:</label><div class="form-row"><input type="text" bind:value={pakExtOutput} readonly /><button class="btn" on:click={browsePakExtOutput}>Select</button></div><div class="form-hint">Le fichier liste <code>&lt;NOM&gt;_list.txt</code> sera généré automatiquement dans ce dossier</div></div>
        <div class="form-actions">{#if running}<span class="running-indicator"></span> Running...{:else}<button class="btn btn-primary" on:click={startPakExtract} disabled={!pakExtSource || !pakExtOutput}>Start Extract</button>{/if}</div>

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
            <label class="checkbox-label"><input type="radio" bind:group={pakFontRepUseList} value={true} /> Fichier liste (<code>*_list.txt</code>)</label>
            <label class="checkbox-label"><input type="radio" bind:group={pakFontRepUseList} value={false} /> Dossier de fichiers</label>
          </div>
          {#if pakFontRepUseList}
            <div class="form-row"><input type="text" bind:value={pakFontRepListFile} placeholder="FONT__INFO_list.txt" readonly /><button class="btn" on:click={browsePakFontRepListFile}>Select</button></div>
            <div class="form-hint">Fichier liste généré lors de l'extraction (ex : FONT__INFO_list.txt)</div>
          {:else}
            <div class="form-row"><input type="text" bind:value={pakFontRepInput} readonly /><button class="btn" on:click={browsePakFontRepInput}>Select</button></div>
            <div class="form-hint">⚠ Le mode dossier peut échouer selon lucksystem — préférer le fichier liste</div>
          {/if}
        </div>
        <div class="form-group"><label>Output PAK :</label><div class="form-row"><input type="text" bind:value={pakFontRepOutput} readonly /><button class="btn" on:click={browsePakFontRepOutput}>Select</button></div></div>
        <div class="form-actions">
          {#if running}
            <span class="running-indicator"></span> Running...
          {:else}
            <button class="btn btn-primary" on:click={startPakFontReplace}
              disabled={!pakFontRepSource || !pakFontRepOutput || (pakFontRepUseList ? !pakFontRepListFile : !pakFontRepInput)}>
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
          {#if imgExpBatch}<div class="form-hint">All CZ files will be converted to PNG</div>{/if}
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
          <div class="about-subtitle">Fork · Yoremi-v3.20 GUI</div>
          <div class="about-desc">
            Interface graphique pour LuckSystem, l'outil de traduction de visual novels Visual Art's / Key.<br>
            Inclut des correctifs CZ (CZ1, CZ4), script, PAK, et une interface subprocess.
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
          <div class="about-version">v3.20 GUI · Wails + Svelte</div>
        </div>
      {/if}
    </div>
  </div>

  <!-- CONSOLE (shared by all views) -->
  {/if}
  {#if activeView !== 'hub' && activeView !== 'about_global'}
  <div class="console-wrapper">
    <div class="console-header">
      <span>Console Output</span>
      <div style="display:flex;gap:6px;align-items:center">
        {#if running}
          <button class="console-stop" on:click={stopProcess}>■ Stop</button>
        {/if}
        <button class="console-clear" on:click={clearConsole}>Clear</button>
      </div>
    </div>
    <div class="console" bind:this={consoleEl}>
      {#each consoleLines as line}<div class={line.cls}>{line.text}</div>{/each}
    </div>
  </div>
  {/if}
</div>
