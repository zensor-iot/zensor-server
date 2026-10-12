/**
 * Zensor — Tailwind preset (v1.0.0)
 *
 * Uso (tailwind.config.js):
 *   module.exports = {
 *     presets: [require('./design-system/tailwind.preset.js')],
 *     content: ['./src/**\/*.{html,js,jsx,ts,tsx,vue}'],
 *   };
 *
 * Y en tu CSS global:  @import './design-system/tokens.css';
 *
 * Reglas:
 *  - Los colores SEMÁNTICOS (canvas, surface, ink, ok, attention, live...) apuntan a variables CSS,
 *    así cambian solos entre tema claro y oscuro. Úsalos en componentes.
 *  - La paleta cruda (moss-800, clay-500...) existe pero NO se usa en componentes.
 *  - Se desactivan colores y escalas por defecto de Tailwind para que una LLM no pueda salirse del sistema.
 */
const tokens = require('./tokens.json');

// {"$value": "..."} -> "..."
const flat = (obj) =>
  Object.fromEntries(Object.entries(obj).map(([k, v]) => [k, v.$value]));

const raw = Object.fromEntries(
  Object.entries(tokens.color.palette).map(([family, steps]) => [family, flat(steps)])
);

// --- tabla de roles (espejo de tokens.json -> font.roles) ---
const ROLE = {
  display: ['display', 'display', 'bold', 'tight', 'display'],
  h1: ['display', '4xl', 'bold', 'tight', 'heading'],
  'h1-mobile': ['display', '3xl', 'bold', 'tight', 'heading'],
  h2: ['display', '2xl', 'bold', 'tight', 'heading'],
  title: ['display', 'xl', 'bold', 'snug', 'normal'],
  h3: ['display', 'lg', 'bold', 'snug', 'normal'],
  body: ['body', 'base', 'regular', 'normal', 'normal'],
  'body-sm': ['body', 'md', 'regular', 'normal', 'normal'],
  label: ['body', 'md', 'medium', 'snug', 'normal'],
  caption: ['body', 'sm', 'regular', 'normal', 'normal'],
  'data-lg': ['mono', '2xl', 'medium', 'tight', 'normal'],
  'data-md': ['mono', 'xl', 'medium', 'tight', 'normal'],
  'data-sm': ['mono', 'md', 'medium', 'tight', 'normal'],
};

const v = (name) => `var(--zs-${name})`;

