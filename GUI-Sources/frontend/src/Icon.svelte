<script>
  export let name = '';
  export let size = 18;
</script>

<svg
  class="app-icon"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.8"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
>
  {#if name === 'sparkles'}
    <path d="M12 6c0 4-2 6-6 6 4 0 6 2 6 6 0-4 2-6 6-6-4 0-6-2-6-6Z" />
    <path d="M5.5 2.5c0 1.7-1.3 3-3 3 1.7 0 3 1.3 3 3 0-1.7 1.3-3 3-3-1.7 0-3-1.3-3-3Z" />
    <path d="M18.5 3.5c0 1.4-1.1 2.5-2.5 2.5 1.4 0 2.5 1.1 2.5 2.5 0-1.4 1.1-2.5 2.5-2.5-1.4 0-2.5-1.1-2.5-2.5Z" />
    <path d="M5 16c0 1.1-.9 2-2 2 1.1 0 2 .9 2 2 0-1.1.9-2 2-2-1.1 0-2-.9-2-2Z" />
  {:else if name === 'archive'}
    <path d="M4 7h16v13H4z" />
    <path d="M3 3h18v4H3zM9 11h6" />
  {:else if name === 'scripts'}
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z" />
    <path d="M14 2v6h6M9 13l-2 2 2 2M15 13l2 2-2 2" />
  {:else if name === 'settings'}
    <circle cx="12" cy="12" r="3" />
    <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-2.83 2.83-.06-.06a1.7 1.7 0 0 0-1.88-.34 1.7 1.7 0 0 0-1.03 1.56V21h-4v-.08A1.7 1.7 0 0 0 8.95 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06-2.83-2.83.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.56-1H3v-4h.08A1.7 1.7 0 0 0 4.6 8.95a1.7 1.7 0 0 0-.34-1.88L4.2 7l2.83-2.83.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.56V3h4v.08A1.7 1.7 0 0 0 15.05 4.6a1.7 1.7 0 0 0 1.88-.34l.06-.06L19.82 7l-.06.06A1.7 1.7 0 0 0 19.4 9c.18.61.72 1 1.36 1H21v4h-.08A1.7 1.7 0 0 0 19.4 15Z" />
  {:else if name === 'data'}
    <ellipse cx="12" cy="5" rx="8" ry="3" />
    <path d="M4 5v7c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12v7c0 1.7 3.6 3 8 3s8-1.3 8-3v-7" />
  {:else if name === 'mobile'}
    <rect x="6" y="2" width="12" height="20" rx="2" />
    <path d="M10 5h4M11 18h2" />
  {:else if name === 'video'}
    <rect x="3" y="5" width="15" height="14" rx="2" />
    <path d="m18 10 3-2v8l-3-2zM8 9l5 3-5 3z" />
  {:else if name === 'image'}
    <rect x="3" y="3" width="18" height="18" rx="2" />
    <circle cx="8.5" cy="8.5" r="1.5" />
    <path d="m21 15-5-5L5 21" />
  {:else if name === 'animation'}
    <rect x="3" y="5" width="18" height="14" rx="2" />
    <path d="M7 5v14M17 5v14M3 9h4M3 15h4M17 9h4M17 15h4" />
  {:else if name === 'audio'}
    <path d="M9 18V5l11-2v13" />
    <circle cx="6" cy="18" r="3" />
    <circle cx="17" cy="16" r="3" />
  {:else if name === 'save'}
    <path d="M5 3h12l4 4v14H3V3Z" />
    <path d="M7 3v6h9V3M7 21v-7h10v7" />
  {:else if name === 'babel'}
    <circle cx="12" cy="12" r="9" />
    <path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18" />
  {:else if name === 'font'}
    <path d="M4 7V4h16v3M9 20h6M12 4v16" />
  {:else if name === 'dialogues'}
    <path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4Z" />
    <path d="M8 9h8M8 13h5" />
  {:else if name === 'tools'}
    <path d="M14.7 6.3a4 4 0 0 0-5-5L12 3.6 8.4 7.2 6.1 4.9a4 4 0 0 0 5 5L19 17.8a2.1 2.1 0 0 1-3 3l-7.9-7.9a4 4 0 0 0-5-5l2.3 2.3-3.6 3.6-2.3-2.3a4 4 0 0 0 5 5" />
  {:else if name === 'info'}
    <circle cx="12" cy="12" r="9" />
    <path d="M12 11v5M12 8h.01" />
  {:else if name === 'arrow-left'}
    <path d="m15 18-6-6 6-6M9 12h11" />
  {:else if name === 'terminal'}
    <rect x="3" y="4" width="18" height="16" rx="2" />
    <path d="m7 9 3 3-3 3M13 15h4" />
  {:else if name === 'chevron-down'}
    <path d="m6 9 6 6 6-6" />
  {:else if name === 'chevron-up'}
    <path d="m18 15-6-6-6 6" />
  {:else if name === 'trash'}
    <path d="M3 6h18M8 6V4h8v2M19 6l-1 15H6L5 6M10 11v5M14 11v5" />
  {:else if name === 'stop'}
    <rect x="6" y="6" width="12" height="12" rx="1" />
  {:else if name === 'check'}
    <circle cx="12" cy="12" r="9" />
    <path d="m8 12 2.7 2.7L16.5 9" />
  {:else if name === 'warning'}
    <path d="M10.3 3.6 2.4 17.4A2 2 0 0 0 4.1 20h15.8a2 2 0 0 0 1.7-2.6L13.7 3.6a2 2 0 0 0-3.4 0Z" />
    <path d="M12 9v4M12 17h.01" />
  {:else}
    <circle cx="12" cy="12" r="9" />
  {/if}
</svg>
