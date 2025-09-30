# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Adnexus Studio is a browser-based professional TV ad creation platform for producing 15-second and 30-second CTV/OTT ads. It uses WebCodecs API for high-performance video processing entirely in the browser, with no server-side rendering required.

**Key Technologies:**
- TypeScript with LitElement components
- PixiJS for canvas rendering
- FFmpeg.wasm for video processing
- WebCodecs API for encoding/decoding
- GSAP for animations
- Turtle build system (@benev/turtle)
- Construct/Slate framework (@benev/construct, @benev/slate)

## Directory Structure

```
/s/              - Source TypeScript files (rootDir)
  /components/   - LitElement UI components (omni-*)
  /context/      - Application context and state management
    /controllers/ - Business logic controllers
  /views/        - View components (tooltip, shortcuts, etc.)
  /tools/        - Utility tools and helpers
  /utils/        - Shared utilities
  /icons/        - SVG icon libraries

/x/              - Compiled JavaScript output (outDir)
/assets/         - Static assets (fonts, images, etc.)
```

## Core Architecture

Adnexus Studio follows a unidirectional data flow pattern:

```
Actions → State → Controllers → Components/Views
```

### Key Architectural Components:

1. **Context (`s/context/context.ts`)**
   - `OmniContext` extends Construct's Context
   - Manages historical state (with undo/redo) via `AppCore`
   - Manages non-historical state separately
   - Integrates all controllers
   - Handles localStorage persistence

2. **State Management**
   - Historical state: project data, effects, tracks, animations, transitions
   - Non-historical state: UI state, settings, playback state
   - State changes trigger compositor updates and localStorage saves
   - Undo/redo uses AppCore history system (64 operation limit)

3. **Controllers (`s/context/controllers/`)**
   - `Compositor` - PixiJS canvas rendering and object management
   - `Timeline` - Timeline logic and playback control
   - `Media` - Media file management
   - `VideoExport` - FFmpeg-based video export
   - `Collaboration` - WebRTC real-time collaboration
   - `Shortcuts` - Keyboard shortcut handling

4. **Components**
   - All use LitElement with `omnislate` (Nexus from @benev/slate)
   - Component naming: `Omni*` (OmniTimeline, OmniText, etc.)
   - Panels: `*Panel` classes for Construct editor integration

## Development Commands

### Build & Development
```bash
npm install           # Install dependencies
npm run build         # Production build (turtle-standard + copy assets)
npm start             # Development server with watch (turtle-standard-watch)
npm test              # Run tests with cynic
```

### Deployment
```bash
npm run build-production  # Production build
./deploy.sh               # Interactive deployment script
```

The build outputs to `/x/` directory which is served at `http://localhost:8000` during development.

### Testing
```bash
npm test              # Runs x/tests.test.js with cynic test runner
```

## Build System

Uses **Turtle** build system from @benev/turtle:
- `turtle-standard` - Standard build process
- `turtle-standard-watch` - Watch mode for development
- Rollup-based bundling
- ESM modules with import maps
- TypeScript compilation: `s/` → `x/`

### Build Scripts
- `copy-coi` - Copies coi-serviceworker.js for SharedArrayBuffer support
- `copy-assets` - Copies assets/ to x/assets/
- `prepare-x` - Prepares deployment package

## Key Patterns

### State Actions
Actions are defined in `s/context/actions.ts` and split into:
- `historical_actions` - Tracked in history (undo/redo)
- `non_historical_actions` - Not tracked (UI state, settings)

Use `ZipAction.actualize()` to bind actions to state.

### Component Registration
Components must be registered before use:
```typescript
register_to_dom({OmniManager, LandingPage})
```

Dynamic registration happens in router for editor route.

### Compositor Updates
When state changes affect canvas objects:
1. State change triggers watch.track callback
2. Compositor methods called to update PixiJS objects
3. Canvas re-renders automatically

### Collaboration State
- `collaboration` controller is singleton (preserves state across context refreshes)
- Clients cannot save to localStorage or export
- Host broadcasts state changes via WebRTC

## Browser Requirements

⚠️ **WebCodecs API Required:**
- Chrome/Edge 94+
- Safari 16.4+
- Firefox has limited support

The app checks `window.VideoEncoder` and `window.VideoDecoder` availability at startup.

## CDN Deployment

Deployed to `cdn.ad.nexus/studio/` with DNS `studio.ad.nexus` pointing to CDN.

**Cache Strategy:**
- Static assets (JS/CSS/images): `Cache-Control: public,max-age=31536000,immutable`
- HTML/JSON: `Cache-Control: public,max-age=0,must-revalidate`

**CORS Requirements:**
- Cross-Origin-Opener-Policy: same-origin
- Cross-Origin-Embedder-Policy: require-corp
- coi-serviceworker.js handles SharedArrayBuffer isolation

See DEPLOYMENT.md for detailed deployment instructions.

## Common Development Tasks

### Adding a New Effect Type
1. Define type in `s/context/types.ts`
2. Add manager to `Compositor` in `s/context/controllers/compositor/controller.ts`
3. Create manager class in `s/context/controllers/compositor/parts/`
4. Add state actions for the effect type
5. Update timeline rendering logic
6. Add export handling in VideoExport controller

### Adding a New Panel
1. Create panel class extending Construct's Panel
2. Register in `setupContext()` panels object in `s/main.ts`
3. Add to layouts if needed
4. Create corresponding component if interactive

### Modifying Video Export
Video export uses FFmpeg.wasm with WebCodecs for encoding:
- `VideoExport` controller orchestrates export
- `FFmpegHelper` manages FFmpeg instance
- `get_effect_at_timestamp()` determines which effects are active
- Each effect manager provides frames at specified timestamps

## Integration Points

### Zen AI Integration
Integration points for Zen AI models (from `~/work/zen/`):
- AI-powered content generation (future)
- Style transfer and effects (future)
- Automated copy generation (future)

### Twilio Integration
SDK included (`twilio` package) for:
- Dynamic phone number insertion
- Call tracking in ads

### QR Code Generation
Uses `qrcode` package for direct response campaigns.

## Important Notes

- Do not modify `x/` directory directly (build output)
- State persistence uses localStorage with projectId as key
- Collaboration state is NOT saved to localStorage for clients
- All video processing happens client-side (privacy-first)
- WebCodecs encoding may fail on unsupported browsers - always check `is_webcodecs_supported` signal
- PostHog analytics initialized in main.ts (EU server)

## Related Ecosystems

This project is part of the Adnexus ad tech stack:
- Adnexus Studio (this repo) - Creative tools
- Adnexus DSP - Campaign management
- Adnexus SSP - Inventory monetization
- Adnexus Exchange - Programmatic marketplace