module.exports = {
  theme: {
    // Reemplazos (no "extend"): fuera de estos valores no hay colores/escala.
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      // --- semánticos (usar estos) ---
      canvas: v('bg-canvas'),
      surface: v('bg-surface'),
      sunken: v('bg-sunken'),
      inverse: v('bg-inverse'),
      'inverse-raised': v('bg-inverse-raised'),
      scrim: v('bg-scrim'),
      line: v('border-default'),
      'line-strong': v('border-strong'),
      ink: v('text-primary'),
      muted: v('text-secondary'),
      'on-inverse': v('text-inverse'),
      'on-inverse-muted': v('text-inverse-muted'),
      attention: v('text-attention'),
      brand: v('brand-ink'),
      accent: v('accent'),
      'accent-inverse': v('accent-on-inverse'),
      'on-accent-inverse': v('accent-on-inverse-fg'),
      primary: v('action-primary-bg'),
      'on-primary': v('action-primary-fg'),
      'primary-hover': v('action-primary-hover'),
      focus: v('focus-ring'),
      track: v('track'),
      ok: { bg: v('status-ok-bg'), fg: v('status-ok-fg'), fill: v('status-ok-fill') },
      warn: { bg: v('status-attention-bg'), fg: v('status-attention-fg'), fill: v('status-attention-fill') },
      live: { bg: v('status-live-bg'), fg: v('status-live-fg'), fill: v('status-live-fill') },
      // --- paleta cruda (solo para definir nuevos semánticos) ---
      ...raw,
    },
    fontFamily: {
      display: [tokens.font.family.display.$value],
      body: [tokens.font.family.body.$value],
      mono: [tokens.font.family.mono.$value],
    },
    fontSize: {
      xs: v('text-xs'),
      sm: v('text-sm'),
      md: v('text-md'),
      base: v('text-base'),
      lg: v('text-lg'),
      xl: v('text-xl'),
      '2xl': v('text-2xl'),
      '3xl': v('text-3xl'),
      '4xl': v('text-4xl'),
      display: v('text-display'),
    },
    fontWeight: { regular: '400', medium: '500', bold: '700' },
    lineHeight: {
      tight: v('leading-tight'),
      snug: v('leading-snug'),
      normal: v('leading-normal'),
      relaxed: v('leading-relaxed'),
    },
    letterSpacing: {
      display: v('tracking-display'),
      heading: v('tracking-heading'),
      normal: v('tracking-normal'),
      wide: v('tracking-wide'),
    },
    spacing: {
      0: '0',
      1: v('space-1'),
      2: v('space-2'),
      3: v('space-3'),
      4: v('space-4'),
      5: v('space-5'),
      6: v('space-6'),
      8: v('space-8'),
      10: v('space-10'),
      12: v('space-12'),
      16: v('space-16'),
      touch: v('size-touch-min'),
      control: v('size-control'),
    },
    borderRadius: {
      none: '0',
      sm: v('radius-sm'),
      md: v('radius-md'),
      lg: v('radius-lg'),
      xl: v('radius-xl'),
      '2xl': v('radius-2xl'),
      pill: v('radius-pill'),
    },
    borderWidth: { DEFAULT: v('border-hairline'), 0: '0', control: v('border-control'), focus: v('border-focus') },
    boxShadow: {
      none: 'none',
      raised: v('shadow-raised'),
      overlay: v('shadow-overlay'),
      drawer: v('shadow-drawer'),
      modal: v('shadow-modal'),
    },
    zIndex: { base: '0', sticky: '10', scrim: '40', drawer: '50', toast: '60' },
    opacity: { 0: '0', offline: '0.82', disabled: '0.45', 100: '1' },
    transitionDuration: {
      instant: v('duration-instant'),
      fast: v('duration-fast'),
      base: v('duration-base'),
      slow: v('duration-slow'),
      slower: v('duration-slower'),
    },
    transitionTimingFunction: {
      standard: v('ease-standard'),
      decelerate: v('ease-decelerate'),
      accelerate: v('ease-accelerate'),
      linear: 'linear',
    },
    screens: Object.fromEntries(
      Object.entries(tokens.breakpoint).map(([k, val]) => [k, val.$value])
    ),
    extend: {
      width: { drawer: v('size-drawer'), sidebar: v('size-sidebar') },
      maxWidth: { content: v('size-content-max'), drawer: v('size-drawer') },
      minHeight: { touch: v('size-touch-min'), control: v('size-control') },
      minWidth: { touch: v('size-touch-min') },
      keyframes: {
        'drawer-in': { from: { transform: 'translateX(48px)', opacity: '0' }, to: { transform: 'translateX(0)', opacity: '1' } },
        'scrim-in': { from: { opacity: '0' }, to: { opacity: '1' } },
      },
      animation: {
        'drawer-in': 'drawer-in var(--zs-duration-slow) var(--zs-ease-decelerate) both',
        'scrim-in': 'scrim-in var(--zs-duration-base) var(--zs-ease-standard) both',
      },
    },
  },
  // Dark mode lo gobierna tokens.css: [data-theme="dark"] o prefers-color-scheme.
  darkMode: ['selector', '[data-theme="dark"]'],
  plugins: [
    // Roles tipográficos: class="text-h1", "text-data-md"... (equivalen a .zs-text-* de tokens.css)
    function ({ addComponents }) {
      addComponents(
        Object.fromEntries(
          Object.entries(ROLE).map(([name, [family, size, weight, leading, tracking]]) => [
            `.text-${name}`,
            {
              fontFamily: `var(--zs-font-${family})`,
              fontSize: `var(--zs-text-${size})`,
              fontWeight: `var(--zs-weight-${weight})`,
              lineHeight: `var(--zs-leading-${leading})`,
              letterSpacing: `var(--zs-tracking-${tracking})`,
            },
          ])
        )
      );
    },
  ],
};
